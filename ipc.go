package main

import (
    "fmt"
	"net"
	"bufio"
	"time"
)

func getBlock(conn net.Conn, blocks chan int) {
	for block := range blocks {
		req := "{\"jsonrpc\": \"2.0\", \"method\": \"eth_getBlockByNumber\", \"params\": [\"0x4C4B40\", false], \"id\": 100}\n"

		conn.Write([]byte(req))
		temp, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error", err)
		}
		print(temp)
	}
}

func main() {
	start := time.Now()

	blocks := make(chan int)

	conn, err := net.Dial("unix", "/home/jrush/.local/share/io.parity.ethereum/jsonrpc.ipc")
	if err != nil {
		fmt.Println("Error", err)
	}

	for i := 0; i < 100; i++ {
		go getBlock(conn, blocks)
	}

	for i := 0; i < 100000; i++ {
		blocks <- 1 // just to make it run
	}

	elapsed := time.Since(start)
    fmt.Println("Singular ipc took:", elapsed)

}