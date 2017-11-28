package server

import (
	"bufio"
	"bytes"
	"errors"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
)

type Authenticator struct {
	authdKeys *os.File
}

func NewAuthenticator(authdKeysPath string) (*Authenticator, error) {
	file, err := os.Create(authdKeysPath)
	if err != nil {
		return nil, err
	}
	auth := &Authenticator{
		authdKeys: file,
	}
	return auth, nil
}

func (a *Authenticator) AddAuthdKey(path string) error {
	keyBuf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(keyBuf)
	if err != nil {
		return err
	}
	marshaledKey := ssh.MarshalAuthorizedKey(publicKey)
	a.authdKeys.Seek(0, 2)
	_, err = a.authdKeys.Write(append([]byte{'\n'}, marshaledKey...))
	if err != nil {
		return err
	}
	return nil
}

func (a *Authenticator) Authenticate(key ssh.PublicKey) (bool, error) {
	keyContent := ssh.MarshalAuthorizedKey(key)
	keyContent = bytes.TrimRight(keyContent, "\n")

	a.authdKeys.Seek(0, 0)
	keyReader := bufio.NewReader(a.authdKeys)
	for {
		curKey, err := keyReader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return false, errors.New("Failed to load keys from server")
		}
		if curKey != nil {
			curKey = bytes.TrimRight(curKey, "\n")
		}

		if bytes.Equal(keyContent, curKey) {
			return true, nil
		}
		if err == io.EOF {
			break
		}
	}
	return false, nil
	// return &ssh.Permissions{Extensions: map[string]string{"authenticated": "true"}}, nil
}
