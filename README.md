# my-live
my-live是一个流媒体平台，利用流媒体技术实现视频和音频内容的实时采集、传输和分发。

# WebSocket 流转图
```
浏览器 (test-ws.html)
     ↓ (WebSocket连接 /ws)
Handler (handler.go)
     ↓ 创建
Client 对象 (client.go)
     ├── readPump()  ← 持续读浏览器发来的消息（join、chat）
     └── writePump() ← 持续写消息给浏览器 + 每30秒发心跳ping

     ↓ 重要操作都交给
Hub（单例，大脑） (hub.go)
     ├── rooms map: 房间ID → 该房间的所有Client
     ├── register 通道：有人要加入房间
     ├── unregister通道：有人离开
     └── broadcast通道：有人发弹幕/礼物，要广播

message.go 只是定义消息格式（JSON结构体）
main.go 只负责启动服务器和Hub
```