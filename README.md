# ChatGpt Proxy
An unofficial Golang program that provides both ChatGPT access token and API proxy simultaneously.

一个非官方的Golang程序，同时提供ChatGPT访问令牌和API代理。

## 配置文件
```yaml
host: ""
port: 55555
# 代理 	http://username:password@host:port 或 socks5://username:password@host:port
proxy-url: "http://127.0.0.1:7890"
openai-proxy:
  # 是否开启
  enable: true
  # 账号密码 可选
  email: ""
  password: ""
  # plus账号cookie 解除限制 可选
  puid: ""
  # api前缀
  prefix: "openai-proxy"
api-proxy:
  # 是否开启
  enable: true
  # api前缀
  prefix: "api-proxy"
```
## 快速入门
[根据自己系统版本下载release](https://github.com/popcell/chatgpt-proxy/releases)
填写config配置文件
```shell
# login to get access token
./chatgpt-proxy -c login 
# run proxy server
./chatgpt-proxy

# openai_proxy http://localhost:55555/openai-proxy
# api_proxy http://localhost:55555/api-proxy
```
## docker部署
```shell
wget https://raw.githubusercontent.com/popcell/chatgpt-proxy/master/config.yaml
# 修改config.yaml
docker build -t chatgpt-proxy .
docker run -d --name chatgpt-proxy --restart always -p 55555:55555 -v $(pwd)/config.yaml:/app/config.yaml chatgpt-proxy
```
## docker-compose部署
```shell
wget https://raw.githubusercontent.com/popcell/chatgpt-proxy/master/config.yaml
# 修改config.yaml
docker build -t chatgpt-proxy .
docker-compose up -d
```
## 参考
[通过tls指纹绕过cloudflare](https://github.com/acheong08/ChatGPT-Proxy-V4)

[IOS获取access_token](https://zhile.io/2023/05/19/how-to-get-chatgpt-access-token-via-pkce.html)