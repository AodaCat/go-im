package server

import (
	"net"
	"strings"

	"github.com/AodaCat/go-im/util"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	server *Server
	conn   net.Conn
}

func NewUser(server *Server, conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		server: server,
		conn:   conn,
	}

	go user.ListenMsg()
	return user
}

func (receiver *User) ListenMsg() {
	for {
		util.PrintThreadId("user.ListenMsg")
		if msg, ok := <-receiver.C; ok {
			receiver.SendMsg(msg)
		} else {
			break
		}
	}
}

func (receiver *User) SendMsg(msg string) {
	receiver.conn.Write([]byte(msg + "\r\n"))
}

func (receiver *User) Online() {
	receiver.server.mapLock.Lock()
	receiver.server.OnlineMap[receiver.Name] = receiver
	receiver.server.mapLock.Unlock()
	receiver.server.SysBroadCast(receiver, "online!")
}

func (receiver *User) Offline() {
	receiver.server.mapLock.Lock()
	delete(receiver.server.OnlineMap, receiver.Name)
	receiver.server.mapLock.Unlock()
	receiver.server.SysBroadCast(receiver, "offline!")
}

func (receiver *User) DoMessage(msg string) {
	if msg == "who" {
		receiver.server.mapLock.Lock()
		for _, user := range receiver.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":online!"
			receiver.SendMsg(onlineMsg)
		}
		receiver.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := receiver.server.OnlineMap[newName]
		if ok {
			receiver.SendMsg("this name already exists!")
			return
		}
		receiver.server.mapLock.Lock()
		delete(receiver.server.OnlineMap, receiver.Name)
		receiver.server.OnlineMap[newName] = receiver
		receiver.server.mapLock.Unlock()
		receiver.Name = newName
		receiver.SendMsg("your name changed successfully!")
	} else if len(msg) > 3 && msg[:3] == "to|" {
		ss := strings.Split(msg, "|")
		toName := ss[1]
		if len(toName) < 1 {
			receiver.SendMsg("message error! eg: to|name|content")
			return
		}
		toUser, ok := receiver.server.OnlineMap[toName]
		if !ok {
			receiver.SendMsg("this user dose not exist!")
			return
		}
		content := ss[2]
		if len(content) < 1 {
			receiver.SendMsg("message error! content cannot be empty")
			return
		}
		toMsg := "[" + receiver.Name + "]: " + content
		toUser.C <- toMsg
	} else {
		receiver.server.SysBroadCast(receiver, msg)
	}
}
