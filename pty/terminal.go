package pty

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"strconv"
)

type Terminal struct {
	prompt  []byte
	line    []byte
	channel TerminalReadWriter
}

type TerminalReadWriter interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
}

func NewTerminal(channel ssh.Channel, prompt string) *Terminal {
	return &Terminal{
		prompt:  []byte(prompt),
		channel: channel,
	}
}

func (t *Terminal) Run() {
	t.writePrompt()
	reader := make([]byte, 1, 1)
	for {
		bRead, err := t.channel.Read(reader)
		if err != nil {
			fmt.Println("Error reading buffer.", err)
			return
		}
		if bRead == 0 {
			fmt.Println("No characters in boffer.")
		}

        outs, isPrefix := t.processByte(reader[0])
		bWrite, err := t.channel.Write(outs)
        if err != nil {
            fmt.Printf("Error while writing bytes. %q\n", err.Error())
        } else if bWrite < 1 {
            fmt.Println("No bytes written.")
        }

        if !isPrefix {
            t.writePrompt()
        }
        reader = make([]byte, 1, 1)
    }
}

func (t *Terminal) processByte(char byte) ([]byte, bool) {
    bytes := []byte{}
    var isPrefix bool
    switch char {
    case 127:
        isPrefix = true
        if len(t.line) < 1 {
            return bytes, isPrefix
        }
        bytes = append(bytes, 27, '[')
        bytes = strconv.AppendInt(bytes, int64(len(t.line)), 10)
        bytes = append(bytes, 'D', 27, '[', 'K')
        t.line = t.line[:len(t.line)-1]
        bytes = append(bytes, t.line...)
    case 13:
        isPrefix = false
        bytes = []byte{'\r', '\n'}
        if len(t.line) > 0 {
            bytes = append(bytes, t.line...)
            bytes = append(bytes, '\r', '\n')
        }
        t.line = nil
    default:
        isPrefix = true
        bytes = []byte{char}
        t.line = append(t.line, char)
    }
    return bytes, isPrefix
}

func (t *Terminal) writePrompt() error {
	_, err := t.channel.Write(t.prompt)
	return err
}
