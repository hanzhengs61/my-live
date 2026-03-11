package ws

import (
	"log"
	"sync"
	"time"
)

// Hub 是整个WebSocket的“大脑”（消息中心）
// 主要负责：
// 1.管理所有在线客户端
// 2.接收要广播的消息
// 3.把消息推送给每一个客户端
type Hub struct {
	rooms      map[string]map[*Client]bool // 房间ID -> 该房间的所有客户端
	broadcast  chan Message                // 广播通道：任何人想发消息，就把消息丢进这个通道
	register   chan *Client                // 新客户端注册通道
	unregister chan *Client                // 客户端断开通道
	mu         sync.RWMutex                // 读写锁，防止并发读写clients map崩溃
}

// 单例Hub（整个项目只用一个）
var hub *Hub
var once sync.Once

// GetHub 获取单例Hub
func GetHub() *Hub {
	once.Do(func() {
		hub = &Hub{
			rooms:      make(map[string]map[*Client]bool),
			broadcast:  make(chan Message, 512), // 缓存512条，防止消息太快阻塞
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
	})
	return hub
}

// Start 启动消息中心Hub（在main方法中调用一次）
// 一个死循环，永远在后台运行
func (h *Hub) Start() {
	go func() {
		for {
			select {
			// 有新客户端连接
			case client := <-h.register:
				h.mu.Lock()
				if h.rooms[client.roomID] == nil {
					h.rooms[client.roomID] = make(map[*Client]bool)
				}
				h.rooms[client.roomID][client] = true
				h.mu.Unlock()
				h.broadcastOnlineCount(client.roomID)
				log.Printf("✅ 用户 %d 加入房间 [%s]，当前房间人数: %d", client.userID, client.roomID, len(h.rooms[client.roomID]))

			// 有客户端断开
			case client := <-h.unregister:
				h.mu.Lock()
				if roomClients, ok := h.rooms[client.roomID]; ok {
					if _, exist := roomClients[client]; exist {
						delete(roomClients, client)
						if len(roomClients) == 0 {
							delete(h.rooms, client.roomID)
						}
					}
				}
				h.mu.Unlock()
				h.broadcastOnlineCount(client.roomID)
				log.Printf("❌ 用户 %d 离开房间 [%s]", client.userID, client.roomID)

			// 需要广播的消息来了
			case msg := <-h.broadcast:
				// 把这条消息发送给每一个在线客户端
				h.mu.RLock()
				roomClients, ok := h.rooms[msg.RoomID]
				if !ok || len(roomClients) == 0 {
					h.mu.RUnlock()
					continue
				}

				// 拷贝一份当前房间客户端列表（快照），避免持锁太久 + 避免并发删除导致的closed channel
				clientsSnapshot := make([]*Client, 0, len(roomClients))
				for client := range roomClients {
					clientsSnapshot = append(clientsSnapshot, client)
				}
				h.mu.RUnlock()

				// 现在安全地发送（不会再碰到已关闭的channel）
				for _, client := range clientsSnapshot {
					select {
					case client.send <- msg.ToJSON():
					default:
						// 发送队列满或客户端异常，跳过（不会panic）
					}
				}
			}
		}
	}()
}

// 广播房间在线人数（万人不卡，因为只遍历一次）
func (h *Hub) broadcastOnlineCount(roomID string) {
	h.mu.RLock()
	count := 0
	if clients, ok := h.rooms[roomID]; ok {
		count = len(clients)
	}
	h.mu.RUnlock()

	if count == 0 {
		return
	}

	onlineMsg := Message{
		Type:        TypeOnline,
		RoomID:      roomID,
		OnlineCount: count,
		Timestamp:   time.Now().UnixMilli(),
	}

	h.broadcast <- onlineMsg
}
