package cmdmodel

import (
	"fmt"

	"github.com/byzk-project-deploy/terminal-client/server"
)

type PrintError interface {
	PrintError(err error)
}

type ServerList struct {
	m map[string]*ServerResult
}

func (s *ServerList) Len() int {
	return len(s.m)
}

func (s *ServerList) RangeWWithContextWrapper(ctx PrintError, fn func(name string, info *ServerResult) error) error {
	for k := range s.m {
		info := s.m[k]
		if info.Err != nil {
			ctx.PrintError(fmt.Errorf("连接服务器[%s]失败: %s", k, info.Err.Error()))
			continue
		}

		return fn(k, info)
	}
	return nil
}

func (s *ServerList) Del(name string) {
	delete(s.m, name)
}

func (s *ServerList) Add(name string, r *ServerResult) {
	s.m[name] = r
}

type ServerResult struct {
	Err error
	*server.Info
}
