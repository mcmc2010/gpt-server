package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"

	"mcmcx.com/gpt-server/utils"
)

const (
	API_HTTP_RESULT_ERROR  = -1
	API_HTTP_RESULT_OK     = 0
	API_HTTP_RESULT_FAILED = 1
)

type ChatGPTPermission struct {
	AllowCreateEngine  bool `json:"allow_create_engine"`
	AllowFineTuning    bool `json:"allow_fine_tuning"`
	AllowLogprobs      bool `json:"allow_logprobs"`
	AllowSampling      bool `json:"allow_sampling"`
	AllowSearchIndices bool `json:"allow_search_indices"`
	AllowView          bool `json:"allow_view"`
	//
	Created      uint32 `json:"created"`
	ID           string `json:"id" validate:"required"`
	Object       string `json:"object"`
	IsBlocking   bool   `json:"is_blocking"`
	Organization string `json:"organization"`
}

type ChatGPTModel struct {
	ID         string              `json:"id" validate:"required"`
	Created    uint32              `json:"created"`
	Object     string              `json:"object"`
	Permission []ChatGPTPermission `json:"permission"`
	Root       string              `json:"root"`
	Owned      string              `json:"owned_by"`
}

type ChatGPTRequest struct {
	Prompt   string `json:"prompt"`
	MaxTurns int    `json:"max_turns"`
}

type ChatGPTResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type API_HTTPData struct {
	url        string
	Method     string
	Headers    map[string]string
	Timeout    float64
	SkipVerify bool
	// Request
	Content       string
	ContentLength int
	// Response
	Body   any
	Length int

	//
	ErrorCode    int
	ErrorMessage string
}

func (I *API_HTTPData) data(value any) any {
	switch value.(type) {
	case string:
		{
			var v any
			var text string = value.(string)
			err := json.Unmarshal([]byte(text), &v)
			if err != nil {
				return text
			}
			return v
		}
	case interface{}:
		{
			vv, ok := interface{}(value).([]interface{})
			if !ok {
				return value
			} else {
				vx := map[string]any{}
				for i, v := range vv {
					vx[fmt.Sprintf("%d", i)] = v
				}
			}
		}
	default:
		{
			return nil
		}
	}
	return nil
}

func (I *API_HTTPData) Data() any {
	//
	return I.data(I.Body)
}

func (I API_HTTPData) Get(i any) any {

	v := reflect.ValueOf(&I.Body)
	iv := reflect.ValueOf(i)
	if v.Kind() == reflect.Struct {
		if iv.Kind() != reflect.String {
			return nil
		}

		var count int = v.NumField()
		for n := 0; n < count; n++ {
			field := v.Field(n)
			if !field.IsValid() {
				continue
			}
			vfield := iv.FieldByName(field.String())
			vfield.Set(field.Elem())
		}
	} else if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {

	}
	return nil
}

// HTTP Request
func API_HTTPRequest(base_url string, path string, data *API_HTTPData) *API_HTTPData {
	if data == nil {
		data = &API_HTTPData{
			url:        "",
			Method:     "",
			Headers:    nil,
			Timeout:    5.0,
			SkipVerify: false,
		}
	}

	url, _ := url.JoinPath(base_url, path)
	data.url = url
	if len(data.Method) == 0 {
		data.Method = "GET"
	}

	//
	data.Body = nil
	data.Length = 0
	data.ErrorCode = -1
	data.ErrorMessage = "<null>"

	timeout := 5.0 * 1000
	if data.Timeout > 0.0 {
		timeout = data.Timeout * 1000
	}

	//
	var UserAgent string = `Mozilla/5.0 (%s,%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36`
	var OSName string = runtime.GOOS
	var OSArch string = runtime.GOARCH
	if OSArch == "amd64" {
		OSArch = "x86_64"
	} else if OSArch == "arm64" {
		OSArch = "ARM64"
	} else {
		OSArch = "x86_32"
	}

	user_agent := fmt.Sprintf(UserAgent, OSName, OSArch)

	// Request Method POST:
	if data.Method == "POST" {
		data.ContentLength = len(data.Content)
		if data.ContentLength == 0 {
			data.Content = ""
		}
	}
	//
	var client = &http.Client{
		Timeout: time.Millisecond * time.Duration(timeout),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: data.SkipVerify,
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					cert, err := x509.ParseCertificate(rawCerts[0])
					if err != nil {
						utils.Logger.LogError("(API) TLS Verify Certificate Error: ", err)
						return err
					}

					_, err = cert.Verify(x509.VerifyOptions{})
					if !data.SkipVerify && err != nil {
						utils.Logger.LogError("(API) TLS Verify Certificate Error: ", err)
						return err
					}

					utils.Logger.Log("(API) TLS Verify Certificate : ", cert.NotAfter)
					return nil
				},
			},
		},
	}
	utils.Logger.Log("(API) Request URL: ", data.url)

	//
	request, err := http.NewRequest(data.Method, url, nil)
	if err != nil {
		utils.Logger.LogError("(API) Create Request Error: ", err)
		return nil
	}
	request.Header.Set("Content-Type", "application/json;charset=utf-8")
	request.Header.Set("User-Agent", user_agent)

	if data.Headers != nil {
		for key, val := range data.Headers {
			request.Header.Set(key, val)
		}
	}

	//

	//
	response, err := client.Do(request)
	if err != nil {
		utils.Logger.LogError("(API) Request Error: ", err)
		return nil
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		utils.Logger.LogError("(API) Response Body Error: ", err)
		return nil
	}

	defer response.Body.Close()

	//
	if response.StatusCode != http.StatusOK {
		body := string(bytes)

		data.Body = body
		data.Length = len(body)

		utils.Logger.LogError("(API) Response Body Failed: ",
			fmt.Sprintf("[(%d) Status:%s] ", response.StatusCode, response.Status))

		data.ErrorCode = response.StatusCode
		data.ErrorMessage = response.Status
		return nil
	}

	var body any
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		utils.Logger.LogError("(API) Response Body JSON Format Error: ", err)

		// Text
		body := string(bytes)

		data.Body = body
		data.Length = len(body)

		data.ErrorCode = -2
		data.ErrorMessage = err.Error()
		return nil
	}

	data.Body = body
	data.Length = len(bytes)

	data.ErrorCode = 0
	data.ErrorMessage = ""
	return data
}

var http_base_url = ""
var http_additional_headers map[string]string = map[string]string{}

func API_GPTInit(config Config) {
	http_base_url = config.APIUrl

	http_additional_headers["Openai-Organization"] = config.APIOrganization
	http_additional_headers["Authorization"] = fmt.Sprintf("Bearer %s", config.APIKey)
}

// OpenAI API : Models
// curl https://api.openai.com/v1/models \
// -H "Authorization: Bearer $OPENAI_API_KEY" \
// -H "OpenAI-Organization: YOUR_ORG_ID"
func API_GPTModels() *API_HTTPData {

	//var models []ChatGPTModel

	data := API_HTTPData{
		SkipVerify: true,
		Headers:    http_additional_headers,
		//Body: ChatGPTModel{},
	}
	//data.Get(&models)

	API_HTTPRequest(http_base_url, "/v1/models", &data)
	return &data
}

func API_GPTCompletions() *API_HTTPData {
	data := API_HTTPData{
		Method:     "POST",
		SkipVerify: true,
		Headers:    http_additional_headers,
	}

	API_HTTPRequest(http_base_url, "/v1/chat/completions", &data)
	return &data
}
