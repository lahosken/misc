package server

import (
	"math"
	"math/rand"
)

const (
	fakeRegionBox = (100000 * 9100) + 0
)

func kmPerLat() float64 {
	return 111.1
}

func kmPerLng(lat float64) float64 {
	latRadians := lat * math.Pi / 180.0
	cos := math.Cos(latRadians)
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

func randLatLng() (lat, lng float64) {
	lng = 9999.0
	for lng <= -180.0 || lng >= 180.0 {
		lng = (rand.Float64() * 360.0) - 180.0
	}
	cosLat := 9999.0
	for cosLat <= -1.0 || cosLat >= 1.0 {
		cosLat = (rand.Float64() * 2.0) - 1.0
	}
	latRadians := math.Acos(cosLat)
	lat = (latRadians * 180.0 / math.Pi) - 90.0
	return
}

var bigCities = [...][]float64 {
  []float64{ -37.810, 144.967 }, // Melbourne
  []float64{ -36.852, 174.760 }, // Auckland
  []float64{ -33.884, 151.204 }, // Sydney
  []float64{ -23.544, -46.634 }, // São Paulo
  []float64{  -6.214, 106.848 }, // Jakarta
  []float64{   1.398, 103.870 }, // Singapore
  []float64{   6.530,   3.356 }, // Lagos
  []float64{   7.106, 171.375 }, // Marshall Islands
  []float64{  12.964,  77.584 }, // Bangalore
  []float64{  13.764, 100.536 }, // Bangkok
  []float64{  19.136,  72.918 }, // Mumbai
  []float64{  19.433, -99.138 }, // Mexico City
  []float64{  21.310,-157.860 }, // Honolulu
  []float64{  22.332, 114.186 }, // Hong Kong
  []float64{  22.572, 114.062 }, // Shenzhen
  []float64{  23.133, 113.268 }, // Guangzhou
  []float64{  24.940,  67.122 }, // Karachi
  []float64{  25.804, -80.234 }, // Miami
  []float64{  28.622,  77.234 }, // Delhi
  []float64{  29.557, -95.100 }, // Houston
  []float64{  30.044,  31.236 }, // Cairo
  []float64{  30.598, 114.304 }, // Wuhan
  []float64{  30.274, -97.740 }, // Austin
  []float64{  31.226, 121.466 }, // Shanghai
  []float64{  32.788, -96.794 }, // Dallas
  []float64{  33.450,-112.094 }, // Phoenix
  []float64{  33.760, -84.385 }, // Atlanta
  []float64{  34.054,-118.246 }, // LAX
  []float64{  34.062,-117.324 }, // San Bernadino
  []float64{  34.746, 135.574 }, // Osaka
  []float64{  35.616, -82.566 }, // Asheville
  []float64{  35.698, 139.732 }, // Tokyo
  []float64{  36.170,-115.144 }, // Las Vegas
  []float64{  37.430,-122.169 }, // Stanford
  []float64{  37.570, 126.988 }, // Seoul
  []float64{  37.872,-122.260 }, // Cal
  []float64{  38.632, -90.200 }, // St Louis
  []float64{  38.906, -77.036 }, // Washington DC
  []float64{  39.909, 116.396 }, // Beijing
  []float64{  39.950, -75.150 }, // Philadelphia
  []float64{  40.006,-105.264 }, // Boulder
  []float64{  40.778, -73.966 }, // NYC
  []float64{  40.442, -80.014 }, // Pittsburgh
  []float64{  41.046,  29.036 }, // Istanbul
  []float64{  41.874, -87.761 }, // ORD
  []float64{  42.354, -71.091 }, // Boston
  []float64{  42.390, -83.050 }, // Detroit
  []float64{  43.084, -89.371 }, // Madison Wisconsin
  []float64{  43.748, -79.596 }, // Toronto
  []float64{  44.972, -93.230 }, // Minneapolis
  []float64{  45.413, -75.698 }, // Ottawa
  []float64{  45.584, -73.556 }, // Montreal
  []float64{  45.602,-122.684 }, // Portland Oregon
  []float64{  47.672,-122.258 }, // Seattle
  []float64{  49.272,-123.094 }, // Vancouver BC
  []float64{  50.074,  14.422 }, // Prague
  []float64{  51.508,  -0.095 }, // London
  []float64{  52.415,   4.859 }, // Amsterdam
  []float64{  54.364,  18.568 }, // Gdańsk
  []float64{  55.760,  37.626 }, // Moscow
};

func randLatLngNearCity() (lat, lng float64) {
  whichCity := rand.Intn(len(bigCities))
  lat = bigCities[whichCity][0] + 0.5 * (rand.Float64() - 0.5)
  lng = bigCities[whichCity][1] + 0.5 * (rand.Float64() - 0.5)
  return
}
