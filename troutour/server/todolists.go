package server

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"math/rand"
	"time"
)

const (
	memoCatRDown = 1 // region went away, taking your stuff with it
)

// Things we want to tell the player eventually.
// E.g., "That coffee shop you had a route to--it
// got wiped off the map, taking your routes with it."
type Memo struct {
	RecipientID string
	Category    int               `datastore:",noindex"`
	Details     map[string]string `datastore:"-"`
	DetailsJSON []byte            `datastore:",noindex"`
	When        time.Time         //  TODO: cron job to clean up messages >(some time) old
}

func (m *Memo) Load(ps []datastore.Property) error {
	err := datastore.LoadStruct(m, ps)
	if err != nil {
		return err
	}
	err = json.Unmarshal(m.DetailsJSON, &m.Details)
	if err != nil {
		return err
	}
	return nil
}

func (m *Memo) Save() ([]datastore.Property, error) {
	mm := []byte{}
	mm, err := json.Marshal(m.Details)
	if err != nil {
		return []datastore.Property{}, err
	}
	m.DetailsJSON = mm
	return datastore.SaveStruct(m)
}

func RenderMemos(memos []Memo) (htmls []string) {
	already := map[string]bool{}
	for ix, memo := range memos {
		switch memo.Category {
		case memoCatRDown:
			regionName := memo.Details["region.Name"]
			alreadyID := fmt.Sprintf("%d:%s", memoCatRDown, regionName)
			if already[alreadyID] {
				continue
			}
			already[alreadyID] = true
			effects := map[string]int{}
			for otherIx := ix; otherIx < len(memos); otherIx++ {
				otherMemo := memos[otherIx]
				if otherMemo.Category == memoCatRDown &&
					otherMemo.Details["region.Name"] == regionName {
					effects[otherMemo.Details["object"]]++
				}
			}
			if len(effects) > 1 {
				html := fmt.Sprintf("☯The region %s disappeared off the entertainment circuit, taking your contacts and resources with it.", regionName)
				htmls = append(htmls, html)
				continue
			}
			object := memo.Details["object"]
			switch object {
			case "route":
				html := ""
				if effects["route"] == 1 {
					html = fmt.Sprintf("☯The region %s disappeared off the entertainment circuit, taking your route with it.", regionName)
				} else {
					html = fmt.Sprintf("☯The region %s disappeared off the entertainment circuit, taking your routes with it.", regionName)
				}
				htmls = append(htmls, html)
				continue
			case "npc":
				html := ""
				if effects["npc"] == 1 {
					html = fmt.Sprintf("☯The region %s disappeared off the entertainment circuit, taking your client with it.", regionName)
				} else {
					html = fmt.Sprintf("☯The region %s disappeared off the entertainment circuit, taking your clients with it.", regionName)
				}
				htmls = append(htmls, html)
				continue
			default:
				continue
			}
		default:
			continue
		}
	}
	return
}

func fetchMemos(ctx context.Context, userID string) (err error, htmls []string) {
	if userID == "" {
		return
	}
	q := datastore.NewQuery("Memo").
		Filter("RecipientID =", userID).
		Limit(50)
	memos := []Memo{}
	keys, err := q.GetAll(ctx, &memos)
	if err != nil {
		return
	}
	if len(keys) < 1 {
		return
	}
	htmls = RenderMemos(memos)
	err = datastore.DeleteMulti(ctx, keys)
	return
}

// Datastore to keep track of places we want to "region up"
type RupTodo struct {
	Lat float64
	Lng float64
}

func addRupTodo(ctx context.Context, username string, lat float64, lng float64) {
	stringID := fmt.Sprintf("%s:%d", username, rand.Intn(10))
	key := datastore.NewKey(ctx, "RupTodo", stringID, 0, nil)
	ntd := RupTodo{lat, lng}
	datastore.Put(ctx, key, &ntd) // in case of failure, meh, just drop it
}

// Datastore to keep track of places we want to get 4sq info
// Keyed by username. Each user can "inspire" one place at a time.
type FsqTodo struct {
	Lat float64
	Lng float64
}

func addFsqTodo(ctx context.Context, stringID string, lat float64, lng float64) {
	key := datastore.NewKey(ctx, "FsqTodo", stringID, 0, nil)
	ftd := FsqTodo{lat, lng}
	datastore.Put(ctx, key, &ftd) // in case of failure, meh, just drop it
}

type ClumpAdjTodo struct {
	// Redundant w/Clump's own Lat, Lng.
	// Takes a smidgeon more space but saves us a lookup. Worth it? ¯\_(ツ)_/¯
	Lat float64
	Lng float64
}

func addClumpAdjTodo(ctx context.Context, stringID string, lat float64, lng float64) error {
	key := datastore.NewKey(ctx, "ClumpAdjTodo", stringID, 0, nil)
	ftd := FsqTodo{lat, lng}
	_, err := datastore.Put(ctx, key, &ftd)
	return err
}

type ClumpDownTodo struct {
	ClumpID string
}

func addClumpDownTodo(ctx context.Context, cid string) {
	key := datastore.NewKey(ctx, "ClumpDownTodo", cid, 0, nil)
	cdtd := ClumpDownTodo{cid}
	datastore.Put(ctx, key, &cdtd)
}
