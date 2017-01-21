package server

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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
	stringID := endpoint1 + "|" + endpoint2
	return datastore.NewKey(ctx, "Route", stringID, 0, builderKey)
}

func fetchUsersOwnRoutes(ctx context.Context, userID string) (rts map[string]Route, err error) {
	rts = map[string]Route{}
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
			return rts, err
		}
		rts[rtKey.StringID()] = route
	}
	return
}
