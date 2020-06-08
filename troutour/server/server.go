package server

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"html/template"
	"net/http"
	"strconv"
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

	http.HandleFunc("/configstore", configstore)  // key/cert/etc storage
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
	late := time.Now().Add(8 * time.Minute)
	dirtyClumps := map[int32]bool{}
	cronClumpAdj(ctx, dirtyClumps)
	if !time.Now().Before(late) {
		return
	}
	cronRegionUp(ctx, dirtyClumps)
	if !time.Now().Before(late) {
		return
	}
	downedClumpCount := cronClumpDown(ctx, dirtyClumps)
	if !time.Now().Before(late) {
		return
	}
	cronFsq(ctx, dirtyClumps)
	if !time.Now().Before(late) {
		return
	}
	if downedClumpCount < 1 {
		doomClumps(ctx, dirtyClumps)
	}
	if !time.Now().Before(late) {
		return
	}
	cronCleanupCheckins(ctx)
	time.Sleep(30 * time.Second)
}

var delayedDoQueueFunc = delay.Func("whatDoesKeyDo", doQueue)

func cronEnqueue(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	stats, err := taskqueue.QueueStats(ctx, []string{"default"})
	if err != nil {
		log.Errorf(ctx, "Tried to check queue stats, got %v", err)
		return
	}
	if len(stats) < 1 {
		log.Errorf(ctx, "Tried to check queue stats, got no stats")
		return
	}
	if stats[0].Tasks > 0 {
		// something's already queued, so don't pile on
		return
	}
	delayedDoQueueFunc.Call(ctx)
}

// Handy function for one-time admin tasks.
func magic(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Content-Type", "application/json")
	centerLat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	centerLng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if centerLat < 1.0 && centerLat > -1.0 {
		centerLat = 33.8131
	}
	if centerLng < 1.0 && centerLng > -1.0 {
		centerLng = -117.9219
	}
	places, err := fetchPlaces(ctx, centerLat, centerLng, 1.0)
	js, err := json.Marshal(struct {
		ErrS   string
		Places map[string]FsqVenue
	}{
		fmt.Sprintf("Got err %v", err),
		places,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}
