package algo

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoEncode(t *testing.T) {
	res := GeoHash(GeoCoord{Lat: 39.92324, Lon: 116.3906}, 6)
	assert.Equal(t, "wx4g0e", res)
}

func TestGeoDist(t *testing.T) {
	res := GeoHaversineDist(
		GeoCoord{Lat: 41.507483, Lon: -99.436554},
		GeoCoord{Lat: 38.504048, Lon: -98.315949},
	)
	assert.Equal(t, strconv.FormatFloat(res/1000, 'f', 2, 64), "347.33")
}

func TestGeoGetNeighbors(t *testing.T) {
	n := GeoGetNeighbors("gbse")

	assert.ElementsMatch(
		t,
		[]string{"gbsk", "gbss", "gbsu", "gbs7", "gbsg", "gbs6", "gbsd", "gbsf"},
		n,
	)
}
