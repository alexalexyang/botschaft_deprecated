package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/alexalexyang/botschaft/models"
)

// Struct to catch put data from Overpass API. Used by getPOIs().
type jsonStruct struct {
	Version   float64                  `json:"version"`
	Generator string                   `json:"generator"`
	Osm3s     map[string]string        `json:"osm3s"`
	Elements  []map[string]interface{} `json:"elements"`
	Radius    float64
}

// Get data from Overpass QL API and put it into a *jsonStruct, to be formatted by formatPOIData().
func getPOIs(bots map[string]models.BotBaseProfile) *jsonStruct {

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
	check(err)

	// Read data into []byte.
	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	// Transform to a Golang struct we can use.
	result := &jsonStruct{}
	json.Unmarshal(body, &result)

	return result
}

// This creates a map of POIs, with their coords as keys and their maps of tags as values.
// Format the *jsonStruct from getPOIs() into a map we can more easily use.
func formatPOIData(poiData *jsonStruct) map[models.LatLonStruct]map[string]string {

	// Key: struct of POI location coords, value: map of POI tags.
	poiDataMap := make(map[models.LatLonStruct]map[string]string)

	for _, item := range poiData.Elements {

		// Key: struct of POI identifiers.
		latlon := models.LatLonStruct{item["lat"].(float64), item["lon"].(float64), int(item["id"].(float64))}

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
