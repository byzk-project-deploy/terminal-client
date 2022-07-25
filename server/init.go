package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type ServerInfo struct {
	name    string
	network string
	address string
	conn    net.Conn
	*bufio.ReadWriter
}

func (s *ServerInfo) IP() string {
	addrStr := s.address
	i := strings.LastIndexByte(addrStr, ':')
	if i == -1 {
		return addrStr
	}
	return addrStr[:i]
}

func (s *ServerInfo) ConnToRW() (*bufio.ReadWriter, error) {
	if s.conn != nil {
		s.conn.Close()
	}

	conn, err := net.Dial(s.network, s.address)
	if err != nil {
		return nil, fmt.Errorf("打开服务[%s]失败, 错误信息: %s", s.name, err.Error())
	}
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	s.conn = conn
	s.ReadWriter = rw
	return s.ReadWriter, nil
}

var (
	unixFilePath = filepath.Join(os.TempDir(), ".bypt.socket")

	// serverMap = make(map[string]*ServerInfo, 8)
)

func NewUnixServerInfo() *ServerInfo {
	return NewServerInfo("unix")
}

func NewServerInfo(name string) *ServerInfo {
	// if serverInfo, ok := serverMap[name]; ok {
	// 	serverInfo.conn.Close()
	// }
	network := "tcp"
	address := name
	if name == "unix" {
		network = "unix"
		address = unixFilePath
	}

	return &ServerInfo{
		name:    name,
		network: network,
		address: address,
	}
}
