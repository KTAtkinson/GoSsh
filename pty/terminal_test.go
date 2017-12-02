package pty

import (
    "bytes"
    "testing"
    "fmt"
)

type testChannel struct {
    reader *bytes.Buffer
    writer *bytes.Buffer
}

func (c *testChannel) WriteReader(p []byte) (int, error) { return c.reader.Write(p) }

func (c *testChannel) Read(p []byte) (int, error) { return c.reader.Read(p) }

func (c *testChannel) Write(p []byte) (int, error) { return c.writer.Write(p) }

func (c *testChannel) ReadWriter(p []byte) (int, error) { return c.writer.Read(p) }

func TestProcessByte(t *testing.T) {
    cases := []struct {
        curLine []byte
        writeTo byte
        expectedBytes []byte
        expectedLn []byte
        isPrefix bool
    } {
        {[]byte{}, byte('h'), []byte("h"), []byte("h"), true},
        {[]byte("hello worl"), byte('d'), []byte("d"), []byte("hello world"), true},
        {[]byte("hello"), byte(127), []byte("\x1b[5D\x1b[Khell"), []byte("hell"), true},
        {[]byte(""), byte(127), []byte{}, []byte{}, true},
        {nil, byte(127), []byte{}, []byte{}, true},
        {[]byte("hello world"), byte(13), []byte("\r\nhello world\r\n"), nil, false},
        {[]byte(""), byte(13), []byte("\r\n"), nil, false},
        {nil, byte(13), []byte("\r\n"), nil, false},
    }

    for _, tcase := range cases {
        term := &Terminal{
            line: tcase.curLine,
        }
        processed_bytes, isPrefix := term.processByte(tcase.writeTo)
        if !bytes.Equal(processed_bytes, tcase.expectedBytes) {
            t.Errorf("Expected %q bytes, got %q.", tcase.expectedBytes, processed_bytes )
        }
        if !bytes.Equal(term.line, tcase.expectedLn) {
            t.Errorf("Expected terminal line to be %q, found %q.", tcase.expectedLn, term.line)
        }
        if isPrefix != tcase.isPrefix {
            t.Errorf("Expected is prefix to be %t, was %t", tcase.isPrefix, isPrefix)
        }
    }
}

func TestRun(t *testing.T) {
    tests := []struct {
        writeOut []byte
        expectedOut []byte
    } {
        {[]byte{'a'}, []byte{'a'}},
    }

    for _, tcase := range tests {
        readBuf := bytes.NewBuffer([]byte{})
        writeBuf := bytes.NewBuffer([]byte{})
        channel := &testChannel{
            reader: readBuf,
            writer: writeBuf,
        }
        term := &Terminal{
            channel: channel,
        }
        nRead, err := channel.WriteReader(tcase.writeOut)
        fmt.Printf("READER: %q\n", channel.reader)
        if err != nil {
            t.Errorf("Error while writing to read buffer.", err.Error())
        } else if nRead < 1 {
            t.Errorf("No bytes written to reader.")
        }

        term.Run()
        fmt.Printf("Writer: %q\n", channel.writer)
        wroteOut := channel.writer.Bytes()
        if !bytes.Equal(wroteOut, tcase.expectedOut) {
            t.Errorf("Expected %q to be written to buffer, found %q instead.", tcase.expectedOut, wroteOut)
        }
    }
}
