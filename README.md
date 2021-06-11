# SimpleChatGo: Simple chat written in Go

A simple chat written in Go.

Project is still WIP.

SimpleChatGo is built using a concurrent TCP server and a number of TCP clients. Each connection is handled by a goroutine. Messages are stored using PostgreSQL.
Every user gets all the previously sent messages on joining.

SimpleChatGo uses [tui-go](https://github.com/marcusolsson/tui-go) for its beautiful terminal interface.

![Screenshot](example/screenshot.png)

## Installing required packages

```
go get github.com/marcusolsson/tui-go
go get github.com/lib/pq
```

