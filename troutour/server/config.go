package main

import (
	"encoding/json"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"net/http"
)

type ConfigStore struct {
	V string `datastore:",noindex"`
}

var configCache map[string]string = map[string]string{}

func getConfig(k string, ctx context.Context) string {
	v, found := configCache[k]
	if found {
		return v
	}
	store := ConfigStore{}
	key := datastore.NewKey(ctx, "ConfigStore", k, 0, nil)
	datastore.Get(ctx, key, &store)
	configCache[k] = store.V
	return store.V
}

// a key-value store to hold config. E.g., 4sq API keys
func configstore(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	magicWord := r.URL.Query().Get("m")
	configKey := r.URL.Query().Get("k")
	configValue := r.URL.Query().Get("v")
	magicStore := ConfigStore{}
	store := ConfigStore{}
	magicKey := datastore.NewKey(ctx, "ConfigStore", "magic", 0, nil)
	key := datastore.NewKey(ctx, "ConfigStore", configKey, 0, nil)
	datastore.Get(ctx, magicKey, &magicStore)
	w.Header().Set("Content-Type", "application/json")
	if magicWord != magicStore.V {
		http.Error(w, "Bad magic word.", http.StatusUnauthorized)
		return
	}
	if len(configValue) > 0 {
		store.V = configValue
		_, err := datastore.Put(ctx, key, &store)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			js, _ := json.Marshal(map[string]string{configKey: store.V})
			w.Write(js)
		}
		configCache = map[string]string{}
	} else {
		datastore.Get(ctx, key, &store)
		js, _ := json.Marshal(map[string]string{configKey: store.V})
		w.Write(js)
	}
}
