package server

import (
	"fmt"
	"github.com/ktatkinson/GoSsh/pty"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
)

func Start(auth *Authenticator) {
	listener, err := net.Listen("tcp4", ":22")
	fmt.Printf("Will listen at port %v\n", listener.Addr())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(fmt.Sprintf("Listener not accepting conntections: %s", err))
			os.Exit(1)
		}

		privateKey, err := getHostKey(".ssh/key.rsa")
		if err != nil {
			fmt.Println("Failed to load host key.")
			os.Exit(1)
		}

		serverConf := &ssh.ServerConfig{PublicKeyCallback: makeAuthCallback(auth)}
		serverConf.AddHostKey(privateKey)
		_, chans, reqs, err := ssh.NewServerConn(conn, serverConf)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error connecting to client: %v", err))
		}

		go ssh.DiscardRequests(reqs)
		go handleChannels("", chans)
	}
}

func handleChannels(key string, chans <-chan ssh.NewChannel) {
	for nChan := range chans {
		if nChan.ChannelType() != "session" {
			nChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		ch, reqs, err := nChan.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection because of %s", err)
		}
		ch.Write([]byte("Authenticated successfully! Welcome to GoSsh.\r\n"))
		term := pty.NewTerminal(ch, ">>> ")
		go handleRequests(reqs, ch)
		go term.Run()
	}
}

func handleRequests(ins <-chan *ssh.Request, ch ssh.Channel) {
	defer ch.Close()
	for req := range ins {
		switch req.Type {
		case "shell":
			if len(req.Payload) == 0 {
				req.Reply(true, nil)
			}
		}
		fmt.Println("Request type:", req.Type)
		fmt.Println("Payload:", string(req.Payload))
	}
}

func getHostKey(keyPath string) (ssh.Signer, error) {
	privateKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	private, err := ssh.ParsePrivateKey(privateKey)
	return private, err
}

func makeAuthCallback(auth *Authenticator) func(ssh.ConnMetadata, ssh.PublicKey) (*ssh.Permissions, error) {
	return func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		authd, err := auth.Authenticate(key)
		if err != nil {
			return nil, err
		} else if !authd {
			return nil, nil
		}

		return &ssh.Permissions{Extensions: map[string]string{"authenticated": "true"}}, nil
	}
}
