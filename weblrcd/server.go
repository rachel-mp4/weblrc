package main

import (
	"fmt"
	"log"
	events "weblrc"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client is a model for a client's connection, and their evtChannel, the queue of LRCEvents that have yet to be written to the connection
type Client struct {
	conn    *websocket.Conn
	evtChan chan events.LRCEvent
}

// Evt is a model for an lrc event from a specific client
type Evt struct {
	client *Client
	evt    events.LRCEvent
}

var (
	clients      = make(map[*Client]bool)
	clientToID   = make(map[*Client]uint32)
	lastID       = uint32(0)
	eventChannel = make(chan Evt, 100)
	clientsMu    sync.Mutex
	prod         bool = false
	wm = append([]byte{byte(events.EventPing)}, []byte("Welcome To The Beginning Of The Rest Of Your Life")...)

)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}
	defer conn.Close()
	client := &Client{conn: conn, evtChan: make(chan events.LRCEvent)}
	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()


	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); clientWriter(client) }()
	go func() { defer wg.Done(); listenToClient(client) }()
	client.evtChan <- wm
	wg.Wait()

	clientsMu.Lock()
	delete(clients, client)
	close(client.evtChan)
	clientsMu.Unlock()
	conn.Close()
	logDebug("Closed connection")
}


func main() {
	go broadcaster()
	wm, _ = events.GenServerEvent(wm, 0)
	http.HandleFunc("/ws", handler)
	log.Fatal(http.ListenAndServe(":927", nil))
}

// listenToClient polls the clients connection and then sends any daya it recieves to the degunker.
// If the connection closes, it closes readChan, which causes degunker to close quit
func listenToClient(client *Client) {
	for {
		_, evt, err := client.conn.ReadMessage()
		if err != nil {
			return
		}
		logDebug(fmt.Sprintf("read %x", evt))
		eventChannel <- Evt{client, evt[1:]}
	}
}

// clientWriter takes an event from the clients event channel, and writes it to the tcp connection.
// If the degunker runs into an error, or if the client's eventChannel closes, then this returns
func clientWriter(client *Client) {
	for {
		evt, ok := <-client.evtChan
		if !ok {
			return
		}
		client.conn.WriteMessage(websocket.BinaryMessage, evt)
	}
}

// broadcaster takes an event from the events channel, and broadcasts it to all the connected clients individual event channels
func broadcaster() {
	for evt := range eventChannel {
		logDebug(fmt.Sprintf("recieved %x from %x", evt.evt, evt.client))
		id := clientToID[evt.client]
		if id == 0 {
			if !(events.IsInit(evt.evt) || events.IsPing(evt.evt)) {
				logDebug(fmt.Sprintf("skipped %x", evt.evt))
				continue
			}
			clientToID[evt.client] = lastID + 1
			lastID += 1
			id = lastID
		}
		if events.IsPing(evt.evt) {
			evt.client.evtChan <- events.ServerPong
			continue
		}
		if events.IsPub(evt.evt) {
			clientToID[evt.client] = 0
		}
		bevt, eevt := events.GenServerEvent(evt.evt, id)
		logDebug("success")

		clientsMu.Lock()
		for client := range clients {
			evtToSend := bevt
			if client == evt.client {
				evtToSend = eevt
			}
			select {
			case client.evtChan <- evtToSend:
				logDebug(fmt.Sprintf("b %x", bevt))
			default:
				logDebug("k")
				err := client.conn.Close()
				if err != nil {
					delete(clients, client)
				}
			}
		}
		clientsMu.Unlock()
	}
}

// logDebug debugs unless in production
func logDebug(s string) {
	if !prod {
		log.Println(s)
	}
}
