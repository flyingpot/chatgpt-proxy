package api

import (
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/flyingpot/funcaptcha"
	"io"
	"log"
	nethttp "net/http"
	"os"
	"strings"

	"github.com/acheong08/endless"
	"github.com/gin-gonic/gin"
)

var (
	handler *gin.Engine
	jar     = tlsclient.NewCookieJar()
	options = []tlsclient.HttpClientOption{
		tlsclient.WithTimeoutSeconds(360),
		tlsclient.WithClientProfile(tlsclient.Safari_IOS_16_0),
		tlsclient.WithNotFollowRedirects(),
		tlsclient.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}
	client, _      = tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	httpProxy      = os.Getenv("http_proxy")
	arkoseTokenUrl = os.Getenv("ARKOSE_TOKEN_URL")
	PORT           string
)

const (
	userAgent             = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"
	contentType           = "application/x-www-form-urlencoded"
	defaultRole           = "user"
	gpt4Model             = "gpt-4"
	openaiHost            = "chat.openai.com"
	defaultArkoseTokenUrl = "https://funcaptcha.vercel.app/api/token"
)

type CreateConversationRequest struct {
	Action                     string    `json:"action"`
	Messages                   []Message `json:"messages"`
	Model                      string    `json:"model"`
	ParentMessageID            string    `json:"parent_message_id"`
	ConversationID             *string   `json:"conversation_id"`
	PluginIDs                  []string  `json:"plugin_ids"`
	TimezoneOffsetMin          int       `json:"timezone_offset_min"`
	ArkoseToken                string    `json:"arkose_token"`
	HistoryAndTrainingDisabled bool      `json:"history_and_training_disabled"`
	AutoContinue               bool      `json:"auto_continue"`
}

type Message struct {
	Author  Author  `json:"author"`
	Content Content `json:"content"`
	ID      string  `json:"id"`
}

type Author struct {
	Role string `json:"role"`
}

type Content struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

func init() {
	if httpProxy != "" {
		client.SetProxy(httpProxy)
		println("Proxy set:" + httpProxy)
	}

	if arkoseTokenUrl == "" {
		arkoseTokenUrl = defaultArkoseTokenUrl
	}

	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "9090"
	}
	handler = gin.Default()
	handler.Use(Cors())

	handler.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, ChatGPT!")
	})

	handler.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	handler.Any("/api/*path", proxy)

	gin.SetMode(gin.ReleaseMode)
}

func GetHandler() *gin.Engine {
	return handler
}

func Run() {
	endless.ListenAndServe(os.Getenv("HOST")+":"+PORT, handler)
}

// entrypoint for vercel
func Handler(w nethttp.ResponseWriter, r *nethttp.Request) {
	handler.ServeHTTP(w, r)
}

func proxy(c *gin.Context) {
	// Remove _cfuvid cookie from session
	jar.SetCookies(c.Request.URL, []*http.Cookie{})

	var request_url string
	var err error
	var request_method string
	var request *http.Request
	var response *http.Response

	if c.Param("path") == "/conversation_limit" {
		request_url = "https://" + openaiHost + "/public-api" + c.Param("path") + "?" + c.Request.URL.RawQuery
	} else if c.Request.URL.RawQuery != "" {
		request_url = "https://" + openaiHost + "/backend-api" + c.Param("path") + "?" + c.Request.URL.RawQuery
	} else {
		request_url = "https://" + openaiHost + "/backend-api" + c.Param("path")
	}
	request_method = c.Request.Method

	var body io.Reader
	if c.Param("path") == "/conversation" {
		var cRequest CreateConversationRequest
		if err := c.BindJSON(&cRequest); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		if cRequest.ConversationID == nil || *cRequest.ConversationID == "" {
			cRequest.ConversationID = nil
		}

		if len(cRequest.Messages) != 0 {
			if cRequest.Messages[0].Author.Role == "" {
				cRequest.Messages[0].Author.Role = defaultRole
			}
		}

		if strings.HasPrefix(cRequest.Model, gpt4Model) {
			result, err := funcaptcha.GetToken(&funcaptcha.GetTokenOptions{
				PKey: "35536E1E-65B4-4D96-9D97-6ADB7EFF8147",
				SURL: "https://tcr9i.chat.openai.com",
			})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			cRequest.ArkoseToken = result.Token
		}
		jsonBytes, _ := json.Marshal(cRequest)
		body = bytes.NewBuffer(jsonBytes)
	} else {
		body = c.Request.Body
	}

	request, err = http.NewRequest(request_method, request_url, body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	request.Header.Set("Authorization", GetAccessTokenFromHeader(c.Request.Header))
	request.Header.Set("user-agent", userAgent)

	response, err = client.Do(request)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()
	// Get status code
	c.Status(response.StatusCode)
	if response.StatusCode > 299 {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Printf("Could not read response body: %v\n", err)
		} else {
			log.Printf("Request failed with status code: %d, status: %s, body: %s\n", response.StatusCode, http.StatusText(response.StatusCode), string(bodyBytes))
		}
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")

	buf := make([]byte, 4096)
	for {
		n, err := response.Body.Read(buf)
		if n > 0 {
			_, writeErr := c.Writer.Write(buf[:n])
			if writeErr != nil {
				log.Printf("Error writing to client: %v", writeErr)
				break
			}
		}

		c.Writer.Flush()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading from response body: %v", err)
			break
		}
	}
}

func GetAccessTokenFromHeader(header nethttp.Header) string {
	// pandora will pass X-Authorization header
	// but maybe other project will use Authorization header to pass access token
	xAuth := header.Get("X-Authorization")
	if xAuth == "" {
		return header.Get("Authorization")
	} else {
		return xAuth
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Accept,Origin,Content-Length,Content-Type,Authorization,X-Authorization,X-Requested-With,Access-Control-Request-Method,Access-Control-Request-Headers,Content-Disposition")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
