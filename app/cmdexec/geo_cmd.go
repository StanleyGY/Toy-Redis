package cmdexec

import (
	"strconv"
	"strings"

	"github.com/stanleygy/toy-redis/app/algo"
	"github.com/stanleygy/toy-redis/app/resp"
)

type geoCmdExecutor struct{}

/*
Syntax: GEOADD key longitude latitude member [longitude latitude member ...]
*/
func (e geoCmdExecutor) parseGeoAddCmdArgs(cmdArgs []*resp.RespValue, key *string, longitudes *[]float64, latitudes *[]float64, members *[]string) error {
	if len(cmdArgs) < 4 {
		return ErrInvalidArgs
	}
	var i int

	*key = cmdArgs[0].BulkStr
	for i = 1; i < len(cmdArgs); i += 3 {
		if i+2 >= len(cmdArgs) {
			return ErrInvalidArgs
		}
		lon, err := strconv.ParseFloat(cmdArgs[i].BulkStr, 64)
		if err != nil {
			return err
		}
		lat, err := strconv.ParseFloat(cmdArgs[i+1].BulkStr, 64)
		if err != nil {
			return err
		}
		*longitudes = append(*longitudes, lon)
		*latitudes = append(*latitudes, lat)
		*members = append(*members, cmdArgs[i+2].BulkStr)
	}

	if i < len(cmdArgs) {
		return ErrInvalidArgs
	}
	return nil
}

func (e geoCmdExecutor) executeGeoAddCmd(c *ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key        string
		longitudes []float64
		latitudes  []float64
		members    []string
	)
	err := e.parseGeoAddCmdArgs(cmdArgs, &key, &longitudes, &latitudes, &members)
	if err != nil {
		AddErrorReplyEvent(c, err)
		return
	}

	// Look up store at key
	store, found := db.GeoStore[key]
	if !found {
		db.GeoStore[key] = make(map[string]*GeoStoreValue)
		store = db.GeoStore[key]
	}

	// Execute cmd
	for i := 0; i < len(members); i++ {
		c := algo.GeoCoord{
			Lat: latitudes[i],
			Lon: longitudes[i],
		}
		store[members[i]] = &GeoStoreValue{
			Coord: c,
			Hash:  algo.GeoHash(c, algo.GeoMaxPrecision),
		}
	}
	AddIntegerReplyEvent(c, len(members))
}

/*
Syntax: GEODIST key member1 member2 [M | KM]
Reply:
  - Null reply: one or both of the elements are missing
  - Bulk string reply: distance as double
*/
func (e geoCmdExecutor) parseGeoDistCmd(cmdArgs []*resp.RespValue, key *string, m1 *string, m2 *string, unit *string) error {
	if len(cmdArgs) < 3 || len(cmdArgs) > 4 {
		return ErrInvalidArgs
	}
	*key = cmdArgs[0].BulkStr
	*m1 = cmdArgs[1].BulkStr
	*m2 = cmdArgs[2].BulkStr

	if len(cmdArgs) == 4 {
		*unit = strings.ToUpper(cmdArgs[3].BulkStr)
		if *unit != "M" && *unit != "KM" {
			return ErrInvalidArgs
		}
	}
	return nil
}

func (e geoCmdExecutor) executeGeoDistCmd(c *ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key  string
		m1   string
		m2   string
		unit string = "KM"
	)
	err := e.parseGeoDistCmd(cmdArgs, &key, &m1, &m2, &unit)
	if err != nil {
		AddErrorReplyEvent(c, err)
		return
	}

	// Look up store at key
	store, found := db.GeoStore[key]
	if !found {
		AddNullBulkStringReplyEvent(c)
		return
	}

	// Look up members
	v1 := store[m1]
	v2 := store[m2]
	dist := algo.GeoHaversineDist(v1.Coord, v2.Coord)

	if unit == "KM" {
		dist /= 1000.0
	}
	AddBulkStringReplyEvent(c, strconv.FormatFloat(dist, 'f', -1, 64))
}

/*
Syntax: GEOHASH key member [member ...]
Reply:
  - Array reply
*/
func (e geoCmdExecutor) parseGeoHashCmd(cmdArgs []*resp.RespValue, key *string, members *[]string) error {
	if len(cmdArgs) < 2 {
		return ErrInvalidArgs
	}
	*key = cmdArgs[0].BulkStr
	for i := 1; i < len(cmdArgs); i++ {
		*members = append(*members, cmdArgs[i].BulkStr)
	}
	return nil
}

func (e geoCmdExecutor) executeGeoHashCmd(c *ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key     string
		members []string
	)

	err := e.parseGeoHashCmd(cmdArgs, &key, &members)
	if err != nil {
		AddErrorReplyEvent(c, err)
		return
	}

	// Look up store at key
	store, found := db.GeoStore[key]
	if !found {
		AddEmptyArrayReplyEvent(c)
		return
	}

	// Collect all hashes
	var res []*resp.RespValue
	for _, m := range members {
		v, found := store[m]
		if found {
			res = append(res, resp.MakeBulkString(v.Hash))
		} else {
			res = append(res, resp.MakeNilBulkString())
		}
	}
	AddArrayReplyEvent(c, res)
}

/*
Syntax: GEORADIUS key longitude latitude radius
*/
func (e geoCmdExecutor) parseGeoRadiusCmdArgs(cmdArgs []*resp.RespValue, key *string, longitude *float64, latitude *float64, radius *float64) error {
	if len(cmdArgs) < 4 {
		return ErrInvalidArgs
	}
	var err error
	*key = cmdArgs[0].BulkStr
	*longitude, err = strconv.ParseFloat(cmdArgs[1].BulkStr, 64)
	if err != nil {
		return err
	}
	*latitude, err = strconv.ParseFloat(cmdArgs[2].BulkStr, 64)
	if err != nil {
		return err
	}
	*radius, err = strconv.ParseFloat(cmdArgs[3].BulkStr, 64)
	if err != nil {
		return err
	}
	return nil
}

func (e geoCmdExecutor) executeGeoRadiusCmd(c *ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key       string
		longitude float64
		latitude  float64
		radius    float64
	)
	err := e.parseGeoRadiusCmdArgs(cmdArgs, &key, &longitude, &latitude, &radius)
	if err != nil {
		AddErrorReplyEvent(c, err)
		return
	}

	// Look up store at key
	store, found := db.GeoStore[key]
	if !found {
		AddEmptyArrayReplyEvent(c)
		return
	}

	// This searching algorithm takes O(N^2) and unoptimized. Redis uses skip list to speed up
	// the geospatial query.
	originCoord := algo.GeoCoord{
		Lat: latitude,
		Lon: longitude,
	}
	var res []*resp.RespValue
	for m, v := range store {
		if algo.GeoHaversineDist(v.Coord, originCoord) < radius {
			res = append(res, resp.MakeBulkString(m))
		}
	}
	AddArrayReplyEvent(c, res)
}

func (e geoCmdExecutor) Execute(c *ClientInfo, cmdName string, cmdArgs []*resp.RespValue) {
	switch cmdName {
	case "GEOADD":
		e.executeGeoAddCmd(c, cmdArgs)
	case "GEODIST":
		e.executeGeoDistCmd(c, cmdArgs)
	case "GEOHASH":
		e.executeGeoHashCmd(c, cmdArgs)
	case "GEORADIUS":
		e.executeGeoRadiusCmd(c, cmdArgs)
	}
}
