package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
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
	Radius    string
}

type botCurrentInfo struct {
	Lat    float64
	Lon    float64
	Radius string
	POIs   map[latlonStruct]map[string]string
}

func newBot() {
	// Get user location for bot automatically.
	// Or, let user enter bot location?
	// User also enters bot name and other details.
}

// Initialise a bot
func createBot(lat float64, lon float64, radius string) botCurrentInfo {
	// TODO: Add bot to DB.
	bot := botCurrentInfo{
		Lat:    lat,
		Lon:    lon,
		Radius: radius,
	}
	return bot
}

// Get data from Overpass QL API and put it into a *jsonStruct, to be formatted by formatPOIData().
func getPOIs(botLat float64, botLon float64, radius string, poiKey string, poiValue string) *jsonStruct {
	lat := strconv.FormatFloat(botLat, 'f', 7, 64)
	lon := strconv.FormatFloat(botLon, 'f', 7, 64)

	// Get POIs from Overpass.
	link := "https://overpass-api.de/api/interpreter?data=[out:json];node(around:{radius},{lat},{lon})[{poiKey}={poiValue}];out;"
	replacements := strings.NewReplacer("{radius}", radius, "{lat}", lat, "{lon}", lon, "{poiKey}", poiKey, "{poiValue}", poiValue)
	query := replacements.Replace(link)

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

// Format the *jsonStruct from getPOIs() into a map we can more easily use.
func formatPOIData(poiData *jsonStruct) map[latlonStruct]map[string]string {

	// Key: struct of bot location coords, value: map of POI tags.
	poiDataMap := make(map[latlonStruct]map[string]string)

	for _, item := range poiData.Elements {

		// Key: struct of bot location coords.
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

	mochimochi := createBot(41.718621, 44.795495, "1000")
	iceicebaby := createBot(41.717785, 44.794949, "1000")
	whateverbot := createBot(41.718417, 44.797915, "1000")

	allBots := make(map[string]botCurrentInfo)
	allBots["mochimochi"] = mochimochi
	allBots["iceicebaby"] = iceicebaby
	allBots["whateverbot"] = whateverbot

	for _, botInfo := range allBots {
		pois := getPOIs(botInfo.Lat, botInfo.Lon, botInfo.Radius, "amenity", "restaurant")
		poiDataMap := formatPOIData(pois)
		botInfo.POIs = poiDataMap
	}

	t, err := template.ParseFiles("views/base.gohtml", "views/index.gohtml", "views/allbots.gohtml")
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "base", allBots)

}

func myBotHandler(w http.ResponseWriter, r *http.Request) {

	myBot := createBot(41.718621, 44.795495, "1000")

	pois := getPOIs(myBot.Lat, myBot.Lon, myBot.Radius, "amenity", "restaurant")

	poiDataMap := formatPOIData(pois)

	myBot.POIs = poiDataMap

	fmt.Println(myBot.Lat)
	fmt.Println(myBot.Lon)
	fmt.Println(myBot.Radius)

	for k, v := range myBot.POIs {
		fmt.Println(k)
		for k2, v2 := range v {
			fmt.Println(k2, v2)
		}
		fmt.Println("\n")
	}

	t, err := template.ParseFiles("views/base.gohtml", "views/index.gohtml", "views/mybot.gohtml")
	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "base", myBot)
}
