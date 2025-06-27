package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// * 預設 cli 主機 127.0.0.1 和端口 7989
var (
	host    = flag.String("host", "127.0.0.1", "JsonDB host")
	port    = flag.String("port", "7989", "JsonDB port")
	command = flag.String("c", "", "Execute single command and exit")
)

func main() {
	flag.Parse()

	addr := net.JoinHostPort(*host, *port)
	if *command == "" {
		// * 非單次命令才顯示歡迎訊息
		fmt.Printf("Connecting to JsonDB at %s\n", addr)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", addr, err)
		fmt.Fprintf(os.Stderr, "Make sure JsonDB is running.")
		os.Exit(1)
	}
	defer conn.Close()

	// * 單次命令
	if *command != "" {
		err := exec(conn, *command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("Connected to JsonDB")
	err = cli(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func exec(conn net.Conn, cmd string) error {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	buffer := make([]byte, 4096)
	_, err := reader.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading connecting: %v", err)
	}

	_, err = writer.WriteString(cmd + "\n")
	if err != nil {
		return fmt.Errorf("error sending: %v", err)
	}
	// * 送出指令
	writer.Flush()

	n, err := reader.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	res := string(buffer[:n])
	res = strings.TrimSuffix(res, "jsondb[0]> ")
	res = strings.TrimSpace(res)
	fmt.Println(res)

	writer.WriteString("quit\n")
	writer.Flush()

	return nil
}

func cli(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	stdinReader := bufio.NewReader(os.Stdin)

	buffer := make([]byte, 1024)
	n, err := reader.Read(buffer)
	if err != nil {
		return fmt.Errorf("error reading welcome message: %v", err)
	}

	welcome := string(buffer[:n])
	fmt.Print(welcome)

	for {
		// * 等待用戶輸入
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				break
			}
			return fmt.Errorf("error reading input: %v", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if strings.EqualFold(input, "quit") || strings.EqualFold(input, "bye") || strings.EqualFold(input, "exit") {
			writer.WriteString("quit\n")
			writer.Flush()

			buffer := make([]byte, 1024)
			n, err := reader.Read(buffer)
			if err == nil {
				fmt.Print(string(buffer[:n]))
			}
			break
		}

		_, err = writer.WriteString(input + "\n")
		if err != nil {
			return fmt.Errorf("error sending command: %v", err)
		}
		writer.Flush()

		buffer := make([]byte, 4096)
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed")
				return nil
			}
			return fmt.Errorf("error reading response: %v", err)
		}

		res := string(buffer[:n])
		fmt.Print(res)
	}

	fmt.Println("Disconnected from JsonDB")
	return nil
}
