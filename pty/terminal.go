package pty

import (
    "golang.org/x/crypto/ssh"
    "strconv"
    "fmt"
)

type Terminal struct {
    prompt []byte
    line []byte
    channel ssh.Channel
}

func NewTerminal(channel ssh.Channel, prompt string) *Terminal {
    return &Terminal{
        prompt: []byte(prompt),
        channel: channel,
    }
}

func (t *Terminal) Run() {
    t.WritePrompt()
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

        char := reader[0]
        switch char {
        case 127:
            reout := []byte{27, '['}
            reout = strconv.AppendInt(reout, int64(len(t.line)), 10)
            reout = append(reout, 'D', 27, '[', 'K')
            t.line = t.line[:len(t.line)-1]
            reout = append(reout, t.line...)
            t.channel.Write(reout)
        case 13:
            outs := []byte{'\r', '\n'}
            if t.line != nil {
            outs = append(outs, t.line...)
            outs = append(outs, '\r', '\n')
            }

            t.channel.Write(outs)
            t.line = nil

            t.WritePrompt()
        default:
            t.channel.Write(reader)
            t.line = append(t.line, char)
            reader = make([]byte, 1, 1)
        }
    }
}

func (t *Terminal) WritePrompt() error {
    _, err := t.channel.Write(t.prompt)
    return err
}
