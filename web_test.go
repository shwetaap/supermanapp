package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateEntry(t *testing.T) {

	// Request for first entry
	var jsonStr = []byte(`{"username": "bob", "unix_timestamp":1514761200,"event_uuid":"85ad929a-db03-4bf4-9541-8f728fa12e40", "ip_address":"91.207.175.104"}`)

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RequestHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"currentGeo":{"lat":34.0549,"lon":-118.2578,"radius":200}}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// Request for second entry

	jsonStr = []byte(`{"username": "bob", "unix_timestamp":1514851200,"event_uuid":"85ad929a-db03-4bf4-9541-8f728fa12e42", "ip_address":"24.242.71.20"}`)

	req, err = http.NewRequest("POST", "/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(RequestHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"currentGeo":{"lat":30.4293,"lon":-97.7207,"radius":5},"traveltoCurrentGeoSuspicious":false,"precedingIpAccess":{"ip_address":"91.207.175.104","speed":40.74263197298201,"lat":34.0549,"lon":-118.2578,"radius":200,"timestamp":1514761200}}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// Request for third entry

	jsonStr = []byte(`{"username": "bob", "unix_timestamp":1514764800,"event_uuid":"85ad929a-db03-4bf4-9541-8f728fa12e41", "ip_address":"206.81.252.7"}`)

	req, err = http.NewRequest("POST", "/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(RequestHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected = `{"currentGeo":{"lat":39.211,"lon":-76.8362,"radius":5},"traveltoCurrentGeoSuspicious":true,"travelfromCurrentGeoSuspicious":false,"precedingIpAccess":{"ip_address":"91.207.175.104","speed":2098.4527861218294,"lat":34.0549,"lon":-118.2578,"radius":200,"timestamp":1514761200},"subsequentIpAccess":{"ip_address":"24.242.71.20","speed":54.84361833707941,"lat":30.4293,"lon":-97.7207,"radius":5,"timestamp":1514851200}}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	os.Remove("./superman.db")
}
