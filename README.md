#Zach Palmer Portfolio

# Description
This is my portfolio app, it is basically an online version of my resume. My idea is to have this be the landing page
for a web app that is made up of smaller projects of mine. Currently on the portfolio page, I talk about myself, my experience,
and my skills. I have a section that uses a library I built named Zypher. Zypher is a encryption tool I built that combines cyphers and 
hashing. It's main use is for hashing plaintext like passwords before they are stored in a database, I built it in hopes to make hashing 
collisions less likely and it harder for the hashes to be cracked. I also created a schedule component that lists the days I am available
and the tasks I have on my schedule. That way if a recruiter does like what they see they can see what I have going on to make it easier to 
setup a time to talk. Users can also add task requests to my schedule that will then send me a email letting me know someone wants to have a meeting
with me. The schedule uses a websocket that way my schedule is updated in realtime so as anyone adds tasks to it everyone else will be able to see the
most up to date version of my schedule. I also have a page that has a copy of my resume. I will be added more pages to the project as I go along. Each page 
will be it's own project, have it's own ec2 instance, database, etc. 


# Quicstart 
This app really is not meant to be pulled down locally and ran. If you want to checkout the code or use postman to send http requests to through service.
Follow these steps *(Note all of the following commands should be ran in the terminal in the project root dir)*: 
- ##  __*First step is to ensure you docker download on your machine.*__ 
- ## __*Next create a docker network so a Redis container and the portfolio-service container can communicate by running:*__
```
docker network create portfolio-service-network
```
- ## __*Next pull the official Redis image and start a Redis container with the image (Note in this step make sure the container is named redis-service):*__
```
docker pull redis
docker run --name redis-service -d -p 6379:6379 redis
```

- ## __*Finally, build the portfolio-service image and then run the service container:*__
```
docker build -t zrp-portfolio .
docker run -d -p 8081:8081 --name zrp-service-mc1 --network portfolio-service-network 
```

- ## __*Now you can send http requests to the service through postman with the url `http://localhost/zypher:8081?txt=""shft=""shftcount=""hshcount=""alt=""ignspace=""restricthash=""` of course with valid values for the query parameters.__*

The Zypher expects the following url qeury parameters:
 - __*txt*__ - the plaintext to be encrypted
 - __*shft*__ - determines how many characters a given character in the plaintext should be shifted
 - __*shftcount*__ - determines the number of times the plaintext is iterated over and shifted
 - __*hshcount*__ - determines how many times the plaintext is hashed, hashing happens after the text has been cyphered
 - __*alt*__ -flag to alternate the shift direction, if true will alternate shifting character +x and -x
 - __*ignspace*__ - flag to either keep spaces or encrypt them
 - __*restricthash*__ - flag to keep characters within hex values

The endpoint that gets my portfolio data is `http://localhost/about:8081` this is a http GET method so it does not require query parameters or a request body, so simply calling this endpoint in postman will return the data.

 
 
