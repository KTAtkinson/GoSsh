package main

import (
	"flag"
	"fmt"
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

	switch action {
	case "start":
		start()
	case "add-key":
		fmt.Println(flag.Arg(0))
		err := addAuthedKey(flag.Arg(1))
		if err != nil {
			fmt.Printf("Error while loading new key. %s", err)
		}
	}
}
