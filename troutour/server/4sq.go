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
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"hash/adler32"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
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
func cronFsq(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<p>cronFsq`)
	ctx := appengine.NewContext(r)

	// RegionBox values of requests we've handled this time.
	// We don't want an impatient user in one city mashing their phone button
	// to block up the pipeline and "starve" users in other cities.
	// If we see many "todo" items in one RegionBox, only handle one of them
	// this time; leave the rest for next time.
	alreadyBoxes := map[int32]bool{}

	now := time.Now()
	rand.Seed(time.Now().Unix())

	// 4sq says don't use things more than 30 days old
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)

	/*
	 * Loop through our "todo"s
	 */
	fmt.Fprintf(w, `<p>ToDos:`)
	for cursor := datastore.NewQuery("FsqTodo").Run(ctx); ; {
		ftd := FsqTodo{}
		key, err := cursor.Next(&ftd)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "reading fsq todos got error %v", err)
			continue
		}

		fmt.Fprintf(w, `<br>considering %v %v %v`, ftd.Lat, ftd.Lng, key)
		if alreadyBoxes[latLng2RegionBox(ftd.Lat, ftd.Lng)] {
			fmt.Fprintf(w, ` but alreadyBoxed`)
			continue
		}
		alreadyBoxes[latLng2RegionBox(ftd.Lat, ftd.Lng)] = true

		// What venues nearby do we already know about?
		venues, err := fetchFsqVenues(ctx, ftd.Lat, ftd.Lng, 3.0)
		if err != nil {
			fmt.Fprintf(w, ` Fetched nearby, got err %v`, err)
			continue
		}
		fmt.Fprintf(w, ` Fetched nearby, got %v`, len(venues))

		// query the 4sq API
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
		formValues := url.Values{
			"ll":     {fmt.Sprintf("%f,%f", ftd.Lat, ftd.Lng)},
			"radius": {fmt.Sprintf("%d", radiusM)},
			"time":   {"any"},
			"day":    {"any"},
			"limit":  {"50"},
		}
		authUrlValues(&formValues, ctx)
		resp, err := urlfetch.Client(ctx).Get(
			"https://api.foursquare.com/v2/venues/explore?" +
				formValues.Encode())
		if err != nil {
			log.Errorf(ctx, "Couldn't fetch 4sq data, got %v", err)
			continue
		}
		defer resp.Body.Close()

		// Count new venues discovered by this query.
		// If there are many, maybe we should query again nearby.
		newFound := 0

		// Struct specifying which values to salvage from the Json.
		// https://developer.foursquare.com/docs/venues/explore
		// https://developer.foursquare.com/docs/responses/venue
		js := struct {
			Response struct {
				Groups []struct {
					Items []struct {
						Venue struct {
							ID           string
							Name         string
							CanonicalUrl string
							Location     struct {
								Lat float64
								Lng float64
							}
							Stats map[string]int64
						}
					}
				}
			}
		}{}
		readerBuf := new(bytes.Buffer)

		// could hook up json.Decoder to resp.Body directly instead of this
		// silly buffer; only reason not to is (unused) printf debugging below.
		// TODO 2017 if the printf ain't useful anymore
		readerBuf.ReadFrom(resp.Body)

		err = json.Unmarshal(readerBuf.Bytes(), &js)
		if err != nil {
			log.Errorf(ctx, "Couldn't decode 4sq JSON, got %v", err)
			continue
		}

		// 4sq says don't use data more than 30 days old.

		for _, fgrp := range js.Response.Groups {
			for _, fitem := range fgrp.Items {
				jv := fitem.Venue
				lat := jv.Location.Lat
				lng := jv.Location.Lng
				venue := FsqVenue{
					ID:           jv.ID,
					Name:         jv.Name,
					Lat:          lat,
					Lng:          lng,
					UsersCount:   jv.Stats["usersCount"],
					FsqUrl:       jv.CanonicalUrl,
					RecentUpdate: now,
				}
				if venue.Score() < min4sqScore {
					continue
				}
				// Our cheesy map shortcuts break down if too too close to the poles
				// or the antimeridian. So ignore venues nearby.
				if math.Abs(lat) > 80.0 || math.Abs(lng) > 179.5 {
					continue
				}
				// Similarly, ignore "null island"
				if math.Abs(lat) < 0.1 && math.Abs(lng) < 0.1 {
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

		// Check for (and GC) stale data that we've stored.
		rmKeys := []*datastore.Key{}
		for _, venue := range venues {
			if thirtyDaysAgo.Before(venue.RecentUpdate) { // not stale, don't rm
				continue
			}
			// Maybe wasn't updated just because outside 4sq explore radius?
			// If so, don't GC it.
			if dist(ftd.Lat, ftd.Lng, venue.Lat, venue.Lng) > float64(radiusM)/1000.0 {
				continue
			}
			rmKeys = append(rmKeys, venue.Key(ctx))
		}
		if len(rmKeys) > 0 {
			fmt.Fprintf(w, ` Found %v old venues, attempting to delete`, len(rmKeys))
			err = datastore.DeleteMulti(ctx, rmKeys)
			if err != nil {
				fmt.Fprintf(w, ` Got err %v`, err)
			}
		}

		if newFound > 2 {
			// We found new venues. Good idea to create new regions around here.
			username := strings.Split(key.StringID(), ":")[0]
			if !strings.HasPrefix(username, "_4sq/") {
				username = "_4sq/" + username
			}
			oneItem := js.Response.Groups[len(js.Response.Groups)/2].Items[0]
			addRupTodo(ctx, username, oneItem.Venue.Location.Lat, oneItem.Venue.Location.Lng)
			// We found new venues. Maybe more lurk nearby?
			// Tweak this todo's coords and try it again next time.
			for true {
				// random-walk
				ftd.Lat += -0.01 + (0.02 * rand.Float64())
				ftd.Lng += -0.01 + (0.02 * rand.Float64())
				if !alreadyBoxes[latLng2RegionBox(ftd.Lat, ftd.Lng)] {
					break
				}
			}
			addFsqTodo(ctx, username, ftd.Lat, ftd.Lng)
		}

		// We handled this todo, so rm it
		datastore.Delete(ctx, key)
	}

	/*
	 * Look for stale stored venue data.
	 * Find some? Create "todo" items to look at next time.
	 */
	if len(alreadyBoxes) > 0 {
		return
	}
	oldFVQ := datastore.NewQuery("FsqVenue").
		Filter("RecentUpdate <", thirtyDaysAgo).
		Limit(100)
	for cursor := oldFVQ.Run(ctx); ; {
		venue := FsqVenue{}
		_, err := cursor.Next(&venue)
		if err == datastore.Done {
			break
		}
		fmt.Fprintf(w, `<p>Considering old venue`)
		if err != nil {
			fmt.Fprintf(w, ` hit err %v`, err)
			log.Errorf(ctx, "Looking for old FsqVenues, hit %v", err)
			break
		}
		if alreadyBoxes[latLng2RegionBox(venue.Lat, venue.Lng)] {
			fmt.Fprintf(w, ` ...but too close`)
			continue
		}
		if rand.Float64() > 0.3 {
			fmt.Fprintf(w, ` ...but random`)
			continue
		}
		alreadyBoxes[latLng2RegionBox(venue.Lat, venue.Lng)] = true
		addFsqTodo(ctx, "_oldVen/"+venue.ID, venue.Lat, venue.Lng)
		fmt.Fprintf(w, ` added`)
	}
}

// Add "boilerplate" fields to a request URL string:
// client info, version date, mode. As instructed by
// https://developer.foursquare.com/overview/versioning
// v: url.Values to enhance
func authUrlValues(v *url.Values, ctx context.Context) {
	fsq_client_id, fsq_client_secret := fsqConfig(ctx)
	v.Add("client_id", fsq_client_id)
	v.Add("client_secret", fsq_client_secret)
	v.Add("v", "20160704") // version
	v.Add("m", "foursquare")
}

func (v FsqVenue) Score() int64 {
	return v.UsersCount
}

func fsqConfig(ctx context.Context) (ID, Secret string) {
	clientConfig := getConfig("4sq_client", ctx)
	fields := strings.Split(clientConfig, "|")
	return fields[0], fields[1]
}

func id2Color(id string) (r, g, b float32) {
	h := float64(adler32.Checksum([]byte(id)))
	r = float32(math.Mod(h/1.0, 0.5) + 0.5)
	g = float32(math.Mod(h/100.0, 1.0))
	b = float32(math.Mod(h/10000.0, 1.0))
	return
}
