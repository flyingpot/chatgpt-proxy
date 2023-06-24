package api

import (
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
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
	client, _ = tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	httpProxy = os.Getenv("http_proxy")
	PORT      string
)

const (
	userAgent   = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"
	contentType = "application/x-www-form-urlencoded"
	defaultRole = "user"
	gpt4Model   = "gpt-4"
	openaiHost  = "chat.openai.com"
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
			url := "https://tcr9i.chat.openai.com/fc/gt2/public_key/35536E1E-65B4-4D96-9D97-6ADB7EFF8147"
			payload := "bda=eyJjdCI6IjJ3U2IwKzdkTVU4aE5lK2owdmExWUEvd2Q5MEFLeE5RemovSlRLL3hMNG42NFFWQVpzSU5YMVB5UjlBejUvbFlXL0VrYVJMeU8wSHdLeDJrdnNyNlBndUFhMUdqV3B3akVoQmxKTVNpVXREdE5ITGVxLyt6eW1mNG1QVTY5WVFQZzBSSVF0VWhzeGNwQ0Z0YzVnZTdDN3R1OHdkN05WNDBoM0xxZ2xNWlhwVDBiZ0hLN1hwMkpRMjR3MG9nUE5zL0UwWXhrN1c5bTdVTVR2R3NZK2ZsNzFRcjBBcm9qdG1DamdIOTBEaXhuVnJYYXFvWFUyQjVWekNaajhXVWVoOVZuUG9GeFBmcmpVWnZldGRaaXU4QXEzK3hzYWF0YXg1QnNzeWtIL1o2WFlCcGZXRVJWYnZldG9ybkRhQmd2M0tsSXA5VTJOcDd0STFsdnlIYXN5UjlxMDN4V3J1ZGJvVXNSV3E0QjFOalZOMGhpT0xEbVQ0NERsc2ptZWJNZzNCa2srYlBmd001WkdtYi9wT1hpUXhOYkcwYW1BelRFYkU3Vk1OZFV1RlBIZ3NxdTNlRDRGSnplTFZ5Zlp6UTE4UHNHVVMrMVdjSUFNTjZ5bmw2RWNqcEtXMDFLc0Q0RTBob25qc0xvREhaVS90dS9MT1hHRVJPd1l4dW1FZnUzRVdoZ3dUN0tuUi9FYy9xZXpQRXZuaTNrUVI3SVJ1bjZsR1pST2Evb2FXNnRWRis0K0YvV21uVEwyaDc0K0kwY1RNRkl0bFVZU00rR0JWcUd3WVlDemhEWFora3NNM2JWWjFzdWkrR29BckxvK0NkckN3NW1rc2pNM0hleDBVUEhySjhETGdxS043eW9rcnVObDk1N3hJaWlDZzdaMzF1M1lTNzgzdk1rZ1c4c3ltWVZaL3BxRDlGTUVGTXF1SkFzZVRBSFE2L09haDRVWGRIWjdkTTZNalNpR0ZRR3J3R0ZyVVBOTjQxMisyVkoxbjRJeHVNS0FCUytJQnl3R2J6UWtyOHdpcGZ0V1R6bm5JMk5BSWlYZy9sYVlUQ2RGNlpsL3VKb3I1L2JiUGN2RktTK011UXJPL2tkb1V6ZEJaTW56bWtqTTEyTm12M3BMQk9qcGw0VmRPSmJWZU0zUUplT0Rtc3htZ3ppS3hUa1FSamZCS3RMdlpjTlNTalU0TUpmZ09PY3I3S0Rqa2UvRnlDYSszWUNjVlZwNHA1QWMvN01oYUVZTmV1RXFxT2xTVytrR2pDdDVvdXpMMkIvODhKWHBsZk9nc1YvS3FyRVprTEVHd1R5UEwyK1Rzd3JJRWJyZElGOENSM2t3S2xaSVhHVVhVS1hHTDYzS2YwbjNvUHQ2ZXhqcWVZd0VnYWxwWk9ram5xKzdGWENXMmNNRk1Mb2kxRXR4YUtOWXJjblZBZ2c4aXkzdVBha1NxTkZma3BiM0dIV25OcVpZL2tnVzRBejlSeW85OVZwSlZoak5hamdnMG9YdTRUcnJoVHdoTklpdzczbC9hZ2pGRENEZmZ2eng1MWRpWHZUVDhEY2lReUgxMDFXQ0g2WHAwejVyTDIxU0Y1cCtNZ1BZZVNsSzBqSkdmWjE3VnZtdjYxOVdycVdXMTJqTUNZdkhJK2NmeDV2ZGR5TFZWNSt1NVdUckNydUlhOS9LUGd2MVNMdzRSLzg3VS96YXRiTFg3NU1neUlqT1FVSkIwZ2R4MGFScnZiWGx5LzVZUlZGRDZBRGc5S2dPSnVZWmhVdWg4NktYc0JDa2ZudnZMcjI3My9ialMyTjZ5WW9zZVNpMWZRT29KMmNoNjEzTEtmSjN2UjVIV1QrY0RYRlEvSTAxZDBMRXlZTXBtSmJLVHNhOG5FQkZHeWxpQTFWNk5FMXpnSEJZRXozVHlKalVwalRPQXMxWmdUb0ttb1FlMmk0OUgrY1FIZFRpdytrVi8vNVFOS1prRzNtcDNyQVRJYkVJTU12UlBLTlF1YThHQWZieXV2dHRmNkdyTlhHeldPdHBwbm5PNnFpSkJaZXRCT0dwWkZZTU1mcVNldGpYU2pabFdxMXVTSzZrUjc3dHdpa3Zta2dwQnRLdktiTDVrald3eElMVGdneDBmakM5Qjh1azYweitJZVNxenF2bnBuOXB5ZzlVZ0Fxb0xncHY3Um56ZEsySHJuY3ZzNkhTc3JGTW9iYWRvR3hpdk1QN204UC9rTVg3VjJBMEZRMVREUFNPSWFnMDFSWDVQOXZjZ3psVFhienZyMzl5VXhSWlRsUFVwNzhud3lFQzVWUzB4RUJzaXExTzVGUFdIV2R0Z3A1RkE2RGFrUnZIZzBLR0VWVWJzVzZ2cG9wUGpadW8yeFRTdlVxRWcwMGQwZ25BTmdUUnV3cnV3anhtWGpUa21OSCtWS3NkTklIQ1lWd3ZkOHZBa2hGTlF0MFRhV1NydVZoeTZNcmx2L1hTRzZLRVBwUUdQcWs4TFlZNHpocjlRUXltNDZjQUk1a25MaWlkY3pYNC8rR1E1S2xGUVZrUU9EK1lyRnduVElEZUZvWHB5RDBDRXdHcDZyME5aemU2THp2WWNQSVBDMStKTFJLYktqQTFjZnlndEVRdnFNUzh0ZFoyaHR5d0d6REUvbWlJM05ZNW9WVjFCRlZNNWZoMURjOGUvMyttUlFleEhkc0ZOaVVqQkFQeURZeC96Um81MlA1NnI1bGgyYjZQaFpCN3p3VVpwWlk2MGlWSnZiSU05R0pzUHhFaStiRFNiRkNJYk1NbUJFZWNpQk9STUdUeTFwdXpRZVlscElQR2hXNWRrdGxpMVNhTm1zWlNNNHJOeTRGenRxN1ZTODlUVjBsTHZENTlGUnZmOW8rTi85cWV6Ti9OcUpXYUZ2dnJNQmJ5MHZhUzR3cVVhemFPVnM4MnYwbUI3Nm9tL1V4UksyZEtQSG9jSWx5VVlpWFNXQVNEdWxJWVpsU1M2VXRrSjRkUlFETHYxaGgxRy9YNGNqUXpVV1lGTS83TjZFcFgyUzVBd2VjQS8xNGhzNTk3TzhJUHVFaWUxdWtJOGJSQUJCbDlGKzNVUzUxUW1mN1FUUXliakRNUUtLdjJ3VmlTamhBZE1TMCthRjJWNnNGMDB4V2xwR0I5Y291Y0RpVlNQZ2xmblFSakZaSUhoSVFVckJ2dkZqRGlmVWJheU9tNExaOG1QZWNuWjV0UTRDUkFhdEV0MStUbmFROWU4ZGlNZ2p1aW5WdlZMVUkyVGRJNU8yUkx1VHNTaUdJdVh2T003dlVHUnB4WTdEaURkTk5aUk9KY1RWeHduWEo5STlxZFI0YTh5T3hGSVdRRUg2WXVlNVV4VGJ0d3NjWG05V3A1K0pFNnpIKzJsdW5vMjE0UWZsTkl2UTRtVFJ6dHlCUlBLOGpuNWZOYTF1Rytoa25NNks3bEdnelA4bmRoL0JoR0ZnMjdjaCtFdG8wUHhsTURZd2NUcy9zRCt5Z0tMZU5qcDIrd1JTSmZObzdUSXdPWU5PMStXWDBlZVRiczBDbS9EMHc2aS9nUzJUdFhBK0NYc1M0UkI2UTRoc0txWlBOOEJZQ3Zsd2pSWDR2YWpMOVVxSGRFaVArb3IvWmNmNmROVTR2M3cvdGVRTDdMdi8rSlFIY0F0Y3VtbStzWlpDL09UYXd6eDVQREhGRk1lUG93L0NQd2VZb09uSkJEeHV5cHI2YkNrT2FGazN0K2U3TnlucmhoMFhGNkpKQnBSZXQvbjFITFh0dGZRQ3l6Mm5uTTAwOFM4Rm43Y2xZVFo2Y05GUENtcG1TV21ZYitCOFk4REJVdk1NZ3dpYjdKYklCSm5WMlplbWtQU2dZWVAvN09XaVlQNUd1RWxvcEdEeHMrR0gvYU4wMVE3N0JHT1gyMVlNUlQ5NytSS0gzekVlNUZyc0JDTm9aUU5kRHluTjlzbitUQ3JMWkJMZWhFQTJ0VXNSWmR3dmw1aS9lMlgyZSszSHBFb2Fkc0RwLzdNM1F4ckRpeGgwd3ovSjlEQVVsaFF4eHR4NUpNd1JZazBwRkpUV2N2QWFjbEo5OU9uR1RFMFdSZ3cxcnYvVTdZNHhhNHlabGQ0WmpMOUtBVGkvNkpWWlNBVFAzMnViZXN5TVliSmtMcnVONTZ6UytidHZ5bmNMUjl4dmY4R0lTd2dLN1dZVDZCTzUzT3haUmdRWXlHc3I0SXFxMkpWT3QweEtGQXB5ak1xNUZ2eVNNbkZ3WXZEMXYyelNpN3hRWUNNbmZ2NjlSWCtYZFJ3aVVvRUtEYURialNNN3VDOU41MTFzQlRJT2NsODd6emszMTdBcnJhNGhxemhoWnlPU1crU1ZmSG56Y1AyY0l2TnV1UjNscktlbGdNNmJYSk1FQThpV0lxbG1mclh1TXNYTHJhMUFiYy9kVnpNVXVzeFpEMENzbjZsOWVGVlh3eW9kZDdPT0s5Wlk4Nzd5STlMN1BVTDFxL2hrakw4SjY0OW5KMEdKOGUrVnA5dVJqS2JhMkVhU0JZS2pHbjhEdWl3ejZPQTdHc3YzT2hyckFKVXQranBOdGVCYVB2Tll6eU0wUjMxZzgwMFpvQ1FjNkFIVzVROWNIRGFlZWY4N2lQYm9sV0FRbVVrajhNR1IrUmV3Z3E0ZEcwdjZiUThkbTBaUzh0L0tqa0tsTHZYRDNRMFdjRnFpL2JEU25IVitDeVk4Q2ljdTJ2UFp2clBOUlptdEFtSEhzdGxIVjhuR0R1emg3d21ZNHp4eFZOUTJ1M2syQTlxeVI3SHBTcTZUaFFUbFpXVDluSFgwd2FxSTVKZlNlOXlKN0g5LzIwRmlWRjB1WGNHM0hHV0pkN1cxM1dVVjdnbkxQOStOcEptanJvZ2pBZVZIa2ptMElFblM3S0RhYjhiMjY5NjZZcWtsM0lzZFhJanNFNlFML1VqMFVPL3B3dW9DYWhGNnBUNVpBTFpXcXg5eUJXcUdwa1pBYUJha0tvdThHRndmL0pLUnRDdFd5ZVduQ1M4RS8wRTJOUEpraTdyZVZUTjhHTEFGRkVDMnlud2dsdzRDbE1YdUhnQWIwMUpVL0JmYVhYRVRGR1p3U05TejVaN0gvU1FReTE1UjgwL05TWVB3cHEzUUFGcWFIMmNIY3Zrbk0xL3Awd0hrVmQvQWd3b25OVXlaZEpUenZHZmsxRWJBQTlpeGtNc0k5cCtmR3kxVUoxODZIdmt3Wk9QSTdiL3N5VEtQdGh4TEM2Nk9scXI4S0MrQ1laSlFHN3VuMGgzUHV0WHJ0Zk1BM2hySDlLMm1VSk1XNHdhSDE1bjUwcTd0bmdoK3QvVzNxbUpmU1hPTTVTaENGRkdIeG5SRlRqdThmT3BYMUNYL2U3Q1NGNDdmWXBrTkpXM0MyVU1DMkZqeERMYU9ucVNvYWI1dVg3am9XbGo1MFFxdWZERWdqK2dQQmZBRlBJUTcydkZka3BnVVVCaCtrS09YWWtSRmlNczIvZ1B0ejRPSzh3MWViQXZNcjhiRi9XOVY0NG53eUxXZ2pYdEUwSWVHZ09QZklwc2tCK2cvU25YVlZjQ3B2TkcyZUlmc1l5VXJaNzV1bW5kWkhVcnQrbmFIRGhGZnVPODNORUNQYXhCem1rUFB5MTR4UDBpcjNNOXAreXc5YU5vUGJwcVRleTB6a3l0Wk5nUVExTEFjc0E3QVZ3a0RQQndYTHh3OWpJeXBYSWcyekY1SE40NDJOZUh5alVNWFBTMTFvNitqTlcyTDFRajlDZXVBMFJDZHo2UVVkTERJcUpPQ0l2RWNYSFQrMmpmcDZRWjRJTmlsOFcxV0p4N1V6ZzZkQmpIZmF4Ym4xeUtGb1VmVFZFUTZiaW42ajBlR0FIdlZidTk0RFBCTEllK1JRU2pTQXA3M3NGNGZzbkxicVh3WkJxcStuemZDZ1hMd1lTVFF0Y2N3R2hkbkV0ZlVKU3FmaUlNdTJuT1hmZWFDNklEbkdRRE1sVVJpWmpZc2FKeTcxTTBhR3MwcTdmbFJhSkl4U3F6S1krVUNZVXRwT1E0Q3ZFNzNjNHhHVjFKL0w1K21pb00vQkZyM2NGYWg0SW01NTRWaUVIa0dKTjVrTktPWTlLNVNreDdCQ0xEQi9HdUIrTGh3Yk1GQ0JsQlk4Vk1taitxK3BzQ3M3VTllVi9lSERtUXBiRHhkbmhvR1d1b3k4NVFwRVIrb3VRVmZSeXhGN1p1c1FvaTA5UUJhSVhUTWQ1VTV0RStETHpldjVFVFR4K09IY1paSDRaSVg2bUlHWExNNG5WMHZ4ZzFJNGZvTEFJL1BQV3Vla010VlkxM3FkMkZZY1J2aG5mNXA0UTBLSmZESnYxVjNuRzJIZnc2S0llZ0p4eXI2elp2RUQ3WDc1R0VoQThCSkdncitoZUhWNC9qMVVybGUzM1hUbmlqTkNLZ2tCZGptMkR0enNoTkQzcjMyVU4xRHFXV1hvb0NobkthRDdFMXl5UHovYnNvT1M5QUM1RVBmUllsQ0ZkN0liTjRxUE5CZzJGcnB0ZSsvNFAwM2hlUjVwTkNHanU5RzVOemtzdlVlNGpzPSIsIml2IjoiYmIxNmY3NDJhNDk5ZDYwYTI0NTY4NTI5MGJlMzQ0ZjQiLCJzIjoiYWFhYmI3YjVkNDJmMjEwMCJ9&public_key=35536E1E-65B4-4D96-9D97-6ADB7EFF8147&site=https%3A%2F%2Fchat.openai.com&userbrowser=Mozilla%2F5.0%20(X11%3B%20Linux%20x86_64%3B%20rv%3A114.0)%20Gecko%2F20100101%20Firefox%2F114.0&capi_version=1.5.2&capi_mode=lightbox&style_theme=default&rnd=0.2304346700108898"
			req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
			req.Header.Set("Content-Type", contentType)
			resp, err := client.Do(req)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			responseMap := make(map[string]string)
			json.NewDecoder(resp.Body).Decode(&responseMap)
			cRequest.ArkoseToken = responseMap["token"]
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
