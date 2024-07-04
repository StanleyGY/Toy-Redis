package algo

import (
	"math"
	"strings"
)

const (
	GeoMaxPrecision = 12
)

var (
	geoHashEncodings       = "0123456789bcdefghjkmnpqrstuvwxyz"
	geoEarthRadiusInMeters = 6371000.0
)

type GeoCoord struct {
	Lat float64
	Lon float64
}

func (c GeoCoord) AddDist(yOffset, xOffset float64) GeoCoord {
	latDelta := yOffset / geoEarthRadiusInMeters * (180 / math.Pi)
	lonDelta := xOffset / geoEarthRadiusInMeters * (360 / math.Pi)

	return GeoCoord{
		Lat: c.Lat + latDelta,
		Lon: c.Lon + lonDelta,
	}
}

type GeoBoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

func (b GeoBoundingBox) Contains(c GeoCoord) bool {
	return b.MinLat <= c.Lat && c.Lat <= b.MaxLat && b.MinLon <= c.Lon && c.Lon <= b.MaxLon
}

func (b GeoBoundingBox) Center() GeoCoord {
	lat, lon := (b.MinLat+b.MaxLat)/2, (b.MinLon+b.MaxLon)/2
	return GeoCoord{Lat: lat, Lon: lon}
}

func (b GeoBoundingBox) LatDelta() float64 {
	return b.MaxLat - b.MinLat
}

func (b GeoBoundingBox) LonDelta() float64 {
	return b.MaxLon - b.MinLon
}

func MakeGeoBoundingBox(geoHash string) *GeoBoundingBox {
	v := geoBase32Decode(geoHash)
	precision := len(geoHash)

	center := geoDecodeCoord(v, precision)
	latRange := math.Ldexp(180.0, -precision*5/2)
	lonRange := math.Ldexp(360.0, -(precision*5 - precision*5/2))

	return &GeoBoundingBox{
		MinLat: center.Lat - latRange/2,
		MaxLat: center.Lat + latRange/2,
		MinLon: center.Lon - lonRange/2,
		MaxLon: center.Lon + lonRange/2,
	}
}

func geoBase32Decode(geoHash string) int64 {
	var v int64 = 0
	for i := 0; i < len(geoHash); i++ {
		ind := strings.Index(geoHashEncodings, string(geoHash[i]))
		v <<= 5
		v |= int64(ind)
	}
	return v
}

func geoBase32Encode(v int64, precision int) string {
	// Group every 5-bit and encode into an ASCII character.
	// The highest 2 bit is not used.
	b := make([]byte, precision)
	for i := precision - 1; i >= 0; i-- {
		b[i] = geoHashEncodings[v&0b11111]
		v >>= 5
	}
	return string(b)
}

func geoDecodeCoord(v int64, precision int) GeoCoord {
	var (
		minLat float64 = -90
		maxLat float64 = 90
		minLon float64 = -180
		maxLon float64 = 180
		midLat float64
		midLon float64
	)

	for i := precision*5 - 1; i >= 0; i-- {
		if (precision*5-1-i)%2 == 1 {
			// Compute latitude
			midLat = (minLat + maxLat) / 2.0
			if v&(1<<i) == 0 {
				// Pushed a zero bit
				maxLat = midLat
			} else {
				// Pushed a one bit
				minLat = midLat
			}
		} else {
			// Compute longitude
			midLon = (minLon + maxLon) / 2.0
			if v&(1<<i) == 0 {
				// Pushed a zero bit
				maxLon = midLon
			} else {
				// Pushed a one bit
				minLon = midLon
			}
		}
	}
	midLat = (minLat + maxLat) / 2.0
	midLon = (minLon + maxLon) / 2.0
	return GeoCoord{Lat: midLat, Lon: midLon}
}

func geoEncodeCoord(coord GeoCoord, precision int) int64 {
	// Interleaving bits of latitude and longitude
	var (
		minLat float64 = -90
		maxLat float64 = 90
		minLon float64 = -180
		maxLon float64 = 180
	)
	var res int64

	for i := 0; i < precision*5; i++ {
		if i%2 == 1 {
			// Compute latitude
			midLat := (minLat + maxLat) / 2.0
			if coord.Lat < midLat {
				// Push a zero bit
				res <<= 1
				maxLat = midLat
			} else {
				// Push a one bit
				res = res<<1 | 1
				minLat = midLat
			}
		} else {
			// Compute longitude
			midLon := (minLon + maxLon) / 2.0
			if coord.Lon < midLon {
				// Push a zero bit
				res <<= 1
				maxLon = midLon
			} else {
				// Push a one bit
				res = res<<1 | 1
				minLon = midLon
			}
		}
	}
	return res
}

func GeoHash(coord GeoCoord, precision int) string {
	// Tutorial: https://yatmanwong.medium.com/geohash-implementation-explained-2ed9627a61ff
	if precision > GeoMaxPrecision {
		precision = GeoMaxPrecision
	}
	res := geoEncodeCoord(coord, precision)
	return geoBase32Encode(res, precision)
}

func GeoHaversineDist(coordA GeoCoord, coordB GeoCoord) float64 {
	// Tutorial: https://www.movable-type.co.uk/scripts/latlong.html
	latARad := coordA.Lat * math.Pi / 180
	latBRad := coordB.Lat * math.Pi / 180
	deltaLat := (coordB.Lat - coordA.Lat) * math.Pi / 180
	deltaLon := (coordB.Lon - coordA.Lon) * math.Pi / 180

	a := math.Pow(math.Sin(deltaLat/2), 2) + math.Cos(latARad)*
		math.Cos(latBRad)*math.Pow(deltaLon/2, 2)

	return geoEarthRadiusInMeters * 2.0 * math.Asin(math.Sqrt(a))
}

func GeoGetNeighbors(geoHash string) []string {
	// `geoHash` encodes a box of possible coordinates (aka. bounding box)
	// The length of `geoHash` determines the precision
	box := MakeGeoBoundingBox(geoHash)

	center := box.Center()
	latDelta := box.LatDelta()
	lonDelta := box.LonDelta()
	precision := len(geoHash)

	return []string{
		// North west
		GeoHash(GeoCoord{Lat: center.Lat + latDelta, Lon: center.Lon - lonDelta}, precision),
		// North
		GeoHash(GeoCoord{Lat: center.Lat + latDelta, Lon: center.Lon}, precision),
		// North east
		GeoHash(GeoCoord{Lat: center.Lat + latDelta, Lon: center.Lon + lonDelta}, precision),

		// West
		GeoHash(GeoCoord{Lat: center.Lat, Lon: center.Lon - lonDelta}, precision),
		// East
		GeoHash(GeoCoord{Lat: center.Lat, Lon: center.Lon + lonDelta}, precision),

		// South west
		GeoHash(GeoCoord{Lat: center.Lat - latDelta, Lon: center.Lon - lonDelta}, precision),
		// South
		GeoHash(GeoCoord{Lat: center.Lat - latDelta, Lon: center.Lon}, precision),
		// South east
		GeoHash(GeoCoord{Lat: center.Lat - latDelta, Lon: center.Lon + lonDelta}, precision),
	}
}
