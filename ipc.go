package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
	"io/ioutil"
	"net"
	"bufio"
	"time"
)


type Params []interface{}

type Payload struct {
    Jsonrpc string        `json:"jsonrpc"`
    Method  string        `json:"method"`
    Params                `json:"params"`
    ID      int           `json:"id"`
}

type Result struct {
    block int
    body string
}

func getBlock(block int) string {
    // Process blocks untill the blocks channel closes
    hexBlockNum := fmt.Sprintf("0x%x", block)

    data := Payload{
        "2.0",
        "eth_getBlockByNumber",
        Params{hexBlockNum, false},
        2,
    }

    payloadBytes, err := json.Marshal(data)
    if err != nil {
        fmt.Println("Error:", err)
        return ""
    }

    body := bytes.NewReader(payloadBytes)

    req, err := http.NewRequest("POST", "http://localhost:8545", body)
    if err != nil {
        fmt.Println("Error:", err)
        return ""
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)

    if err != nil {
        fmt.Println("Error:", err)
        return ""
    }

    defer resp.Body.Close()

    body1, err := ioutil.ReadAll(resp.Body)
    return string(body1)
}

func main() {
	start := time.Now()

	conn, err := net.Dial("unix", "/home/jrush/.local/share/io.parity.ethereum/jsonrpc.ipc")
	if err != nil {
		fmt.Println("Error", err)
	}

	for i := 0; i < 10000; i++ {
		req := "{\"jsonrpc\": \"2.0\", \"method\": \"eth_getBlockByNumber\", \"params\": [\"0x4C4B40\", false], \"id\": 100}\n"

		conn.Write([]byte(req))
		_, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error", err)
		}
	}

	elapsed := time.Since(start)
	fmt.Println("Pipe time:", elapsed)

	start1 := time.Now()
	
	for i := 0; i < 10000; i++ {
		getBlock(5000000)
	}
	
	elapsed1 := time.Since(start1)
	fmt.Println("Curl took time:", elapsed1)
	
}