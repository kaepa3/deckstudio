package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connection from local network / mobile devices
	},
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Println("New WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("WebSocket client disconnected")
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

type IncomingMessage struct {
	Type     string `json:"type"`
	ButtonID string `json:"button_id"`
}

type OutgoingInitMessage struct {
	Type   string  `json:"type"`
	Config *Config `json:"config"`
}

type OutgoingAppMessage struct {
	Type    string `json:"type"`
	AppName string `json:"app_name"`
	Title   string `json:"title"`
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	hub := newHub()
	go hub.run()

	// Active window monitoring goroutine
	go func() {
		var lastApp string
		var lastTitle string
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			info, err := GetActiveAppInfo()
			if err != nil || info.AppName == "" {
				continue
			}

			if info.AppName != lastApp || info.Title != lastTitle {
				lastApp = info.AppName
				lastTitle = info.Title

				msg := OutgoingAppMessage{
					Type:    "active_app",
					AppName: info.AppName,
					Title:   info.Title,
				}
				data, err := json.Marshal(msg)
				if err == nil {
					hub.broadcast <- data
				}
			}
		}
	}()

	// Serve Static UI Files
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// WebSocket Endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		client := &Client{conn: conn, send: make(chan []byte, 256)}
		hub.register <- client

		// Send initial config to new client
		initMsg := OutgoingInitMessage{
			Type:   "init",
			Config: cfg,
		}
		if initData, err := json.Marshal(initMsg); err == nil {
			client.send <- initData
		}

		// Also send current active app
		if info, err := GetActiveAppInfo(); err == nil && info.AppName != "" {
			appMsg := OutgoingAppMessage{
				Type:    "active_app",
				AppName: info.AppName,
				Title:   info.Title,
			}
			if appData, err := json.Marshal(appMsg); err == nil {
				client.send <- appData
			}
		}

		// Writer goroutine
		go func() {
			defer func() {
				conn.Close()
			}()
			for message := range client.send {
				conn.WriteMessage(websocket.TextMessage, message)
			}
		}()

		// Reader loop
		defer func() {
			hub.unregister <- client
			conn.Close()
		}()

		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var inMsg IncomingMessage
			if err := json.Unmarshal(msgBytes, &inMsg); err != nil {
				log.Printf("Failed to unmarshal WS message: %v", err)
				continue
			}

			if inMsg.Type == "trigger" && inMsg.ButtonID != "" {
				btn := cfg.FindButtonByID(inMsg.ButtonID)
				if btn != nil {
					log.Printf("Triggering button: %s (keys: %v)", btn.ID, btn.Keys)
					go func(b *ButtonConfig) {
						if err := ExecuteButtonAction(b); err != nil {
							log.Printf("Error executing key action: %v", err)
						}
					}(btn)
				} else {
					log.Printf("Button ID not found in config: %s", inMsg.ButtonID)
				}
			}
		}
	})

	localIP := getLocalIP()
	port := cfg.Server.Port
	log.Println("==================================================")
	log.Printf(" 🎛️  DeckStudio Server is running!")
	log.Printf(" 🌐  Local Access:  http://localhost:%d", port)
	log.Printf(" 📱  Mobile Access: http://%s:%d", localIP, port)
	log.Println("==================================================")

	addr := fmt.Sprintf(":%d", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
