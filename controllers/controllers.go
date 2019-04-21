package controllers

import (
	"html/template"
	"math"
	"net/http"

	"github.com/alexalexyang/botschaft/models"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
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

// Should split into model for updating of POIs and then maybe controller for returning newAllBots
// Update each bot with map of nearest POIs.
func updateBotPOIMap(bots map[string]models.BotBaseProfile) map[string]models.BotBaseProfile {
	pois := getPOIs(bots)
	poiDataMap := formatPOIData(pois)

	newAllBots := make(map[string]models.BotBaseProfile)

	for name, bot := range bots {
		// fmt.Println("Bot in range: ", bot)
		poiMapPerBot := make(map[models.LatLonStruct]map[string]string)

		for k, v := range poiDataMap {
			distance := haversine(k.Lat, k.Lon, bot.Lat, bot.Lon)
			if nearestPOIs(distance, bot.Radius) == true {
				poiMapPerBot[k] = v
				// fmt.Println("Innermost:", k)
			}
		}
		// bot.POIs = poiMapPerBot
		newAllBots[name] = bot
	}
	return newAllBots
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	allbots := models.GetAllBots()

	allbots = updateBotPOIMap(allbots)

	t, err := template.ParseFiles("views/base.gohtml", "views/index.gohtml", "views/allbots.gohtml")
	check(err)

	t.ExecuteTemplate(w, "base", allbots)

}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("views/base.gohtml", "views/createuser.gohtml")
	check(err)

	if r.Method != http.MethodPost {
		t.ExecuteTemplate(w, "base", nil)
		return
	}

	jsonMap := make(map[string]interface{})

	jsonMap["UserID"] = r.FormValue("userid")
	jsonMap["Name"] = r.FormValue("name")
	jsonMap["Age"] = r.FormValue("age")
	jsonMap["Gender"] = r.FormValue("gender")
	jsonMap["City"] = r.FormValue("city")
	jsonMap["Country"] = r.FormValue("country")

	models.CreateInserttoDB("users", jsonMap)

	http.Redirect(w, r, "http://localhost:3000", http.StatusSeeOther)

}

func CreateBotHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("views/base.gohtml", "views/createbot.gohtml")
	check(err)

	if r.Method != http.MethodPost {
		t.ExecuteTemplate(w, "base", nil)
		return
	}

	jsonMap := make(map[string]interface{})

	jsonMap["UserID"] = r.FormValue("userid")
	jsonMap["BotID"] = r.FormValue("botid")
	jsonMap["Name"] = r.FormValue("name")
	jsonMap["Lat"] = r.FormValue("lat")
	jsonMap["Lon"] = r.FormValue("lon")
	jsonMap["Radius"] = 10

	models.CreateInserttoDB("bots", jsonMap)

	http.Redirect(w, r, "http://localhost:3000", http.StatusSeeOther)

}

// func CreateHandler(w http.ResponseWriter, r *http.Request) {
// 	body, err := ioutil.ReadAll(r.Body)
// 	check(err)
// 	defer r.Body.Close()

// 	var jsonMap map[string]interface{}

// 	err = json.Unmarshal(body, &jsonMap)
// 	check(err)
// 	models.CreateInserttoDB("test", jsonMap)

// }
