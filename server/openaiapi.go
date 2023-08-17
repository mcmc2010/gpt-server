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
	"strconv"
	"strings"
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
	Payload    any
	// Request
	Content       string
	ContentLength int
	// Response
	Body     any
	BodyType string
	Length   int
	Chunked  bool

	//
	ErrorCode    int
	ErrorMessage string
}

func (I *API_HTTPData) OSName() string {
	name := runtime.GOOS
	if strings.ContainsAny(name, "android") {
		name = "Android"
	} else if strings.ContainsAny(name, "darwin") {
		name = "Macintosh"
	} else if strings.ContainsAny(name, "freebsd") || strings.ContainsAny(name, "openbsd") {
		name = "UNIX"
	} else if strings.ContainsAny(name, "linux") {
		name = "Linux"
	} else {
		name = "Windows"
	}
	return name
}

func (I *API_HTTPData) OSArch() string {
	var arch string = runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	} else if arch == "arm64" {
		arch = "ARM64"
	} else if arch == "arm" {
		arch = "ARM"
	} else {
		arch = "x86_32"
	}
	return arch
}

func (I *API_HTTPData) UserAgent() string {
	sys := fmt.Sprintf("(%s, %s)", I.OSName(), I.OSArch())

	user_agent := fmt.Sprintf(`Mozilla/5.0 %s AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36`,
		sys)
	return user_agent
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
	data.Chunked = false
	data.ErrorCode = -1
	data.ErrorMessage = "(null)"

	timeout := 5.0 * 1000
	if data.Timeout > 0.0 {
		timeout = data.Timeout * 1000
	}

	//
	user_agent := data.UserAgent()

	// Request Method POST:
	if data.Method == "POST" {
		switch data.Payload.(type) {
		case string:
			data.Content = data.Payload.(string)
		case int16:
		case int32:
		case int8:
			data.Content = fmt.Sprintf("%d", data.Payload)
		case nil:
			data.Content = ""
		default:
			bytes, err := json.Marshal(data.Payload)
			if err != nil {
				data.Content = fmt.Sprintf("%+v", data.Payload)
			} else {
				data.Content = string(bytes)
			}
		}
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

	var payload io.Reader = nil
	if data.Method == "POST" && data.ContentLength > 0 {
		payload = strings.NewReader(data.Content)
	}

	//
	utils.Logger.Log("(API) Request (", data.Method, ") URL: ", data.url)

	request, err := http.NewRequest(data.Method, url, payload)
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
	response, err := client.Do(request)
	if err != nil {
		utils.Logger.LogError("(API) Request Error: ", err)
		return nil
	}

	var value string = ""
	value = response.Header.Get("Content-Length")
	if len(value) == 0 {
		value = "0"
	}
	content_length, _ := strconv.ParseInt(value, 10, 64)

	var chunked bool = false
	value = response.Header.Get("Transfer-Encoding")
	if len(value) > 0 && value == "chunked" {
		chunked = true
	}

	data.BodyType = "json"
	data.Length = int(content_length)
	data.Chunked = chunked

	if data.Chunked {
		result := API_HTTPReadableStream(response, data)
		if result < 0 {
			return nil
		}
	} else {
		result := API_HTTPReadable(response, data)
		if result < 0 {
			return nil
		}
	}

	//
	if response.StatusCode != http.StatusOK {

		utils.Logger.LogError("(API) Response Body Failed: ",
			fmt.Sprintf("[(%d) Status:%s] ", response.StatusCode, response.Status))

		data.ErrorCode = response.StatusCode
		data.ErrorMessage = response.Status
		return nil
	}

	if data.BodyType != "json" {
		utils.Logger.LogError("(API) Response Body JSON Format Error: ", err)

		data.ErrorCode = -2
		data.ErrorMessage = "JSON Format error"
		return nil
	}

	data.ErrorCode = 0
	data.ErrorMessage = ""
	return data
}

func API_HTTPReadable(response *http.Response, data *API_HTTPData) int {

	//
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		data.ErrorCode = -2
		data.ErrorMessage = err.Error()

		utils.Logger.LogError("(API) Response Body Error: ", err)
		return -1
	}

	defer response.Body.Close()

	length := len(bytes)

	data.BodyType = "json"
	data.Body = nil
	data.Length = length

	var body any = nil
	if length > 0 {
		err = json.Unmarshal(bytes, &body)
		if err != nil {
			text := string(bytes)

			data.Body = text
			data.BodyType = "text"
			data.Length = len(text)
		} else {
			data.Body = body
		}
	}
	return 0
}

func API_HTTPReadableStream(response *http.Response, data *API_HTTPData) int {
	return 0
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

func API_GPTCompletions(payload any) *API_HTTPData {
	data := API_HTTPData{
		Method:     "POST",
		SkipVerify: true,
		Headers:    http_additional_headers,
		Payload:    payload,
	}

	API_HTTPRequest(http_base_url, "/v1/chat/completions", &data)
	return &data
}
