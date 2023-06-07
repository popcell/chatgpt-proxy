package openai_proxy

import (
	http "github.com/bogdanfinn/fhttp"
	tlsClient "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
	log "github.com/popcell/chatgpt-proxy/utils/logger"
	"io"
)

var (
	jar     = tlsClient.NewCookieJar()
	options = []tlsClient.HttpClientOption{
		tlsClient.WithTimeoutSeconds(360),
		tlsClient.WithClientProfile(tlsClient.Safari_IOS_16_0),
		tlsClient.WithNotFollowRedirects(),
		tlsClient.WithCookieJar(jar),
	}
	client, _ = tlsClient.NewHttpClient(tlsClient.NewNoopLogger(), options...)
	userAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 14_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Mobile/15E148 Safari/604.1"
	Puid      = ""
)

func InitOpenAiConfig(proxy string, puid string) {
	// http://username:password@host:port 或 socks5://username:password@host:port
	err := client.SetProxy(proxy)
	if err != nil {
		log.Error("proxy set error: ", err.Error())
	}
	Puid = puid
}
func OpenAiProxy(c *gin.Context) {
	// Remove _cfuvid cookie from session
	jar.SetCookies(c.Request.URL, []*http.Cookie{})
	var url string
	var err error
	var requestMethod string
	var request *http.Request
	var response *http.Response

	if c.Request.URL.RawQuery != "" {
		url = "https://chat.openai.com/backend-api" + c.Param("path") + "?" + c.Request.URL.RawQuery
	} else {
		url = "https://chat.openai.com/backend-api" + c.Param("path")
	}
	requestMethod = c.Request.Method

	request, err = http.NewRequest(requestMethod, url, c.Request.Body)
	if err != nil {
		log.Error("NewRequest Error: ", err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	XAuthorization := c.Request.Header.Get("Authorization")

	request.Header.Set("Host", "chat.openai.com")
	request.Header.Set("Origin", "https://chat.openai.com/chat")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Keep-Alive", "timeout=360")
	request.Header.Set("Authorization", XAuthorization)
	request.Header.Set("sec-ch-ua", "\"Chromium\";v=\"112\", \"Brave\";v=\"112\", \"Not:A-Brand\";v=\"99\"")
	request.Header.Set("sec-ch-ua-mobile", "?0")
	request.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	request.Header.Set("sec-fetch-dest", "empty")
	request.Header.Set("sec-fetch-mode", "cors")
	request.Header.Set("sec-fetch-site", "same-origin")
	request.Header.Set("sec-gpc", "1")
	request.Header.Set("user-agent", userAgent)
	if Puid != "" {
		request.Header.Set("cookie", "_puid="+Puid+";")
	}

	response, err = client.Do(request)
	if err != nil {
		log.Error("Request Error: ", err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()
	c.Header("Content-Type", response.Header.Get("Content-Type"))
	// Get status code
	c.Status(response.StatusCode)

	buf := make([]byte, 4096)
	for {
		n, err := response.Body.Read(buf)
		if n > 0 {
			_, writeErr := c.Writer.Write(buf[:n])
			if writeErr != nil {
				log.Error("Error writing to client: %v", writeErr)
				break
			}
			c.Writer.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("Error reading from response body: %v", err)
			break
		}
	}
}
