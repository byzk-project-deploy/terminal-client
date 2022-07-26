package server

import (
	"bufio"
	"fmt"
	transport_stream "github.com/go-base-lib/transport-stream"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Info struct {
	name    string
	network string
	address string
	conn    net.Conn
}

func (s *Info) IP() string {
	addrStr := s.address
	i := strings.LastIndexByte(addrStr, ':')
	if i == -1 {
		return addrStr
	}
	return addrStr[:i]
}

func (s *Info) ConnToStream() (*transport_stream.Stream, error) {
	if s.conn != nil {
		s.conn.Close()
	}

	conn, err := net.Dial(s.network, s.address)
	if err != nil {
		return nil, fmt.Errorf("打开服务[%s]失败, 错误信息: %s", s.name, err.Error())
	}
	s.conn = conn
	rw := bufio.NewReadWriter(bufio.NewReader(s.conn), bufio.NewWriter(s.conn))
	return transport_stream.NewStream(rw), nil
}

var (
	unixFilePath = filepath.Join(os.TempDir(), ".bypt.socket")

	// serverMap = make(map[string]*ServerInfo, 8)
)

func NewUnixServerInfo() *Info {
	return NewServerInfo("unix")
}

func NewServerInfo(name string) *Info {
	network := "tcp"
	address := name
	if name == "unix" {
		network = "unix"
		address = unixFilePath
	}

	return &Info{
		name:    name,
		network: network,
		address: address,
	}
}
