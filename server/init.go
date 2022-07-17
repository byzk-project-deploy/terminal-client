package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

type ServerInfo struct {
	conn net.Conn
	*bufio.ReadWriter
}

var (
	unixFilePath = filepath.Join(os.TempDir(), ".bypt.socket")

	serverMap = make(map[string]*ServerInfo, 8)
)

// func InitUnixServer(app *grumble.App) {
// 	if _, err := OpenUnixServer(); err != nil {
// 		app.PrintError(fmt.Errorf("打开本地Unix文件失败: %s", err.Error()))
// 		app.Close()
// 		os.Exit(6)
// 	}
// }

// func RWUnixStream() *ServerInfo {
// 	s, _ := RWStream("unix")
// 	return s
// }

// func RWStream(name string) (*ServerInfo, bool) {
// 	s, ok := serverMap[name]
// 	return s, ok
// }

func OpenUnixServer() (*ServerInfo, error) {
	return OpenServer("unix")
}

func OpenServer(name string) (*ServerInfo, error) {
	if serverInfo, ok := serverMap[name]; ok {
		serverInfo.conn.Close()
	}
	network := "tcp"
	address := name
	if name == "unix" {
		network = "unix"
		address = unixFilePath
	}

	s, err := net.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("打开服务[%s]失败, 错误信息: %s", name, err.Error())
	}
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	serverInfo := &ServerInfo{
		ReadWriter: rw,
		conn:       s,
	}
	serverMap[name] = serverInfo
	return serverInfo, nil
}
