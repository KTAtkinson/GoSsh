package server

import (
	"fmt"
	"github.com/ktatkinson/GoSsh/pty"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
    "errors"
)

type authenticator interface {
    Authenticate(ssh.PublicKey) (bool, error)
    AddAuthdKey(string) (error)
}

type Server struct {
    authenticator
    ip string
    port int

    requestsHandler func(<-chan *ssh.Request, ssh.Channel)
    hostKey ssh.Signer
}

func New(ip string, port int, hostKeyPath string, keyAuthenticator authenticator) (*Server, error) {
    key, err := getHostKey(hostKeyPath)
    if err != nil {
        return nil, err
    }
    return &Server{
        authenticator: keyAuthenticator,
        ip: ip,
        port: port,
        requestsHandler: handleRequests,
        hostKey: key,
    }, nil
}

func (s *Server) HostAddr() string  {
    return fmt.Sprintf("%s:%d", s.ip, s.port)
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp4", s.HostAddr())
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

		serverConf := &ssh.ServerConfig{PublicKeyCallback: s.AuthenticateKey}
		serverConf.AddHostKey(s.hostKey)
		_, chans, reqs, err := ssh.NewServerConn(conn, serverConf)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error connecting to client: %v", err))
		}

		go ssh.DiscardRequests(reqs)
		go s.handleChannels("", chans)
	}
}

func (s *Server) handleChannels(key string, chans <-chan ssh.NewChannel) {
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
		go s.requestsHandler(reqs, ch)
		go term.Run()
	}
}

func(s *Server) AuthenticateKey(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
    authd, err := s.Authenticate(key)
    if err != nil {
        return nil, err
    } else if !authd {
        return nil, errors.New("Failed to authenticate (public key)")
    }

    return &ssh.Permissions{Extensions: map[string]string{"authenticated": "true"}}, nil
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
