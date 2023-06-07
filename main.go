package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/popcell/chatgpt-proxy/api"
	"github.com/popcell/chatgpt-proxy/api/api_proxy"
	"github.com/popcell/chatgpt-proxy/api/openai_proxy"
	"github.com/popcell/chatgpt-proxy/middleware"
	log "github.com/popcell/chatgpt-proxy/utils/logger"
	"github.com/spf13/viper"
	"os"
)

var (
	isLogin string
)

func initFlag() {
	flag.StringVar(&isLogin, "c", "", "Use -c login to login")
	flag.StringVar(&isLogin, "cmd", "", "Use -c login to login")

	// 自定义帮助信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	// 解析命令行参数
	flag.Parse()
}
func main() {
	initFlag()
	// 设置配置文件名和路径
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic("Read config failed: ", err)
	}
	host := viper.GetString("host")
	port := viper.GetInt("port")

	proxyUrl := viper.GetString("proxy-url")

	openAiProxyEnable := viper.GetBool("openai-proxy.enable")
	openAiEmail := viper.GetString("openai-proxy.email")
	openAiPassword := viper.GetString("openai-proxy.password")
	puid := viper.GetString("openai-proxy.puid")
	openAiProxyPrefix := viper.GetString("openai-proxy.prefix")

	apiProxyEnable := viper.GetBool("api-proxy.enable")
	apiProxyPrefix := viper.GetString("api-proxy.prefix")

	// 输出解析结果
	if isLogin == "login" {
		api.InitConfig(proxyUrl)
		api.Login(openAiEmail, openAiPassword)
		return
	}

	gin.SetMode(gin.ReleaseMode)
	gin.ForceConsoleColor()
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello,World!")
	})
	if openAiProxyEnable {
		openai_proxy.InitOpenAiConfig(proxyUrl, puid)
		router.Any(fmt.Sprintf("/%s/*path", openAiProxyPrefix), openai_proxy.OpenAiProxy)
		fmt.Printf("openAiProxy run on %s:%d/%s\n", host, port, openAiProxyPrefix)
	}
	if apiProxyEnable {
		api_proxy.InitApiProxyConfig(proxyUrl, apiProxyPrefix)
		var chatGptUrl = fmt.Sprintf("/%s/*path", apiProxyPrefix)
		router.Any(chatGptUrl, func(c *gin.Context) {
			//requestDump, err := httputil.DumpRequest(c.Request, true)
			//if err != nil {
			//	log.Info("Failed to dump request: ", err)
			//} else {
			//	log.Info("Request: ", time.Now().Format("2006-01-02 15:04:05"), string(requestDump))
			//}
			api_proxy.ReverseProxy.ServeHTTP(c.Writer, c.Request)
		})
		fmt.Printf("chatGptApiProxy run on %s:%d/%s\n", host, port, apiProxyPrefix)
	}

	err = router.Run(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Error("run server err: ", err.Error())
	}
}
