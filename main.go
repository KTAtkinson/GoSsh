package main

import (
	"flag"
	"fmt"
	"github.com/KTAtkinson/GoSsh/server"
	"os"
    "io/ioutil"
    "golang.org/x/crypto/ssh"
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
    hostKey := getHostKey("/.ssh/key.rsa")
	srvr := server.New(ip, port, hostKey, auth)

	switch action {
	case "start":
		err = srvr.Start()
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
	case "add-key":
		fmt.Println(flag.Arg(0))
		err = srvr.AddAuthdKey(flag.Arg(1))
		if err != nil {
			fmt.Printf("Error while loading new key. %s", err)
		}
	}
}

func getHostKey(keyPath string) (ssh.Signer) {
	privateKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		fmt.Println("Failed to load host key:", err.Error())
        os.Exit(1)
    }
	private, err := ssh.ParsePrivateKey(privateKey)
    if err != nil {
        fmt.Println("Failed to load key from file:", err.Error())
        os.Exit(1)
    }
	return private
}
