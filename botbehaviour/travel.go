package botbehaviour

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type bot struct {
	ID     int
	Name   string
	Lat    float64
	Lon    float64
	Radius float64
	Pois   []poi
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
func GetTravelBots() []bot {

	// Select all travelbots.
	db, err := sql.Open("sqlite3", "database.db")
	check(err)

	query := `SELECT BotID, Name, Radius, Lat, Lon FROM bots WHERE bottype="travelbot";`
	rows, err := db.Query(query)
	check(err)
	defer rows.Close()

	var bots []bot

	for rows.Next() {
		b := bot{}

		err = rows.Scan(&b.ID, &b.Name, &b.Radius, &b.Lat, &b.Lon)
		check(err)
		bots = append(bots, b)
	}
	err = rows.Err()
	check(err)
	db.Close()
	return bots
}

// Concatenates separate queries for POIs from each bot in []bot into a single long query,
func createOSMQuery(bots []bot) string {
	poiType := "amenity"
	poiSubType := "restaurant"

	var pointsBuffer bytes.Buffer
	for _, bot := range bots {
		lat := strconv.FormatFloat(bot.Lat, 'f', 6, 64)
		lon := strconv.FormatFloat(bot.Lon, 'f', 6, 64)
		radius := strconv.FormatFloat(bot.Radius, 'f', 6, 64)
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

// Finds the nearest POIs to a bot using haversine() and withinBotRadius().
func getNearestPOIs(bots []bot, pois []poi, radius float64) []bot {
	newBotsSlice := []bot{}
	for _, bot := range bots {
		for _, poi := range pois {
			distance := haversine(bot.Lat, bot.Lon, poi.Lat, poi.Lon)
			if withinBotRadius(distance, radius) == true {
				bot.Pois = append(bot.Pois, poi)
			}
		}
		newBotsSlice = append(newBotsSlice, bot)
	}
	return newBotsSlice
}

// Insert POIs within bot radius to botpois table set to "maybe".
func insertBotPOIsDB(bots []bot) {
	db, err := sql.Open("sqlite3", "database.db")
	check(err)
	for _, bot := range bots {
		for _, poi := range bot.Pois {
			id := strconv.Itoa(poi.ID)
			lat := strconv.FormatFloat(poi.Lat, 'f', 6, 64)
			lon := strconv.FormatFloat(poi.Lon, 'f', 6, 64)

			statement := `INSERT INTO botpois (botid, osmid, latitude, longitude, visitype) values ($1, $2, $3, $4, $5);`
			check(err)
			_, err = db.Exec(statement, bot.ID, id, lat, lon, `maybe`)

			tags := poi.Tags
			statement = `INSERT INTO taginfo (
					botid,
					osmid,
					amenity,
					name,
					name_en,
					addr_housenumber,
					addr_street,
					opening_hours,
					phone,
					cuisine,
					description,
					internet_access,
					smoking,
					wheelchair
					) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`
			check(err)
			_, err = db.Exec(statement,
				bot.ID,
				id,
				tags["amenity"],
				tags["name"],
				tags["name:en"],
				tags["addr:housenumber"],
				tags["addr:street"],
				tags["opening_hours"],
				tags["phone"],
				tags["cuisine"],
				tags["description"],
				tags["internet_access"],
				tags["smoking"],
				tags["wheelchair"])
		}
	}
	db.Close()
}

// CREATE TABLE taginfo (
// 	botid,
// 	osmid TEXT,
// 	amenity TEXT,
// 	name TEXT,
// 	"name_en" TEXT,
// 	"addr_housenumber" TEXT,
// 	"addr_street" TEXT,
// 	opening_hours TEXT,
// 	phone TEXT,
// 	cuisine TEXT,
// 	description TEXT,
// 	internet_access TEXT,
// 	smoking TEXT,
// 	wheelchair TEXT
//   );

// Insert current location as "visited" in botpois. Replace current location with randomly picked location. Delete all "maybe" pois.
func pickNewPOI(bots []bot) []bot {
	newBotsSlice := []bot{}

	db, err := sql.Open("sqlite3", "database.db")
	check(err)

	for _, bot := range bots {
		// Insert current location as "visited" in botpois.
		id := strconv.Itoa(bot.ID)
		lat := strconv.FormatFloat(bot.Lat, 'f', 6, 64)
		lon := strconv.FormatFloat(bot.Lon, 'f', 6, 64)

		statement := `INSERT INTO botpois (botid, latitude, longitude, visitype) values ($1, $3, $4, $5);`

		check(err)
		_, err = db.Exec(statement, id, lat, lon, `visited`)

		// Select all "maybe" pois.

		type poi struct {
			ID   sql.NullInt64     `json:"id"`
			Lat  float64           `json:"lat"`
			Lon  float64           `json:"lon"`
			Tags map[string]string `json:"tags"`
			// VisitType: potential or visited.
			VisitType string
		}

		pois := []poi{}
		rows, err := db.Query(`SELECT osmid, latitude, longitude FROM botpois WHERE visitype="maybe" AND botid=$1;`, bot.ID)
		check(err)
		defer rows.Close()
		for rows.Next() {
			poi := poi{}
			err = rows.Scan(&poi.ID, &poi.Lat, &poi.Lon)
			check(err)
			pois = append(pois, poi)
		}
		err = rows.Err()
		check(err)

		// Replace current location with randomly picked location.
		if len(pois) > 0 {
			rand.Seed(time.Now().Unix())
			randomPOI := pois[rand.Intn(len(pois))]
			bot.Lat = randomPOI.Lat
			bot.Lon = randomPOI.Lon
		}
		newBotsSlice = append(newBotsSlice, bot)
		statement = `UPDATE bots SET Lat=?,Lon=? WHERE BotID=?;`
		_, err = db.Exec(statement, bot.Lat, bot.Lon, bot.ID)

	}
	db.Close()
	return newBotsSlice
}

// Delete all "maybe" pois.
func refresh() {
	db, err := sql.Open("sqlite3", "database.db")
	check(err)
	statement := `DELETE FROM botpois WHERE visitype="maybe"; DELETE FROM taginfo;`
	_, err = db.Exec(statement)
}

type tagsStruct struct {
	Amenity          string
	Name             string
	Name_en          string
	Addr_housenumber string
	Addr_street      string
	Opening_hours    string
	Phone            string
	Cuisine          string
	Description      string
	Internet_access  string
	Smoking          string
	Wheelchair       string
}

func getTags(db *sql.DB, workingPOI poi) poi {
	newPOI := poi{}

	// Get data from table taginfo.
	rows, err := db.Query(`SELECT 
	amenity,
	name,
	name_en,
	addr_housenumber,
	addr_street,
	opening_hours,
	phone,
	cuisine,
	description,
	internet_access,
	smoking,
	wheelchair
	FROM taginfo WHERE osmid=$1;`, workingPOI.ID)
	check(err)
	defer rows.Close()

	for rows.Next() {

		tags := tagsStruct{}

		err = rows.Scan(
			&tags.Amenity,
			&tags.Name,
			&tags.Name_en,
			&tags.Addr_housenumber,
			&tags.Addr_street,
			&tags.Opening_hours,
			&tags.Phone,
			&tags.Cuisine,
			&tags.Description,
			&tags.Internet_access,
			&tags.Smoking,
			&tags.Wheelchair)
		err = rows.Err()
		check(err)

		workingPOI.Tags = make(map[string]string)
		workingPOI.Tags["Amenity"] = tags.Amenity
		workingPOI.Tags["Name"] = tags.Name
		workingPOI.Tags["Name_en"] = tags.Name_en
		workingPOI.Tags["Addr_housenumber"] = tags.Addr_housenumber
		workingPOI.Tags["Addr_street"] = tags.Addr_street
		workingPOI.Tags["Opening_hours"] = tags.Opening_hours
		workingPOI.Tags["Phone"] = tags.Phone
		workingPOI.Tags["Cuisine"] = tags.Cuisine
		workingPOI.Tags["Description"] = tags.Description
		workingPOI.Tags["Internet_access"] = tags.Internet_access
		workingPOI.Tags["Smoking"] = tags.Smoking
		workingPOI.Tags["Wheelchair"] = tags.Wheelchair

		newPOI = workingPOI
	}
	return newPOI
}

func GetTravelPlans() []byte {
	bots := GetTravelBots()
	newBotsSlice := []bot{}

	db, err := sql.Open("sqlite3", "database.db")
	check(err)

	for _, bot := range bots {

		// Get data from table botpois.
		rows, err := db.Query(`SELECT osmid, latitude, longitude FROM botpois WHERE visitype="maybe" AND botid=$1`, bot.ID)
		check(err)
		defer rows.Close()

		poisTempSlice := []poi{}

		for rows.Next() {
			poi := poi{}
			var IDint64 sql.NullInt64

			err = rows.Scan(&IDint64, &poi.Lat, &poi.Lon)
			check(err)
			poi.ID = int(IDint64.Int64)

			poisTempSlice = append(poisTempSlice, poi)
		}
		err = rows.Err()
		check(err)

		for _, poi := range poisTempSlice {
			poi = getTags(db, poi)
			bot.Pois = append(bot.Pois, poi)
		}

		newBotsSlice = append(newBotsSlice, bot)
	}
	db.Close()

	botsJSON, err := json.Marshal(newBotsSlice)
	check(err)

	// for _, bot := range newBotsSlice {
	// 	for _, poi := range bot.Pois {
	// 		for k, v := range poi.Tags {
	// 			fmt.Println(k, v)
	// 		}
	// 	}
	// }
	return botsJSON
}

func GoTravel() {

	for {
		travelBots := GetTravelBots()
		query := createOSMQuery(travelBots)
		pois := getPOIs(query)
		travelBots = getNearestPOIs(travelBots, pois, 1000)
		insertBotPOIsDB(travelBots)
		travelBots = pickNewPOI(travelBots)
		time.Sleep(1800 * time.Second)
		refresh()

		for _, bot := range travelBots {
			fmt.Println(bot.Name, bot.Lat, bot.Lon)
			fmt.Println("")
		}
	}
}
