package main

import (
    "fmt"
	"net"
	"bufio"
	"time"
)

func getBlock(conn net.Conn, blocks chan int) {
	for _ = range blocks {
		req := "{\"jsonrpc\": \"2.0\", \"method\": \"eth_getBlockByNumber\", \"params\": [\"0x4C4B40\", false], \"id\": 100}\n"

		conn.Write([]byte(req))
		_, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error", err)
		}
	}
}


func concurrentIPC() {
	conn, err := net.Dial("unix", "/home/jrush/.local/share/io.parity.ethereum/jsonrpc.ipc")
	if err != nil {
		fmt.Println("Error", err)
	}
    blocks := make(chan int)

    // Make 250 block processors
    for i := 0; i < 250; i++ {
        go getBlock(conn, blocks)
    }

    // Send the blocks to be processed
    for i := 0; i < 100000; i++ {
        blocks <- 1
    }
}

func sequentialHTTP() {
	conn, err := net.Dial("unix", "/home/jrush/.local/share/io.parity.ethereum/jsonrpc.ipc")
	if err != nil {
		fmt.Println("Error", err)
	}
    blocks := make(chan int)

    // Only make one block processor
    for i := 0; i < 1; i++ {
        go getBlock(conn, blocks)
    }

    // Send the blocks to be processed
    for i := 0; i < 100000; i++ {
        blocks <- 1
    }
}

func main() {
	start := time.Now()
	concurrentIPC()
	elapsed := time.Since(start)
	fmt.Println("Concurrent ipc took:", elapsed)
	
	start = time.Now()
	sequentialHTTP()
	elapsed = time.Since(start)
    fmt.Println("Sequential ipc took:", elapsed)
}