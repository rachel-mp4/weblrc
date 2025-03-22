package client

import (
	"io"
	"log"
	"net"
	"time"
	"weblrc"

	"github.com/gorilla/websocket"
)

var (
	pingChannel = make(chan struct{})
)

type LRCCommand struct {
	n   int
	buf events.LRCEvent
}

// ConnectToChannel attempts to connect to a url, and if it succeeds, it sets up a listener, chatter, and pinger, and returns the connection
func ConnectToChannel(url string, quit chan struct{}, send chan events.LRCEvent) *websocket.Conn {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:927/ws", nil)
	if err != nil {
		log.Fatal(err)
	}

	eventChan := make(chan []byte, 100)
	go chat(conn, send)
	go relayToParser(eventChan)
	go listen(conn, eventChan)
	go pinger(send)
	return conn
}

func relayToParser(eventChan chan events.LRCEvent) {
	for {
		evt, ok := <-eventChan
		if !ok {
			return
		}
		addToCmdLog(evt)
		parseCommand(evt)
	}
}



func chat(conn *websocket.Conn, send chan []byte) {
	for {
		msg, ok := <-send
		if !ok {
			return
		}
		conn.WriteMessage(websocket.BinaryMessage, msg)
	}
}

// listen listens for LRCEvents and then acts on them accordingly
func listen(conn *websocket.Conn, eventChan chan []byte) {
	for {
		_, e, err := conn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				log.Fatal("Read error:", err)
			} else {
				log.Println("Server closed")
			}
		}
		eventChan <- e
	}
}

func parseCommand(e events.LRCEvent) {
	e = e[1:]
	switch events.ParseEventType(e) {
	case events.EventPing:
		if len([]byte(e)) > 5 {
			setWelcomeMessage(string(e[5:]))
		} else {
			setWelcomeMessage("Fail")
		}
	case events.EventPong:
		go ponged()
	case events.EventInit:
		id, color, name, isFromMe := events.ParseInitEvent(e)
		initMsg(id, color, name, true, isFromMe)
	case events.EventPub:
		pubMsg(events.ParsePubEvent(e))
	case events.EventInsert:
		insertIntoMsg(events.ParseInsertEvent(e))
	case events.EventDelete:
		deleteFromMessage(events.ParseDeleteEvent(e))
	}
}

func pinger(send chan events.LRCEvent) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		ping(send)
	}
}

func ping(send chan events.LRCEvent) {
	p := make([]byte, 2)
	copy(p, events.ClientPing)
	t0 := time.Now()
	send <- p
	<-pingChannel
	t1 := time.Now()
	setPingTo(int(t1.Sub(t0).Milliseconds()))
}

func ponged() {
	pingChannel <- struct{}{}
}

// dial dials the url
func dial(url string) (net.Conn, error) {
	return net.Dial("tcp", ":927")
}

// hangUp closes the connection if it exists
func hangUp(conn *websocket.Conn) {
	if conn != nil {
		conn.Close()
	}
}

// deNagle disables Nagle's algorithm, causing the connection to send tcp packets as soon as possible at the cost of increasing overhead by not pooling events
func deNagle(conn net.Conn) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetNoDelay(true)
}
