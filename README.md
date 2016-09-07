# Switchboard

Switchboard is a message broker for the web. 

Start up the server and POST a message via http or connect via web socket and send a message that way.  The recipient of your message will get it immediately if they are connected via web socket.  Otherwise they can just poll using a GET http request to retrieve all of their messages.

This is a simple way for you to communicate with computers inside your home network when you are out of the house.