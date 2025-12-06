package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	addr := flag.String("addr", "http://127.0.0.1:7777", "control API address")
	token := flag.String("token", "", "bearer token for control API")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("usage: agentctl [flags] status|config|service <start|stop|restart|status|install>")
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	switch cmd {
	case "status":
		get(*addr+"/status", *token)
	case "config":
		get(*addr+"/config", *token)
	case "service":
		if flag.NArg() < 2 {
			fmt.Println("service action required: start|stop|restart|status|install")
			os.Exit(1)
		}
		action := flag.Arg(1)
		postEmpty(fmt.Sprintf("%s/service/%s", *addr, action), *token)
	default:
		fmt.Println("unknown command", cmd)
		os.Exit(1)
	}
}

func authReq(req *http.Request, token string) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func get(url, token string) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	authReq(req, token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("request error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(strings.TrimSpace(string(body)))
	if resp.StatusCode >= 300 {
		os.Exit(1)
	}
}

func postEmpty(url, token string) {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	authReq(req, token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("request error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		fmt.Println(pretty.String())
	} else {
		fmt.Println(strings.TrimSpace(string(body)))
	}
	if resp.StatusCode >= 300 {
		os.Exit(1)
	}
}
