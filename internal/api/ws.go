package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WSManager 管理所有WebSocket客户端
type WSManager struct {
	clients    map[*websocket.Conn]bool // 客户端连接映射
	broadcast  chan []byte              // 广播消息通道
	register   chan *websocket.Conn     // 注册客户端通道
	unregister chan *websocket.Conn     // 注销客户端通道
}

// NewWSManager 创建新的WebSocket管理器实例
func NewWSManager() *WSManager {
	return &WSManager{
		clients:    make(map[*websocket.Conn]bool), // 初始化客户端映射
		broadcast:  make(chan []byte),              // 初始化广播通道
		register:   make(chan *websocket.Conn),     // 初始化注册通道
		unregister: make(chan *websocket.Conn),     // 初始化注销通道
	}
}

// Run 启动WebSocket管理循环
func (m *WSManager) Run() {
	for {
		select {
		// 处理新客户端注册
		case conn := <-m.register:
			m.clients[conn] = true
			log.Println("New WS client connected")
			
		// 处理客户端注销
		case conn := <-m.unregister:
			if _, ok := m.clients[conn]; ok {
				delete(m.clients, conn) // 从客户端映射中删除
				conn.Close()            // 关闭连接
			}
			
		// 处理广播消息
		case msg := <-m.broadcast:
			// 向所有客户端发送消息
			for conn := range m.clients {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					delete(m.clients, conn) // 删除发送失败的客户端
					conn.Close()            // 关闭连接
				}
			}
		}
	}
}

// ServeWS HTTP处理函数，用于升级WebSocket连接
// w: HTTP响应写入器
// r: HTTP请求
func (m *WSManager) ServeWS(w http.ResponseWriter, r *http.Request) {
	// WebSocket升级器，允许所有来源
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	
	// 将HTTP连接升级为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS upgrade error:", err)
		return
	}
	
	// 将新连接注册到管理器
	m.register <- conn
}