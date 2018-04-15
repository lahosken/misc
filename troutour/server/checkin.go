package server

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"html"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	overplentifulRoutes = 2000
	plentifulPrisms     = 50
)

type RecentCheckin struct {
	RegionID string `datastore:",noindex"`
	UserID   string
	T        time.Time
}

// for now, checking in doesn't really do much except report status.
// anyhow, here's a handy utility for json-izing status.
func statusJson(s ...string) []byte {
	retval, _ := json.Marshal(map[string][]string{"msgs": s})
	return []byte(retval)
}

// returns nearest region within pinging distance (or foundP= false if
// no region in pinging distance)
func nearestRegion(centerLat float64, centerLng float64, regions map[string]Region) (r Region, foundP bool) {
	rDist := pingMaxRangeKm
	for _, region := range regions {
		if region.LifecycleState != rlsActive {
			continue
		}
		d := dist(centerLat, centerLng, region.Lat, region.Lng)
		if d < rDist {
			r = region
			rDist = d
			foundP = true
		}
	}
	return
}

func fetchInventory(ctx context.Context, userID string) (i UserInventory, err error) {
	userKey := newUserKey(ctx, userID)
	inventoryKey := datastore.NewKey(ctx, "UserInventory", userID, 0, userKey)
	err = datastore.Get(ctx, inventoryKey, &i)
	if err == datastore.ErrNoSuchEntity {
		err = nil
		return
	}
	if err != nil {
		log.Errorf(ctx, "error loading inventory, hit %v", err)
		return // bail!
	}
	return
}

func makeResponseNPCs(npcs map[int64]*NPC, regions map[string]Region, youID string) []ResponseNPC {
	rv := []ResponseNPC{}
	for _, npc := range npcs {
		_, found := regions[npc.RegionID]
		if npc.RegionID == "" || found {
			rv = append(rv, ResponseNPC{npc.RegionID, npc.AgentID == youID})
		}
	}
	return rv
}

// compute which regions is this region connected to by the user's routes.
// returns a map regionID -> distance in "hops"
func computeConnectedRegions(rootID string, routes map[string]Route) (rv map[string]int) {
	rv = map[string]int{rootID: 0}
	for iter := 0; iter < 1000; iter++ {
		found := false
		for fromReg, hops := range rv {
			if hops != iter {
				continue
			}
			for _, rt := range routes {
				toReg := ""
				if rt.EndIDs[0] == fromReg {
					toReg = rt.EndIDs[1]
				}
				if rt.EndIDs[1] == fromReg {
					toReg = rt.EndIDs[0]
				}
				if toReg == "" {
					continue
				}
				if rv[toReg] > 0 { // if we already visited, don't revisit
					continue
				}
				rv[toReg] = iter + 1
				found = true
			}
		}
		if !found {
			break
		}
	}
	return
}

func fetchRoutableRegionIDs(ctx context.Context, thisRegion Region, regions map[string]Region, existingRoutes map[string]Route) (routables []string, err error) {
	clumpKeys := []*datastore.Key{datastore.NewKey(ctx, "Clump", thisRegion.Clump, 0, nil)}
	cas, err := fetchClumpAdjsByEndID(ctx, thisRegion.Clump)
	if err != nil {
		log.Errorf(ctx, "error fetching adjacent clump info, %v", err)
		return // could keep going, but... meh I dunno
	}
	for _, ca := range cas {
		if ca.EndIDs[0] != thisRegion.Clump {
			clumpKeys = append(clumpKeys,
				datastore.NewKey(ctx, "Clump", ca.EndIDs[0], 0, nil))
		}
		if ca.EndIDs[1] != thisRegion.Clump {
			clumpKeys = append(clumpKeys,
				datastore.NewKey(ctx, "Clump", ca.EndIDs[1], 0, nil))
		}
	}
	clumps := make([]Clump, len(clumpKeys))
	err = datastore.GetMulti(ctx, clumpKeys, clumps)
	if err != nil {
		log.Errorf(ctx, "ERROR FETCHING nearby clumps %v", err)
		return // could keep going, but... meh I dunno
	}
	for _, clump := range clumps {
		for _, kid := range clump.Kids {
			if kid == thisRegion.ID {
				continue
			}
			if regions[kid].LifecycleState == rlsEbbing {
				continue
			}
			alreadyHasRoute := false
			for _, route := range existingRoutes {
				if route.EndIDs[0] == kid && route.EndIDs[1] == thisRegion.ID {
					alreadyHasRoute = true
					break
				}
				if route.EndIDs[1] == kid && route.EndIDs[0] == thisRegion.ID {
					alreadyHasRoute = true
					break
				}
			}
			if alreadyHasRoute {
				continue
			}

			routables = append(routables, kid)
		}
	}
	return
}

// trade in prisms for routes; don't persist
func checkinPrisms2Routes(thisRegion Region, newRoutes *[]Route, routableRegionIDs *[]string, userID string, inventory *UserInventory) {
	for rrix := len(*routableRegionIDs) - 1; rrix >= 0; rrix-- {
		routableRegionID := (*routableRegionIDs)[rrix]
		for pix, prism := range inventory.Prisms {
			if prism == routableRegionID {
				*newRoutes = append(*newRoutes, Route{userID, []string{thisRegion.ID, routableRegionID}})
				inventory.Prisms = append(inventory.Prisms[:pix], inventory.Prisms[pix+1:]...)
				*routableRegionIDs = append((*routableRegionIDs)[:rrix], (*routableRegionIDs)[rrix+1:]...)
				break
			}
		}
	}
}

func checkinAndCheckTooSoon(ctx context.Context, region Region, userID string, recent map[string]RecentCheckin) (tooSoonP bool) {
	if _, found := recent[region.ID]; found {
		return true
	}
	rc := RecentCheckin{
		RegionID: region.ID,
		UserID:   userID,
		T:        time.Now(),
	}
	rcKey := datastore.NewKey(ctx, "RecentCheckin", "", 0, nil)
	if _, err := datastore.Put(ctx, rcKey, &rc); err != nil {
		log.Errorf(ctx, "Couldn't save recent checkin, got %v", err)
	}
	return false
}

func fetchUserRecentCheckins(ctx context.Context, userID string) (m map[string]RecentCheckin, err error) {
	now := time.Now()
	anHourAgo := now.Add(-time.Hour)
	m = map[string]RecentCheckin{}
	cq := datastore.NewQuery("RecentCheckin").Filter("UserID =", userID)
	for cursor := cq.Run(ctx); ; {
		rc := RecentCheckin{}
		_, err = cursor.Next(&rc)
		if err == datastore.Done {
			err = nil
			break
		}
		if err != nil {
			log.Errorf(ctx, "ERROR FETCHING recent checkins %v", err)
			return
		}
		if !anHourAgo.Before(rc.T) {
			continue
		}
		m[rc.RegionID] = rc
	}
	return
}

func predictProjection(lat, lng float64, checkins map[string]RecentCheckin, thisRegion Region, regions map[string]Region) (predictLat, predictLng float64, predictionP bool) {
	offsetLat := 0.0
	offsetLng := 0.0
	for _, checkin := range checkins {
		thatRegion, found := regions[checkin.RegionID]
		if !found {
			continue
		}
		d := dist(thisRegion.Lat, thisRegion.Lng, thatRegion.Lat, thatRegion.Lng)
		if d < regionsTooCloseKm {
			continue
		}
		offsetLat += (thisRegion.Lat - thatRegion.Lat) / d
		offsetLng += (thisRegion.Lng - thatRegion.Lng) / d
	}
	d := dist(lat, lng, lat+offsetLat, lng+offsetLng)
	if d < 0.001 {
		return
	}
	predictLat = lat + 0.5*(offsetLat/d)
	predictLng = lng + 0.5*(offsetLng/d)
	predictionP = true
	return
}

func checkinDoomBusyRegion(ctx context.Context, routes map[string]Route, regions map[string]Region) {
	endCount := map[string]int{}
	for _, route := range routes {
		for _, endID := range route.EndIDs {
			endCount[endID]++
		}
	}
	mostCount := 0
	mostEndID := ""
	rand.Seed(time.Now().Unix())
	for endID, count := range endCount {
		if count > mostCount && rand.Float64() < 0.01 {
			mostEndID = endID
			mostCount = count
		}
	}
	if mostEndID == "" {
		return
	}
	doomedRegion := Region{}
	found := false
	if doomedRegion, found = regions[mostEndID]; !found {
		doomedRegionKey := datastore.NewKey(ctx, "Region", mostEndID, 0, nil)
		err := datastore.Get(ctx, doomedRegionKey, &doomedRegion)
		if err != nil && err != datastore.ErrNoSuchEntity {
			log.Errorf(ctx, "Couldn't load doomed region, got %v", err)
			return
		}
	}
	if doomedRegion.Clump == "" {
		// Ugh, we have many lingering routes referring to an ex-Region.
		// Instead of marking that Region's clump for downing, rm these
		// garbage routes.
		q := datastore.NewQuery("Route").Filter("EndIDs =", mostEndID).KeysOnly().Limit(100)
		rmKeys, _ := q.GetAll(ctx, nil)
		datastore.DeleteMulti(ctx, rmKeys)
		return
	} else {
		addClumpDownTodo(ctx, doomedRegion.Clump)
	}
}

func checkin(w http.ResponseWriter, r *http.Request, userID string, sessionID string) {
	w.Header().Set("Content-Type", "application/json")
	if userID == "" {
		w.Write(statusJson(`You're not logged in so why check in?!?`))
		return
	}
	ctx := appengine.NewContext(r)
	s := ``
	centerLat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	centerLng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	token := r.URL.Query().Get("token")
	rand.Seed(time.Now().Unix())
	err, regions := fetchRegs(ctx, centerLat, centerLng, pingMaxRangeKm)
	if err != nil {
		log.Errorf(ctx, "checkin couldn't fetch nearby regions, hit %v", err)
		w.Write(statusJson(`Don't know about anything interesting nearby, hit error. Better luck next time?`))
		return
	}

	thisRegion, foundP := nearestRegion(centerLat, centerLng, regions)
	if !foundP {
		addRupTodo(ctx, userID, centerLat, centerLng)
		w.Write(statusJson(`Nothing nearby. Better luck next time?`))
		return
	}

	inventory, err := fetchInventory(ctx, userID)
	if err != nil {
		s += fmt.Sprintf("<p>Couldn't fetch inventory, got err %v", err)
		w.Write(statusJson(s))
		return // bail!
	}

	recentCheckins, _ := fetchUserRecentCheckins(ctx, userID)
	if checkinAndCheckTooSoon(ctx, thisRegion, userID, recentCheckins) {
		// Maybe the client didn't get our previous response and is retrying?
		// (_Probably_ it's just the user being impatient and mashing buttons, but
		//  _maybe_ this is a retry?) Check to see if we have a cached response:
		item, err := memcache.Get(ctx, fmt.Sprintf("CHK/%s/%s", sessionID, token))
		if err == nil {
			w.Write(item.Value) // It's a retry! Spew cached JSON
		} else {
			// It's an impatient user. Counsel patience.
			w.Write(statusJson(fmt.Sprintf(`⌛&nbsp;I was here (%s) less than an hour&nbsp;ago.`, html.EscapeString(thisRegion.Name))))
		}
		return
	}

	drama := false // if true, "enough" "interesting" things have happened and we shouldn't encourage any more

	routes, err := fetchUsersOwnRoutes(ctx, userID)
	if err != nil {
		s = fmt.Sprintf("<p><b>ERROR</b> FETCHING user's own routes %v", err) + s
		// keep going, I guess?
	}
	if len(routes) > overplentifulRoutes {
		checkinDoomBusyRegion(ctx, routes, regions)
	}

	npcs := map[int64]*NPC{}
	err = fetchNPCsNearby(ctx, centerLat, centerLng, 1.0, npcs)
	err = fetchNPCsByAgent(ctx, userID, npcs)
	if err != nil {
		log.Errorf(ctx, "Failed to load user's own NPCs, got %v. Bailing.", err)
		s += fmt.Sprintf("<p>Couldn't fetch user's NPCs, got err %v", err)
		w.Write(statusJson(s))
		return
	}
	userNPCCount := len(npcs)
	if err != nil {
		log.Errorf(ctx, "Failed to load nearby NPCs, got %v", err)
	}

	travelingNPCs := []*NPC{}
	thisRegionNPCs := []*NPC{}
	for _, npc := range npcs {
		if npc.RegionID == "" && npc.AgentID == userID {
			travelingNPCs = append(travelingNPCs, npc)
		}
		if npc.RegionID == thisRegion.ID {
			thisRegionNPCs = append(thisRegionNPCs, npc)
		}
	}

	connectedRegions := computeConnectedRegions(thisRegion.ID, routes)

	/*
	 * Trade in prisms for routes.
	 */
	routableRegionIDs, err := fetchRoutableRegionIDs(ctx, thisRegion, regions, routes)
	if err != nil {
		log.Errorf(ctx, "fetchRoutableRegionIDs got err %v", err)
	}
	newRoutes := []Route{}
	checkinPrisms2Routes(thisRegion, &newRoutes, &routableRegionIDs, userID, &inventory)
	// If user didn't have many (or any?) routes yet, worth reporting.
	if len(newRoutes)*4 > len(routes) {
		if len(newRoutes) == 1 {
			s += `/&nbsp;Established new ⛗<em>Route</em>&nbsp;/`
		} else {
			s += fmt.Sprintf(`/&nbsp;Established %d new ⛗<em>Routes</em>&nbsp;/`, len(newRoutes))
		}
		drama = true
	}

	/*
	 * If a surplus of prisms, maybe turn in 10 of them to a place
	 * we don't have an actual prism for.
	 */
	if !drama && len(newRoutes) == 0 &&
		len(inventory.Prisms) > plentifulPrisms &&
		len(routableRegionIDs) > 0 {
		wildcardIx := rand.Intn(len(routableRegionIDs))
		wildcard := routableRegionIDs[wildcardIx]
		routableRegionIDs = append(routableRegionIDs[:wildcardIx], routableRegionIDs[wildcardIx+1:]...)
		region, found := regions[wildcard]
		if found {
			s += fmt.Sprintf("/ traded in some old 💎Prisms for ⛗Route to %s /", html.EscapeString(region.Name))
		} else {
			s += fmt.Sprintf("/ traded in some old 💎Prisms for new ⛗Route /")
		}
		inventory.Prisms = inventory.Prisms[10:]
		drama = true
		newRoutes = append(newRoutes, Route{userID, []string{thisRegion.ID, wildcard}})
	}

	unloadedRegions := []Region{}
	for ix := len(newRoutes) - 1; ix >= 0; ix-- {
		route := newRoutes[ix]
		for _, end := range route.EndIDs {
			_, found := regions[end]
			if !found {
				rKey := datastore.NewKey(ctx, "Region", end, 0, nil)
				r := Region{}
				rerr := datastore.Get(ctx, rKey, &r)
				if rerr != nil || r.LifecycleState != rlsActive {
					// TODO better error handling? (seen no such entity once)
					log.Infof(ctx, "canceling Route to unloaded region %v, got rerr %v Lifecycle %v", end, rerr, r.LifecycleState)
					newRoutes = append(newRoutes[:ix], newRoutes[ix+1:]...)
					break
				}
				unloadedRegions = append(unloadedRegions, r)
			}
		}
	}

	unloadedRegionsForBoxes := map[string]Region{}
	for _, r := range unloadedRegions {
		_, already := unloadedRegionsForBoxes[r.ID]
		if already {
			continue
		}
		rerr, m := fetchRegs(ctx, r.Lat, r.Lng, 0.0)
		if rerr != nil {
			continue
		}
		for k, v := range m {
			unloadedRegionsForBoxes[k] = v
		}
	}
	unloadedResponseRegions := makeResponseRegions(unloadedRegionsForBoxes)

	if len(inventory.Prisms) < 100 || !stSliceContains(inventory.Prisms, thisRegion.ID) {
		if len(inventory.Prisms)+2*len(routes) < 6 {
			s += `/&nbsp;Got a 💎<em>Prism</em>. Visit some region nearby to transform that Prism into a ⛗Route back here.&nbsp;/`
			drama = true
		}
		inventory.Prisms = append(inventory.Prisms, thisRegion.ID)
	}

	for bix, bountyID := range inventory.Bounties {
		if thisRegion.ID != bountyID {
			continue
		}
		inventory.Trophies++
		inventory.Bounties = append(inventory.Bounties[:bix], inventory.Bounties[bix+1:]...)
		drama = true
		s += fmt.Sprintf("/Picked up that 🏆Trophy, now have %d/", inventory.Trophies)
		break
	}

	if !drama {
		for bix, bountyID := range inventory.Bounties {
			if _, found := regions[bountyID]; found {
				continue
			}
			inventory.Bounties = append(inventory.Bounties[:bix], inventory.Bounties[bix+1:]...)
			s += fmt.Sprintf("/Strayed too far from that awards show, missed it/")
			drama = true
			break
		}
	}

	if !drama && userNPCCount > 1 && inventory.Coins > 3 && rand.Intn(2*len(inventory.Bounties)+2) == 0 {
		predictLat, predictLng, predictionP := predictProjection(centerLat, centerLng, recentCheckins, thisRegion, regions)
		if predictionP {
			for attemptCount := 0; attemptCount < 13; attemptCount++ {
				rightPlaceRegion, found1 := nearestRegion(predictLat, predictLng, regions)
				_, found2 := recentCheckins[rightPlaceRegion.ID]
				found3 := stSliceContains(inventory.Bounties, rightPlaceRegion.ID)
				if rightPlaceRegion.ID != thisRegion.ID && found1 && !found2 && !found3 {
					inventory.Bounties = append(inventory.Bounties, rightPlaceRegion.ID)
					s += fmt.Sprintf(`/<span class="trophy">🏆</span>Go to award show at %s for a 🏆Trophy/`, html.EscapeString(rightPlaceRegion.Name))
					drama = true
					break
				}
				predictLat += (-0.2 + (0.4 * rand.Float64())) / kmPerLat()
				predictLng += (-0.2 + (0.4 * rand.Float64())) / kmPerLng(centerLat)
			}
		}
	}

	if !drama && len(travelingNPCs) > 0 && len(thisRegionNPCs) == 0 {
		npc := travelingNPCs[rand.Intn(len(travelingNPCs))]
		npc.RegionID = thisRegion.ID
		npc.RegionBox = thisRegion.RegionBox
		if rand.Float64() < 0.01 && rand.Float64() < inventory.Cred-0.5 {
			s += fmt.Sprintf("/&nbsp;🎤&nbsp;Client settles down here &amp; produces <em>weird</em> art; area is doomed to fall off the entertainment circuit.&nbsp;/")
			addClumpDownTodo(ctx, thisRegion.Clump)
		} else {
			s += fmt.Sprintf("/&nbsp;🎤&nbsp;Client settles down here.&nbsp;/")
		}
		drama = true
	}

	cred := inventory.Cred + inventory.TransientCred

	if !drama && len(travelingNPCs) == 0 && len(thisRegionNPCs) != 0 {
		npc := thisRegionNPCs[rand.Intn(len(thisRegionNPCs))]
		neededCred := 1.0
		if cred > neededCred && (npc.AgentID != userID && strings.HasPrefix(npc.AgentID, "_dev")) {
			if inventory.Cred < neededCred {
				inventory.TransientCred -= (neededCred - inventory.Cred)
			}
			npc.AgentID = userID
			npc.RegionID = ""
			npc.RegionBox = fakeRegionBox
			s += fmt.Sprintf("/&nbsp;🎤&nbsp;Client pulls up stakes to walk with you.&nbsp;/")
			drama = true
		}
		if cred < neededCred {
			inventory.TransientCred += 0.1 + rand.Float64()*(neededCred-cred)
			if inventory.Cred > 1.0 {
				inventory.TransientCred += 0.1 * rand.Float64() * (inventory.Cred - 1.0)
			}
			if userNPCCount < 2 {
				if inventory.Cred < 0.02 && rand.Float64() < 0.3 {
					s += `/&nbsp;Tried and failed to recruit a 🎤&nbsp;Client. Eventually, you'll earn professional credibility and thus recruit more easily.&nbsp;/`
					drama = true

				}
				inventory.TransientCred += 0.1
			}
		}
	}

	if len(travelingNPCs) > 0 {
		keys := []*datastore.Key{}
		for _, npc := range travelingNPCs {
			keys = append(keys, datastore.NewKey(ctx, "NPC", "", npc.ID, nil))
		}
		_, err = datastore.PutMulti(ctx, keys, travelingNPCs)
		if err != nil {
			log.Errorf(ctx, "NPC failed to persist changes because %v", err)
			// weird but keep going I guess
		}
	}
	if len(thisRegionNPCs) > 0 {
		keys := []*datastore.Key{}
		for _, npc := range thisRegionNPCs {
			keys = append(keys, datastore.NewKey(ctx, "NPC", "", npc.ID, nil))
		}
		_, err = datastore.PutMulti(ctx, keys, thisRegionNPCs)
		if err != nil {
			log.Errorf(ctx, "NPC failed to persist changes because %v", err)
			// weird but keep going I guess
		}
	}

	if !drama && len(travelingNPCs) == 0 && len(thisRegionNPCs) == 0 && rand.Float64() < 0.70 {
		newRecruitNPC := randNPC()
		newRecruitNPC.AgentID = userID
		newRecruitKey := datastore.NewKey(ctx, "NPC", "", 0, nil)
		neededCred := 1.0
		if cred > neededCred {
			if inventory.Cred < neededCred {
				inventory.TransientCred -= (neededCred - inventory.Cred)
			}
			_, err = datastore.Put(ctx, newRecruitKey, &newRecruitNPC)
			if err == nil {
				s += fmt.Sprintf("/&nbsp;You scouted for talent and found a new 🎤&nbsp;Client.&nbsp;/")
				npcs[0] = &newRecruitNPC
				drama = true
			} else {
				log.Errorf(ctx, "Failed to persist new recruit because %v", err)
			}
		}
		if cred < neededCred {
			if inventory.Cred < 0.02 && rand.Float64() < 0.3 {
				s += `/&nbsp;Tried and failed to recruit a 🎤&nbsp;Client. Eventually, you'll earn professional credibility and thus recruit more easily.&nbsp;/`
				drama = true
			}
			inventory.TransientCred += 0.1 + rand.Float64()*(neededCred-cred)
		}
	}

	log.Infof(ctx, "drama=%v", drama)
	if !drama && inventory.Coins > 5 && inventory.Trophies > 5 {
		credCost := math.Pow(1.05, 100.0*inventory.Cred) * 1000.0
		log.Infof(ctx, "credCost=%v", credCost)
		if float64(inventory.Coins*inventory.Trophies) > 2.0*credCost {
			coinSpend := (int64(credCost) / inventory.Trophies) + 1
			inventory.Cred += 0.01 * float64(coinSpend*inventory.Trophies) / credCost
			inventory.Coins -= coinSpend
			inventory.Trophies = 0
			s += fmt.Sprintf("/🎉You threw a party to show off your 🏆Trophies and are now more professionally credible./")
		}
		drama = true
	}

	ambientEarnings := 0.0
	connectedClientEarnings := 0.0
	unconnectedClientEarnings := 0.0

	// if player has accumulated trophies but is in a place where it's
	// difficult to earn coins, help 'em out a little. Heck, help 'em
	// out a little even if they're not in a place where it's difficult
	// to earn coins.
	ambientEarnings += math.Max(float64(inventory.Trophies)-3.0, 0.0)

	for _, npc := range npcs {
		if npc.AgentID != userID {
			ambientEarnings += 0.4
			continue
		}
		d, found := connectedRegions[npc.RegionID]
		if found {
			connectedClientEarnings += 1.0 + (inventory.Cred / (float64(d) + 1.0))
		} else {
			unconnectedClientEarnings += 0.5
		}
	}
	earnings := int64(math.Sqrt(ambientEarnings + connectedClientEarnings + unconnectedClientEarnings))
	if earnings > 0 {
		bestEarnings := ambientEarnings
		subject := "Local performers"
		if connectedClientEarnings > bestEarnings {
			bestEarnings = connectedClientEarnings
			subject = "Your clients"
		}
		if unconnectedClientEarnings > bestEarnings {
			bestEarnings = unconnectedClientEarnings
			subject = "Your far-flung clients"
		}
		inventory.Coins += earnings
		log.Infof(ctx, "connectedClientEarnings= %0.2f , unconnectedClientEarnings= %0.2f", connectedClientEarnings, unconnectedClientEarnings)
		s += fmt.Sprintf("/%s paid 💰%d, now you have 💰%d/", subject, earnings, inventory.Coins)
	}

	err = datastore.RunInTransaction(ctx, func(context context.Context) error {
		userKey := newUserKey(ctx, userID)
		inventoryKey := datastore.NewKey(ctx, "UserInventory", userID, 0, userKey)
		_, err := datastore.Put(ctx, inventoryKey, &inventory)
		if err != nil {
			log.Errorf(ctx, "error saving inventory, hit %v", err)
			s = s + fmt.Sprintf("<p>Couldn't save inventory, got err %v", err)
			return err
		}
		for _, route := range newRoutes {
			routeKey := newRouteKey(ctx, userID, route.EndIDs[0], route.EndIDs[1])
			_, err = datastore.Put(ctx, routeKey, &route)
			if err != nil {
				log.Errorf(ctx, "error saving new route, hit %v", err)
				s = s + fmt.Sprintf("<p>Couldn't save new route, got err %v", err)
				return err
			}
		}
		return nil
	}, nil)

	s = fmt.Sprintf(`At region %s./`, html.EscapeString(thisRegion.Name)) + s

	newReportedRoutes := []ResponseRoute{}
	for _, route := range newRoutes {
		newReportedRoutes = append(newReportedRoutes, ResponseRoute{
			[]string{route.EndIDs[0], route.EndIDs[1]}})
	}
	oldReportedRoutes := []ResponseRoute{}
	for _, route := range routes {
		e0, found0 := regions[route.EndIDs[0]]
		e1, found1 := regions[route.EndIDs[1]]
		if e0.LifecycleState != rlsActive || e1.LifecycleState != rlsActive {
			continue
		}
		if !(found0 || found1) {
			continue
		}
		oldReportedRoutes = append(oldReportedRoutes, ResponseRoute{
			[]string{route.EndIDs[0], route.EndIDs[1]}})
	}

	reportedNPCs := makeResponseNPCs(npcs, regions, userID)

	response := struct {
		Checkin   []string                      `json:"chkn,omitempty"`
		Regions   map[string]([]ResponseRegion) `json:"regs,omitempty"`
		OldRoutes []ResponseRoute               `json:"orts,omitempty"`
		NewRoutes []ResponseRoute               `json:"nrts,omitempty"`
		Messages  []string                      `json:"msgs,omitempty"`
		Inventory ResponseUserInventory         `json:"inv"`
		NPCs      []ResponseNPC                 `json:"npcs,omitempty"`
	}{
		[]string{thisRegion.ID},
		unloadedResponseRegions,
		oldReportedRoutes,
		newReportedRoutes,
		[]string{s},
		inventory.makeResponseInventory(),
		reportedNPCs,
	}
	js, _ := json.Marshal(response)
	w.Write(js)

	// In case the client doesn't receive the response OK,
	// keep a copy in memcache; maybe we can serve it
	// when the client retries.
	// Remove a couple of big-and-not-vital fields:
	response.NPCs = []ResponseNPC{}
	response.Inventory.Prisms = []string{}
	item := memcache.Item{
		Key:        fmt.Sprintf("CHK/%s/%s", sessionID, token),
		Object:     response,
		Expiration: 40 * time.Second, // enough for a few retries?  ¯\_(ツ)_/¯
	}
	err = memcache.JSON.Set(ctx, &item)
	if err != nil {
		log.Errorf(ctx, "Memcache couldn't stash response, got %v", err)
	}
}

// func cronCleanupCheckins(w http.ResponseWriter, r *http.Request) {
// 	ctx := appengine.NewContext(r)
func cronCleanupCheckins(ctx context.Context) {
	now := time.Now()
	anHourAgo := now.Add(-time.Hour)
	rcq := datastore.NewQuery("RecentCheckin").
		Filter("T <", anHourAgo).
		Limit(500).
		KeysOnly()
	keys, err := rcq.GetAll(ctx, nil)
	if err != nil {
		log.Errorf(ctx, "Can't get keys to clean up, hit err %v", err)
	}
	for _, key := range keys {
		datastore.Delete(ctx, key)
	}
}
