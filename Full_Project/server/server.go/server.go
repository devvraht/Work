package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ViewerClient struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
}

var (
	viewers     = make(map[*ViewerClient]bool)
	viewersLock sync.Mutex
	upgrader    = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

func handleStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	log.Println("Drone connected:", conn.RemoteAddr())

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Drone disconnected")
			break
		}

		// Broadcast frame to all viewers
		viewersLock.Lock()
		for v := range viewers {
			go func(c *ViewerClient, m []byte) {
				c.Mutex.Lock()
				defer c.Mutex.Unlock()
				c.Conn.WriteMessage(websocket.TextMessage, m)
			}(v, msg)
		}
		viewersLock.Unlock()
	}
}

func handleView(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}

	client := &ViewerClient{Conn: conn}
	viewersLock.Lock()
	viewers[client] = true
	viewersLock.Unlock()
	log.Println("Viewer connected:", conn.RemoteAddr())

	defer func() {
		viewersLock.Lock()
		delete(viewers, client)
		viewersLock.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Println("Viewer disconnected")
			break
		}
	}
}

func main() {
	http.HandleFunc("/stream", handleStream)
	http.HandleFunc("/view", handleView)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
