package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// General purpose struct for GPS coordinates.
type latlonStruct struct {
	Lat float64
	Lon float64
	ID  int
}

// Struct to catch put data from Overpass API. Used by getPOIs().
type jsonStruct struct {
	Version   float64                  `json:"version"`
	Generator string                   `json:"generator"`
	Osm3s     map[string]string        `json:"osm3s"`
	Elements  []map[string]interface{} `json:"elements"`
	Radius    float64
}

type botOwner struct {
	OwnerID   int
	OwnerName string
}

type botStruct struct {
	OwnerID int
	BotID   int
	Name    string
	Likes   map[string]bool
	// Visits take OSM POI ID as key and map of lat and lon as values.
	SuggestedVisits map[latlonStruct]map[string]string
	UpcomingVisits  map[int]map[string]float64
	PastVisits      map[int]map[string]float64
	// Friends takes BotID as key and true as value.
	Friends map[int]bool
	Lat     float64
	Lon     float64
	Radius  float64
	// Messages the bot sends to others. It takes the other party's ID as key. The value takes datetimestamp as key, and message as string.
	Messages map[int]map[string]string

	// latlonStruct contains the location of the POI.
	// The value map[string]string contains the POI tags.
	POIs map[latlonStruct]map[string]string
}

// Initialise a bot
func createBot(name string, lat float64, lon float64, radius float64) botStruct {
	// TODO: Add bot to DB.
	bot := botStruct{
		Lat:    lat,
		Lon:    lon,
		Radius: radius,
		// POIs:   make(map[latlonStruct]map[string]string),
	}
	return bot
}

func getAllBots() map[string]botStruct {
	radius := float64(1000)
	mochimochi := createBot("mochimochi", 41.718621, 44.795495, radius)
	iceicebaby := createBot("iceicebaby", 41.717785, 44.794949, radius)
	whateverbot := createBot("whateverbot", 41.718417, 44.797915, radius)
	rustabot := createBot("rustabot", 41.705064, 44.789050, radius)
	benny := createBot("benny", 42.705064, 46.789050, radius)

	allBots := make(map[string]botStruct)
	allBots["mochimochi"] = mochimochi
	allBots["iceicebaby"] = iceicebaby
	allBots["whateverbot"] = whateverbot
	allBots["rustabot"] = rustabot
	allBots["benny"] = benny

	return allBots
}

func haversine(lonFrom float64, latFrom float64, lonTo float64, latTo float64) float64 {
	earthRadius := float64(6371)

	var deltaLat = (latTo - latFrom) * (math.Pi / 180)
	var deltaLon = (lonTo - lonFrom) * (math.Pi / 180)

	var a = math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latFrom*(math.Pi/180))*math.Cos(latTo*(math.Pi/180))*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func nearestPOIs(distance float64, radius float64) bool {
	if distance < radius/1000 {
		return true
	} else {
		return false
	}
}

// Update each bot with map of nearest POIs.
func updateBotPOIMap(bots map[string]botStruct) map[string]botStruct {
	pois := getPOIs(bots)
	poiDataMap := formatPOIData(pois)

	newAllBots := make(map[string]botStruct)

	for name, bot := range bots {
		// fmt.Println("Bot in range: ", bot)
		poiMapPerBot := make(map[latlonStruct]map[string]string)

		for k, v := range poiDataMap {
			distance := haversine(k.Lat, k.Lon, bot.Lat, bot.Lon)
			if nearestPOIs(distance, bot.Radius) == true {
				poiMapPerBot[k] = v
				// fmt.Println("Innermost:", k)
			}
		}
		bot.POIs = poiMapPerBot
		newAllBots[name] = bot
	}
	return newAllBots
}

// Get data from Overpass QL API and put it into a *jsonStruct, to be formatted by formatPOIData().
func getPOIs(bots map[string]botStruct) *jsonStruct {

	// Template string for part of query for each bot location.
	pointTemplate := "node(around:{radius},{lat},{lon})[{poiKey}={poiValue}];"

	// Insert POI type into template.
	poiKey := "amenity"
	poiValue := "restaurant"
	replacements := strings.NewReplacer("{poiKey}", poiKey, "{poiValue}", poiValue)
	point := replacements.Replace(pointTemplate)

	// Concatenate all parts.
	var buffer bytes.Buffer
	for _, botInfo := range bots {
		lat64 := botInfo.Lat
		lon64 := botInfo.Lon
		lat := strconv.FormatFloat(lat64, 'f', 6, 64)
		lon := strconv.FormatFloat(lon64, 'f', 6, 64)

		radius := strconv.FormatFloat(botInfo.Radius, 'f', 0, 64)

		replacements := strings.NewReplacer("{radius}", radius, "{lat}", lat, "{lon}", lon, "{poiKey}", poiKey, "{poiValue}", poiValue)
		newPoint := replacements.Replace(point)
		buffer.WriteString(newPoint)
	}

	// Insert parts into main query.
	link := "https://overpass-api.de/api/interpreter?data=[out:json];({buffer});out;"
	replacements = strings.NewReplacer("{buffer}", buffer.String())
	query := replacements.Replace(link)

	// Send a GET request to Overpass to get all POIs for all bot locations.
	resp, err := http.Get(query)
	if err != nil {
		log.Fatalln(err)
	}

	// Read data into []byte.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// Transform to a Golang struct we can use.
	result := &jsonStruct{}
	json.Unmarshal(body, &result)

	return result
}

// This creates a map of POIs, with their coords as keys and their maps of tags as values.
// Format the *jsonStruct from getPOIs() into a map we can more easily use.
func formatPOIData(poiData *jsonStruct) map[latlonStruct]map[string]string {

	// Key: struct of POI location coords, value: map of POI tags.
	poiDataMap := make(map[latlonStruct]map[string]string)

	for _, item := range poiData.Elements {

		// Key: struct of POI identifiers.
		latlon := latlonStruct{item["lat"].(float64), item["lon"].(float64), int(item["id"].(float64))}

		// Value: map of POI tags.
		tagsMap := make(map[string]string)
		tagsInterface := item["tags"].(map[string]interface{})

		for k, v := range tagsInterface {
			tagsMap[k] = v.(string)
		}
		poiDataMap[latlon] = tagsMap
	}
	return poiDataMap
}

func main() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/mybot", myBotHandler)
	http.ListenAndServe(":3000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	allbots := getAllBots()

	allbots = updateBotPOIMap(allbots)

	t, err := template.ParseFiles("views/base.gohtml", "views/index.gohtml", "views/allbots.gohtml")
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "base", allbots)

}

func myBotHandler(w http.ResponseWriter, r *http.Request) {

}
