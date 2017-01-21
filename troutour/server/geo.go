package server

import (
	"math"
)

const (
	fakeRegionBox = (100000 * 9100) + 0
)

func kmPerLat() float64 {
	return 111.1
}

func kmPerLng(lat float64) float64 {
	latRad := lat * math.Pi / 180.0
	cos := math.Cos(latRad)
	if cos <= 0.01 {
		cos = 0.01
	}
	return 111.1 * cos
}

func dist(lat0, lng0, lat1, lng1 float64) float64 {
	meanLat := (lat0 + lat1) / 2.0
	dX := kmPerLng(meanLat) * (lng0 - lng1)
	dY := kmPerLat() * (lat0 - lat1)
	return math.Sqrt((dX * dX) + (dY * dY))
}

// we index regions by "regionbox"--rectangles .01 degree on a side.
func latLng2RegionBox(lat, lng float64) int32 {
	latX100 := int32(math.Floor(lat * 100.0))
	lngX100 := int32(math.Floor(lng * 100.0))
	return (100000 * latX100) + lngX100
}

func nearbyRegionRanges(centerLat, centerLng, lKm float64) [][]int32 {
	retval := [][]int32{}
	if math.Abs(centerLat) > 80.0 || math.Abs(centerLng) > 179.5 {
		return retval
	}
	if math.Abs(centerLat) < 0.5 && math.Abs(centerLng) < 0.5 {
		return retval
	}
	latPerKm := 1.0 / kmPerLat()
	lngPerKm := 1.0 / kmPerLng(centerLat)
	south := centerLat - (latPerKm * lKm)
	north := centerLat + (latPerKm * lKm)
	west := centerLng - (lngPerKm * lKm)
	east := centerLng + (lngPerKm * lKm)
	prevRowStart := int32(math.MaxInt32)
	for lat := south; lat <= north; lat += 0.009 {
		rowStart := latLng2RegionBox(lat, west)
		if rowStart == prevRowStart {
			continue
		}
		prevRowStart = rowStart
		rowEnd := latLng2RegionBox(lat, east)
		retval = append(retval, []int32{rowStart, rowEnd})
	}
	return retval
}

// We index clumps by "clumpbox"--rectangles .05 degree on a side.
// Similar to regionboxes, but less fine on the granularity
func latLng2ClumpBox(lat, lng float64) int32 {
	latX20 := int32(math.Floor(lat * 20.0))
	lngX20 := int32(math.Floor(lng * 20.0))
	return (100000 * latX20) + lngX20
}

func nearbyClumpRanges(centerLat, centerLng, lKm float64) [][]int32 {
	retval := [][]int32{}
	if math.Abs(centerLat) > 80.0 || math.Abs(centerLng) > 179.5 {
		return retval
	}
	if math.Abs(centerLat) < 0.5 && math.Abs(centerLng) < 0.5 {
		return retval
	}
	latPerKm := 1.0 / kmPerLat()
	lngPerKm := 1.0 / kmPerLng(centerLat)
	south := centerLat - (latPerKm * lKm)
	north := centerLat + (latPerKm * lKm)
	west := centerLng - (lngPerKm * lKm)
	east := centerLng + (lngPerKm * lKm)
	prevRowStart := int32(math.MaxInt32)
	for lat := south; lat <= north; lat += 0.009 {
		rowStart := latLng2ClumpBox(lat, west)
		if rowStart == prevRowStart {
			continue
		}
		prevRowStart = rowStart
		rowEnd := latLng2ClumpBox(lat, east)
		retval = append(retval, []int32{rowStart, rowEnd})
	}
	return retval
}

func segPOrientation(p1Lat, p1Lng, p2Lat, p2Lng, xLat, xLng float64) float64 {
	return ((p1Lat - p2Lat) * (xLng - p1Lng)) - ((p1Lng - p2Lng) * (xLat - p1Lat))
}

func segsIntersectP(a1Lat, a1Lng, a2Lat, a2Lng, b1Lat, b1Lng, b2Lat, b2Lng float64) bool {

	// If both endpoints of one segment are on the same side of another segment,
	// then those segments don't intersect. This orientation function determines
	// which side of segment p1-p2 point 'x' falls on. If two points have the
	// same pos/neg sign "orientation" to some seg, those two points are on the
	// same side.

	// if same side (orientation same sign), product will be >0
	if segPOrientation(a1Lat, a1Lng, a2Lat, a2Lng, b1Lat, b1Lng)*
		segPOrientation(a1Lat, a1Lng, a2Lat, a2Lng, b2Lat, b2Lng) > 0.0 {
		return false
	}
	// if same side (segPOrientation same sign), product will be >0
	if segPOrientation(b1Lat, b1Lng, b2Lat, b2Lng, a1Lat, a1Lng)*
		segPOrientation(b1Lat, b1Lng, b2Lat, b2Lng, a2Lat, a2Lng) > 0.0 {
		return false
	}
	return true
}
