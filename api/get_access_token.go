package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	log "github.com/popcell/chatgpt-proxy/utils/logger"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Auth0 struct {
	SessionToken string
	Email        string
	Password     string
	ApiPrefix    string
}

var (
	client         *http.Client
	proxyTransport *http.Transport
)

func GenerateCodeVerifier() string {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		panic(err)
	}
	encodedToken := base64.RawURLEncoding.EncodeToString(token)
	codeVerifier := strings.TrimRight(encodedToken, "=")
	return codeVerifier
}

func GenerateCodeChallenge(codeVerifier string) string {
	hash := sha256.New()
	hash.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	codeChallenge = strings.TrimRight(codeChallenge, "=")
	return codeChallenge
}
func CheckEmail(email string) bool {
	emailRegex := `^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,7}$`
	return regexp.MustCompile(emailRegex).MatchString(email)
}

func NewAuth0(email, password string) *Auth0 {
	if CheckEmail(email) == false {
		panic("Invalid email")
	}
	return &Auth0{
		Email:    email,
		Password: password,
	}
}
func (auth Auth0) run() {
	auth.GetLoginUrl()
}
func (auth Auth0) GetLoginUrl() {
	codeVerifier := GenerateCodeVerifier()
	codeChallenge := GenerateCodeChallenge(codeVerifier)
	reqUrl := "https://auth0.openai.com/authorize?client_id=pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh&audience=https%3A%2F%2Fapi.openai.com%2Fv1&redirect_uri=com.openai.chat%3A%2F%2Fauth0.openai.com%2Fios%2Fcom.openai.chat%2Fcallback&scope=openid%20email%20profile%20offline_access%20model.request%20model.read%20organization.read%20offline&response_type=code&code_challenge=" + codeChallenge + "&code_challenge_method=S256&prompt=login"
	auth.Login(codeVerifier, reqUrl)
}

func (auth Auth0) Login(codeVerifier string, reqUrl string) {
	// 创建 HTTP 请求
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://ios.chat.openai.com/")

	if err != nil {
		fmt.Println("Login req Error:", err)
		return
	}

	// 发送 HTTP 请求
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Login Do Error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		parsedURL, _ := url.Parse(resp.Request.URL.String())
		queryValues := parsedURL.Query()
		state := queryValues.Get("state")
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		auth.CheckUsername(codeVerifier, state)
	} else {
		fmt.Println("Login Do Error")
	}

}

func (auth Auth0) CheckUsername(codeVerifier string, state string) {
	reqUrl := fmt.Sprintf("https://auth0.openai.com/u/login/identifier?state=%s", state)
	// 创建 HTTP 请求
	formData := url.Values{
		"state":                       {state},
		"username":                    {auth.Email},
		"js-available":                {"true"},
		"webauthn-available":          {"true"},
		"is-brave":                    {"false"},
		"webauthn-platform-available": {"false"},
		"action":                      {"default"},
	}

	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(formData.Encode()))
	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	req.Header.Set("Referer", reqUrl)
	req.Header.Set("Origin", "https://auth0.openai.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		fmt.Println("CheckUsername Error:", err)
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("CheckUsername Error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusFound {
		auth.CheckPassword(codeVerifier, state)
	} else {
		fmt.Println("CheckUsername Error")
	}

}
func (auth Auth0) CheckPassword(codeVerifier string, state string) {
	reqUrl := fmt.Sprintf("https://auth0.openai.com/u/login/password?state=%s", state)
	// 创建 HTTP 请求
	formData := url.Values{
		"state":    {state},
		"username": {auth.Email},
		"password": {auth.Password},
		"action":   {"default"},
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		fmt.Println("CheckPassword req Error:", err)
		return
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	req.Header.Set("Referer", reqUrl)
	req.Header.Set("Origin", "https://auth0.openai.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("CheckPassword Error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if !strings.HasPrefix(location, "/authorize/resume?") {
			fmt.Println("Login failed.", err)
			return
		}
		auth.GetLocation(codeVerifier, location, reqUrl)
	} else if resp.StatusCode == http.StatusBadRequest {
		fmt.Println("Wrong email or password.", err)
		return
	} else {
		fmt.Println("Error Login.", err)
		return
	}

}

func (auth Auth0) GetLocation(codeVerifier string, location string, ref string) {
	reqUrl := fmt.Sprintf("https://auth0.openai.com%s", location)

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		fmt.Println("GetLocation Error:", err)
		return
	}
	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	req.Header.Set("Referer", ref)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("GetLocation Error:", err)
		return
	}
	if resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		if !strings.HasPrefix(location, "com.openai.chat://auth0.openai.com/ios/com.openai.chat/callback?") {
			fmt.Println("Login failed GetLocation", err)
			return
		}
		auth.GetAccessToken(codeVerifier, resp.Header.Get("Location"))
	} else if resp.StatusCode == 400 {
		fmt.Println("Wrong email or password.", err)
		return
	} else {
		fmt.Println("Error Login.", err)
		return
	}
}

func (auth Auth0) GetAccessToken(codeVerifier string, callbackUrl string) {

	parsedURL, _ := url.Parse(callbackUrl)
	queryValues := parsedURL.Query()
	if reqError, ok := queryValues["error"]; ok {
		errorStr := fmt.Sprintf("%s: %s", reqError[0], queryValues["error_description"][0])
		fmt.Println("errorStr", errorStr)
		return
	} else if _, ok := queryValues["code"]; !ok {
		fmt.Println("Error get code from callback url.")
		return
	}
	reqUrl := "https://auth0.openai.com/oauth/token"
	// 创建 HTTP 请求
	formData := url.Values{
		"redirect_uri":  {"com.openai.chat://auth0.openai.com/ios/com.openai.chat/callback"},
		"grant_type":    {"authorization_code"},
		"client_id":     {"pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh"},
		"code":          {queryValues["code"][0]},
		"code_verifier": {codeVerifier},
	}

	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		fmt.Println("GetAccessToken Error:", err)
		return
	}
	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("GetAccessToken Error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		accessToken := result["access_token"].(string)
		expiresAt := time.Now().Add(time.Second * time.Duration(result["expires_in"].(float64))).Add(-5 * time.Minute)
		expiresIn := expiresAt.Format("2006-01-02T15:04:05.000Z")
		fmt.Println(accessToken)
		fmt.Println(expiresIn)
		err := os.WriteFile("access_token.txt", []byte(accessToken), 0644)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
		fmt.Println("accessToken written access_token.txt successfully")
	} else {
		fmt.Println("Error check email.")
	}
	return
}
func InitConfig(proxyUrl string) {
	if strings.HasPrefix(proxyUrl, "http") {
		proxyTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyUrl)
			},
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
		proxyTransport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		}
		if err != nil {
			log.Error("InitApiProxyConfig error: ", err.Error())
		}
	}
	// 创建 cookie 容器
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println("Cookiejar Error:", err)
		return
	}

	client = &http.Client{
		Transport: proxyTransport,
		Timeout:   time.Second * 100,
		Jar:       jar,
	}
}
func Login(email string, password string) {
	auth0 := NewAuth0(email, password)
	auth0.run()
}
