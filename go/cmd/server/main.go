package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"go-jsondb/internal/command"
	"go-jsondb/internal/server"
)

func main() {
	fmt.Println("JsonDB starting on 127.0.0.1:7989")

	server, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	parser := command.NewParser()

	listener, err := net.Listen("tcp", "127.0.0.1:7989")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Println("JsonDB ready to connect")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error connect: %v", err)
			continue
		}

		go newConn(conn, server, parser)
	}
}

func newConn(conn net.Conn, jsondbServer *server.Server, parser *command.Parser) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	fmt.Printf("New client connected: %s\n", addr)

	session := jsondbServer.NewClient()
	reader := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	writer.WriteString("JsonDB Go 0.1.0\n")
	writer.WriteString("Type 'help' for available commands or 'quit' to exit\n")
	writer.WriteString("jsondb[0]> ")
	writer.Flush()

	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())

		// * 無內容，直接顯示提示符
		if line == "" {
			writer.WriteString(fmt.Sprintf("jsondb[%d]> ", session.GetDB()))
			writer.Flush()
			continue
		}

		if strings.EqualFold(line, "quit") {
			writer.WriteString("Bye\n")
			writer.Flush()
			break
		}

		var res string
		cmd, err := parser.Parse(line)
		if err != nil {
			res = fmt.Sprintf("Error: %v", err)
		} else {
			res = session.Exec(cmd)
		}

		writer.WriteString(res + "\n")
		writer.WriteString(fmt.Sprintf("jsondb[%d]> ", session.GetDB()))
		writer.Flush()
	}

	if err := reader.Err(); err != nil {
		log.Printf("Error reading from client %s: %v", addr, err)
	}

	fmt.Printf("Client disconnected: %s\n", addr)
}
