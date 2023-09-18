package dblite

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// TODO connect with muliple clients

type Notify struct {
	message     string
	messageType int
	err         bool
}

var Channel = make(chan Notify, 100)

// DemonNet listens incoming queries form ws & send result
func Resever(w http.ResponseWriter, r *http.Request) {

	var upgrader = websocket.Upgrader{} // default options

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		eLog.Print("when upgrade ", err)
		return
	}
	defer c.Close()

	var note Notify
	var message []byte

	for {
		if note.err {
			return
		}
		note.messageType, message, err = c.ReadMessage()
		if err != nil {
			iLog.Println("ReadMessage ", err)
			note.err = true
			Channel <- note
			return
		}

		// Hande all of Queries
		note.message = HandleQueries(string(message))

		Channel <- note

	}
}

// DemonNet listens incoming queries form ws & send result
func Sender(w http.ResponseWriter, r *http.Request) {

	var upgrader = websocket.Upgrader{} // default options

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)

		return
	}
	defer c.Close()

	var note Notify

	for {
		note = <-Channel
		if note.err {
			return
		}
		// send result to client
		err = c.WriteMessage(note.messageType, []byte(note.message))
		if err != nil {
			fmt.Println("ERROR! :Panic WriteMessage ", err)
			note.err = true
			Channel <- note
			return
		}
	}
}

// ws listens incoming queries form ws & send result
func Ws(w http.ResponseWriter, r *http.Request) {

	var upgrader = websocket.Upgrader{} // default options

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("ERROR! :Panic ReadMessage ", err)
			break
		}

		log.Printf("Recve: %s", message)

		// Hande all of Queries
		start := time.Now()
		result := HandleQueries(string(message)) + "\n" + time.Since(start).String()

		// send result to client
		err = c.WriteMessage(messageType, []byte(result))
		if err != nil {
			fmt.Println("ERROR! :Panic WriteMessage ", err)
			break
		}
	}

}