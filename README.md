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
    
    Expected output: 
    
    curl -X POST -H "Content-Type: application/json" -d '{"username": "bob", "unix_timestamp":1514851200,"event_uuid":"85ad929a-db03-4bf4-9541-8f728fa12e42", "ip_address":"24.242.71.20"}' http://localhost:8000/
    
    Expected output: 
    
    curl -X POST -H "Content-Type: application/json" -d '{"username": "bob", "unix_timestamp":1514764800,"event_uuid":"85ad929a-db03-4bf4-9541-8f728fa12e41", "ip_address":"206.81.252.6"}' http://localhost:8000/
    
    Expected output: 
        
 
    
    