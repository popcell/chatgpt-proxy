package api_proxy

import (
	"crypto/tls"
	"fmt"
	log "github.com/popcell/chatgpt-proxy/utils/logger"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var (
	ReverseProxy   *httputil.ReverseProxy
	proxyTransport = &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: time.Second * 60,
		IdleConnTimeout:       time.Second * 60,
		TLSHandshakeTimeout:   time.Second * 60,
	}
)

func InitApiProxyConfig(proxyUrl string, apiPrefix string) {

	if strings.HasPrefix(proxyUrl, "http") {
		proxyTransport.Proxy =
			func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyUrl)
			}
	} else if strings.HasPrefix(proxyUrl, "socks5") {
		u, err := url.Parse(proxyUrl)
		if err != nil {
			log.Error("InitApiProxyConfig error: ", err.Error())
			return
		}
		username := u.User.Username()
		password, _ := u.User.Password()
		host := u.Hostname()
		address := fmt.Sprintf("%s:%s", host, u.Port())
		dialer, err := proxy.SOCKS5("tcp", address, &proxy.Auth{User: username, Password: password}, proxy.Direct)
		proxyTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
		if err != nil {
			log.Error("InitApiProxyConfig error: ", err.Error())
		}
	}
	target, err := url.Parse("https://api.openai.com")
	if err != nil {
		log.Error("ApiProxy error: ", err)
	}
	ReverseProxy = httputil.NewSingleHostReverseProxy(target)
	ReverseProxy.Transport = proxyTransport
	ReverseProxy.Director = func(req *http.Request) {
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = strings.Replace(req.URL.Path, apiPrefix, "", -1)
	}

	// 打印HTTP请求和响应的日志
	//ReverseProxy.ModifyResponse = func(resp *http.Response) error {
	//	// 打印HTTP响应的日志
	//	responseDump, err := httputil.DumpResponse(resp, true)
	//	if err != nil {
	//		log.Info("Failed to dump response: ", err)
	//	} else {
	//		log.Info("Response: ", string(responseDump))
	//	}
	//	return nil
	//}
}
