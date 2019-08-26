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

type UserData struct {
	Unix_timestamp int64 `json:"unix_timestamp"`
	Username  string `json:"username"`
	Event_uuid string	`json:"event_uuid"`
	Ip_address string `json:"ip_address"`
}

type Geocurrentloc struct {
	Lat float64    `json:"lat"`
	Lon float64    `json:"lon"`
	Radius uint16  `json:"radius"`
}

type IpAccess struct {
	Ip string   `json:"ip_address,omitempty"`
	Speed float64	`json:"speed,omitempty"`
	Lat float64    `json:"lat,omitempty"`
	Lon float64    `json:"lon,omitempty"`
	Radius uint16  `json:"radius,omitempty"`
	Timestamp int64 `json:"timestamp,omitempty"`
}
type ResponseData struct {
	CurrentGeo *Geocurrentloc `json:"currentGeo"`
	TraveltoCurrentGeoSuspicious *bool `json:"traveltoCurrentGeoSuspicious,omitempty"`
	TravelFromCurrentGeoSuspicious *bool `json:"travelfromCurrentGeoSuspicious,omitempty"`
	PrecedingIpAccess *IpAccess `json:"precedingIpAccess,omitempty"`
	SubsequentIpAccess *IpAccess `json:"subsequentIpAccess,omitempty"`
}



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
	  // If you are using strings that may be invalid, check that ip is not nil
	  ip := net.ParseIP(ipaddr)
	  record, err := db.City(ip)
	  if err != nil {
		  log.Fatal(err)
	  }
	  //println(record.Location)
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
	  fmt.Printf("%s\n", reqBody)
	  err = json.Unmarshal(reqBody, &data)

	  if err != nil {
		  log.Fatal("Decoding error: ", err)
	  }

	  fmt.Println(data.Username, data.Event_uuid)
	  //w.Write([]byte("Received a POST request\n"))



	  // Data base actions
	  //os.Remove("./superman.db")
	  db, err := sql.Open("sqlite3", "./superman.db")
	  checkErr(err)
	  statement, err := db.Prepare("create table if not exists logindata (uuid text not null primary key, username text not null, unix_timestamp integer not null, ip_address text not null, lat real not null, lon real not null, radius not null)")
	  checkErr(err)
	  statement.Exec()



	  // Get the Latitude & Longitude from Geoip Db

	  location := Fetch_loc_ip(data.Ip_address)
	  println(location)

	  //Insert the location & Input data into the db

	  stmt, err := db.Prepare("INSERT INTO logindata(uuid, username, unix_timestamp, ip_address, lat, lon, radius) values(?,?,?,?,?,?,?)")
	  checkErr(err)

	  res, err := stmt.Exec(data.Event_uuid, data.Username, data.Unix_timestamp, data.Ip_address, location.Lat, location.Lon, location.Radius)
	  checkErr(err)

	  id, err := res.LastInsertId()
	  checkErr(err)
	  fmt.Println(id)


	  //Check if insertion is correct

	  // query
	  rows, err := db.Query("SELECT * FROM logindata")
	  checkErr(err)
	  var event_uuid string
	  var username string
	  var unix_timestamp int64
	  var ip_address string
	  var lat float64
	  var lon float64
	  var radius uint16


	  for rows.Next() {
		  err = rows.Scan(&event_uuid, &username, &unix_timestamp, &ip_address, &lat, &lon, &radius)
		  checkErr(err)
		  fmt.Println(event_uuid)
		  fmt.Println(username)
		  fmt.Println(unix_timestamp)
		  fmt.Println(ip_address)
		  fmt.Println(lat)
		  fmt.Println(lon)
		  fmt.Println(radius)
	  }

	  rows.Close()


	  // Search for previous and subsequent nearest entries from the db


	  //Previous Entry
	  var prevdistance = math.MaxFloat64
	  //var prevspeed int64
	  origin := haversine.Coord{Lat: location.Lat, Lon: location.Lon}
	  //origin_time := time.Unix(data.Unix_timestamp, 0)
	  //var destination_time int64
	  //var destination_radius uint16
	  var preipaccess IpAccess
	  //stmt, err = db.Prepare("SELECT unix_timestamp, ip_address, lat, lon, radius FROM logindata where username = ? and unix_timestamp < ?  ORDER BY unix_timestamp")
	  //checkErr(err)
	  prevrows, err := db.Query("SELECT unix_timestamp, ip_address, lat, lon, radius FROM logindata where username = ? and unix_timestamp < ?  ORDER BY unix_timestamp", data.Username, data.Unix_timestamp)
	  checkErr(err)
	  defer prevrows.Close()
	  for prevrows.Next(){
		  err = prevrows.Scan(&unix_timestamp, &ip_address, &lat, &lon, &radius)
		  destination := haversine.Coord{Lat: lat, Lon: lon}
		  mi, _ := haversine.Distance(origin, destination)
		  if mi <= prevdistance {
			  prevdistance = mi
			  //destination_time = unix_timestamp
			  //destination_radius = radius
			  preipaccess.Ip = ip_address
			  preipaccess.Lat = lat
			  preipaccess.Lon = lon
			  preipaccess.Radius = radius
			  preipaccess.Timestamp = unix_timestamp
		  }

	  }
	  if prevdistance != math.MaxFloat64 && prevdistance >= 0 {
		  if prevdistance == 0 {
			  //prevspeed = 0
			  preipaccess.Speed = 0
		  } else {
			  timediff := float64(data.Unix_timestamp-preipaccess.Timestamp) / 3600

			  preipaccess.Speed = (prevdistance - float64(location.Radius) - float64(preipaccess.Radius)) / timediff

			  //prevspeed = int64((prevdistance - float64(location.Radius) - float64(preipaccess.Radius)) / timediff)
		  }
		  // println(prevspeed)
	  }

	  //Subsequent Entry
	  var nextdistance = math.MaxFloat64
	  // var nextspeed int64
	  var postipaccess IpAccess
	  nextrows, err := db.Query("SELECT unix_timestamp, ip_address, lat, lon, radius FROM logindata where username = ? and unix_timestamp > ?  ORDER BY unix_timestamp", data.Username, data.Unix_timestamp)
	  checkErr(err)
	  defer nextrows.Close()
	  for nextrows.Next(){
		  err = nextrows.Scan(&unix_timestamp, &ip_address, &lat, &lon, &radius)
		  destination := haversine.Coord{Lat: lat, Lon: lon}
		  mi, _ := haversine.Distance(origin, destination)
		  if mi <= nextdistance {
			  nextdistance = mi
			  //destination_time = unix_timestamp
			  //destination_radius = radius
			  postipaccess.Ip = ip_address
			  postipaccess.Lat = lat
			  postipaccess.Lon = lon
			  postipaccess.Radius = radius
			  postipaccess.Timestamp = unix_timestamp
		  }

	  }
	  if nextdistance != math.MaxFloat64 && nextdistance >= 0 {
		  if nextdistance == 0{
			  //nextspeed = 0
			  postipaccess.Speed = 0
		  } else {
			  timediff := float64(postipaccess.Timestamp-data.Unix_timestamp) / 3600
			  postipaccess.Speed = (nextdistance - float64(location.Radius) - float64(postipaccess.Radius)) / timediff
			  //nextspeed = int64((nextdistance - float64(location.Radius) - float64(postipaccess.Radius)) / timediff)
		  }
		  // println(nextspeed)
	  }
	  // Construct the response
	  //var responseJson []byte
	  var responsepacket ResponseData
	  var fromflag bool
	  var toflag bool

	  // TODO: Case when no preceding or subsequent locations are found
	  // TODO: Pointer access of data gives an error while retrieving data
	  if prevdistance == math.MaxFloat64 && nextdistance == math.MaxFloat64{
		  //responseJson, err = json.Marshal(*location)
		  responsepacket.CurrentGeo = location
		  /*responsepacket.TravelFromCurrentGeoSuspicious = nil
		  responsepacket.TraveltoCurrentGeoSuspicious = nil
		  responsepacket.PrecedingIpAccess = nil
		  responsepacket.SubsequentIpAccess = nil*/

		  //responsepacket.CurrentGeo.radius = location.radius
		  //responsepacket.CurrentGeo.Lon = location.Lon
		  //responsepacket.CurrentGeo.Lat = location.Lat
	  } else if prevdistance == math.MaxFloat64{
		  responsepacket.CurrentGeo = location
		  responsepacket.SubsequentIpAccess = &postipaccess
		  if postipaccess.Speed <= 500 {
			  fromflag = false
			  //responsepacket.TravelFromCurrentGeoSuspicious = &fromflag
		  } else {
			  fromflag = true
			  //responsepacket.TravelFromCurrentGeoSuspicious = &fromflag
		  }
		  responsepacket.TravelFromCurrentGeoSuspicious = &fromflag
	  } else if nextdistance == math.MaxFloat64 {
		  responsepacket.CurrentGeo = location
		  responsepacket.PrecedingIpAccess = &preipaccess
		  //responsepacket.PrecedingIpAccess = preipaccess
		  if preipaccess.Speed <= 500 {
			  toflag = false
			  //responsepacket.TraveltoCurrentGeoSuspicious = false
		  } else {
			  toflag = true
			  //responsepacket.TraveltoCurrentGeoSuspicious = true
		  }
		  responsepacket.TraveltoCurrentGeoSuspicious = &toflag
	  } else {

		  //var responsepacket responseData
		  responsepacket.CurrentGeo = location
		  if preipaccess.Speed <= 500 {
			  toflag = false
			  //responsepacket.TraveltoCurrentGeoSuspicious = false
		  } else {
			  toflag = true
			  //responsepacket.TraveltoCurrentGeoSuspicious = true
		  }
		  responsepacket.TraveltoCurrentGeoSuspicious = &toflag
		  if postipaccess.Speed <= 500 {
			  fromflag = false
			  //responsepacket.TravelFromCurrentGeoSuspicious = false
		  } else {
			  fromflag = true
			  //responsepacket.TravelFromCurrentGeoSuspicious = true
		  }
		  responsepacket.TravelFromCurrentGeoSuspicious = &fromflag
		  responsepacket.PrecedingIpAccess = &preipaccess
		  responsepacket.SubsequentIpAccess = &postipaccess
	  }// TODO: Radius not showing in json output, but present in the "geocurrentloc" struct
	  responseJson, err := json.Marshal(responsepacket)
	  if err != nil {
		  fmt.Fprintf(w, "Error: %s", err)
	  }

	  w.Header().Set("Content-Type", "application/json")

	  w.Write(responseJson)

  }
  func main() {
	      http.HandleFunc("/", RequestHandler)
	      http.ListenAndServe(":8000", nil)
	  }
