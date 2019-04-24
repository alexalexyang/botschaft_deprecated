package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/alexalexyang/botschaft/botbehaviour"
	"github.com/alexalexyang/botschaft/models"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// BotsTravel ---------------------------------------------------------------

func BotsTravelHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("views/base.gohtml", "views/index.gohtml", "views/botbehaviour/travel.gohtml")
	check(err)

	// Get travelbots - impt parts: bot, and its pois.
	// Marshal into json.
	// Test first with bot only without pois.
	bots := botbehaviour.GetTravelPlans()

	t.ExecuteTemplate(w, "base", string(bots))
}

// CRUD handlers ---------------------------------------------------------------

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("views/base.gohtml", "views/crud/createuser.gohtml")
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

	t, err := template.ParseFiles("views/base.gohtml", "views/crud/createbot.gohtml")
	check(err)

	if r.Method != http.MethodPost {
		t.ExecuteTemplate(w, "base", nil)
		return
	}

	jsonMap := make(map[string]interface{})

	jsonMap["UserID"] = r.FormValue("userid")
	jsonMap["BotID"] = r.FormValue("botid")
	jsonMap["Name"] = r.FormValue("name")
	jsonMap["Lat"] = r.FormValue("latitude")
	jsonMap["Lon"] = r.FormValue("longitude")
	jsonMap["Radius"] = 100

	fmt.Println(jsonMap)
	models.CreateInserttoDB("bots", jsonMap)

	http.Redirect(w, r, "http://localhost:3000", http.StatusSeeOther)

}

func CreateBotPoisHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("views/base.gohtml", "views/crud/createbotpois.gohtml")
	check(err)

	if r.Method != http.MethodPost {
		t.ExecuteTemplate(w, "base", nil)
		return
	}

	jsonMap := make(map[string]interface{})

	jsonMap["bsid"] = r.FormValue("bsid")
	jsonMap["osmid"] = r.FormValue("osmid")
	jsonMap["botid"] = r.FormValue("botid")
	jsonMap["latitude"] = r.FormValue("latitude")
	jsonMap["longitude"] = r.FormValue("longitude")
	// visitype typo
	jsonMap["visitype"] = r.FormValue("visittype")

	fmt.Println(jsonMap)
	models.CreateInserttoDB("botpois", jsonMap)

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
