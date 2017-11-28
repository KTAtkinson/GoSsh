package pty

import (
    "bytes"
    "testing"
)

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
            channel: bytes.NewBuffer([]byte{}),
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
