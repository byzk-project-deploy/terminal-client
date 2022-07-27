package cmdmodel

import (
	"fmt"
	transport_stream "github.com/go-base-lib/transport-stream"

	"github.com/byzk-project-deploy/terminal-client/server"
)

type RangeCallback func(name string, stream *transport_stream.Stream, info *ServerConn) error
type AsyncRangeCallback func(name string, stream *transport_stream.Stream, info *ServerConn) error

type PrintError interface {
	PrintError(err error)
}

type ServerList struct {
	m map[string]*ServerConn
}

func (s *ServerList) Len() int {
	return len(s.m)
}

func (s *ServerList) AsyncRangeWithContextWrapper(ctx PrintError, fn RangeCallback) {
	panic("not implement me")
}

func (s *ServerList) RangeWithContextWrapper(ctx PrintError, fn RangeCallback) error {
	for k := range s.m {
		info := s.m[k]
		stream, err := info.ConnToStream()
		if err != nil {
			if ctx != nil {
				ctx.PrintError(fmt.Errorf("连接服务器[%s]失败: %s", k, err.Error()))
			}
			continue
		}

		return fn(k, stream, info)
	}
	return nil
}

func (s *ServerList) Del(name string) {
	delete(s.m, name)
}

func (s *ServerList) Add(name string, r *ServerConn) {
	s.m[name] = r
}

type ServerConn struct {
	*server.Info
}

func newServerList() *ServerList {
	return &ServerList{
		m: make(map[string]*ServerConn, 8),
	}
}
