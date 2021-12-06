package main

import (
	"github.com/AodaCat/go-im/server"
	"github.com/AodaCat/go-im/util"
)

func main() {
	util.PrintThreadId("main")
	newServer := server.NewServer(
		"127.0.0.1",
		8001,
	)
	newServer.Start()
}
