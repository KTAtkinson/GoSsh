package main

import (
	"flag"
	"fmt"
	"github.com/ktatkinson/GoSsh/server"
	"os"
)

func main() {
	var ip string
	var port int
	flag.StringVar(&ip, "ip", "0.0.0.0", "IP address at which the server should run.")
	flag.IntVar(&port, "port", 22, "Integer port number at which the servers should run")
	flag.Parse()

	action := flag.Arg(0)
	fmt.Println(fmt.Printf("ACTION: %v", action))
	if action == "" {
		action = "start"
	}

	auth, err := server.NewAuthenticator("/.ssh/authorized_keys")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    srvr, err := server.New(ip, port, "/.ssh/key.rsa", auth)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	switch action {
	case "start":
		srvr.Start()
	case "add-key":
		fmt.Println(flag.Arg(0))
		err = srvr.AddAuthdKey(flag.Arg(1))
		if err != nil {
			fmt.Printf("Error while loading new key. %s", err)
		}
	}
}
