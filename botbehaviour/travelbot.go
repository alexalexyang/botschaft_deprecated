package travelbot

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type bot struct {
	id  int
	lat float64
	lon float64
	// radius float64
	pois []poi
}

type poi struct {
	ID   int               `json:"id"`
	Lat  float64           `json:"lat"`
	Lon  float64           `json:"lon"`
	Tags map[string]string `json:"tags"`
	// VisitType: potential or visited.
	VisitType string
}

type jsonStruct struct {
	Pois []poi `json:"elements"`
}

// Collects all bots with travel in their drive in a []bot.
func getTravelbots() []bot {

	// Select all travelbots.
	db, err := sql.Open("sqlite3", "database.db")
	check(err)

	// Create table if not exists.
	query := `SELECT BotID, Lat, Lon FROM bots WHERE bottype="travelbot";`
	rows, err := db.Query(query)
	check(err)
	defer rows.Close()

	var bots []bot

	for rows.Next() {
		b := bot{}

		err = rows.Scan(&b.id, &b.lat, &b.lon)
		check(err)
		bots = append(bots, b)
	}
	err = rows.Err()
	check(err)
	db.Close()
	return bots
}

// Concatenates separate queries from each bot in []bot into a single long query,
func createOSMQuery(bots []bot) string {
	poiType := "amenity"
	poiSubType := "restaurant"
	radius := "1000"

	var pointsBuffer bytes.Buffer
	for _, bot := range bots {
		lat := strconv.FormatFloat(bot.lat, 'f', 6, 64)
		lon := strconv.FormatFloat(bot.lon, 'f', 6, 64)
		pointTemplate := "node(around:{radius},{lat},{lon})[{poiType}={poiSubType}];"
		replacements := strings.NewReplacer("{radius}", radius, "{lat}", lat, "{lon}", lon, "{poiType}", poiType, "{poiSubType}", poiSubType)
		point := replacements.Replace(pointTemplate)
		pointsBuffer.WriteString(point)
	}

	queryTemplate := "https://overpass-api.de/api/interpreter?data=[out:json];({points});out;"
	replacements := strings.NewReplacer("{points}", pointsBuffer.String())

	return replacements.Replace(queryTemplate)
}

// Query OSM with single long query from createOSMQuery() to get all POIs near all travel bots. Collect into a []poi.
func getPOIs(query string) []poi {

	// Send a GET request to Overpass to get all POIs for all bot locations.
	resp, err := http.Get(query)
	check(err)

	// Read data into []byte.
	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	// fmt.Println(string(body))

	// Transform to a Golang struct we can use.
	result := &jsonStruct{}
	json.Unmarshal(body, &result)

	return result.Pois
}

// Used to find distance between two points on Earth in km.
func haversine(lonA float64, latA float64, lonB float64, latB float64) float64 {
	earthRadius := float64(6371)

	var deltaLat = (latB - latA) * (math.Pi / 180)
	var deltaLon = (lonB - lonA) * (math.Pi / 180)

	var a = math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latA*(math.Pi/180))*math.Cos(latB*(math.Pi/180))*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Used with Haversine formula to find out if a POI is within the radius of a bot, in km.
func withinBotRadius(haversineDistance float64, radius float64) bool {
	if haversineDistance < radius/1000 {
		return true
	} else {
		return false
	}
}

func getNearestPOIs(bots []bot, pois []poi, radius float64) []bot {
	newBotsSlice := []bot{}
	for _, bot := range bots {
		for _, poi := range pois {
			distance := haversine(bot.lat, bot.lon, poi.Lat, poi.Lon)
			if withinBotRadius(distance, radius) == true {
				bot.pois = append(bot.pois, poi)
			}
		}
		newBotsSlice = append(newBotsSlice, bot)
	}
	return newBotsSlice
}

func insertBotPOIsDB(bots []bot) {
	db, err := sql.Open("sqlite3", "database.db")
	check(err)
	for _, bot := range bots {
		for _, poi := range bot.pois {
			id := strconv.Itoa(poi.ID)
			lat := strconv.FormatFloat(poi.Lat, 'f', 6, 64)
			lon := strconv.FormatFloat(poi.Lon, 'f', 6, 64)

			statement := `INSERT INTO botpois (botid, osmid, latitude, longitude, visitype) values ($1, $2, $3, $4, $5);`

			check(err)
			_, err = db.Exec(statement, bot.id, id, lat, lon, `maybe`)
		}
	}
	db.Close()
}

func pickNewPOI() {
	// Insert current location/poi as "visited" in botpois.
	// Select all "maybe" pois. Pick one at random.
	// Replace current location with it.
	// Delete all "maybe" pois.
}

func travelBotGoroutine() {
	// Bring everything together and start it.
}

func allTogether() {
	travelbots := getTravelbots()
	query := createOSMQuery(travelbots)
	pois := getPOIs(query)
	travelbots = getNearestPOIs(travelbots, pois, 1000)
	insertBotPOIsDB(travelbots)

	//
}
