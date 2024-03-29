Website screenshot as a service. <br>
Dependencies:<br>
1. MongoDB as metadta and file storage. https://www.mongodb.com/
2. NATS as message queue. https://nats.io/
3. Headless Chrome. can be installed or in docker container https://hub.docker.com/r/justinribeiro/chrome-headless/

Components:<br>

1. API
2. Capture.

both backend components can be in one instance

3. Screenshotctl (command line tool to interact with backend). at the moment supports only request for creation screenshots


backend part can be run in three modes:

1. api - application serves http request for view and request screenshot but will not do screenshot. expected that there is another instance or instances in capture mode connected to the same nats cluster.
2. capture - application subscribes requests for screenshots. process screenshot making and store result to mongo gridfs.
3. standalone - both components starts in the same instance.

Scaling notes:
  both components can be up simultaneously in any number of replicas.
  communication between api and capture go through nats message queue which can be scaled pretty easily
  
  in case of scaling api service. api instances should be behind some load balancer like Nginx or Caddy server.
  
  storage scaling depends on mongodb scaling features like replication and sharding
  
  
usage:<br>
  api:<br>
     
     ./screenshot --address=:9000 --queue=nats://localhost:4222 --database=mongodb://localhost:27017 --mode=api
    
  capture: <br>
      
      ./screenshot --queue=nats://localhost:4222 --database=mongodb://localhost:27017 --chrome=http://localhost:9222 --mode=capture
      
  standlone:<br>
  
       ./screenshot --address=:9000 --queue=nats://localhost:4222 --database=mongodb://localhost:27017 --chrome=http://localhost:9222 --mode=standalone    
          

  screenshotctl: <br>
       
       ./screenshotctl --backend=http://localhost:9000 --urls="http://google.com;http://facebook.com"
       

to try it just run <br>
        
        docker-compose up 
        
in root of repo. Expected that docker-compose installed on machine (https://docs.docker.com/compose/install/)   

above command will up server side with dependencies

to build client (expected golang installed (go1.13)): <br>
    
       go mod download  && go build -o screenshot  ./screenshotctl/cmd/screenshotctl 
       
   client usage:<br>
      
      ./screenshot --backend=http://localhost:9000 --urls="http://google.com;http://facebook.com"
      ./screenshot --backend=http://localhost:9000 -f={path to file with urls}
      export SCREENSHOT_BACKEND=http://localhost:9000 && ./screenshot -f={path to file with urls}
      
      
testing: repo contains codeship files, so to test can be executed with required dependencies via jet cli  https://documentation.codeship.com/pro/jet-cli/installation/

      jet-cli steps    
      