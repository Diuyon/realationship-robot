# realationship-robot

# 启动
修改 main.go 中的 APPID & APPTOKEN。

## 目录结构
```text
.
├── README.md
├── business  实现各个事件回调函数的地方
│   ├── at_message.go  at事件的回调函数实现
│   └── define.go
├── client
│   ├── client.go  对外提供功能的 client，结合 http & wss client 实现各种管理和功能
│   ├── define.go
│   ├── handler.go  管理事件回调函数的地方, 目前只支持 AT & Reaction 两个事件的回调，等后面抽象出 APIClient，就把这个文件放到 business 目录下
│   ├── http_client.go  http client 实现
│   └── ws_client.go  wss client 实现
├── dto  定义流转数据的结构的地方
│   ├── emoji.go
│   ├── message.go
│   ├── message_create.go
│   ├── message_reaction.go
│   ├── stream.go
│   ├── user.go
│   ├── websocket_event.go
│   ├── websocket_intent.go
│   ├── websocket_opcode.go
│   └── websocket_payload.go
├── errs
│   └── err.go
├── go.mod
├── go.sum
├── log
│   └── log.go
├── main.go
├── util
│   ├── http.go
│   ├── message.go
│   ├── token.go
│   └── util.go
└── version
    └── version.go

```