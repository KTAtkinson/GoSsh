package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
)

func addAuthedKey(path string) error {
	authWriter, err := os.Create(".ssh/authorized_keys")
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(authWriter)

	newKey, err := os.Open(path)
	if err != nil {
		return err
	}
	keyBuf := make([]byte, 4096)
	read, err := newKey.Read(keyBuf)
	if read == 0 {
		return errors.New("No key in key file.")
	}
	if err != nil {
		return err
	}
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(keyBuf)
	marshaledKey := ssh.MarshalAuthorizedKey(publicKey)
	writer.Write(marshaledKey)
	writer.Flush()
	return err
}

func getHostKey(keyPath string) (ssh.Signer, error) {
	privateBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	return private, err
}

func authenticate(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	fmt.Print("Authorizing keys...")
	authedKeys, err := os.Open(".ssh/authorized_keys")
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	keyReader := bufio.NewReader(authedKeys)
	keyContent := ssh.MarshalAuthorizedKey(key)
	keyContent = bytes.TrimRight(keyContent, "\n")
	fmt.Printf("Client %s offered key:\n%s\n", conn.SessionID(), keyContent)

	for {
		var curKey []byte
		var err error
		for err == nil {
			var prefix []byte
			var isPrefix bool
			prefix, isPrefix, err = keyReader.ReadLine()
			curKey = append(curKey, prefix...)

			if !isPrefix {
				break
			}
		}

		if err == io.EOF {
			return nil, errors.New("Failed to authenticate")
		} else if err != nil {
			return nil, errors.New("Failed to load keys from server")
		}

		fmt.Printf("Found key:\n%s\n", curKey)
		if bytes.Equal(keyContent, curKey) {
			break
		}
	}
	return &ssh.Permissions{Extensions: map[string]string{"authenticated": "true"}}, nil
}
