package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/AodaCat/go-im/util"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	SysMsg    chan string
}

func NewServer(ip string, port int) *Server {

	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		SysMsg:    make(chan string),
	}
	return server
}

func (receiver *Server) Start() {

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", receiver.Ip, receiver.Port))
	if err != nil {
		fmt.Println("net.Listen err")
		return
	}
	defer listen.Close()
	go receiver.ListenMsg()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen.Accept err")
			return
		}
		go receiver.Handler(conn)
	}
}

func (receiver *Server) Handler(conn net.Conn) {

	util.PrintThreadId("server.Handler")
	user := NewUser(receiver, conn)
	user.Online()
	isLive := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		for {
			util.PrintThreadId("receiver.SysBroadCast")
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
			}
			if err != nil {
				if err == io.EOF {
					fmt.Println("conn closed!")
				} else if _, ok:=err.(*net.OpError); ok {
					fmt.Println("conn.Read err: conn closed!")
				} else {
					fmt.Println("conn.Read err:", err)
				}
				return
			}
			msg := string(buf[:n-1])
			user.DoMessage(msg)
			isLive <- true
		}
	}()
	for {
		select {
		case <-isLive:

		case <-time.After(300 * time.Second):
			user.SendMsg("you are forced offline!")
			close(user.C)
			err := conn.Close()
			if err != nil {
				fmt.Println("conn.Close err:", err)
			}
			return
		}
	}
}

func (receiver *Server) SysBroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg
	receiver.SysMsg <- sendMsg
}

func (receiver *Server) ListenMsg() {
	for {
		util.PrintThreadId("service.ListenMsg")
		if msg, ok := <-receiver.SysMsg; ok {
			receiver.mapLock.Lock()
			for _, user := range receiver.OnlineMap {
				user.C <- msg
			}
			receiver.mapLock.Unlock()
		} else {
			break
		}
	}
}
