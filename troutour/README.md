Troubadour Tour Board
=====================

A game you play by wandering around a bustling area with your
phone "checking in" to places as you go.

It's a Google App Engine app written in Go. You want the App Engine Go SDK.

Not included w/this source code: need to set app keys for
using foursquare API and for OpenID Google login. For local
testing, you don't need the Google login, but you still need
a 4sq API key. I have a handy .sh script that looks like

    curl 'localhost:8080/configstore?k=4sq_client&v=SECRET|REDACTED'

...but instead of SECRET and REDACTED it has the app ID and key.