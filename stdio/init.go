package stdio

import (
	"bytes"
	"context"
	"io"
	"os"
)

type stdinWrapper struct {
	bufData *bytes.Buffer
	d       chan []byte
	stop    chan struct{}
}

func (s *stdinWrapper) loop() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-s.stop:
			return
		default:
			n, _ := os.Stdin.Read(buf)
			if n > 0 {
				s.d <- buf[:n]
			}
		}
	}
}

func (s *stdinWrapper) Read(p []byte) (n int, err error) {
	return s.ReadWithContext(p, nil)
}

func (s *stdinWrapper) ReadWithContext(p []byte, ctx context.Context) (n int, err error) {
	var (
		data    []byte
		dataLen int

		l int
	)

	if s.bufData.Len() > 0 {
		data = s.bufData.Bytes()
		s.bufData.Reset()
		goto Read
	}
	if ctx != nil {
		select {
		case data = <-s.d:
			// data = d
		case <-ctx.Done():
			return 0, io.EOF
		}
	} else {
		data = <-s.d
	}
Read:
	dataLen = len(data)
	l = copy(p, data)
	if l < dataLen {
		s.bufData.Write(data[l:])
	}
	return l, nil
}

func (s *stdinWrapper) Close() error {
	go func() {
		s.stop <- struct{}{}
	}()
	// os.Stdin.Close()
	return nil
}

var Stdin = &stdinWrapper{
	d:       make(chan []byte, 8),
	stop:    make(chan struct{}, 1),
	bufData: &bytes.Buffer{},
}

func init() {
	go Stdin.loop()
}
