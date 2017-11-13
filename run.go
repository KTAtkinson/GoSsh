package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:")
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

		serverConf := &ssh.ServerConfig{PublicKeyCallback: authenticate}
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
        go handleRequests(reqs, ch)
        go echoIncomingBytes(ch)
    }
}


func handleRequests(ins <-chan *ssh.Request, ch ssh.Channel) {
    defer ch.Close()
    for req := range ins {
        switch req.Type {
            case "shell" :
                if len(req.Payload) == 0 {
                    req.Reply(true, nil)
                }
        }
        fmt.Println("Request type:", req.Type)
    }
}


func echoIncomingBytes(ch ssh.Channel) {
    ch.Write([]byte{'>', '>', '>', ' '})
    reader := make([]byte, 12, 12)
    var outs []byte
    for {
        bRead, err := ch.Read(reader)
        if err != nil {
            fmt.Println("Error reading buffer.", err)
            return
        }
        if bRead > 0 {
            ch.Write(reader)
            outs = append(outs, reader[:bRead]...)
            reader = make([]byte, 12, 12)
        }
        bLast := outs[len(outs)-1]
        if bLast == 13 {
            ch.Write([]byte{'\n'})
            ch.Write(outs)
            ch.Write([]byte{'\r', '\n', '>', '>', '>', ' '})
            outs = nil
        }
    }
}
