package main

import (
    "fmt"
	"net"
	"bufio"
	"time"
)

func getBlock(conn net.Conn) {
	req := "{\"jsonrpc\": \"2.0\", \"method\": \"eth_getBlockByNumber\", \"params\": [\"0x4C4B40\", false], \"id\": 100}\n"

	conn.Write([]byte(req))
	temp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error", err)
	}
	print(temp)
}

func main() {
	start := time.Now()

	conn, err := net.Dial("unix", "/home/jrush/.local/share/io.parity.ethereum/jsonrpc.ipc")
	if err != nil {
		fmt.Println("Error", err)
	}

	for i := 0; i < 100000; i++ {
		getBlock(conn)
	}
	elapsed := time.Since(start)
    fmt.Println("Singular ipc took:", elapsed)

}