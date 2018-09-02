package server

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"hash/fnv"
	"time"
)

/* When a user "checks in" at a location, we load existing route info
 * and also want to keep track of new/changed routes created/edited
 * during checkin.
 */
type CheckinRoutesPersistable struct {
	Routes         map[string]Route
	RouteKeys      map[string]*datastore.Key
	NewRoutes      map[string]Route
	AppendedRoutes map[string]Route
}

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

func hasNewRoutesP(crp CheckinRoutesPersistable) (newP bool) {
	return len(crp.NewRoutes)+len(crp.AppendedRoutes) > 0
}

func fetchUsersOwnRoutes(ctx context.Context, userID string) (crp CheckinRoutesPersistable, err error) {
	crp.Routes = map[string]Route{}
	crp.RouteKeys = map[string]*datastore.Key{}
	crp.NewRoutes = map[string]Route{}
	crp.AppendedRoutes = map[string]Route{}
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
			return crp, err
		}
		sid := rtKey.StringID()
		crp.Routes[sid] = route
		crp.RouteKeys[sid] = rtKey
	}
	return
}

func crpAddLeg(crp CheckinRoutesPersistable, hereRegionID string, thereRegionID string, userID string) {
	appendedRoute := false
	for oldRouteKey, oldRoute := range crp.Routes {
		if len(oldRoute.EndIDs) > 9 { // don't keep appending to "long" strand
			continue
		}
		if oldRoute.EndIDs[len(oldRoute.EndIDs)-1] == thereRegionID {
			oldRoute.EndIDs = append(oldRoute.EndIDs, hereRegionID)
			crp.AppendedRoutes[oldRouteKey] = oldRoute
			appendedRoute = true
			break
		}
	}
	if !appendedRoute {
		crp.NewRoutes[thereRegionID] = Route{userID, []string{thereRegionID, hereRegionID}}
	}
}

func crpPersist(ctx context.Context, crp CheckinRoutesPersistable, userID string) (err error) {
	for _, route := range crp.NewRoutes {
		routeKey := newRouteKey(ctx, userID, route.EndIDs[0], route.EndIDs[1])
		_, err = datastore.Put(ctx, routeKey, &route)
		if err != nil {
			log.Errorf(ctx, "error saving new route, hit %v", err)
			return
		}
	}
	for k, route := range crp.AppendedRoutes {
		routeKey := crp.RouteKeys[k]
		_, err = datastore.Put(ctx, routeKey, &route)
		if err != nil {
			log.Errorf(ctx, "error saving appended route, hit %v", err)
			return
		}
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
