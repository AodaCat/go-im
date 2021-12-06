package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var serverIp string
var serverPort int

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, port int) (client *Client) {
	client = &Client{
		ServerIp:   serverIp,
		ServerPort: port,
		flag:       999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, port))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return
}

func (receiver *Client) menu() bool {
	var input int
	fmt.Println("1.all")
	fmt.Println("2.one")
	fmt.Println("3.update name")
	fmt.Println("0.exit")
	fmt.Scanln(&input)
	if input < 0 && input > 3 {
		fmt.Println("input error!")
		return false
	}
	receiver.flag = input
	return true
}

func (receiver *Client) UpdateName() bool {
	fmt.Println("please input new name:")
	fmt.Scanln(&receiver.Name)
	sendMsg := "rename|" + receiver.Name + "\n"
	_, err := receiver.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (receiver *Client) PublicChat() {
	var chatMsg string
	fmt.Println("please input message:")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := receiver.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("please input message(exit):")
		fmt.Scanln(&chatMsg)
	}
}

func (receiver *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := receiver.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

func (receiver *Client) PrivateChat() {
	receiver.SelectUser()
	var toUser string
	var chatMsg string
	fmt.Println("please select user(exit):")
	fmt.Scanln(&toUser)
	for toUser != "exit" {
		fmt.Println("please input message:")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + toUser + "|" + chatMsg + "\n\n"
				_, err := receiver.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println("please input message:")
			fmt.Scanln(&chatMsg)
		}
		toUser = ""
		receiver.SelectUser()
		fmt.Println("please select user(exit):")
		fmt.Scanln(&toUser)
	}
}

func (receiver *Client) Start() {
	for receiver.flag != 0 {
		for !receiver.menu() {
		}
		switch receiver.flag {
		case 1:
			receiver.PublicChat()
		case 2:
			receiver.PrivateChat()
		case 3:
			receiver.UpdateName()
		}
	}
}

func (receiver *Client) DealResp() {
	io.Copy(os.Stdout, receiver.conn)
	//等价于此写法
	//for {
	//	buf:=make([]byte, 4096)
	//	receiver.conn.Read(buf)
	//	fmt.Println(buf)
	//}
}

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "server ip(default:127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8001, "server port(default:8001)")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("client creation error!")
		return
	}
	go client.DealResp()
	fmt.Println("client created successfully!")
	client.Start()
}
