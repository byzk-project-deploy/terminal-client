package config

import (
	"os"
	"path/filepath"

	"github.com/byzk-project-deploy/terminal-client/errors"
	"github.com/byzk-project-deploy/terminal-client/user"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	configDirPath  = filepath.Join(user.HomeDir(), ".bypt")
	configFilePath = filepath.Join(configDirPath, "client.toml")
)

var (
	currentConfig *Info
)

func init() {
	viper.SetConfigFile(configFilePath)
	viper.SetConfigType("toml")
	viper.SetDefault("system.callShellPath", "/usr/bin/bash")
	viper.SetDefault("system.callShellArgs", []string{"-c"})

	stat, err := os.Stat(configFilePath)
	if err != nil || stat.IsDir() {
		if err = os.MkdirAll(configDirPath, 0755); err != nil {
			errors.ExitConfigFileCreatEmpty.Println("创建空的默认配置文件失败: %s", err.Error())
		}

		if err = viper.WriteConfigAs(configFilePath); err != nil {
			errors.ExitConfigFileWriteToEmpty.Println("写出默认配置失败: %s", err.Error())
		}
	}

	if err = viper.ReadInConfig(); err != nil {
		errors.ExitConfigFileRead.Println("读取配置文件内容失败: %s", err.Error())
	}

	if err = viper.Unmarshal(&currentConfig); err != nil {
		errors.ExitConfigFileParser.Println("配置文件解析失败: %s", err.Error())
	}
	viper.OnConfigChange(configOnChange)
	viper.WatchConfig()
}

func configOnChange(in fsnotify.Event) {
	var tempConfig *Info
	if err := viper.Unmarshal(&tempConfig); err == nil {
		currentConfig = tempConfig
	}
}

func Current() *Info {
	return currentConfig
}
