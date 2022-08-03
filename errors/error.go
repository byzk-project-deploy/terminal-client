package errors

import (
	"fmt"
	"os"
)

type ExitCode uint8

const (
	// ExitGetSysCurrentUser 获取系统当前用户失败
	ExitGetSysCurrentUser ExitCode = iota + 1
	// ExitConfigFileCreatEmpty 创建空的配置文件失败
	ExitConfigFileCreatEmpty
	// ExitConfigFileRead 配置文件读取失败
	ExitConfigFileRead
	// ExitConfigFileWriteToEmpty 默认配置写出失败
	ExitConfigFileWriteToEmpty
	// ExitConfigFileParser 配置解析失败
	ExitConfigFileParser
	// ExitTlsConfig tls配置失败
	ExitTlsConfig
)

func (e ExitCode) Println(formatStr string, args ...any) {
	_, _ = os.Stderr.Write([]byte(fmt.Sprintf(formatStr, args...)))
	e.Exit()
}

func (e ExitCode) Exit() {
	os.Exit(int(e))
}
