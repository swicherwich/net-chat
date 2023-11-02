package main

import (
	"log"
	"net"
)

const Port = ":8080"

type PierEvent int

const (
	PierConnected = iota
	PierSendText
	PierDisconnected
)

type PierMessage struct {
	Conn  net.Conn
	Event PierEvent
	Data  []byte
}

func handleConnection(conn net.Conn, out chan PierMessage) {
	conn.Write([]byte(`
Hello fellow, here what you can do 
:send to send a message
:quit to disconnect
`))

	b := make([]byte, 64)
	for {
		conn.Read(b)
		data := b[6:]
		command := b[:5]
		switch string(command) {
		case ":send":
			pier := PierMessage{
				Conn:  conn,
				Event: PierSendText,
				Data:  data,
			}
			out <- pier
		case ":quit":
			pier := PierMessage{
				Conn:  conn,
				Event: PierDisconnected,
			}
			out <- pier
		default:
			conn.Write([]byte("Unknown command\n"))
		}
	}
}

func server(out chan PierMessage) {
	piers := map[string]net.Conn{}

	for {
		pier := <-out
		pierIP := pier.Conn.RemoteAddr().String()
		switch pier.Event {
		case PierConnected:
			piers[pierIP] = pier.Conn
			log.Printf("Pier %s connected\n", pierIP)
		case PierSendText:
			log.Printf("Pier send message: %s\n", string(pier.Data))
			go broadcast(pier.Data, piers, pier.Conn)
		case PierDisconnected:
			delete(piers, pierIP)
			log.Printf("Pier %s disconnected\n", pierIP)
		}
	}
}

func broadcast(data []byte, piers map[string]net.Conn, except ...net.Conn) {
	for k, v := range piers {
		for _, c := range except {
			if k != c.RemoteAddr().String() {
				v.Write(data)
			}
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatalf("Cannot listen on %s port", Port)
	}

	out := make(chan PierMessage)
	go server(out)

	for {
		conn, _ := ln.Accept()

		pier := PierMessage{
			Conn:  conn,
			Event: PierConnected,
		}
		out <- pier
		go handleConnection(conn, out)
	}
}
