package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	/* region LifecycleState values */
	rlsUnknown = 0 // Wasn't already in our Region store. New?
	rlsDormant = 1 // Inactive, invisible
	rlsNascent = 2 // Transitioning dormant->active
	rlsActive  = 3 // Player can see it, hurrah
	rlsEbbing  = 4 // Transitioning active -> dormant
	rlsGC      = 5 // Transitioning dormant -> (deleted)

	clumpMaxRadiusKm   = 1.0
	pingMaxRangeKm     = 1.0
	clumpCreateRangeKm = pingMaxRangeKm + clumpMaxRadiusKm
	regCreateRangeKm   = clumpCreateRangeKm + clumpMaxRadiusKm

	regionsTooCloseKm = 0.150

	minRegionsPerClump = 1
	maxRegionsPerClump = 5
)

var (
	errNoVenuesButTombstone = errors.New("no close venues; we've queried for them before tho")
	errNotEnoughVenues      = errors.New("too few venues even semi-near")
	errNoCloseVenue         = errors.New("no venue within pinging distance")
)

type Region struct {
	ID             string `datastore:",noindex"`
	RegionBox      int32
	LifecycleState int8
	Clump          string
	Name           string  `datastore:",noindex"`
	LatX10000      int32   `datastore:",noindex"`
	LngX10000      int32   `datastore:",noindex"`
	Lat            float64 `datastore:"-"`
	Lng            float64 `datastore:"-"`
	FsqUrl         string  `datastore:",noindex"`
}

func (r *Region) Load(ps []datastore.Property) error {
	err := datastore.LoadStruct(r, ps)
	if err != nil {
		return err
	}
	r.Lat = float64(r.LatX10000) / 10000.0
	r.Lng = float64(r.LngX10000) / 10000.0
	return nil
}

func (r *Region) Save() ([]datastore.Property, error) {
	r.LatX10000 = int32(r.Lat * 10000.0)
	r.LngX10000 = int32(r.Lng * 10000.0)
	r.RegionBox = latLng2RegionBox(r.Lat, r.Lng)
	return datastore.SaveStruct(r)
}

func cronRegionUp(ctx context.Context, dirtyClumps map[int32]bool) {
	ninetySecondsFromStart := time.Now().Add(90 * time.Second)

	todoList := []RupTodo{}
	ntdq := datastore.NewQuery("RupTodo").Limit(100)
	todoKeys, err := ntdq.GetAll(ctx, &todoList)
	if err != nil {
		log.Errorf(ctx, "fetching to-dos got %v", err)
		return
	}
	for ix, ntd := range todoList {
		if ninetySecondsFromStart.Before(time.Now()) {
			break
		}
		log.Infof(ctx, "cronRegionUp %v", ix)
		clumpBox := latLng2ClumpBox(todoList[ix].Lat, todoList[ix].Lng)
		if dirtyClumps[clumpBox] {
			continue
		}
		err, addedCount := regionUp(ctx, ntd.Lat, ntd.Lng)
		switch err {
		case errNotEnoughVenues:
			sid := todoKeys[ix].StringID()
			addFsqTodo(ctx, sid, ntd.Lat, ntd.Lng)
			datastore.Delete(ctx, todoKeys[ix]) // if 4sq finds new things, it'll create a todo hereabouts.
			dirtyClumpsMarkDirty(dirtyClumps, todoList[ix].Lat, todoList[ix].Lng, 7.0)

			continue

		case errNoCloseVenue:
			sid := todoKeys[ix].StringID()
			addFsqTodo(ctx, sid, ntd.Lat, ntd.Lng)
		case nil:
			// pass
		default:
			log.Errorf(ctx, "attempted to activate but hit %v", err)
			continue
		}
		if addedCount > 0 {
			dirtyClumpsMarkDirty(dirtyClumps, todoList[ix].Lat, todoList[ix].Lng, 7.0)

			ntd.Lat += -0.005 + (rand.Float64() * 0.010)
			ntd.Lng += -0.005 + (rand.Float64() * 0.010)
			datastore.Put(ctx, todoKeys[ix], &ntd)
		} else {
			if rand.Float64() < 0.3 {
				sid := todoKeys[ix].StringID()
				addFsqTodo(ctx, sid, ntd.Lat, ntd.Lng)
			}
			err := datastore.Delete(ctx, todoKeys[ix])
			if err != nil {
				log.Errorf(ctx, "couldn't delete todo item, got %v", err)
			}
		}
	}
}

// Helper function for regionUp.
// find venues that aren't too crowded with other venues;
// make (but don't persist) nascent regions for them.
func regionUpNascentize(lat float64, lng float64, venues map[string]FsqVenue, regions map[string]Region, clumps map[string]*Clump) {

	vidsByScore := []string{}
	for vid, venue := range venues {
		_, already := regions[vid]
		if already {
			continue
		}
		if venue.FsqUrl == "about:tombstone" {
			continue
		}
		if dist(venue.Lat, venue.Lng, lat, lng) > regCreateRangeKm {
			continue
		}
		if rand.Float64() > 0.5 {
			continue
		}

		// insertion sort:
		ix := 0
		for ; ix < len(vidsByScore) && venues[vidsByScore[ix]].Score() > venue.Score(); ix++ {
		}
		vidsByScore = append(vidsByScore[:ix], append([]string{vid}, vidsByScore[ix:]...)...)
	}

	for _, vid := range vidsByScore {
		venue := venues[vid]

		// consider other regions and potential-regions.
		// If we would be too close to 'em, don't add.
		// If we're closer to an existing region's clump-center than that
		//   existing region is, don't add: that would put us in an existing
		//   clump and we don't want to change those.
		reject := false
		for _, otherRegion := range regions {
			if !(otherRegion.LifecycleState == rlsNascent ||
				otherRegion.LifecycleState == rlsActive) {
				continue
			}
			if dist(venue.Lat, venue.Lng, otherRegion.Lat, otherRegion.Lng) < regionsTooCloseKm {
				reject = true
				break
			}
			if otherRegion.LifecycleState == rlsActive {
				clump, found := clumps[otherRegion.Clump]
				if found {
					thisClumpDist := dist(venue.Lat, venue.Lng, clump.Lat, clump.Lng)
					rClumpDist := dist(otherRegion.Lat, otherRegion.Lng, clump.Lat, clump.Lng)
					if thisClumpDist < rClumpDist {
						reject = true
						break
					}
				}
			}
		}
		if reject {
			continue
		}
		regions[venue.ID] = Region{
			ID:             venue.ID,
			LifecycleState: rlsNascent,
			Name:           venue.Name,
			Lat:            venue.Lat,
			Lng:            venue.Lng,
			FsqUrl:         venue.FsqUrl,
		}
	}
}

func regionUp(ctx context.Context, centerLat float64, centerLng float64) (err error, addedCount int) {
	err, regions := fetchRegs(ctx, centerLat, centerLng, regCreateRangeKm+regionsTooCloseKm)
	if err != nil {
		return
	}
	venues, err := fetchFsqVenues(ctx, centerLat, centerLng, regCreateRangeKm)
	if err != nil {
		log.Errorf(ctx, "ERROR FETCHING venues %v", err)
		return
	}

	if len(venues) < minRegionsPerClump {
		err = errNotEnoughVenues
		return
	}
	closeVenueCount := 0
	for _, venue := range venues {
		if dist(venue.Lat, venue.Lng, centerLat, centerLng) < pingMaxRangeKm {
			closeVenueCount++
		}
	}
	err, clumps := fetchClumps(ctx, centerLat, centerLng, clumpMaxRadiusKm*3.5)
	if err != nil {
		return
	}

	// "nascentize" : add (but don't persist) some nascent regions in `regions`
	regionUpNascentize(centerLat, centerLng, venues, regions, clumps)
	nascentCount := 0
	for _, region := range regions {
		if region.LifecycleState == rlsNascent {
			nascentCount++
		}
	}
	if nascentCount < minRegionsPerClump {
		if closeVenueCount < 1 {
			err = errNoCloseVenue
		}
		return
	}

	// "clumpify" : try to group nascent regions into clumps.
	tempClumpNum := 0
	tcid := func() string {
		return fmt.Sprintf("temp%d", tempClumpNum)
	}
	clumps[tcid()] = &Clump{tcid(), 0, 0, 0, centerLat, centerLng, 1, []string{}, time.Now()}
	tempClumpNum++
	maxIters := 16
	for iter := 0; iter < maxIters; iter++ {
		for _, clump := range clumps {
			if clump.Tmp == 0 {
				continue
			}
			// if a nascent clump has too many kids, create another nearby clump
			if len(clump.Kids) > maxRegionsPerClump {
				lat := regions[clump.Kids[maxRegionsPerClump/2]].Lat
				lng := regions[clump.Kids[maxRegionsPerClump/2]].Lng
				clumps[tcid()] = &Clump{tcid(), 0, 0, 0, lat, lng, 1, []string{}, time.Now()}
				tempClumpNum++
			}
			// if we're exactly halfway done, try to "jiggle" away too-small clumps
			if iter == maxIters/2 && len(clump.Kids) < minRegionsPerClump {
				delete(clumps, clump.ID)
			}
		}
		// clear previous iteration state
		for ix, clump := range clumps {
			if clump.Tmp == 0 {
				continue
			}
			clumps[ix].Kids = []string{}
		}
		// for each nascent region, assign to closest nascent clump
		for _, region := range regions {
			if region.LifecycleState != rlsNascent {
				continue
			}
			clumpIDsByDist := clumpIDsByDistance(region.Lat, region.Lng, clumps)
			if len(clumpIDsByDist) < 1 {
				continue
			}
			clump := clumps[clumpIDsByDist[0]]
			if clump.Tmp != 1 {
				continue
			}
			if dist(region.Lat, region.Lng, clump.Lat, clump.Lng) > clumpMaxRadiusKm {
				continue
			}
			clumps[clumpIDsByDist[0]].Kids = append(clump.Kids, region.ID)
		}

		// for each nascent clump, re-center it to the midpoint of its kids
		for id, clump := range clumps {
			if clump.Tmp == 0 || (len(clump.Kids) < 1) {
				continue
			}
			accumLat := 0.0
			accumLng := 0.0
			for _, kid := range clump.Kids {
				accumLat += regions[kid].Lat
				accumLng += regions[kid].Lng
			}
			clumps[id].Lat = accumLat / float64(len(clump.Kids))
			clumps[id].Lng = accumLng / float64(len(clump.Kids))
		}
	}

	for _, clump := range clumps {
		if clump.Tmp != 1 {
			continue
		}
		if len(clump.Kids) < minRegionsPerClump ||
			len(clump.Kids) > maxRegionsPerClump {
			continue
		}
		if dist(centerLat, centerLng, clump.Lat, clump.Lng) > clumpCreateRangeKm {
			continue
		}

		wouldInterfereWithExistingClump := false
		for _, region := range regions {
			if region.LifecycleState != rlsActive {
				continue
			}
			otherClump, found := clumps[region.Clump]
			if !found {
				continue
			}
			if dist(region.Lat, region.Lng, otherClump.Lat, otherClump.Lng) >
				dist(region.Lat, region.Lng, clump.Lat, clump.Lng) {
				wouldInterfereWithExistingClump = true
				break
			}
		}
		if wouldInterfereWithExistingClump {
			continue
		}

		clump.SetID()
		clumpKey := datastore.NewKey(ctx, "Clump", clump.ID, 0, nil)
		transactionOptions := datastore.TransactionOptions{}
		transactionOptions.XG = true
		err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
			_, err := datastore.Put(ctx, clumpKey, clump)
			if err != nil {
				log.Errorf(ctx, "putting clump failed %v", err)
				return err
			}
			for _, kid := range clump.Kids {
				region := regions[kid]
				region.Clump = clump.ID
				region.LifecycleState = rlsActive
				kidKey := datastore.NewKey(ctx, "Region", region.ID, 0, nil)
				_, err := datastore.Put(ctx, kidKey, &region)
				if err != nil {
					log.Errorf(ctx, "putting new region failed %v", err)
					return err
				}
				memcache.Delete(ctx, fmt.Sprintf("rgs/%d", latLng2RegionBox(region.Lat, region.Lng)))
			}
			return nil
		}, &transactionOptions)
		if err != nil {
			return
		}
		addClumpAdjTodo(ctx, clump.ID, clump.Lat, clump.Lng)
		addedCount++
	}
	if closeVenueCount < 1 {
		err = errNoCloseVenue
	}
	return
}

// Region info as we present it in response
type ResponseRegion struct {
	Name   string    `json:"name"`  // E.g. "Bud's Taco Shack"
	ID     string    `json:"id"`    // 4sq id
	Color  []float32 `json:"color"` // E.g. [0.1, 0.8, 1.0] ([rgb])
	Lat    float64   `json:"lat"`
	Lng    float64   `json:"lng"`
	FsqUrl string    `json:"fsq"` // 4sq Canonical URL
}

type ResponseRoute struct {
	EndIDs []string `json:"ends"`
}

func makeResponseRegions(regs map[string]Region) map[string]([]ResponseRegion) {
	m := map[string]([]ResponseRegion){}

	for _, region := range regs {
		if region.LifecycleState != rlsActive {
			continue
		}
		r, g, b := id2Color(region.ID)
		rregion := ResponseRegion{
			region.Name,
			region.ID,
			[]float32{r, g, b},
			region.Lat,
			region.Lng,
			region.FsqUrl,
		}
		boxs := fmt.Sprintf("%d", region.RegionBox)
		_, found := m[boxs]
		if !found {
			m[boxs] = []ResponseRegion{}
		}
		m[boxs] = append(m[boxs], rregion)
	}

	return m
}

// the basic status heartbeat
func pace(w http.ResponseWriter, r *http.Request, userID string, sessionID string) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Content-Type", "application/json")
	centerLat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	centerLng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	err, regions := fetchRegs(ctx, centerLat, centerLng, pingMaxRangeKm)

	regionBoxes := makeResponseRegions(regions)
	for _, boxRange := range nearbyRegionRanges(centerLat, centerLng, pingMaxRangeKm) {
		for boxix := boxRange[0]; boxix <= boxRange[1]; boxix++ {
			boxs := fmt.Sprintf("%d", boxix)
			_, found := regionBoxes[boxs]
			if !found {
				regionBoxes[boxs] = []ResponseRegion{}
			}
		}
	}

	closeEnoughKm := (regionsTooCloseKm + pingMaxRangeKm) / 2.0
	atLeastOneCloseEnough := false
	for _, region := range regions {
		if region.LifecycleState != rlsActive {
			continue
		}
		if dist(region.Lat, region.Lng, centerLat, centerLng) < closeEnoughKm {
			atLeastOneCloseEnough = true
			break
		}
	}
	if !atLeastOneCloseEnough {
		addRupTodo(ctx, userID, centerLat, centerLng)
	}

	_, messages, goneRegs := fetchMemos(ctx, userID)

	js, err := json.Marshal(struct {
		Regions  map[string]([]ResponseRegion) `json:"regs"`
		Messages []string                      `json:"msgs,omitempty"`
		GoneRegs []string                      `json:"rmregs,omitempty"`
	}{
		regionBoxes,
		messages,
		goneRegs,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func probe(w http.ResponseWriter, r *http.Request, userID string, sessionID string) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Content-Type", "application/json")
	centerLat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	centerLng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	closeEnoughKm := (regionsTooCloseKm + pingMaxRangeKm) / 2.0

	err, regions := fetchRegs(ctx, centerLat, centerLng, closeEnoughKm)

	regionBoxes := makeResponseRegions(regions)

	atLeastOneCloseEnough := false
	for _, region := range regions {
		if region.LifecycleState != rlsActive {
			continue
		}
		if dist(region.Lat, region.Lng, centerLat, centerLng) < closeEnoughKm {
			atLeastOneCloseEnough = true
			break
		}
	}
	if !atLeastOneCloseEnough {
		addRupTodo(ctx, userID, centerLat, centerLng)
	}

	js, err := json.Marshal(struct {
		Regions map[string]([]ResponseRegion) `json:"regs,omitempty"`
	}{
		regionBoxes,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func fetchRegs(ctx context.Context, lat float64, lng float64, rangeKm float64) (err error, m map[string]Region) {
	m = map[string]Region{}
	for _, boxRange := range nearbyRegionRanges(lat, lng, rangeKm) {
		cacheMissedP := false
		for boxix := boxRange[0]; boxix <= boxRange[1]; boxix++ {
			fetched := []Region{}
			_, cerr := memcache.Gob.Get(ctx, fmt.Sprintf("rgs/%d", boxix), &fetched)
			if cerr != nil {
				if cerr != memcache.ErrCacheMiss {
					log.Errorf(ctx, "fetchRegs cache get hit %v", cerr)
				}
				cacheMissedP = true
				break
			}
			for _, region := range fetched {
				m[region.ID] = region
			}
		}
		if cacheMissedP {
			rq := datastore.NewQuery("Region").
				Filter("RegionBox >=", boxRange[0]).
				Filter("RegionBox <=", boxRange[1])
			fetched := []Region{} // TODO this looks more and more like GetAll
			for cursor := rq.Run(ctx); ; {
				region := Region{}
				_, err = cursor.Next(&region)
				if err == datastore.Done {
					err = nil
					break
				}
				if err != nil {
					log.Errorf(ctx, "ERROR FETCHING regions %v", err)
					return
				}
				fetched = append(fetched, region)
			}
			for _, region := range fetched {
				m[region.ID] = region
			}
			byBox := map[int32]([]Region){}
			for boxix := boxRange[0]; boxix <= boxRange[1]; boxix++ {
				byBox[boxix] = []Region{}
			}
			for _, region := range fetched {
				byBox[region.RegionBox] = append(byBox[region.RegionBox], region)
			}
			for boxix := boxRange[0]; boxix <= boxRange[1]; boxix++ {
				item := memcache.Item{
					Key:        fmt.Sprintf("rgs/%d", boxix),
					Object:     byBox[boxix],
					Expiration: 1 * time.Hour,
				}
				merr := memcache.Gob.Set(ctx, &item)
				if merr != nil {
					log.Errorf(ctx, "Memcache couldn't stash regs, got %v", err)
				}
			}
		}
	}
	return
}
