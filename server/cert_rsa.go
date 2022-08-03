//go:build rsa

package server

import (
	"github.com/byzk-project-deploy/terminal-client/errors"
	"github.com/tjfoc/gmsm/gmtls"
	"github.com/tjfoc/gmsm/x509"
)

func getTlsConfig() *gmtls.Config {
	pair, err := gmtls.X509KeyPair(clientPemCert, clientPemKey)
	if err != nil {
		errors.ExitTlsConfig.Println("解析通信证书失败: %s", err.Error())
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(rootPemCert)

	return &gmtls.Config{
		Certificates:       []gmtls.Certificate{pair},
		ClientAuth:         gmtls.RequireAndVerifyClientCert,
		RootCAs:            pool,
		InsecureSkipVerify: false,
		ServerName:         hostname,
	}

}
