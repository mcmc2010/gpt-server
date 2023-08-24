package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	//"net/http/httputil"
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
	Body           any
	BodyType       string
	Length         int
	StreamUsed     bool
	ChunkedUsed    bool
	StreamCallback func(index int, buffer *[]byte, length int)

	//Time
	tick         int64
	elapsed_time int //milliseconds

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

func (I *API_HTTPData) StartTime() {
	I.tick = time.Now().UnixMilli()
	I.elapsed_time = 0
}

func (I *API_HTTPData) EndTime() {
	I.elapsed_time = int(time.Now().UnixMilli() - I.tick)
}

func (I *API_HTTPData) UserAgent() string {
	sys := fmt.Sprintf("(%s, %s)", I.OSName(), I.OSArch())

	user_agent := fmt.Sprintf(`Mozilla/5.0 %s AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36`,
		sys)
	return user_agent
}

func (I *API_HTTPData) data(value any) any {
	switch tt := value.(type) {
	case []byte:
		{
			var v any
			err := json.Unmarshal(value.([]byte), &v)
			if err != nil {
				return value
			}
			return v
		}
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
			println("%+v", tt)

			vv, ok := interface{}(value).([]interface{})
			if !ok {
				return value
			} else {
				vx := map[string]any{}
				for i, v := range vv {
					vx[fmt.Sprintf("%d", i)] = v
				}
				return vx
			}
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
			//
			StreamUsed:     false,
			StreamCallback: nil,
		}
	}

	url, _ := url.JoinPath(base_url, path)
	data.url = url
	if len(data.Method) == 0 {
		data.Method = http.MethodGet
	}

	//
	data.Body = nil
	data.Length = 0

	data.ErrorCode = -1
	data.ErrorMessage = "(null)"

	data.StartTime()
	defer data.EndTime()

	timeout := 5.0 * 1000
	if data.Timeout > 0.0 {
		timeout = data.Timeout * 1000
	}

	//
	user_agent := data.UserAgent()

	//
	var dialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	var client = &http.Client{
		Timeout: time.Millisecond * time.Duration(timeout),
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
			DialContext:       dialer.DialContext,
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

	// Request Method POST:
	if data.Method == http.MethodPost {
		//

		//
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

	var payload io.Reader = nil
	if data.Method == http.MethodPost && data.ContentLength > 0 {
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
	if data.StreamUsed {
		request.Header.Set("Accept", "text/event-stream")
	}

	if data.Headers != nil {
		for key, val := range data.Headers {
			request.Header.Set(key, val)
		}
	}

	//response, err := client.Post(url, "application/json;charset=utf-8", payload)
	//response, err := http.DefaultClient.Do(request)
	response, err := client.Do(request)
	if err != nil {
		utils.Logger.LogError("(API) Request Error: ", err)
		return nil
	}

	//
	var value string = ""
	value = response.Header.Get("Content-Length")
	if len(value) == 0 {
		value = "0"
	}
	content_length, _ := strconv.ParseInt(value, 10, 64)

	value = response.Header.Get("Content-Type")
	if len(value) > 0 && value == "text/event-stream" {
		data.StreamUsed = true
	}

	value = response.Header.Get("Transfer-Encoding")
	if len(value) > 0 && value == "chunked" {
		data.ChunkedUsed = true
	}

	data.BodyType = "json"
	data.Length = int(content_length)

	if data.StreamUsed {
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

func API_HTTPReadableStreamChunkedData(reader *io.Reader, chunks *[][]byte) int {
	var result int = 1
	var bytes []byte = make([]byte, 256) //make([]byte, 4096)
	count, err := (*reader).Read(bytes)
	if err != nil {
		if err == io.EOF {
			result = 0
		} else {
			result = -1
		}
	}

	if count > 0 {
		*chunks = append(*chunks, bytes[0:count])
	}

	// 如果数据还没有全部读取，则继续读取下一块数据
	if result > 0 {
		result = API_HTTPReadableStreamChunkedData(reader, chunks)
		if result < 0 {
			return -1
		}
	} else if result < 0 {
		return -1
	}

	return 0
}

func API_HTTPReadableStreamChunked(reader *io.Reader, data *API_HTTPData, buffer *[]byte) int {

	*buffer = make([]byte, 0)

	//
	var chunks [][]byte
	var chunks_count int = 0

	result := API_HTTPReadableStreamChunkedData(reader, &chunks)
	if result < 0 {
		data.ErrorCode = -2
		data.ErrorMessage = "Read stream chunked error."

		utils.Logger.LogError("(API) Response Body Error: ", "Read stream chunked error.")
		return -1
	}

	chunks_count = len(chunks)
	for i := 0; i < chunks_count; i++ {
		temp := chunks[i]
		*buffer = append(*buffer, temp...)
	}

	length := len(*buffer)
	if length > 0 {
		if data.Body == nil {
			data.Length = 0
			data.Body = []byte{}
		}
		data.Body = append(data.Body.([]byte), *buffer...)
		data.Length += length
	}

	return length
}

func API_HTTPReadableStreamAsync(reader *io.Reader, data *API_HTTPData) int {

	//
	var buffer []byte
	var index int = 0
	var length int = 0

	length = API_HTTPReadableStreamChunked(reader, data, &buffer)
	if length < 0 {
		if data.StreamCallback != nil {
			data.StreamCallback(-1, nil, 0)
		}
		return -1
	}
	if data.StreamCallback != nil {
		data.StreamCallback(index, &buffer, length)
	}

	index++
	return 0
}

func API_HTTPReadableStream(response *http.Response, data *API_HTTPData) int {

	reader := io.Reader(response.Body) //httputil.NewChunkedReader(response.Body)

	go API_HTTPReadableStreamAsync(&reader, data)

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

func API_GPTCompletions(payload any, ondata func(int, *[]byte, int)) *API_HTTPData {
	data := API_HTTPData{
		Method:     "POST",
		SkipVerify: true,
		Headers:    http_additional_headers,
		Payload:    payload,
		//
		StreamUsed: true,
		StreamCallback: func(index int, buffer *[]byte, length int) {
			if ondata != nil {
				ondata(index, buffer, length)
			}
		},
	}

	API_HTTPRequest(http_base_url, "/v1/chat/completions", &data)
	return &data
}
