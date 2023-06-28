// implements the main MOS API listener using websockets
// implementation according to MOS Protocol version 4.0
// https://mosprotocol.com/wp-content/MOS-Protocol-Documents/MOS-Protocol-Version-4.0.pdf

package listener

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// process connections

var PORT = ":8081"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HELLO")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Client connected from ", r.Host)

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Error while upgrading socket connection: ", err)
		return
	}
}

defer ws.Close()

for {
	mt, message, err := ws.ReadMessage()
	if err != nil {
		log.println("Error while reading message from ", r.Host, ", error: ", err)
		break
	}

	// at this point we have received a message from the client
	// this is for debugging only:

	fmt.Println(message)

}
