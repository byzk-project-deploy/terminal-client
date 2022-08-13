package server

import (
	"bufio"
	_ "embed"
	"fmt"
	transport_stream "github.com/go-base-lib/transport-stream"
	"github.com/tjfoc/gmsm/gmtls"
	"net"
	"os"
	"path/filepath"
)

const hostname = "unix"

//go:embed root.crt
var rootPemCert []byte

//go:embed client.crt
var clientPemCert []byte

//go:embed client.key
var clientPemKey []byte

var infoMap = make(map[string]*Info, 8)

type Info struct {
	name      string
	alias     string
	network   string
	address   string
	conn      net.Conn
	tlsConfig *gmtls.Config
}

func (s *Info) Name() string {
	return s.name
}

func (s *Info) ConnToStream() (*transport_stream.Stream, error) {
	s.Close()

	//if s.tlsConfig == nil {
	//	s.tlsConfig = GetTlsConfig()
	//}

	var err error
	s.conn, err = gmtls.Dial(s.network, s.address, s.tlsConfig)
	if err != nil {
		s.conn = nil
		return nil, fmt.Errorf("打开服务[%s]失败, 错误信息: %s", s.name, err.Error())
	}

	rw := bufio.NewReadWriter(bufio.NewReader(s.conn), bufio.NewWriter(s.conn))
	return transport_stream.NewStream(rw), nil
}

func (s *Info) Close() {
	if s.conn != nil {
		_ = s.conn.Close()
	}
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

	if info, ok := infoMap[name]; ok {
		return info
	}

	info := &Info{
		name:    name,
		network: network,
		address: address,
	}

	if info.network == "unix" {
		info.tlsConfig = GetTlsConfig()
	} else {
		_, _ = GeneratorTlsClientConfig("")
	}

	infoMap[name] = info
	return info
}
