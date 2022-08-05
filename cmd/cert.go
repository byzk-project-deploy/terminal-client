//go:build inside

package cmd

import (
	"fmt"
	"github.com/byzk-project-deploy/grumble"
	serverclientcommon "github.com/byzk-project-deploy/server-client-common"
	"github.com/byzk-project-deploy/terminal-client/server"
	"github.com/byzk-project-deploy/terminal-client/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type CertResult struct {
	CertPem       string
	PrivateKeyPem string
}

var (
	keyPairCmd = &grumble.Command{
		Name:     "keypair",
		Help:     "密钥工具",
		LongHelp: "基于当前服务器跟正书生成及导出证书和密钥方便客户端与插件的调试，内部私有命令",
		Flags: func(f *grumble.Flags) {
			f.String("o", "outFile", "", "将证书和私钥导出到文件, 私钥及证书文件的名称将采用outFile.[key或crt]来命名")
			f.String("t", "type", "plugin", "证书及密钥类型, 可选值: client(客户端)、plugin(插件)")
			f.String("a", "pluginAuthor", "", "当type选项为plugin时的插件作者名称")
			f.String("n", "pluginName", "", "当type选项为plugin时的插件名称")
		},
		Run: func(c *grumble.Context) (err error) {
			t := strings.ToLower(c.Flags.String("type"))
			if t == "" {
				t = "plugin"
			}

			var options *serverclientcommon.KeypairGeneratorInfo
			switch t {
			case "client":
				options = &serverclientcommon.KeypairGeneratorInfo{
					Type: t,
				}
			case "plugin":
				pluginAuthor := c.Flags.String("pluginAuthor")
				pluginName := c.Flags.String("pluginName")

				if pluginAuthor == "" {
					if pluginAuthor, err = c.ShellTools.Prompt(utils.PromptNotEmptyVerify("作者名称")); err != nil {
						_, _ = c.App.Println("操作已取消")
						return nil
					}
				}

				if pluginName == "" {
					if pluginName, err = c.ShellTools.Prompt(utils.PromptNotEmptyVerify("插件名称")); err != nil {
						_, _ = c.App.Println("操作已取消")
						return nil
					}
				}

				options = &serverclientcommon.KeypairGeneratorInfo{
					Type:   t,
					Author: pluginAuthor,
					Name:   pluginName,
				}
			default:
				return fmt.Errorf("不支持的密钥类型: %s", t)
			}

			unixServerInfo := server.NewUnixServerInfo()
			stream, err := unixServerInfo.ConnToStream()
			if err != nil {
				return err
			}

			var res *CertResult
			resData, err := serverclientcommon.CmdKeyPair.ExchangeWithData(options, stream)
			if err != nil {
				return err
			}

			if err = resData.UnmarshalJson(&res); err != nil {
				return fmt.Errorf("转换服务器响应数据失败: %s", err.Error())
			}

			outFile := c.Flags.String("outFile")
			if outFile == "" {
				_, _ = c.App.Println("证书PEM字符串:")
				_, _ = c.App.Println(res.CertPem)
				_, _ = c.App.Println()
				_, _ = c.App.Println()
				_, _ = c.App.Println("私钥PEM字符串:")
				_, _ = c.App.Println(res.PrivateKeyPem)
			} else {
				if strings.HasSuffix(outFile, "/") {
					outFile += "keypair"
				}

				_ = os.MkdirAll(filepath.Dir(outFile), 0755)
				if err := ioutil.WriteFile(outFile+".crt", []byte(res.CertPem), 0655); err != nil {
					return fmt.Errorf("写出证书到文件[%s]失败: %s", outFile+".crt", err.Error())
				}

				if err := ioutil.WriteFile(outFile+".key", []byte(res.PrivateKeyPem), 0655); err != nil {
					return fmt.Errorf("写出证书到文件[%s]失败: %s", outFile+".key", err.Error())
				}
			}

			return
		},
	}
)

func init() {
	defaultCommands = append(defaultCommands, keyPairCmd)
}
