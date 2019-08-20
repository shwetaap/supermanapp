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

type userData struct {
	Unix_timestamp int64 `json: "unix_timestamp"`
	Username  string `json: "username"`
	Event_uuid string	`json: "event_uuid"`
	Ip_address string `json: "ip_address"`
}

type geocurrentloc struct {
	Lat float64    `json: "lat"`
	Lon float64    `json: "lon"`
	radius float64  `json: "radius"`
}

type ipAccess struct {
	Ip string   `json: "ip_address,omitempty"`
	Speed float64	`json: "speed,omitempty"`
	Lat float64    `json: "lat,omitempty"`
	Lon float64    `json: "lon,omitempty"`
	radius float64  `json: "radius,omitempty"`
	Timestamp int64 `json: "timestamp,omitempty"`
}
type responseData struct {
	CurrentGeo *geocurrentloc `json: "currentGeo"`
	TraveltoCurrentGeoSuspicious bool `json: "traveltoCurrentGeoSuspicious,omitempty"`
	TravelFromCurrentGeoSuspicious bool `json: "travelfromCurrentGeoSuspicious,omitempty"`
	PrecedingIpAccess ipAccess `json: "precedingIpAccess,omitempty"`
	SubsequentIpAccess ipAccess `json: "subsequentIpAccess,omitempty"`
}



  func requestHandler(w http.ResponseWriter, r *http.Request) {
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
		     case "POST":
		              reqBody, err := ioutil.ReadAll(r.Body)
		              if err != nil {
						  log.Fatal("Error reading the body", err)
			              }
		              var data userData
		              fmt.Printf("%s\n", reqBody)
		              err = json.Unmarshal(reqBody, &data)

				 	  //decoder := json.NewDecoder(r.Body)
				 	  //fmt.Printf("%s\n", decoder)

				 	  //err = decoder.Decode(&data)
				 	  if err != nil {
						  log.Fatal("Decoding error: ", err)
				 	  }

				 	  fmt.Println(data.Username, data.Event_uuid)
		              w.Write([]byte("Received a POST request\n"))



		              // Data base actions
		              //os.Remove("./superman.db")
		              db, err := sql.Open("sqlite3", "./superman.db")
		              checkErr(err)
		              statement, err := db.Prepare("create table if not exists logindata (uuid text not null primary key, username text not null, unix_timestamp integer not null, ip_address text not null, lat real not null, lon real not null, radius not null)")
				 	  checkErr(err)
				 	  statement.Exec()



		              // Get the Latitude & Longitude from Geoip Db
		              //var location *geoloc
		              location := fetch_loc_ip(data.Ip_address)
		              println(location)

		              //Insert the location & Input data into the db

				      stmt, err := db.Prepare("INSERT INTO logindata(uuid, username, unix_timestamp, ip_address, lat, lon, radius) values(?,?,?,?,?,?,?)")
				      checkErr(err)

				 res, err := stmt.Exec(data.Event_uuid, data.Username, data.Unix_timestamp, data.Ip_address, location.Lat, location.Lon, location.radius)
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
				 var radius float64

/*				 var unix_timestamp int64
				 var ip_address string
				 var latprev float64
				 var lonprev float64
				 var radius uint16

				 var unix_timestamp int64
				 var ip_address string
				 var latsucc float64
				 var lonsucc float64
				 var radius uint16*/

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
				 var prevspeed int64
				 origin := haversine.Coord{Lat: location.Lat, Lon: location.Lon}
				 //origin_time := time.Unix(data.Unix_timestamp, 0)
				 //var destination_time int64
				 //var destination_radius uint16
				 var preipaccess ipAccess
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
						 preipaccess.radius = radius
						 preipaccess.Timestamp = unix_timestamp
					 }

				 }
				 if prevdistance != math.MaxFloat64 && prevdistance >= 0 {
				 	if prevdistance == 0 {
				 		prevspeed = 0
					} else {
						timediff := float64(data.Unix_timestamp-preipaccess.Timestamp) / 3600

						prevspeed = int64((prevdistance - float64(location.radius) - float64(preipaccess.radius)) / timediff)
					}
					 println(prevspeed)
				 }

				 //Subsequent Entry
				 var nextdistance = math.MaxFloat64
				 var nextspeed int64
				 var postipaccess ipAccess
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
						 postipaccess.radius = radius
						 postipaccess.Timestamp = unix_timestamp
					 }

				 }
				 if nextdistance != math.MaxFloat64 && nextdistance >= 0 {
				 	if nextdistance == 0{
				 		nextspeed = 0
					} else {
						timediff := float64(postipaccess.Timestamp-data.Unix_timestamp) / 3600

						nextspeed = int64((nextdistance - float64(location.radius) - float64(postipaccess.radius)) / timediff)
					}
					 println(nextspeed)
				 }
				// Construct the response
				//var responseJson []byte
				var responsepacket responseData

				 // TODO: Case when no preceding or subsequent locations are found
				 // TODO: Pointer access of data gives an error while retrieving data
				if prevdistance == math.MaxFloat64 && nextdistance == math.MaxFloat64{
					//responseJson, err = json.Marshal(*location)
					responsepacket.CurrentGeo = location
					//responsepacket.CurrentGeo.radius = location.radius
					//responsepacket.CurrentGeo.Lon = location.Lon
					//responsepacket.CurrentGeo.Lat = location.Lat
				} else if prevdistance == math.MaxFloat64{
					responsepacket.CurrentGeo = location
					responsepacket.SubsequentIpAccess = postipaccess
					if nextspeed <= 500 {
						responsepacket.TravelFromCurrentGeoSuspicious = false
					} else {
						responsepacket.TravelFromCurrentGeoSuspicious = true
					}

				} else if nextdistance == math.MaxFloat64 {
					responsepacket.CurrentGeo = location
					responsepacket.PrecedingIpAccess = preipaccess
					if prevspeed <= 500 {
						responsepacket.TraveltoCurrentGeoSuspicious = false
					} else {
						responsepacket.TraveltoCurrentGeoSuspicious = true
					}
				} else {

					//var responsepacket responseData
					responsepacket.CurrentGeo = location
					if prevspeed <= 500 {
						responsepacket.TraveltoCurrentGeoSuspicious = false
					} else {
						responsepacket.TraveltoCurrentGeoSuspicious = true
					}
					if nextspeed <= 500 {
						responsepacket.TravelFromCurrentGeoSuspicious = false
					} else {
						responsepacket.TravelFromCurrentGeoSuspicious = true
					}
					responsepacket.PrecedingIpAccess = preipaccess
					responsepacket.SubsequentIpAccess = postipaccess
				}// TODO: Radius not showing in json output, but present in the "geocurrentloc" struct
				 responseJson, err := json.Marshal(responsepacket)
				 if err != nil {
				 fmt.Fprintf(w, "Error: %s", err)
				 }

				 w.Header().Set("Content-Type", "application/json")

				 w.Write(responseJson)

		     default:
		              w.WriteHeader(http.StatusNotImplemented)
		              w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
		     }

	  }

  func checkErr(err error) {
	  if err != nil {
		  panic(err)
	  }
  }

/*  func  Location(ip net.IP) (*Location, error) {
	city, err := g.reader.City(ip)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lookup city")
	}
	return &Location{
		AccuracyRadius: city.Location.AccuracyRadius,
		Latitude:       city.Location.Latitude,
		Longitude:      city.Location.Longitude,
		MetroCode:      city.Location.MetroCode,
		TimeZone:       city.Location.TimeZone,
	}, nil
}*/

  func fetch_loc_ip(ipaddr string) (*geocurrentloc) {
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
	  var location geocurrentloc
	  location.radius = float64(record.Location.AccuracyRadius)
	  location.Lat = record.Location.Latitude
	  location.Lon = record.Location.Longitude
	  return &location
  }

  func main() {
	      http.HandleFunc("/", requestHandler)
	      http.ListenAndServe(":8000", nil)
	  }