#Zach Palmer Portfolio

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

This app really is not meant to be pulled down locally and ran. If you want to checkout the code or use postman to send http requests to my service.
First download docker then pull the image for the service from zdev19/portfolio-service then build a container from the image, and map the container to port 8080
on your local machine. Then you will be able to send requests via postman to the service.

The Zypher expects the following url qeury parameters:
 *txt* - the plaintext to be encrypted
 *shft* - determines how many characters a given character in the plaintext should be shifted
 *shftcount* - determines the number of times the plaintext is iterated over and shifted
 *hshcount* - determines how many times the plaintext is hashed, hashing happens after the text has been cyphered
 *alt* - flag to alternate the shift direction, if true will alternate shifting character +x and -x
 *ignspace* - flag to either keep spaces or encrypt them
 *restricthash* - flag to keep characters within hex values