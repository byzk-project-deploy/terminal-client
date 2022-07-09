package config

import (
	"github.com/byzk-project-deploy/terminal-client/errors"
	"github.com/byzk-project-deploy/terminal-client/user"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
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
}

func Current() *Info {
	return currentConfig
}
