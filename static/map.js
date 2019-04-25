// Load the map.
var mymap = L.map('map').fitWorld();
mymap.locate({ setView: true, maxZoom: 16 });

L.tileLayer('https://api.tiles.mapbox.com/v4/{id}/{z}/{x}/{y}.png?access_token={accessToken}', {
    attribution: 'Map data &copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors, <a href="https://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="https://www.mapbox.com/">Mapbox</a>',
    maxZoom: 18,
    id: 'mapbox.streets',
    accessToken: 'pk.eyJ1IjoiYWxleGFsZXh5YW5nIiwiYSI6ImNqdTg2aGhtcTA2dzgzem9iMzk2ejNkbmoifQ.agUSBaqeL35dVXu9rBGPjA'
}).addTo(mymap);

// Do stuff on the map.

for (i = 0; i < bots.length; i++) {
    console.log("Bot is: ", bots[i].Name)

    var name = bots[i].Name
    var lat = bots[i].Lat
    var lon = bots[i].Lon

    var botCircle = L.circle([lat, lon], {
        color: 'green',
        fillColor: 'green',
        fillOpacity: 0.2,
        weight: 0.6,
        radius: 50
    }).addTo(mymap);

    botText = '<h3>' + bots[i].Name + '</h3>' +
        '</p>I\'m here!</h3>' +
        '<p>' + lat + ', ' + lon + '</p>'

    botCircle.bindPopup(botText);

    var pois = bots[i].Pois
        // In case there are no POIs, we check for null.
    if (pois != null) {
        for (j = 0; j < pois.length; j++) {
            // console.log(pois[j])
            var circle = L.circle([pois[j].lat, pois[j].lon], {
                color: 'red',
                fillColor: '#f03',
                fillOpacity: 0.2,
                weight: 0.6,
                radius: 10
            }).addTo(mymap);

            // console.log(pois[j].tags)

            text = 'Amenity: ' + pois[j].tags.Amenity + '</br>' +
                pois[j].lat + ', ' + pois[j].lon + '</br>' +
                '<h3>Name: ' + pois[j].tags.Name_en + '</h3>' +
                '<p>Description: ' + pois[j].tags.Description + '</p>' +
                'Address: </br>' +
                '<p>' + pois[j].tags.Addr_housenumber + " " + pois[j].tags.Addr_street + '</p>' +
                '<p>Opening hours: ' + pois[j].tags.Opening_hours + '</p>' +
                '<p>Phone: ' + pois[j].tags.Phone + '</p>' +
                '<p>Cuisine: ' + pois[j].tags.Cuisine + '</p>' +
                '<p>Internet: ' + pois[j].tags.Internet + '</p>' +
                '<p>Wheelchair: ' + pois[j].tags.Wheelchair + '</p>' +
                '<p>Smoking: ' + pois[j].tags.Smoking + '</p>'

            circle.bindPopup(text);
        }



    }


}