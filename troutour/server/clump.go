package server

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"html/template"
	"net/http"
	"time"
)

const (
	clumpAdjReachKm = 5.0
)

// A Clump refers to some nearby Regions
type Clump struct {
	ID        string `datastore:",noindex"`
	ClumpBox  int32
	LatX10000 int32   `datastore:",noindex"`
	LngX10000 int32   `datastore:",noindex"`
	Lat       float64 `datastore:"-"`
	Lng       float64 `datastore:"-"`
	Tmp       int     `datastore:"-"`
	Kids      []string
}

func (clump *Clump) SetID() {
	clump.ID = fmt.Sprintf("%0.4f,%0.4f", clump.Lat, clump.Lng)
}

type ClumpAdj struct {
	EndIDs []string
}

func (ca ClumpAdj) ID() string {
	if ca.EndIDs[0] < ca.EndIDs[1] {
		return fmt.Sprintf("%s|%s", ca.EndIDs[0], ca.EndIDs[1])
	} else {
		return fmt.Sprintf("%s|%s", ca.EndIDs[1], ca.EndIDs[0])
	}
}

func newAdj(id1, id2 string) ClumpAdj {
	if id1 < id2 {
		return ClumpAdj{[]string{id1, id2}}
	} else {
		return ClumpAdj{[]string{id2, id1}}
	}
}

func (c *Clump) Load(ps []datastore.Property) error {
	err := datastore.LoadStruct(c, ps)
	if err != nil {
		return err
	}
	c.Lat = float64(c.LatX10000) / 10000.0
	c.Lng = float64(c.LngX10000) / 10000.0
	return nil
}

func (c *Clump) Save() ([]datastore.Property, error) {
	c.LatX10000 = int32(c.Lat * 10000.0)
	c.LngX10000 = int32(c.Lng * 10000.0)
	c.ClumpBox = latLng2ClumpBox(c.Lat, c.Lng)
	return datastore.SaveStruct(c)
}

func adjifyComputeAddRms(centerID string, ctx context.Context, clumps map[string]*Clump, queryAdjs func(context.Context, string) ([]ClumpAdj, error)) (err error, adds []ClumpAdj, rmIDs []string) {

	loadedAdjs := map[string]ClumpAdj{}
	foundAdjs, err := queryAdjs(ctx, centerID)
	if err != nil {
		log.Errorf(ctx, "Couldn't load initial batch of edges, hit %v", err)
		return
	}
	for _, adj := range foundAdjs {
		loadedAdjs[adj.ID()] = adj
	}
	centerClump := clumps[centerID]

	// order clumps by dist from centerClump; if we consider
	// them in order nearest->furthest, we get a nice optimization:
	// we load adjs that "block" the further clumps before we consider
	// those further clumps
	clumpIDsByDist := clumpIDsByDistance(centerClump.Lat, centerClump.Lng, clumps)

	// rmIDSet: Why a set, you ask? if todo->near is blocked (and should "break through"
	// adj between far1, far2, we "notice" the blocking far1-far2 adj twice,
	// once when we load far1's adjs, once when we load far2's.
	rmIDSet := map[string]bool{}
	for _, clumpID := range clumpIDsByDist {
		if clumpID == centerClump.ID {
			continue
		}
		adj := newAdj(centerID, clumpID)
		already, unblocked, blockerIDs := adjifyFindBlockers(adj, loadedAdjs, clumps)
		if (!already) && (!unblocked) {
			continue
		}
		if (!already) && unblocked {
			adds = append(adds, adj)
			loadedAdjs[adj.ID()] = adj
			for _, blockerID := range blockerIDs {
				rmIDSet[blockerID] = true
				delete(loadedAdjs, blockerID)
			}
		}
		newlyLoadedAdjs, qerr := queryAdjs(ctx, clumpID)
		if qerr != nil {
			err = qerr
			log.Errorf(ctx, "Hit error loading more adjs %v", err)
			return
		}
		for _, adj := range newlyLoadedAdjs {
			if _, toBeRemoved := rmIDSet[adj.ID()]; toBeRemoved {
				continue
			}
			already, unblocked, blockerIDs := adjifyFindBlockers(adj, loadedAdjs, clumps)
			// if we "should not add it", but it's already persisted, then actually we gotta rm it
			if (!already) && (!unblocked) {
				rmIDSet[adj.ID()] = true
			}
			if unblocked {
				// we "should add" it, but we're loading it from persistent storage.
				// so don't append it to our list of 'adds', but do put it into our local data structure:
				loadedAdjs[adj.ID()] = adj
				for _, blockerID := range blockerIDs {
					rmIDSet[blockerID] = true
					delete(loadedAdjs, blockerID)
				}
			}
		} // next newlyLoadedAdj
	} // next furthest clump
	for rmID, _ := range rmIDSet {
		rmIDs = append(rmIDs, rmID)
	}
	return
}

func fetchClumpAdjsByEndID(ctx context.Context, endID string) (cas []ClumpAdj, err error) {
	q := datastore.NewQuery("ClumpAdj").
		Filter("EndIDs =", endID)
	for cursor := q.Run(ctx); ; {
		ca := ClumpAdj{}
		_, err = cursor.Next(&ca)
		if err == datastore.Done {
			err = nil
			break
		}
		if err != nil {
			log.Errorf(ctx, "ERROR FETCHING adjacent clump info %v", err)
			return
		}
		cas = append(cas, ca)
	}
	return
}

func computeClumpAdj(cID string, centerLat float64, centerLng float64, ctx context.Context) (err error) {
	err, clumps := fetchClumps(ctx, centerLat, centerLng, clumpAdjReachKm)
	if err != nil {
		return
	}

	// Thanks to async fun times, we might be asked to compute clumpAdjs for
	// a clump that has meanwhile disappared. That's OK, just exit.
	_, found := clumps[cID]
	if !found {
		return
	}

	queryAdjs := fetchClumpAdjsByEndID

	err, adds, rmIDs := adjifyComputeAddRms(cID, ctx, clumps, queryAdjs)
	if err != nil {
		return
	}
	rmKeys := make([]*datastore.Key, len(rmIDs))
	for ix, rmID := range rmIDs {
		rmKeys[ix] = datastore.NewKey(ctx, "ClumpAdj", rmID, 0, nil)
	}
	err = datastore.DeleteMulti(ctx, rmKeys)
	if err != nil {
		return
	}
	addKeys := make([]*datastore.Key, len(adds))
	for ix, add := range adds {
		addKeys[ix] = datastore.NewKey(ctx, "ClumpAdj", add.ID(), 0, nil)
	}
	_, err = datastore.PutMulti(ctx, addKeys, adds)
	if err != nil {
		return
	}
	return
}

// Given two ClumpAdjs (and a map[clump ID string]Clump), return whether those adjs interset.
// Looks up the "ends" of the ClumpAdjs in the map.
func adjsIntersectP(a ClumpAdj, b ClumpAdj, m map[string]*Clump) bool {
	a1, found := m[a.EndIDs[0]]
	if !found {
		return false
	}
	a2, found := m[a.EndIDs[1]]
	if !found {
		return false
	}
	b1, found := m[b.EndIDs[0]]
	if !found {
		return false
	}
	b2, found := m[b.EndIDs[1]]
	if !found {
		return false
	}
	endSet := map[string]bool{ // if two adjs "share" an endpoint, that's not ∩. Count distinct endpoints:
		a.EndIDs[0]: true,
		a.EndIDs[1]: true,
		b.EndIDs[0]: true,
		b.EndIDs[1]: true,
	}
	if len(endSet) < 4 {
		return false
	}
	return segsIntersectP(a1.Lat, a1.Lng, a2.Lat, a2.Lng, b1.Lat, b1.Lng, b2.Lat, b2.Lng)
}

// Given a ClumpAdj (and a map[clump ID string]Clump), return its length in km.
// Looks up the "ends" of the ClumpAdjs in the map; if not found, returns a "silly-far" distance
func adjLen(adj ClumpAdj, m map[string]*Clump) float64 {
	e1, found := m[adj.EndIDs[0]]
	if !found {
		return 1.0e+21
	}
	e2, found := m[adj.EndIDs[1]]
	if !found {
		return 1.0e+21
	}
	return dist(e1.Lat, e1.Lng, e2.Lat, e2.Lng)
}

// given a clump and a map[clump ID string]Clump of clumps, return a list of clump IDs
// ordered by distance from set of coords, closest to farthest
func clumpIDsByDistance(lat float64, lng float64, clumps map[string]*Clump) []string {
	retval := []string{}
	for _, clump := range clumps {
		insertedP := false
		d := dist(lat, lng, clump.Lat, clump.Lng)
		for insertAt, otherClumpKey := range retval {
			otherClump := clumps[otherClumpKey]
			if d < dist(lat, lng, otherClump.Lat, otherClump.Lng) {
				retval = append(retval, "")
				copy(retval[insertAt+1:], retval[insertAt:])
				retval[insertAt] = clump.ID
				insertedP = true
				break
			}
		}
		if !insertedP {
			retval = append(retval, clump.ID)
		}
	}
	return retval
}

// Considers adding 'adj' to a set of ClumpAdjs in 'm' (referring to clumps in 'clumps').
// If 'adj' is already in there, returns already=true
// If shouldn't add 'adj' because it's blocked by a shorter edge, returns unblocked=false
// If should add 'adj' returns unblocked=true
//    ...if 'adj' is blocked by some longer ClumpAdjs, they're in []blockers; you want to
//    remove those.
// Yep, this is a _darned_ specialized helper function.
func adjifyFindBlockers(adj ClumpAdj, m map[string]ClumpAdj, clumps map[string]*Clump) (already bool, unblocked bool, blockerIDs []string) {
	id := adj.ID()
	if _, already = m[id]; already {
		unblocked = false
		return
	}
	l := adjLen(adj, clumps)
	for otherID, otherAdj := range m {
		if adjsIntersectP(adj, otherAdj, clumps) {
			if l > adjLen(otherAdj, clumps) {
				unblocked = false
				blockerIDs = []string{}
				return
			}
			blockerIDs = append(blockerIDs, otherID)
		}
	}
	unblocked = true
	return
}

func cronClumpAdj(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	// Which clumpBoxes have we already messed around near?
	already := map[int32]bool{}
	todoList := []ClumpAdjTodo{}
	catdq := datastore.NewQuery("ClumpAdjTodo").Limit(100)
	todoKeys, err := catdq.GetAll(ctx, &todoList)
	if err != nil {
		log.Errorf(ctx, "fetching to-dos got %v", err)
	}
	if len(todoKeys) == 0 {
		fmt.Fprintf(w, "<p>OMG DONE")
	} else {
		fmt.Fprintf(w, "<p>I see at least %d TODOs", len(todoKeys))
	}
	doneCount := 0
	for ix, catdKey := range todoKeys {
		clumpBox := latLng2ClumpBox(todoList[ix].Lat, todoList[ix].Lng)
		if already[clumpBox] {
			continue
		}
		stringID := catdKey.StringID()
		err := computeClumpAdj(stringID, todoList[ix].Lat, todoList[ix].Lng, ctx)
		if err != nil {
			fmt.Fprintf(w, "<p>computeClumpAdj hit snag %v", err)
			continue
		}
		err = datastore.Delete(ctx, catdKey)
		if err != nil {
			log.Errorf(ctx, "<p>failed to delete done to-do %v", err)
			return
		}
		doneCount++
		clumpRanges := nearbyClumpRanges(todoList[ix].Lat, todoList[ix].Lng, 10.0)
		for _, clumpRange := range clumpRanges {
			for box := clumpRange[0]; box <= clumpRange[1]; box++ {
				already[box] = true
			}
		}
	}
	fmt.Fprintf(w, "<p>Did some: %d", doneCount)
}

func fetchClumps(ctx context.Context, centerLat float64, centerLng float64, rangeKm float64) (err error, m map[string]*Clump) {
	m = map[string]*Clump{}
	clumpRanges := nearbyClumpRanges(centerLat, centerLng, rangeKm)
	for _, clumpRange := range clumpRanges {
		cq := datastore.NewQuery("Clump").
			Filter("ClumpBox >=", clumpRange[0]).
			Filter("ClumpBox <=", clumpRange[1])
		for cursor := cq.Run(ctx); ; {
			clump := Clump{}
			_, err = cursor.Next(&clump)
			if err == datastore.Done {
				err = nil
				break
			}
			if err != nil {
				log.Errorf(ctx, "ERROR FETCHING clumps %v", err)
				return
			}
			m[clump.ID] = &clump
		}
	}
	return
}

func clumpDown(ctx context.Context, clump Clump, late time.Time) (finishedP bool) {
	err, contentsGone := clumpDownContents(ctx, clump)
	if err != nil {
		log.Errorf(ctx, "Having trouble deleting clump, got %v", err)
	}
	if !contentsGone {
		return
	}
	if !time.Now().Before(late) {
		return
	}
	cas, err := fetchClumpAdjsByEndID(ctx, clump.ID)
	if err != nil {
		log.Errorf(ctx, "Having trouble finding clumpAdjs to rm, got %v", err)
		return
	}
	for _, ca := range cas {
		rmKey := datastore.NewKey(ctx, "ClumpAdj", ca.ID(), 0, nil)
		err = datastore.Delete(ctx, rmKey)
		if err != nil {
			log.Errorf(ctx, "failed to delete ClumpAdj, got %v", err)
			return
		}
		for _, otherClumpID := range ca.EndIDs {
			if otherClumpID == clump.ID {
				continue
			}
			err = addClumpAdjTodo(ctx, otherClumpID, clump.Lat, clump.Lng)
			if err != nil {
				log.Errorf(ctx, "failed to create ClumpAdjToDo, got %v", err)
				return
			}
		}
	}
	finishedP = true
	return
}

// Destroy everything referencing a clump. There's plenty to do, we
// might get halted by a deadline. So do things in such a way that
// we can pick up where we left off.
func clumpDownContents(ctx context.Context, clump Clump) (err error, finishedP bool) {
	oneErr := error(nil) // some errors, we keep going but know we should give up eventually. keep one around so we can fail on it at the end.
	for _, kidID := range clump.Kids {
		kidRegion := Region{}
		kidKey := datastore.NewKey(ctx, "Region", kidID, 0, nil)
		err = datastore.Get(ctx, kidKey, &kidRegion)
		if err == datastore.ErrNoSuchEntity {
			err = nil
		}
		if err != nil {
			log.Errorf(ctx, "Failed to load kid region, got %v", err)
			return
		}
		// try to mark kid region as no longer active to dissuade more things
		// from being associated with it
		if kidRegion.LifecycleState == rlsActive {
			kidRegion.LifecycleState = rlsEbbing
			datastore.Put(ctx, kidKey, &kidRegion)
			memcache.Delete(ctx, fmt.Sprintf("rgs/%d", latLng2RegionBox(kidRegion.Lat, kidRegion.Lng)))
		}
		// Seek and destroy all routes associated with kid region.
		rtq := datastore.NewQuery("Route").Filter("EndIDs =", kidID)
		for cursor := rtq.Run(ctx); ; {
			route := Route{}
			routeKey, err := cursor.Next(&route)
			if err == datastore.Done {
				err = nil
				break
			}
			if err != nil {
				log.Errorf(ctx, "ERROR FETCHING routes to remove %v", err)
				return err, false
			}
			err = datastore.Delete(ctx, routeKey)
			if err != nil {
				// If we fail to remove a route, it's a little messy, but
				// not an emergency. Keep going.
				log.Errorf(ctx, "ERROR removing route %v", err)
				oneErr = err
				continue
			}
			memo := Memo{
				RecipientID: route.BuilderID,
				Category:    memoCatRDown,
				Details: map[string]string{
					"region.Name": template.HTMLEscapeString(kidRegion.Name),
					"object":      "route",
				},
				When: time.Now(),
			}
			_, err = datastore.Put(ctx, datastore.NewKey(ctx, "Memo", "", 0, nil), &memo)
			if err != nil {
				log.Errorf(ctx, "Didn't put memo, got %v", err)
			}
		}

		npcq := datastore.NewQuery("NPC").Filter("RegionID =", kidID)
		for cursor := npcq.Run(ctx); ; {
			npc := NPC{}
			npcKey, err := cursor.Next(&npc)
			if err == datastore.Done {
				err = nil
				break
			}
			if err != nil {
				log.Errorf(ctx, "ERROR FETCHING NPCs to remove %v", err)
				return err, false
			}
			err = datastore.Delete(ctx, npcKey)
			if err != nil {
				// If we fail to remove an NPC, it's a little messy, but
				// not an emergency. Keep going.
				log.Errorf(ctx, "ERROR removing NPC %v", err)
				oneErr = err
				continue
			}
			memo := Memo{
				RecipientID: npc.AgentID,
				Category:    memoCatRDown,
				Details: map[string]string{
					"region.Name": template.HTMLEscapeString(kidRegion.Name),
					"object":      "npc",
				},
				When: time.Now(),
			}
			_, err = datastore.Put(ctx, datastore.NewKey(ctx, "Memo", "", 0, nil), &memo)
			if err != nil {
				log.Errorf(ctx, "Didn't put memo, got %v", err)
			}
		}

		// when there are other things-that-refer-to-regions, remove them too

		// cheesy check: does the region still exist? Or did our previous
		// attempt to load it from the datastore fail?
		if kidRegion.LifecycleState != 0 {
			err = datastore.Delete(ctx, kidKey)
			if err != nil {
				log.Errorf(ctx, "Failed to rm region, got %v", err)
				return
			}
		}
	}
	if oneErr != nil {
		return oneErr, false
	}

	finishedP = true
	return
}

func cronClumpDown(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	late := time.Now().Add(30 * time.Second)
	already := map[int32]bool{}
	q := datastore.NewQuery("ClumpDownTodo")
	for cursor := q.Run(ctx); ; {
		cdtd := ClumpDownTodo{}
		cdtdKey, err := cursor.Next(&cdtd)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "cronClumpDown Next hit error: %v", err)
			break
		}
		clump := Clump{}
		clumpKey := datastore.NewKey(ctx, "Clump", cdtd.ClumpID, 0, nil)
		err = datastore.Get(ctx, clumpKey, &clump)
		if err == datastore.ErrNoSuchEntity {
			datastore.Delete(ctx, cdtdKey)
			continue
		}
		if already[latLng2ClumpBox(clump.Lat, clump.Lng)] {
			continue
		}
		for _, clumpRange := range nearbyClumpRanges(clump.Lat, clump.Lng, 2*clumpAdjReachKm) {
			for ix := clumpRange[0]; ix <= clumpRange[1]; ix++ {
				already[ix] = true
			}
		}
		finishedP := clumpDown(ctx, clump, late)
		if !finishedP {
			continue
		}

		err = datastore.Delete(ctx, clumpKey)
		if err != nil {
			log.Errorf(ctx, "cronClumpDown failed to delete clump, hit %v", err)
			return
		}
		datastore.Delete(ctx, cdtdKey)
		continue
	}
}
