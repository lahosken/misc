package main

import (
	"testing"
)

func TestSegsIntersectPOnes(t *testing.T) {
	oLat := 0.0
	oLng := 0.0
	nLat := 1.0
	nLng := 0.0
	eLat := 0.0
	eLng := 1.0
	dLat := 1.0
	dLng := 1.0
	if segsIntersectP(oLat, oLng, nLat, nLng, eLat, eLng, dLat, dLng) {
		t.Error("segsIntersectP( nward ) false positive")
	}
	if segsIntersectP(nLat, nLng, oLat, oLng, dLat, dLng, eLat, eLng) {
		t.Error("segsIntersectP( sward ) false positive")
	}
	if segsIntersectP(oLat, oLng, nLat, nLng, dLat, dLng, eLat, eLng) {
		t.Error("segsIntersectP( clockwise ) false positive")
	}
	if segsIntersectP(oLat, oLng, eLat, eLng, nLat, nLng, dLat, dLng) {
		t.Error("segsIntersectP( eward ) false positive")
	}
	if !segsIntersectP(oLat, oLng, dLat, dLng, nLat, nLng, eLat, eLng) {
		t.Error("segsIntersectP( x1 ) false negative")
	}
}

func TestGeoSegsIntersectRevert(t *testing.T) {
	a1Lat := 37.7619
	a1Lng := -122.4737
	a2Lat := 37.7673
	a2Lng := -122.4572

	b1Lat := 37.7621
	b1Lng := -122.4670
	b2Lat := 37.7644
	b2Lng := -122.4666

	if !segsIntersectP(a1Lat, a1Lng, a2Lat, a2Lng, b1Lat, b1Lng, b2Lat, b2Lng) {
		t.Error("segsIntersectP() failed to see intersection")
	}
}
