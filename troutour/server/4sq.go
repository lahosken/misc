package server

/*
 * Regions (places in the game) are based on 4square data. We query 4sq and
 * store the results, then use those results when creating Regions.
 * When the regionUp command tries to create regions somewhere, it fetches
 * venue data from our local store. If that fetches no venues, it makes a
 * "todo" to query
 *
 * Given a Region, don't assume we have a stored FsqVenue for that region.
 * We had one back when the region was _created_, but may have since GC'd it.
 *
 * (The "store" is so we don't bump into 4sq rate limits re-querying the same
 * area too often. Those rate limits are pretty generous, tho.)
 *
 */

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"hash/adler32"
	"hash/fnv"
	"math"
	"math/rand"
	// "net/http"
	"net/url"
	"time"
)

const (
	min4sqScore = 50
)

// stored info about a 4square venue
type FsqVenue struct {
	ID           string `datastore:",noindex"`
	RegionBox    int32
	Name         string  `datastore:",noindex"`
	LatX10000    int32   `datastore:",noindex"`
	LngX10000    int32   `datastore:",noindex"`
	Lat          float64 `datastore:"-"`
	Lng          float64 `datastore:"-"`
	UsersCount   int64   `datastore:",noindex"`
	FsqUrl       string  `datastore:",noindex"`
	RecentUpdate time.Time
}

func (fsv *FsqVenue) Load(ps []datastore.Property) error {
	err := datastore.LoadStruct(fsv, ps)
	if err != nil {
		return err
	}
	fsv.Lat = float64(fsv.LatX10000) / 10000.0
	fsv.Lng = float64(fsv.LngX10000) / 10000.0
	return nil
}

func (fsv *FsqVenue) Save() ([]datastore.Property, error) {
	fsv.LatX10000 = int32(fsv.Lat * 10000.0)
	fsv.LngX10000 = int32(fsv.Lng * 10000.0)
	fsv.RegionBox = latLng2RegionBox(fsv.Lat, fsv.Lng)
	return datastore.SaveStruct(fsv)
}

func (fsv *FsqVenue) Key(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, "FsqVenue", fsv.ID, 0, nil)
}

// fetch place info from wikipedia (geonames), google
func fetchPlaces(ctx context.Context, lat float64, lng float64, rangeKm float64) (m map[string]FsqVenue, err error) {
	m = map[string]FsqVenue{}

	now := time.Now()

	if rangeKm < 1.5 {
		// Do we know about any "tombstones" nearby? If so, don't bother talking
		// to APIs, they're not likely to tell us anything.
		venues, fetchFSVErr := fetchFsqVenues(ctx, lat, lng, 0.75)
		if fetchFSVErr == nil {
			sawTombstone := false
			for _, venue := range venues {
				if (venue.FsqUrl == "about:tombstone") &&
					(distKm(lat, lng, venue.Lat, venue.Lng) < 0.75) {
					sawTombstone = true
					break
				}
			}
			if sawTombstone {
				return
			}
		}
	}

	// fetch from geonames
	{
		formValues := url.Values{
			"lat":      {fmt.Sprintf("%f", lat)},
			"lng":      {fmt.Sprintf("%f", lng)},
			"radius":   {fmt.Sprintf("%f", rangeKm)},
			"username": {getConfig("gn_client", ctx)},
			"maxRows":  {"30"},
		}
		resp, err := urlfetch.Client(ctx).Get(
			"http://api.geonames.org/findNearbyWikipediaJSON?" +
				formValues.Encode())
		if err != nil {
			log.Errorf(ctx, "Couldn't fetch geonames data, got %v", err)
			return m, err
		}
		defer resp.Body.Close()
		js := struct {
			Geonames []struct {
				Lat          float64
				Lng          float64
				Rank         int64
				Title        string
				WikipediaUrl string
			}
		}{}
		err = json.NewDecoder(resp.Body).Decode(&js)
		if err != nil {
			log.Errorf(ctx, "Couldn't decode geonames JSON, got %v", err)
			return m, err
		}
		for _, gnitem := range js.Geonames {
			venueIDString := fmt.Sprintf("%s:%s (%f.%f)", gnitem.WikipediaUrl, gnitem.Title, gnitem.Lat, gnitem.Lng)
			hasher := fnv.New64()
			hasher.Write([]byte(venueIDString))
			venue := FsqVenue{
				ID:           fmt.Sprintf("%v", hasher.Sum64()),
				Name:         gnitem.Title,
				Lat:          gnitem.Lat,
				Lng:          gnitem.Lng,
				UsersCount:   gnitem.Rank + 30,
				FsqUrl:       fmt.Sprintf("//%s", gnitem.WikipediaUrl),
				RecentUpdate: now,
			}
			// Our cheesy map shortcuts break down if too too close to the poles
			// or the antimeridian. So ignore venues nearby.
			if math.Abs(gnitem.Lat) > 80.0 || math.Abs(gnitem.Lng) > 179.5 {
				continue
			}
			// Similarly, ignore "null island"
			if math.Abs(gnitem.Lat) < 0.1 && math.Abs(gnitem.Lng) < 0.1 {
				continue
			}
			m[venue.ID] = venue
		}
	}

	// if we found plenty of items in geonames, don't bother to query goog
	if float64(len(m)) > (rangeKm*rangeKm*3.14*2.0)+1.0 {
		return
	}

	// fetch data frome google maps
	{
		formValues := url.Values{
			"key":      {getConfig("google_places_api", ctx)},
			"location": {fmt.Sprintf("%f,%f", lat, lng)},
			"radius":   {fmt.Sprintf("%d", int64(1000.0*rangeKm))},
		}
		resp, err := urlfetch.Client(ctx).Get(
			"https://maps.googleapis.com/maps/api/place/nearbysearch/json?" +
				formValues.Encode())

		if err != nil {
			log.Errorf(ctx, "Couldn't fetch goog data, got %v", err)
			return m, err
		}
		defer resp.Body.Close()
		js := struct {
			Html_Attributions []string
			Results           []struct {
				Geometry struct {
					Location struct {
						Lat float64
						Lng float64
					}
				}
				Name     string
				Place_Id string
			}
		}{}
		err = json.NewDecoder(resp.Body).Decode(&js)
		if err != nil {
			log.Errorf(ctx, "Couldn't decode goog JSON, got %v", err)
			return m, err
		}
		// If there are html_attributions, google license says we
		// have to display them "with results". but since we don't
		// really display results as such, we can't do that. so...
		// if html_attributions is filled in, ignore these unusable results
		if len(js.Html_Attributions) == 0 && len(js.Results) > 0 {
			ranking := len(js.Results) + 2
			for _, gItem := range js.Results {
				// Our cheesy map shortcuts break down if too too close to the poles
				// or the antimeridian. So ignore venues nearby.
				if math.Abs(gItem.Geometry.Location.Lat) > 80.0 || math.Abs(gItem.Geometry.Location.Lng) > 179.5 {
					continue
				}
				// Similarly, ignore "null island"
				if math.Abs(gItem.Geometry.Location.Lat) < 0.1 && math.Abs(gItem.Geometry.Location.Lng) < 0.1 {
					continue
				}
				// goog gives locations that are too far away, so skip those.
				// (It's trying to be helpful: it gives far away places
				//  that enclose our spot. E.g., if you're at the Golden Gate
				//  Bridge, it tells you that you're in San Francisco, though
				//  its SF mark isn't within the requested radius.)
				if distKm(
					gItem.Geometry.Location.Lat, gItem.Geometry.Location.Lng,
					lat, lng) > rangeKm {
					continue
				}
				ranking -= 1
				venueIDString := fmt.Sprintf("g:%s", gItem.Place_Id)
				hasher := fnv.New64()
				hasher.Write([]byte(venueIDString))
				venue := FsqVenue{
					ID:           fmt.Sprintf("%v", hasher.Sum64()),
					Name:         gItem.Name,
					Lat:          gItem.Geometry.Location.Lat,
					Lng:          gItem.Geometry.Location.Lng,
					UsersCount:   int64(ranking),
					FsqUrl:       fmt.Sprintf("https://www.google.com/maps/search/?api=1&query_place_id=%s&query=%s", gItem.Place_Id, gItem.Name),
					RecentUpdate: now,
				}
				m[venue.ID] = venue
			}
		}
	}
	if rangeKm >= 1.0 && len(m) < 1 {
		// We got nuthin'. So save a "tombstone", a fake venue
		// that later on tells us there's no point repeatedly
		// searching around here.
		venue := FsqVenue{
			ID:           fmt.Sprintf("tomb:%v,%v", lat, lng),
			Name:         "about:tombstone",
			Lat:          lat,
			Lng:          lng,
			UsersCount:   0,
			FsqUrl:       fmt.Sprintf("about:tombstone"),
			RecentUpdate: now,
		}
		_, err := datastore.Put(ctx, venue.Key(ctx), &venue)
		if err != nil {
			log.Errorf(ctx, "Couldn't save tombstone, got %v", err)
		}
	}
	return
}

// fetch our stored 4sq data
func fetchFsqVenues(ctx context.Context, lat float64, lng float64, rangeKm float64) (m map[string]FsqVenue, err error) {
	m = map[string]FsqVenue{}
	for _, boxRange := range nearbyRegionRanges(lat, lng, rangeKm) {
		fvq := datastore.NewQuery("FsqVenue").
			Filter("RegionBox >=", boxRange[0]).
			Filter("RegionBox <=", boxRange[1])
		for cursor := fvq.Run(ctx); ; {
			venue := FsqVenue{}
			_, err = cursor.Next(&venue)
			if err == datastore.Done {
				err = nil
				break
			}
			if err != nil {
				log.Errorf(ctx, "ERROR FETCHING venues %v", err)
				return
			}
			m[venue.ID] = venue
		}
	}
	return
}

// We have a "todo" list of lat/lngs. For each of those,
// query 4sq API. Record new data. GC stale data.
func cronFsq(ctx context.Context, dirtyClumps map[int32]bool) {
	// ctx := appengine.NewContext(r)

	now := time.Now()
	late := now.Add(90 * time.Second)
	rand.Seed(time.Now().Unix())

	// 4sq says don't use things more than 30 days old
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)

	/*
	 * Loop through our "todo"s
	 */
	for cursor := datastore.NewQuery("FsqTodo").Run(ctx); ; {
		if !time.Now().Before(late) {
			return
		}
		ftd := FsqTodo{}
		key, err := cursor.Next(&ftd)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "reading fsq todos got error %v", err)
			continue
		}

		if dirtyClumps[latLng2ClumpBox(ftd.Lat, ftd.Lng)] {
			continue
		}
		dirtyClumpsMarkDirty(dirtyClumps, ftd.Lat, ftd.Lng, 10.0)

		// What venues nearby do we already know about?
		venues, err := fetchFsqVenues(ctx, ftd.Lat, ftd.Lng, 3.0)
		if err != nil {
			continue
		}
		sawTombstone := false
		for _, venue := range venues {
			if (venue.FsqUrl == "about:tombstone") &&
				(distKm(ftd.Lat, ftd.Lng, venue.Lat, venue.Lng) < 0.75) {
				sawTombstone = true
				break
			}
		}
		if sawTombstone {
			// We saw a "tombstone" venue, a fake venue that says
			// we've previously asked GeoNames for venues around here and
			// got no results
			datastore.Delete(ctx, key)
			continue
		}

		// query the geonames, google APIs
		// radiusM: Do we already know many venues nearby?
		//  If so, choose small radius, get "denser" knowledge of venues.
		//  If not, choose large radius, find semi-nearby venues.
		radiusM := (10000 / (len(venues) + 4)) + rand.Intn(2000)
		if radiusM < 1000 {
			radiusM = 1000
		}
		if radiusM > 3000 {
			radiusM = 3000
		}

		// Count new venues discovered by this query.
		// If there are many, maybe we should query again nearby.
		newFound := 0

		// BEGIN
		// the part of this "4sq" code that fetches data from geonames
		{
			formValues := url.Values{
				"lat":      {fmt.Sprintf("%f", ftd.Lat)},
				"lng":      {fmt.Sprintf("%f", ftd.Lng)},
				"radius":   {fmt.Sprintf("%f", float64(radiusM)/1000.0)},
				"username": {getConfig("gn_client", ctx)},
				"maxRows":  {"30"},
			}
			resp, err := urlfetch.Client(ctx).Get(
				"http://api.geonames.org/findNearbyWikipediaJSON?" +
					formValues.Encode())
			if err != nil {
				log.Errorf(ctx, "Couldn't fetch geonames data, got %v", err)
				continue
			}
			defer resp.Body.Close()
			js := struct {
				Geonames []struct {
					Lat          float64
					Lng          float64
					Rank         int64
					Title        string
					WikipediaUrl string
				}
			}{}
			err = json.NewDecoder(resp.Body).Decode(&js)
			if err != nil {
				log.Errorf(ctx, "Couldn't decode geonames JSON, got %v", err)
				continue
			}

			if len(js.Geonames) > 0 {
				for _, gnitem := range js.Geonames {
					venueIDString := fmt.Sprintf("%s:%s (%f.%f)", gnitem.WikipediaUrl, gnitem.Title, gnitem.Lat, gnitem.Lng)
					hasher := fnv.New64()
					hasher.Write([]byte(venueIDString))
					venue := FsqVenue{
						ID:           fmt.Sprintf("%v", hasher.Sum64()),
						Name:         gnitem.Title,
						Lat:          gnitem.Lat,
						Lng:          gnitem.Lng,
						UsersCount:   gnitem.Rank + 30,
						FsqUrl:       fmt.Sprintf("//%s", gnitem.WikipediaUrl),
						RecentUpdate: now,
					}
					// Our cheesy map shortcuts break down if too too close to the poles
					// or the antimeridian. So ignore venues nearby.
					if math.Abs(gnitem.Lat) > 80.0 || math.Abs(gnitem.Lng) > 179.5 {
						continue
					}
					// Similarly, ignore "null island"
					if math.Abs(gnitem.Lat) < 0.1 && math.Abs(gnitem.Lng) < 0.1 {
						continue
					}
					_, err := datastore.Put(ctx, venue.Key(ctx), &venue)
					if err != nil {
						log.Errorf(ctx, "Couldn't save venue, got %v", err)
						continue
					}
					_, found := venues[venue.ID]
					if !found {
						newFound++
					}
					venues[venue.ID] = venue
				}
			}
		}
		// END
		// the part of this "4sq" code that actually fetches data from geonames

		// BEGIN
		// the part of this "4sq" code that fetches data from google maps
		if newFound < 4 {
			formValues := url.Values{
				"key":      {getConfig("google_places_api", ctx)},
				"location": {fmt.Sprintf("%f,%f", ftd.Lat, ftd.Lng)},
				"radius":   {fmt.Sprintf("%d", radiusM)},
			}

			resp, err := urlfetch.Client(ctx).Get(
				"https://maps.googleapis.com/maps/api/place/nearbysearch/json?" +
					formValues.Encode())

			if err != nil {
				log.Errorf(ctx, "Couldn't fetch goog data, got %v", err)
				continue
			}
			defer resp.Body.Close()
			js := struct {
				Html_Attributions []string
				Results           []struct {
					Geometry struct {
						Location struct {
							Lat float64
							Lng float64
						}
					}
					Name     string
					Place_Id string
				}
			}{}
			err = json.NewDecoder(resp.Body).Decode(&js)
			if err != nil {
				log.Errorf(ctx, "Couldn't decode goog JSON, got %v", err)
				continue
			}

			// If there are html_attributions, google license says we
			// have to display them "with results". but since we don't
			// really display results as such, we can't do that. so...
			// if html_attributions is filled in, discard these unusable results
			if len(js.Html_Attributions) == 0 && len(js.Results) > 0 {
				ranking := len(js.Results) + 1
				for _, gItem := range js.Results {
					// Our cheesy map shortcuts break down if too too close to the poles
					// or the antimeridian. So ignore venues nearby.
					if math.Abs(gItem.Geometry.Location.Lat) > 80.0 || math.Abs(gItem.Geometry.Location.Lng) > 179.5 {
						continue
					}
					// Similarly, ignore "null island"
					if math.Abs(gItem.Geometry.Location.Lat) < 0.1 && math.Abs(gItem.Geometry.Location.Lng) < 0.1 {
						continue
					}
					// goog gives locations that are too far away, so skip those.
					// (It's trying to be helpful: it gives far away places
					//  that enclose our spot. E.g., if you're at the Golden Gate
					//  Bridge, it tells you that you're in San Francisco, though
					//  its SF mark isn't within the requested radius.)
					if distKm(
						gItem.Geometry.Location.Lat, gItem.Geometry.Location.Lng,
						ftd.Lat, ftd.Lng) > (float64(radiusM) / 1000.0) {
						continue
					}
					ranking -= 1
					venueIDString := fmt.Sprintf("g:%s", gItem.Place_Id)
					hasher := fnv.New64()
					hasher.Write([]byte(venueIDString))
					venue := FsqVenue{
						ID:           fmt.Sprintf("%v", hasher.Sum64()),
						Name:         gItem.Name,
						Lat:          gItem.Geometry.Location.Lat,
						Lng:          gItem.Geometry.Location.Lng,
						UsersCount:   int64(ranking),
						FsqUrl:       fmt.Sprintf("https://www.google.com/maps/search/?api=1&query_place_id=%s&query=%s", gItem.Place_Id, gItem.Name),
						RecentUpdate: now,
					}
					_, err := datastore.Put(ctx, venue.Key(ctx), &venue)
					if err != nil {
						log.Errorf(ctx, "Couldn't save venue, got %v", err)
						continue
					}
					_, found := venues[venue.ID]
					if !found {
						newFound++
					}
					venues[venue.ID] = venue
				}
			}
		}
		// END
		// the part of this "4sq" code that actually fetches data from google maps

		if newFound < 1 {
			// We got nuthin'. So save a "tombstone", a fake venue
			// that later on tells us there's no point repeatedly
			// searching around here.
			venue := FsqVenue{
				ID:           fmt.Sprintf("tomb:%v,%v", ftd.Lat, ftd.Lng),
				Name:         "",
				Lat:          ftd.Lat,
				Lng:          ftd.Lng,
				UsersCount:   0,
				FsqUrl:       fmt.Sprintf("about:tombstone"),
				RecentUpdate: now,
			}
			_, err := datastore.Put(ctx, venue.Key(ctx), &venue)
			if err != nil {
				log.Errorf(ctx, "Couldn't save tombstone, got %v", err)
			}
		}

		// 4sq says don't use data more than 30 days old.
		// Check for (and GC) stale data that we've stored.
		rmKeys := []*datastore.Key{}
		for _, venue := range venues {
			if thirtyDaysAgo.Before(venue.RecentUpdate) { // not stale, don't rm
				continue
			}
			rmKeys = append(rmKeys, venue.Key(ctx))
		}
		if len(rmKeys) > 0 {
			err = datastore.DeleteMulti(ctx, rmKeys)
		}

		/*
			if newFound > 2 {
				// We found new venues. Good idea to create new regions around here.
				username := strings.Split(key.StringID(), ":")[0]
				if !strings.HasPrefix(username, "_4sq/") {
					username = "_4sq/" + username
				}
				addRupTodo(ctx, username, ftd.Lat, ftd.Lng)
				// We found new venues. Maybe more lurk nearby?
				// Tweak this todo's coords and try it again next time.
				for true {
					// random-walk
					ftd.Lat += -0.01 + (0.02 * rand.Float64())
					ftd.Lng += -0.01 + (0.02 * rand.Float64())
				}
				addFsqTodo(
					ctx,
					username,
					ftd.Lat-0.01+(0.02*rand.Float64()),
					ftd.Lng-0.01+(0.02*rand.Float64()))
			}
		*/

		// We handled this todo, so rm it
		datastore.Delete(ctx, key)
	}

	/*
	 * Look for stale stored venue data.
	 * Find some? Create "todo" items to look at next time.
	 */
	oldFVQ := datastore.NewQuery("FsqVenue").
		Filter("RecentUpdate <", thirtyDaysAgo).
		Limit(100)
	for cursor := oldFVQ.Run(ctx); ; {
		if !time.Now().Before(late) {
			return
		}
		venue := FsqVenue{}
		_, err := cursor.Next(&venue)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "Looking for old FsqVenues, hit %v", err)
			break
		}
	}
}

func (v FsqVenue) Score() int64 {
	return v.UsersCount
}

func id2Color(id string) (r, g, b float32) {
	h := float64(adler32.Checksum([]byte(id)))
	r = float32(math.Mod(h/1.0, 0.5) + 0.5)
	g = float32(math.Mod(h/100.0, 1.0))
	b = float32(math.Mod(h/10000.0, 1.0))
	return
}
