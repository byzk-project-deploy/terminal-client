package config

type Info struct {
	// Server *ServerConfig
	System *SystemConfig
}

type ServerConfig struct {
}

// SystemConfig shell环境配置
type SystemConfig struct {
	// CallShellPath 系统调用的shell路径
	CallShellPath string
	// CallShellArgs 系统调用的shell的传入参数
	CallShellArgs []string
}
