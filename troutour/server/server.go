package server

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/delay"
	//	"google.golang.org/appengine/datastore"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"html/template"
	"net/http"
	"time"
)

func init() {
	http.HandleFunc("/", Sessionate(topscreen))
	http.HandleFunc("/a/pace", Sessionate(pace))       // player wants nearby info
	http.HandleFunc("/a/probe", Sessionate(probe))     // player info 1km "ahead"
	http.HandleFunc("/a/checkin", Sessionate(checkin)) // pressed the button

	http.HandleFunc("/a/magic", magic) // handy URL for one-time admin tasks

	// callback for OpenID flow
	http.HandleFunc("/oauth2_callback_goog", Sessionate(oauth2CallbackGoog))
	http.HandleFunc("/logout", Sessionate(logout))

	http.HandleFunc("/configstore", configstore) // key/cert/etc storage
	// http.HandleFunc("/cron/fsq", cronFsq)             // query 4square
	// http.HandleFunc("/cron/rup", cronRegionUp)        // create regions
	// http.HandleFunc("/cron/clumpadj", cronClumpAdj)   // compute "close" regions
	// http.HandleFunc("/cron/clumpdown", cronClumpDown) // destroy regions
	// http.HandleFunc("/cron/ccc", cronCleanupCheckins) // GC checkins
	http.HandleFunc("/cron/enqueue", cronEnqueue) // post job to task queue
}

// Show the main screen.
func topscreen(w http.ResponseWriter, r *http.Request, userID string, sessionID string) {
	ctx := appengine.NewContext(r)
	log.Infof(ctx, "topscreen userID=%v", userID)
	googOAuth2ID, _ := googOAuth2Config(ctx)
	googleAuthURL := ""
	if userID == "" {
		loginToken := actionToken(sessionID, "googAuth")
		googleAuthURL = fmt.Sprintf(`https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&response_type=code&scope=openid&redirect_uri=https://rovercast-1372.appspot.com/oauth2_callback_goog&state=%s|/`, googOAuth2ID, loginToken)
	}
	template.Must(template.New("").Parse(tmplS)).Execute(w, struct {
		UserID        string
		GoogleAuthURL string
	}{
		UserID:        userID,
		GoogleAuthURL: googleAuthURL,
	})
}

func doQueue(ctx context.Context) {
	dirtyClumps := map[int32]bool{}
	cronClumpAdj(ctx, dirtyClumps)
	cronRegionUp(ctx, dirtyClumps)
	cronClumpDown(ctx, dirtyClumps)
	cronFsq(ctx)
	cronCleanupCheckins(ctx)
	time.Sleep(30 * time.Second)
}

var delayedDoQueueFunc = delay.Func("whatDoesKeyDo", doQueue)

func cronEnqueue(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	delayedDoQueueFunc.Call(ctx)
}

// Handy function for one-time admin tasks.
func magic(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	err, addedCount := regionUp(ctx, 37.7746, -122.4101)
	if err != nil {
		fmt.Fprintf(w, "<p>Magic sees err = %v", err)
	}
	fmt.Fprintf(w, "<p>Magic sees addedCount = %v", addedCount)
	/*
		thirtySecondsFromStart := time.Now().Add(30 * time.Second)
		types := []string{"FsqVenue", "Region", "NPC", "Route", "Clump", "ClumpAdj"}
		for {
			if thirtySecondsFromStart.Before(time.Now()) {
				break
			}
			found := 0
			for _, t := range types {
				q := datastore.NewQuery(t).Limit(100).KeysOnly()
				keys, err := q.GetAll(ctx, nil)
				if err != nil {
					fmt.Fprintf(w, "couldn't fetch keys for %s got err %v", t, err)
					return
				}
				err = datastore.DeleteMulti(ctx, keys)
				if err != nil {
					fmt.Fprintf(w, "couldn't delete keys for %s got err %v", t, err)
					return
				}
				found += len(keys)
			}
			fmt.Fprintf(w, "Deleted %d entities", found)
			if found < 100 {
				break
			}
		}
	*/

}
