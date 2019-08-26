package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oschwald/geoip2-golang"
	"github.com/umahmood/haversine"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
)
// Input user data
type UserData struct {
	Unix_timestamp int64 `json:"unix_timestamp"`
	Username  string `json:"username"`
	Event_uuid string	`json:"event_uuid"`
	Ip_address string `json:"ip_address"`
}
// Structure for the location of an IPAddress
type Geocurrentloc struct {
	Lat float64    `json:"lat"`
	Lon float64    `json:"lon"`
	Radius uint16  `json:"radius"`
}
//  Structure for storing preceding and subsequent entries
type IpAccess struct {
	Ip string   `json:"ip_address,omitempty"`
	Speed float64	`json:"speed,omitempty"`
	Lat float64    `json:"lat,omitempty"`
	Lon float64    `json:"lon,omitempty"`
	Radius uint16  `json:"radius,omitempty"`
	Timestamp int64 `json:"timestamp,omitempty"`
}
//ResponseData struct
type ResponseData struct {
	CurrentGeo *Geocurrentloc `json:"currentGeo"`
	TraveltoCurrentGeoSuspicious *bool `json:"traveltoCurrentGeoSuspicious,omitempty"`
	TravelFromCurrentGeoSuspicious *bool `json:"travelfromCurrentGeoSuspicious,omitempty"`
	PrecedingIpAccess *IpAccess `json:"precedingIpAccess,omitempty"`
	SubsequentIpAccess *IpAccess `json:"subsequentIpAccess,omitempty"`
}

// Request Handler

  func RequestHandler(w http.ResponseWriter, r *http.Request) {
	     if r.URL.Path != "/" {
		              http.NotFound(w, r)
		              return
		      }
	     switch r.Method {
	     	 case "GET":
		              for k, v := range r.URL.Query() {
			                     fmt.Printf("%s: %s\n", k, v)
			              }
		              w.Write([]byte("Received a GET request\n"))
		              break
		     case "POST":
		     	      servicerequest(w,r)
		     	      break

		     default:
		              w.WriteHeader(http.StatusNotImplemented)
		              w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
		              break
		     }

	  }

  func checkErr(err error) {
	  if err != nil {
		  panic(err)
	  }
  }

  func Fetch_loc_ip(ipaddr string) (*Geocurrentloc) {
	  db, err := geoip2.Open("GeoLite2-City.mmdb")

	  if err != nil {
		  log.Fatal(err)
	  }
	  defer db.Close()
	  // For Strings that may be invalid, check that ip is not nil
	  ip := net.ParseIP(ipaddr)
	  record, err := db.City(ip)
	  if err != nil {
		  log.Fatal(err)
	  }
	  // set the location values and return the entry
	  var location Geocurrentloc
	  location.Radius = record.Location.AccuracyRadius
	  location.Lat = record.Location.Latitude
	  location.Lon = record.Location.Longitude
	  return &location
  }

  func servicerequest(w http.ResponseWriter, r *http.Request) (){
	  reqBody, err := ioutil.ReadAll(r.Body)
	  if err != nil {
		  log.Fatal("Error reading the body", err)
	  }
	  var data UserData
	  err = json.Unmarshal(reqBody, &data)

	  if err != nil {
		  log.Fatal("Decoding error: ", err)
	  }

	  // Data base actions

	  db, err := sql.Open("sqlite3", "./superman.db")
	  checkErr(err)
	  statement, err := db.Prepare("create table if not exists logindata (uuid text not null primary key, username text not null, unix_timestamp integer not null, ip_address text not null, lat real not null, lon real not null, radius not null)")
	  checkErr(err)
	  statement.Exec()
	  defer db.Close()

	  // Get the Latitude & Longitude from Geoip Db

	  location := Fetch_loc_ip(data.Ip_address)

	  //Insert the location & Input data into the db

	  stmt, err := db.Prepare("INSERT INTO logindata(uuid, username, unix_timestamp, ip_address, lat, lon, radius) values(?,?,?,?,?,?,?)")
	  checkErr(err)

	  res, err := stmt.Exec(data.Event_uuid, data.Username, data.Unix_timestamp, data.Ip_address, location.Lat, location.Lon, location.Radius)
	  checkErr(err)

	  _, err = res.LastInsertId()
	  checkErr(err)

	  var unix_timestamp int64
	  var ip_address string
	  var lat float64
	  var lon float64
	  var radius uint16

	  // Search for previous and subsequent nearest entries from the db


	  //Identify the Previous Entry
	  var prevdistance = math.MaxFloat64
	  // Calculate the distance using the Haversine method
	  origin := haversine.Coord{Lat: location.Lat, Lon: location.Lon}

	  var preipaccess IpAccess
	  // Collect entries with timestamp before the timestamp of the current request
	  prevrows, err := db.Query("SELECT unix_timestamp, ip_address, lat, lon, radius FROM logindata where username = ? and unix_timestamp < ?  ORDER BY unix_timestamp", data.Username, data.Unix_timestamp)
	  checkErr(err)
	  defer prevrows.Close()
	  for prevrows.Next(){
		  err = prevrows.Scan(&unix_timestamp, &ip_address, &lat, &lon, &radius)
		  destination := haversine.Coord{Lat: lat, Lon: lon}

		  //Calculate the distance between origin and destination, for every timestamp before the existing one
		  // and calculate the speed for the shortest distance
		  mi, _ := haversine.Distance(origin, destination)
		  if mi <= prevdistance {
			  prevdistance = mi
			  preipaccess.Ip = ip_address
			  preipaccess.Lat = lat
			  preipaccess.Lon = lon
			  preipaccess.Radius = radius
			  preipaccess.Timestamp = unix_timestamp
		  }

	  }
	  if prevdistance != math.MaxFloat64 && prevdistance >= 0 {
		  if prevdistance == 0 {
			  preipaccess.Speed = 0
		  } else {
		  	// Calculate the time in hours
			  timediff := float64(data.Unix_timestamp-preipaccess.Timestamp) / 3600
			// Calculate the speed
			  preipaccess.Speed = (prevdistance - float64(location.Radius) - float64(preipaccess.Radius)) / timediff
		  }
	  }

	  //Identify the  Subsequent Entry
	  var nextdistance = math.MaxFloat64
	  var postipaccess IpAccess
	  // Collect entries with timestamp beyond the timestamp of the current request
	  nextrows, err := db.Query("SELECT unix_timestamp, ip_address, lat, lon, radius FROM logindata where username = ? and unix_timestamp > ?  ORDER BY unix_timestamp", data.Username, data.Unix_timestamp)
	  checkErr(err)
	  defer nextrows.Close()
	  for nextrows.Next(){
		  err = nextrows.Scan(&unix_timestamp, &ip_address, &lat, &lon, &radius)
		  destination := haversine.Coord{Lat: lat, Lon: lon}
		  mi, _ := haversine.Distance(origin, destination)
		  if mi <= nextdistance {
			  nextdistance = mi
			  postipaccess.Ip = ip_address
			  postipaccess.Lat = lat
			  postipaccess.Lon = lon
			  postipaccess.Radius = radius
			  postipaccess.Timestamp = unix_timestamp
		  }

	  }
	  if nextdistance != math.MaxFloat64 && nextdistance >= 0 {
		  if nextdistance == 0{
			  postipaccess.Speed = 0
		  } else {
			  timediff := float64(postipaccess.Timestamp-data.Unix_timestamp) / 3600
			  postipaccess.Speed = (nextdistance - float64(location.Radius) - float64(postipaccess.Radius)) / timediff
		  }
	  }
	  var responsepacket ResponseData
	  var fromflag bool
	  var toflag bool

	  // Case 1 : When no previous or subsequent entries are found
	  if prevdistance == math.MaxFloat64 && nextdistance == math.MaxFloat64{
		  responsepacket.CurrentGeo = location
	  } else if prevdistance == math.MaxFloat64{
		  // Case 2 : When only subsequent entries are found
		  responsepacket.CurrentGeo = location
		  responsepacket.SubsequentIpAccess = &postipaccess
		  if postipaccess.Speed <= 500 {
			  fromflag = false
		  } else {
			  fromflag = true
		  }
		  responsepacket.TravelFromCurrentGeoSuspicious = &fromflag
	  } else if nextdistance == math.MaxFloat64 {
	  	// Case 3 : When only previous entries are found
		  responsepacket.CurrentGeo = location
		  responsepacket.PrecedingIpAccess = &preipaccess
		  if preipaccess.Speed <= 500 {
			  toflag = false
		  } else {
			  toflag = true
		  }
		  responsepacket.TraveltoCurrentGeoSuspicious = &toflag
	  } else {

		  // Case 4 : When both entries are found
		  responsepacket.CurrentGeo = location
		  if preipaccess.Speed <= 500 {
			  toflag = false
		  } else {
			  toflag = true
		  }
		  responsepacket.TraveltoCurrentGeoSuspicious = &toflag
		  if postipaccess.Speed <= 500 {
			  fromflag = false
		  } else {
			  fromflag = true
		  }
		  responsepacket.TravelFromCurrentGeoSuspicious = &fromflag
		  responsepacket.PrecedingIpAccess = &preipaccess
		  responsepacket.SubsequentIpAccess = &postipaccess
	  }
	  responseJson, err := json.Marshal(responsepacket)
	  if err != nil {
		  fmt.Fprintf(w, "Error: %s", err)
	  }

	  w.Header().Set("Content-Type", "application/json")

	  w.Write(responseJson)

  }
  func main() {

  	// Create a function handler for the incoming request
	      http.HandleFunc("/", RequestHandler)
	      // Create a server to listen to the incoming request
	      http.ListenAndServe(":8000", nil)
	  }
