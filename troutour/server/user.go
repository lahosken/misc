package server

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type UserInventory struct {
	Prisms        []string
	Bounties      []string
	Coins         int64
	Trophies      int64
	Cred          float64
	TransientCred float64
}

type ResponseUserInventory struct {
	Prisms   []string `json:"prisms"`
	Bounties []string `json:"bounties"`
	Coins    int64    `json:"coins"`
	Trophies int64    `json:"trophies"`
	Cred     int64    `json:"cred"`
}

func (ui UserInventory) makeResponseInventory() (rui ResponseUserInventory) {
	rui.Prisms = ui.Prisms
	rui.Bounties = ui.Bounties
	rui.Coins = ui.Coins
	rui.Trophies = ui.Trophies
	rui.Cred = int64(ui.Cred * 100.1)
	return
}

func newUserKey(ctx context.Context, userID string) *datastore.Key {
	return datastore.NewKey(ctx, "User", userID, 0, nil)
}

func stSliceContains(sl []string, st string) bool {
	for _, s := range sl {
		if s == st {
			return true
		}
	}
	return false
}
