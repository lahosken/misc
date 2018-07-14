package server

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"hash/fnv"
	"time"
)

type Route struct {
	BuilderID string
	EndIDs    []string
}

func newRouteKey(ctx context.Context, builderID string, endpoint1 string, endpoint2 string) *datastore.Key {
	builderKey := newUserKey(ctx, builderID)
	if endpoint1 > endpoint2 {
		endpoint1, endpoint2 = endpoint2, endpoint1
	}
	hashable := fmt.Sprintf("%s|%s|%s|%v", builderID, endpoint1, endpoint2, time.Now())
	hasher := fnv.New64()
	hasher.Write([]byte(hashable))
	stringID := fmt.Sprintf("%v", hasher.Sum64())
	return datastore.NewKey(ctx, "Route", stringID, 0, builderKey)
}

func fetchUsersOwnRoutes(ctx context.Context, userID string) (rts map[string]Route, keys map[string]*datastore.Key, err error) {
	rts = map[string]Route{}
	keys = map[string]*datastore.Key{}
	userKey := newUserKey(ctx, userID)
	rtQ := datastore.NewQuery("Route").
		Ancestor(userKey)
	for cursor := rtQ.Run(ctx); ; {
		route := Route{}
		rtKey, err := cursor.Next(&route)
		if err == datastore.Done {
			err = nil
			break
		}
		if err != nil {
			log.Errorf(ctx, "ERROR FETCHING user's own routes %v", err)
			return rts, keys, err
		}
		sid := rtKey.StringID()
		rts[sid] = route
		keys[sid] = rtKey
	}
	return
}

func routeListRemove(before []string, remove []string) (after []string) {
	after = make([]string, len(before))
	removeHash := map[string]bool{}
	for _, r := range remove {
		removeHash[r] = true
	}
	for bix, befores := range before {
		if removeHash[befores] {
			after[bix] = ""
		} else {
			after[bix] = befores
		}
	}
	return
}

// "compact" a route-list by tossing out segments of length 0 or length 1
// (maybe this implementation isn't perfect? it might overlook some
// compact-ible parts. but maybe that's not terrible? i haven't thought about
// it much. maybe instead of searching/replacing patterns, it should build
// up a new sequence from scratch? anyhow...)
func routeListCompact(before []string) (after []string) {
	after = before
	if len(after) > 1 {
		for i := 0; i < len(after)-1; i++ {
			j := i + 1
			for after[i] == after[j] {
				after = append(after[:i], after[j:]...)
			}
		}
	}
	if len(after) > 2 {
		for i := 0; i < len(after)-2; i++ {
			j := i + 2
			for after[i] == after[j] {
				after = append(after[:i], after[j:]...)
			}
		}
	}
	for len(after) > 0 && after[0] == "" {
		after = after[1:]
	}
	for len(after) > 0 && after[len(after)-1] == "" {
		after = after[:len(after)-1]
	}
	for len(after) > 1 && after[1] == "" {
		after = after[2:]
	}
	for len(after) > 1 && after[len(after)-2] == "" {
		after = after[:len(after)-2]
	}
	if len(after) < 2 {
		return []string{}
	}
	return
}
