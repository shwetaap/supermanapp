# supermanapp
Superman Detector

This application is used to identify user accounts that have been compromised.

Usage:

1. Download the application

       git clone https://github.com/shwetaap/supermanapp.git  
       cd supermanapp 
2. Build and run the docker container

        docker build -t my-go-app .
        docker run -p 8000:8000 -it my-go-app
3. Access the application from your host machine


    curl -X POST -H "Content-Type: application/json" -d '{"username": "bob", "unix_timestamp":1514761200,"event_uuid":"85ad929a-db03-4bf4-9541-8f728fa12e40", "ip_address":"91.207.175.104"}' http://localhost:8000/
    
    Expected output: {"currentGeo":{"lat":34.0549,"lon":-118.2578,"radius":200}}
    
    curl -X POST -H "Content-Type: application/json" -d '{"username": "bob", "unix_timestamp":1514851200,"event_uuid":"85ad929a-db03-4bf4-9541-8f728fa12e42", "ip_address":"24.242.71.20"}' http://localhost:8000/
    
    Expected output: {"currentGeo":{"lat":30.4293,"lon":-97.7207,"radius":5},"traveltoCurrentGeoSuspicious":false,"precedingIpAccess":{"ip_address":"91.207.175.104","speed":40.74263197298201,"lat":34.0549,"lon":-118.2578,"radius":200,"timestamp":1514761200}}
    
    curl -X POST -H "Content-Type: application/json" -d '{"username": "bob", "unix_timestamp":1514764800,"event_uuid":"85ad929a-db03-4bf4-9541-8f728fa12e41", "ip_address":"206.81.252.6"}' http://localhost:8000/
    
    Expected output: {"currentGeo":{"lat":39.211,"lon":-76.8362,"radius":5},"traveltoCurrentGeoSuspicious":true,"travelfromCurrentGeoSuspicious":false,"precedingIpAccess":{"ip_address":"91.207.175.104","speed":2098.4527861218294,"lat":34.0549,"lon":-118.2578,"radius":200,"timestamp":1514761200},"subsequentIpAccess":{"ip_address":"24.242.71.20","speed":54.84361833707941,"lat":30.4293,"lon":-97.7207,"radius":5,"timestamp":1514851200}}
        
 
    
    