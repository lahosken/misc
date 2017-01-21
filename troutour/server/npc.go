package server

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// non-player characters.

type NPC struct {
	ID        int64  `datastore:"-"` // helper functions fill this in from ID
	RegionID  string // or "" if traveling with a user/elsewise off the map
	RegionBox int32
	AgentID   string // which user is liege?
}

type ResponseNPC struct {
	RegionID string `json:"reg"`
	Yours    bool   `json:"yrs,omitempty"`
}

func randNPC() (n NPC) {
	n.RegionBox = fakeRegionBox
	return
}

func fetchNPCsNearby(ctx context.Context, lat float64, lng float64, rangeKm float64, accumulate map[int64]*NPC) (err error) {
	for _, boxRange := range nearbyRegionRanges(lat, lng, rangeKm) {
		for ix := boxRange[0]; ix <= boxRange[1]; ix++ {
			q := datastore.NewQuery("NPC").
				Filter("RegionBox >=", boxRange[0]).
				Filter("RegionBox <=", boxRange[1])
			for cursor := q.Run(ctx); ; {
				npc := NPC{}
				key, nerr := cursor.Next(&npc)
				if nerr == datastore.Done {
					break
				}
				if nerr != nil {
					err = nerr
					log.Errorf(ctx, "ERROR FETCHING nearby NPCs %v", err)
					return
				}
				npc.ID = key.IntID()
				accumulate[key.IntID()] = &npc
			}
		}
	}
	return
}

func fetchNPCsByAgent(ctx context.Context, userID string, accumulate map[int64]*NPC) (err error) {
	q := datastore.NewQuery("NPC").
		Filter("AgentID =", userID)
	for cursor := q.Run(ctx); ; {
		npc := NPC{}
		key, nerr := cursor.Next(&npc)
		if nerr == datastore.Done {
			break
		}
		if nerr != nil {
			err = nerr
			log.Errorf(ctx, "ERROR FETCHING nearby NPCs %v", err)
			return
		}
		npc.ID = key.IntID()
		accumulate[key.IntID()] = &npc
	}
	return
}

func fetchNPCsNearbyOld(ctx context.Context, lat float64, lng float64, rangeKm float64) (err error, m map[int64]NPC) {
	m = map[int64]NPC{}
	for _, boxRange := range nearbyRegionRanges(lat, lng, rangeKm) {
		for ix := boxRange[0]; ix <= boxRange[1]; ix++ {
			q := datastore.NewQuery("NPC").
				Filter("RegionBox >=", boxRange[0]).
				Filter("RegionBox <=", boxRange[1])
			for cursor := q.Run(ctx); ; {
				npc := NPC{}
				key, nerr := cursor.Next(&npc)
				if nerr == datastore.Done {
					break
				}
				if nerr != nil {
					err = nerr
					log.Errorf(ctx, "ERROR FETCHING nearby NPCs %v", err)
					return
				}
				npc.ID = key.IntID()
				m[key.IntID()] = npc
			}
		}
	}
	return
}

func fetchNPCsByAgentOld(ctx context.Context, userID string) (err error, m map[int64]NPC) {
	m = map[int64]NPC{}
	q := datastore.NewQuery("NPC").
		Filter("AgentID =", userID)
	for cursor := q.Run(ctx); ; {
		npc := NPC{}
		key, nerr := cursor.Next(&npc)
		if nerr == datastore.Done {
			break
		}
		if nerr != nil {
			err = nerr
			log.Errorf(ctx, "ERROR FETCHING nearby NPCs %v", err)
			return
		}
		npc.ID = key.IntID()
		m[key.IntID()] = npc
	}
	return
}
