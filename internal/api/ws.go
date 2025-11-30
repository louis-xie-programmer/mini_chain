package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WSManager 管理所有 WebSocket 客户端
type WSManager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewWSManager() *WSManager {
	return &WSManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run 启动 WebSocket 管理循环
func (m *WSManager) Run() {
	for {
		select {
		case conn := <-m.register:
			m.clients[conn] = true
			log.Println("New WS client connected")
		case conn := <-m.unregister:
			if _, ok := m.clients[conn]; ok {
				delete(m.clients, conn)
				conn.Close()
			}
		case msg := <-m.broadcast:
			for conn := range m.clients {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					delete(m.clients, conn)
					conn.Close()
				}
			}
		}
	}
}

// ServeWS HTTP 处理函数，用于升级 WebSocket
func (m *WSManager) ServeWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS upgrade error:", err)
		return
	}
	m.register <- conn
}
