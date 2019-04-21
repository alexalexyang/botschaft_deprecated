package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type userStruct struct {
	ID      int
	Name    string
	Age     string
	Gender  string
	City    string
	Country string
}

type BotBaseProfile struct {
	UserID int
	BotID  int
	Name   string
	Lat    float64
	Lon    float64
	Radius float64
}

type BotPOIs struct {
	BSID      int
	OSMID     int
	BotID     int
	Lat       float64
	Lon       float64
	VisitType string
}

type POI struct {
	BSID  int
	OSMID int
	Lat   float64
	Lon   float64
}

type BotFriends struct {
	BotID    int
	FriendID int
}

type BotMessages struct {
	BotID         int
	FriendID      int
	FriendMessage string
}

type BotLikes struct {
	BotID    int
	Activity string
	Thing    string
}

// LatLonStruct is a general purpose struct for GPS coordinates.
type LatLonStruct struct {
	Lat float64
	Lon float64
	ID  int
}

// PrepSQLValues processes incoming JSON and returns two []interfaces{}, one for its keys, and one for values.
func PrepSQLValues(jsonBody map[string]interface{}) ([]interface{}, []interface{}) {

	var columns []interface{}
	var row []interface{}

	for key, value := range jsonBody {
		columns = append(columns, key)
		row = append(row, value)
	}

	return columns, row
}

// Concatenate the columns into a string suitable for an SQL statement.
func concatColtoCreateTable(jsonBody map[string]interface{}) string {
	var buffer bytes.Buffer
	for k, v := range jsonBody {

		var substring string
		switch v.(type) {
		case int:
			fmt.Println("create int: ", v)
			substring = k + ` INTEGER NULL,`
		case float64:
			if v == float64(int(v.(float64))) {
				fmt.Println(" is defo an int.")
				substring = k + ` INTEGER NULL,`
			} else {
				fmt.Println("create float64: ", v)
				substring = k + ` REAL NULL,`
			}
		default:
			fmt.Println("create string: ", v)
			substring = k + ` TEXT NULL,`
		}
		buffer.WriteString(substring)
	}
	bufferString := buffer.String()[:len(buffer.String())-1]
	return bufferString
}

// Concatenate the columns into a string suitable for an SQL statement.
func concatColtoInsert(columns []interface{}) string {
	fmt.Println("columns: ", columns)
	var buffer bytes.Buffer
	for _, item := range columns {
		substring := item.(string) + ","
		buffer.WriteString(substring)
	}
	bufferString := buffer.String()[:len(buffer.String())-1]
	fmt.Println("bufferstring: ", bufferString)
	return bufferString
}

// CreateInserttoDB takes in variable numbers of columns and row values to create and/or insert into a database.
func CreateInserttoDB(dbName string, jsonBody map[string]interface{}) {

	columns, row := PrepSQLValues(jsonBody)

	db, err := sql.Open("sqlite3", "database.db")
	check(err)

	// Create table if not exists.
	queryTemplate := `CREATE TABLE IF NOT EXISTS {dbName} ({buffer});`
	buffer := concatColtoCreateTable(jsonBody)
	replacements := strings.NewReplacer("{dbName}", dbName, "{buffer}", buffer)
	query := replacements.Replace(queryTemplate)

	stmt, err := db.Prepare(query)
	_, err = stmt.Exec()

	// Alter table add column if not exists.

	// Update table where X if exists.

	// Insert values.
	buffer = concatColtoInsert(columns)

	placeholders := strings.Repeat("?,", len(row))
	placeholders = placeholders[:len(placeholders)-1]

	queryTemplate = `INSERT INTO {dbName} ({buffer}) values ({placeholders});`
	replacements = strings.NewReplacer("{dbName}", dbName, "{buffer}", buffer, "{placeholders}", placeholders)
	query = replacements.Replace(queryTemplate)

	stmt, err = db.Prepare(query)
	_, err = stmt.Exec(row...)

	db.Close()
}

// Initialise a bot
func createBot(name string, lat float64, lon float64, radius float64) BotBaseProfile {
	// TODO: Add bot to DB.
	bot := BotBaseProfile{
		Lat:    lat,
		Lon:    lon,
		Radius: radius,
		// POIs:   make(map[controllers.LatLonStruct]map[string]string),
	}
	return bot
}

func GetAllBots() map[string]BotBaseProfile {
	radius := float64(1000)
	mochimochi := createBot("mochimochi", 41.718621, 44.795495, radius)
	iceicebaby := createBot("iceicebaby", 41.717785, 44.794949, radius)
	whateverbot := createBot("whateverbot", 41.718417, 44.797915, radius)
	rustabot := createBot("rustabot", 41.705064, 44.789050, radius)
	benny := createBot("benny", 42.705064, 46.789050, radius)

	allBots := make(map[string]BotBaseProfile)
	allBots["mochimochi"] = mochimochi
	allBots["iceicebaby"] = iceicebaby
	allBots["whateverbot"] = whateverbot
	allBots["rustabot"] = rustabot
	allBots["benny"] = benny

	return allBots
}
