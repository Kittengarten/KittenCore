package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fumiama/terasu/http2"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "url")
		return
	}
	if !strings.HasPrefix(os.Args[1], "https://") {
		fmt.Println("ERROR: invalid url")
		return
	}
	resp, err := http2.Get(os.Args[1])
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("ERROR:", "status code:", resp.StatusCode)
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Print(string(data))
}
