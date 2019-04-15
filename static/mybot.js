// Load the map.
var mymap = L.map('map').setView([lat, lon], 13);

L.tileLayer('https://api.tiles.mapbox.com/v4/{id}/{z}/{x}/{y}.png?access_token={accessToken}', {
    attribution: 'Map data &copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors, <a href="https://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, Imagery © <a href="https://www.mapbox.com/">Mapbox</a>',
    maxZoom: 18,
    id: 'mapbox.streets',
    accessToken: 'pk.eyJ1IjoiYWxleGFsZXh5YW5nIiwiYSI6ImNqdTg2aGhtcTA2dzgzem9iMzk2ejNkbmoifQ.agUSBaqeL35dVXu9rBGPjA'
}).addTo(mymap);

// Do stuff on the map.

for (var poi in pois) {
    var circle = L.circle([pois[poi]["lat"], pois[poi]["lon"]], {
        color: 'red',
        fillColor: '#f03',
        fillOpacity: 0.2,
        weight: 0.6,
        radius: 10
    }).addTo(mymap);

    text = '<p>Amenity: ' + pois[poi]["amenity"] + '</h3>' +
        '<h3>Name: ' + pois[poi]["name:en"] + '</h3>' +
        '<p>Description: ' + pois[poi]["description"] + '</p>' +
        '<p>' + pois[poi]["addr:housenumber"] + " " + pois[poi]["addr:street"] + '</p>' +
        '<p>' + pois[poi]["lat"] + ', ' + pois[poi]["lon"] + '</p>'

    circle.bindPopup(text);
}

var botCircle = L.circle([lat, lon], {
    color: 'green',
    fillColor: 'green',
    fillOpacity: 0.2,
    weight: 0.6,
    radius: 10
}).addTo(mymap);

botText = '<h3>I\'m here!</h3>' +
    '<p>' + lat + ', ' + lon + '</p>'

botCircle.bindPopup(botText);

// Find the bot. The bot is a marker.