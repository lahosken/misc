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

var bigCities = [...][]float64{
	[]float64{-43.528, 172.628}, // Christchurch New Zealand
	[]float64{-41.290, 174.774}, // Wellington
	[]float64{-37.810, 144.967}, // Melbourne
	[]float64{-36.852, 174.760}, // Auckland
	[]float64{-35.310, 149.124}, // Canberra
	[]float64{-34.928, 138.612}, // Adelaide
	[]float64{-34.628, -58.448}, // Buenos Aires
	[]float64{-33.922, 18.418},  // Cape Town
	[]float64{-33.884, 151.204}, // Sydney
	[]float64{-31.954, 115.862}, // Perth
	[]float64{-27.476, 153.028}, // Brisbane
	[]float64{-26.198, 28.038},  // Johannesburg
	[]float64{-23.544, -46.634}, // São Paulo
	[]float64{-19.264, 146.824}, // Townsville
	[]float64{-16.924, 145.750}, // Cairns
	[]float64{-12.426, 130.854}, // Darwin Australia
	[]float64{-12.054, -77.056}, // Lima
	[]float64{-6.918, 107.610},  // Bandung
	[]float64{-6.214, 106.848},  // Jakarta
	[]float64{-1.286, 36.822},   // Nairobi
	[]float64{0.320, 32.580},    // Kampala
	[]float64{1.398, 103.870},   // Singapore
	[]float64{3.134, 101.680},   // Kuala Lumpur
	[]float64{6.530, 3.356},     // Lagos
	[]float64{6.918, 79.868},    // Colombo Sri Lanka
	[]float64{12.964, 77.584},   // Bangalore
	[]float64{13.764, 100.536},  // Bangkok
	[]float64{14.610, 121.012},  // Manila
	[]float64{16.874, 96.202},   // Yangon
	[]float64{17.376, 78.488},   // Hyderabad
	[]float64{18.420, -66.078},  // San Juan Puerto Rico
	[]float64{19.136, 72.918},   // Mumbai
	[]float64{19.433, -99.138},  // Mexico City
	[]float64{19.710, -155.084}, // Hilo
	[]float64{20.236, 85.828},   // Bhubaneswar
	[]float64{22.332, 114.186},  // Hong Kong
	[]float64{22.564, 88.358},   // Kolkata
	[]float64{22.572, 114.062},  // Shenzhen
	[]float64{23.113, -82.368},  // Havana
	[]float64{23.133, 113.268},  // Guangzhou
	[]float64{23.800, 90.416},   // Dhaka
	[]float64{24.940, 67.122},   // Karachi
	[]float64{25.028, 121.530},  // Taipei
	[]float64{25.212, 55.270},   // Dubai
	[]float64{25.700, 32.640},   // Luxor
	[]float64{25.804, -80.234},  // Miami
	[]float64{27.952, -81.458},  // Tampa
	[]float64{28.622, 77.234},   // Delhi
	[]float64{29.557, -95.100},  // Houston
	[]float64{29.976, -90.086},  // New Orleans
	[]float64{30.044, 31.236},   // Cairo
	[]float64{30.598, 114.304},  // Wuhan
	[]float64{30.274, -97.740},  // Austin
	[]float64{30.694, -88.046},  // Mobile
	[]float64{31.226, -110.962}, // Tucson
	[]float64{31.226, 121.466},  // Shanghai
	[]float64{31.762, -106.486}, // El Paso
	[]float64{31.782, 35.220},   // Jerusalem
	[]float64{32.686, -117.102}, // San Diego
	[]float64{32.774, -79.934},  // Charleston
	[]float64{32.788, -96.794},  // Dallas
	[]float64{33.450, -112.094}, // Phoenix
	[]float64{33.760, -84.385},  // Atlanta
	[]float64{33.824, -116.540}, // Palm Springs
	[]float64{34.054, -118.246}, // LAX
	[]float64{34.022, -84.362},  // Roswell
	[]float64{34.062, -117.324}, // San Bernadino
	[]float64{34.664, -86.670},  // Huntsville
	[]float64{34.736, -92.282},  // Little Rock
	[]float64{34.746, 135.574},  // Osaka
	[]float64{35.082, -106.620}, // Albuquerque
	[]float64{35.226, -80.842},  // Charlotte
	[]float64{35.278, -120.658}, // San Luis Obispo
	[]float64{35.616, -82.566},  // Asheville
	[]float64{35.685, -105.938}, // Santa Fe
	[]float64{35.698, 139.732},  // Tokyo
	[]float64{35.778, -78.640},  // Raleigh
	[]float64{35.888, 14.502},   // Malta
	[]float64{36.158, -86.780},  // Nashville
	[]float64{36.170, -115.144}, // Las Vegas
	[]float64{36.736, -119.786}, // Fresno
	[]float64{36.846, -76.290},  // Norfolk
	// []float64{37.210, -93.290},  // Springfield Missouri
	// []float64{37.430, -122.169}, // Stanford
	// []float64{37.542, -77.434},  // Richmond Virginia
	// []float64{37.570, 126.988},  // Seoul
	// []float64{37.872, -122.260}, // Cal
	[]float64{38.258, -85.760},  // Louisville
	[]float64{38.576, -121.480}, // Sacramento
	[]float64{38.632, -90.200},  // St Louis
	[]float64{38.906, -77.036},  // Washington DC
	[]float64{39.098, -94.578},  // Kansas City
	[]float64{39.100, -84.512},  // Cincinnati
	[]float64{39.746, -104.994}, // Denver
	[]float64{39.772, -86.160},  // Indianapolis
	[]float64{39.828, -77.232},  // Gettysburg
	[]float64{39.909, 116.396},  // Beijing
	[]float64{39.950, -75.150},  // Philadelphia
	[]float64{40.006, -105.264}, // Boulder
	[]float64{40.442, -80.014},  // Pittsburgh
	[]float64{40.732, -74.174},  // Newark New Jersey
	[]float64{40.756, -111.890}, // Salt Lake City
	[]float64{40.778, -73.966},  // NYC
	[]float64{41.046, 29.036},   // Istanbul
	[]float64{41.874, -87.761},  // Chicago
	[]float64{41.902, 12.486},   // Rome
	[]float64{42.354, -71.091},  // Boston
	[]float64{42.390, -83.050},  // Detroit
	[]float64{42.650, -73.750},  // Albany New York
	[]float64{42.954, -78.896},  // Buffalo
	[]float64{43.084, -89.371},  // Madison Wisconsin
	[]float64{43.154, -77.596},  // Rochester New York
	[]float64{43.748, -79.596},  // Toronto
	[]float64{43.766, 11.250},   // Florence
	[]float64{44.048, -123.076}, // Eugene Oregon
	[]float64{44.652, -63.582},  // Halifax
	[]float64{44.972, -93.230},  // Minneapolis
	[]float64{45.413, -75.698},  // Ottawa
	[]float64{45.584, -73.556},  // Montreal
	[]float64{45.602, -122.684}, // Portland Oregon
	[]float64{46.054, 14.504},   // Ljubljana, Slovenia
	[]float64{46.862, -113.992}, // Missoula
	[]float64{47.372, 8.538},    // Zurich
	[]float64{47.672, -122.258}, // Seattle
	[]float64{48.136, 11.574},   // Munich
	[]float64{48.210, 16.376},   // Vienna
	[]float64{48.854, 2.344},    // Paris
	[]float64{49.272, -123.094}, // Vancouver BC
	[]float64{49.901, -97.150},  // Winnipeg
	[]float64{50.058, 19.942},   // Krakow
	[]float64{50.074, 14.422},   // Prague
	[]float64{50.542, 30.500},   // Kyiv
	[]float64{50.846, 4.352},    // Brussels
	[]float64{50.910, 6.970},    // Cologne
	[]float64{51.044, -114.068}, // Calgary
	[]float64{51.508, -0.095},   // London
	[]float64{52.116, -106.650}, // Saskatoon
	[]float64{52.222, 21.012},   // Warsaw
	[]float64{52.415, 4.859},    // Amsterdam
	[]float64{53.120, 18.004},   // Bydgoszcz
	[]float64{53.342, -6.266},   // Dublin
	[]float64{53.400, -2.980},   // Liverpool
	[]float64{53.488, -2.242},   // Manchester
	[]float64{53.516, -113.502}, // Edmonton
	[]float64{53.544, 9.994},    // Hamburg
	[]float64{54.683, 25.281},   // Vilnius
	[]float64{55.680, 12.572},   // Copenhagen
	[]float64{55.760, 37.626},   // Moscow
	[]float64{55.860, -4.250},   // Glasgow
	[]float64{55.962, -3.186},   // Edinburgh
}

func randLatLngNearCity() (lat, lng float64) {
	whichCity := rand.Intn(len(bigCities))
	lat = bigCities[whichCity][0] + rand.Float64()*(rand.Float64()-0.5)
	lng = bigCities[whichCity][1] + rand.Float64()*(rand.Float64()-0.5)
	return
}
