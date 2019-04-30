# websocketLab
#### Author: Mason Flint
#### Date: 29 April 2019
## Description:
Command line chat application built using golang and websockets

Built using golang and the following websocket API:
- [Source](https://github.com/gorilla/websocket)
- [Docs](https://godoc.org/github.com/gorilla/websocket)

The application consists of a chat room server and client.
They support the following operations:
- `login <handle> <pass>` - Log in to server
- `newuser <handle> <pass>` - Register new user
- `send <message>` - Send message to clients
- `logout` - Log out from server
