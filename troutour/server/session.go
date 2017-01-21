package server

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Session struct {
	Created time.Time
	UserID  string
}

const (
	sessionLifetimeSeconds = 8388608 // nice round number, ~90 days in seconds
)

var (
	errSessionAlreadyExists = errors.New("Keep going")
)

type sessionHandlerFunc func(w http.ResponseWriter, r *http.Request, userID string, sessionID string)

func actionToken(sessionID string, verb string) (token string) {
	h := sha512.New()
	h.Write([]byte(sessionID))
	h.Write([]byte(verb))
	return fmt.Sprintf("%x", h.Sum([]byte{}))
}

func Sessionate(handler sessionHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		sessionID := ""
		userID := ""
		session := Session{}
		sessionLoadedAlready := false
		c, err := r.Cookie("sid")
		if err == nil {
			sessionID = c.Value
		}

		if sessionID == "" {
			rand.Seed(int64(time.Now().Nanosecond()))
			userID = ""
			if appengine.IsDevAppServer() {
				userID = fmt.Sprintf("_dev.%v", rand.Int())
			}
			session = Session{time.Now(), userID}
			for {
				err := datastore.RunInTransaction(ctx, func(context context.Context) error {
					sessionID = fmt.Sprintf("%d", rand.Int())
					key := datastore.NewKey(ctx, "Session", sessionID, 0, nil)
					err := datastore.Get(ctx, key, &session)
					if err != datastore.ErrNoSuchEntity {
						return errSessionAlreadyExists
					}
					_, err = datastore.Put(ctx, key, &session)
					if err != nil {
						return err
					}
					sessionLoadedAlready = true
					memcache.Set(ctx, &memcache.Item{
						Key:   "Session/" + sessionID,
						Value: []byte(userID),
					})
					return nil
				}, nil)
				if err == nil {
					break
				}
			}
			c = new(http.Cookie)
			c.Name = "sid"
			c.Value = sessionID
			c.Path = "/"
			c.MaxAge = sessionLifetimeSeconds
			http.SetCookie(w, c)
		}
		// sessionID is set.
		if !sessionLoadedAlready {
			cache, err := memcache.Get(ctx, "Session/"+sessionID)
			if err == nil {
				userID = string(cache.Value)
			} else {
				key := datastore.NewKey(ctx, "Session", sessionID, 0, nil)
				err = datastore.Get(ctx, key, &session)
				if err != nil {
					session.Created = time.Now()
					datastore.Put(ctx, key, &session)
				}
				userID = session.UserID
				if userID == "" && appengine.IsDevAppServer() {
					userID = fmt.Sprintf("_dev.%v", sessionID)
				}
				memcache.Set(ctx, &memcache.Item{
					Key:   "Session/" + sessionID,
					Value: []byte(userID),
				})
			}
		}
		handler(w, r, userID, sessionID)
	}
}

func oauth2CallbackGoog(w http.ResponseWriter, r *http.Request, userID string, sessionID string) {
	ctx := appengine.NewContext(r)
	if userID != "" {
		fmt.Fprintf(w, `<p>Mighty strange to see a login attempt when already logged in, yep. TODO this could be better %v <a href="/client/index.html">ARGH</a>`, userID)
		return
	}
	fmt.Fprintf(w, "<html><p>I see args %v", r.URL.Query())
	error := r.URL.Query().Get("error")
	if error == "access_denied" {
		fmt.Fprintf(w, `<p>Got "access denied" from Google; I guess you changed your mind about logging in?`)
		fmt.Fprintf(w, `<p>Redirecting back to main page in a few seconds&hellip; <script>setTimeout(function(){ document.location='/'}, 5 * 1000)</script>`)
	}
	state := r.URL.Query().Get("state")
	stateFields := strings.Split(state, "|")
	if len(stateFields) != 2 {
		fmt.Fprintf(w, `<p>Mighty strange state from el goog, yep "%v" <a href="/client/index.html">ARGH</a>`, state)
		return

	}
	loginToken := stateFields[0]
	if loginToken != actionToken(sessionID, "googAuth") {
		fmt.Fprintf(w, `<p>Got token "%v" but expected "%v", arg <a href="/client/index.html">ARGH</a>`, loginToken, actionToken(sessionID, "googAuth"))
		return

	}
	redirURL := stateFields[1]
	authCode := r.URL.Query().Get("code")

	// Use the authCode to ask for the client ID
	googOAuth2ID, googOAuth2Secret := googOAuth2Config(ctx)
	fmt.Fprintf(w, "<hr>")
	v := url.Values{
		"code":          {authCode},
		"client_id":     {googOAuth2ID},
		"client_secret": {googOAuth2Secret},
		"redirect_uri":  {"https://rovercast-1372.appspot.com/oauth2_callback_goog"}, // wtf
		"grant_type":    {"authorization_code"}}
	fmt.Fprintf(w, "going to fetch ID from goog with values %v", v)
	resp, err := urlfetch.Client(ctx).PostForm(
		"https://www.googleapis.com/oauth2/v4/token", v)
	if err != nil {
		log.Errorf(ctx, "Couldn't get ID from google got err %v", err)
		fmt.Fprintf(w, "<p>Couldn't get ID from google got err %v", err)
		return
	}
	defer resp.Body.Close()
	json0 := struct {
		ID_Token string
	}{}
	err = json.NewDecoder(resp.Body).Decode(&json0)
	if err != nil {
		log.Errorf(ctx, "Couldn't JSON-decode response from google, got err %v", err)
		fmt.Fprintf(w, "<p>Couldn't JSON-decode response from google, got err %v", err)
		return
	}
	fmt.Fprintf(w, "<p>Got id_token: %s \n", json0.ID_Token)
	if strings.Count(json0.ID_Token, ".") != 2 {
		log.Errorf(ctx, "I assume there are 2 dots in %s, but there aren't really", json0.ID_Token)
		fmt.Fprintf(w, "<p>I assume there are 2 dots in %s, but there aren't really", json0.ID_Token)
		return
	}
	jwtClaim, err := base64.RawStdEncoding.DecodeString(strings.Split(json0.ID_Token, ".")[1])
	fmt.Fprintf(w, "<p>Got jwt Claim %s \n", jwtClaim)
	json1 := struct {
		Sub string
	}{}
	err = json.NewDecoder(bytes.NewReader(jwtClaim)).Decode(&json1)
	if err != nil {
		log.Errorf(ctx, "Couldn't JSON-decode JWT part of response from google, got err %v", err)
		fmt.Fprintf(w, "<p>Couldn't JSON-decode JWT part of response from google, got err %v", err)
		return
	}
	fmt.Fprintf(w, "<p>Got sub ID %s \n", json1.Sub)
	userID = "google:" + json1.Sub
	updatedSession := Session{time.Now(), userID}
	key := datastore.NewKey(ctx, "Session", sessionID, 0, nil)
	_, err = datastore.Put(ctx, key, &updatedSession)
	if err != nil {
		log.Errorf(ctx, "Couldn't save userID with session, got %v", err)
		fmt.Fprintf(w, "<p>Couldn't save userID with session, got %v", err)
		return
	}
	fmt.Fprintf(w, "<p>Saved to datastore, yay? \n")
	fmt.Fprintf(w, "<p>Maybe when I'm feeling more confident about this code, I'll redirect you back to your page right away instead of <b>waiting a second</b>. But I've been hearing about how scary OAuth is for, like, the last few years so I guess I'll give myself a chance to look at all this debug spewage&hellip; \n")
	fmt.Fprintf(w, `<script>setTimeout(function(){ document.location='`+redirURL+`'}, 1)</script>`)
	memcache.Set(ctx, &memcache.Item{
		Key:   "Session/" + sessionID,
		Value: []byte(userID),
	})
}

func logout(w http.ResponseWriter, r *http.Request, userID string, sessionID string) {
	ctx := appengine.NewContext(r)
	userID = ""
	updatedSession := Session{time.Now(), userID}
	key := datastore.NewKey(ctx, "Session", sessionID, 0, nil)
	_, err := datastore.Put(ctx, key, &updatedSession)
	if err != nil {
		log.Errorf(ctx, "Couldn't save userID with session, got %v", err)
		fmt.Fprintf(w, "<p>Couldn't save userID with session, got %v", err)
		return
	}
	fmt.Fprintf(w, `<!doctype html>
<html>
<head>
<meta charset="utf-8" /> 
</head>
<body>
<p>Saved to datastore, yay? \n`)
	memcache.Set(ctx, &memcache.Item{
		Key:   "Session/" + sessionID,
		Value: []byte(userID),
	})
	fmt.Fprintf(w, `
<p>Should redirect to <a href="/">main page</a>. If not, click that link.
<script>setTimeout(function(){ document.location='/'}, 1)</script>
</body></html>`)
}

func googOAuth2Config(ctx context.Context) (ID, Secret string) {
	clientConfig := getConfig("goog_oauth2_client", ctx)
	fields := strings.Split(clientConfig, "|")
	if len(fields) > 1 {
		return fields[0], fields[1]
	} else {
		return "", ""
	}
}
