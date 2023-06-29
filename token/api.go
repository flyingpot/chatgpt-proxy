package token

import (
	"bytes"
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"io"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
)

const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

type GetTokenResult struct {
	ChallengeURL          string `json:"challenge_url"`
	ChallengeURLCDN       string `json:"challenge_url_cdn"`
	ChallengeURLCDNSRI    string `json:"challenge_url_cdn_sri"`
	DisableDefaultStyling bool   `json:"disable_default_styling"`
	IFrameHeight          int    `json:"iframe_height"`
	IFrameWidth           int    `json:"iframe_width"`
	KBio                  bool   `json:"kbio"`
	MBio                  bool   `json:"mbio"`
	NoScript              string `json:"noscript"`
	TBio                  bool   `json:"tbio"`
	Token                 string `json:"token"`
}

func GetOpenAIToken(client tlsclient.HttpClient) (string, error) {
	surl := "https://tcr9i.chat.openai.com"
	pkey := "35536E1E-65B4-4D96-9D97-6ADB7EFF8147"

	formData := url.Values{
		"bda": {GetBda(DefaultUserAgent,
			fmt.Sprintf("%s/v2/%s/1.5.2/enforcement.%s.html",
				surl, pkey, Random()), "")},
		"public_key":   {"35536E1E-65B4-4D96-9D97-6ADB7EFF8147"},
		"site":         {"https://chat.openai.com"},
		"userbrowser":  {DefaultUserAgent},
		"capi_version": {"1.5.2"},
		"capi_mode":    {"lightbox"},
		"style_theme":  {"default"},
		"rnd":          {strconv.FormatFloat(rand.Float64(), 'f', -1, 64)},
	}

	form := strings.ReplaceAll(formData.Encode(), "+", "%20")
	form = strings.ReplaceAll(form, "%28", "(")
	form = strings.ReplaceAll(form, "%29", ")")
	req, err := http.NewRequest("POST", surl+"/fc/gt2/public_key/"+pkey, bytes.NewBufferString(form))
	if err != nil {
		return "", err
	}

	req.Header.Set("Origin", surl)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("sec-fetch-mode", "cors")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result GetTokenResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	return result.Token, nil
}
