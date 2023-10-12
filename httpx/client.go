package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ddliu/go-httpclient"
	"mcmcx.com/gpt-server/utils"
)

const (
	HTTP_RESULT_ERROR   = -1
	HTTP_RESULT_OK      = 0
	HTTP_RESULT_FAILED  = 1
	HTTP_RESULT_INVALID = 2
)

var http_base_url = ""
var http_additional_headers map[string]string = map[string]string{}

type HTTPClient2 struct {
	BaseUrl           string
	AdditionalHeaders map[string]string
	//
	Client   *httpclient.HttpClient
	Response *httpclient.Response
	//
	Data *HTTPData2
}

type HTTPData2 struct {
	url        string
	inner_url string
	Method     string
	Headers    map[string]string
	Timeout    float64
	SkipVerify bool
	//Request
	Payload any
	Length  int
	//Response
	ContentType   string
	Content       any
	ContentLength int
	HasStream     bool

	//Time
	tick         int64
	elapsed_time int //milliseconds

	//
	CallbackStream func(index int, buffer *[]byte, length int, data *HTTPData2)

	//
	ErrorCode    int
	ErrorMessage string
}

func (I *HTTPData2) parse_content(value any) any {

	switch value.(type) {
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

func (I *HTTPData2) Data() any {
	//
	return I.parse_content(I.Content)
}

func (I *HTTPData2) StartTime() {
	I.tick = time.Now().UnixMilli()
	I.elapsed_time = 0
}

func (I *HTTPData2) EndTime() int {
	I.elapsed_time = int(time.Now().UnixMilli() - I.tick)
	return I.elapsed_time
}

func (I *HTTPData2) OSName() string {
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

func (I *HTTPData2) OSArch() string {
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

func (I *HTTPData2) UserAgent() string {
	sys := fmt.Sprintf("(%s, %s)", I.OSName(), I.OSArch())

	user_agent := fmt.Sprintf(`Mozilla/5.0 %s AppleWebKit/537.00 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.00`,
		sys)
	return user_agent
}

// HTTP Request
func (I *HTTPClient2) HTTPRequest2(path string, params map[string]any, data *HTTPData2) *HTTPData2 {
	if data == nil {
		data = &HTTPData2{
			url:        "",
			Method:     "",
			Headers:    nil,
			Timeout:    5.0,
			SkipVerify: false,
			Payload:    nil,
		}
	}

	url, _ := url.JoinPath(I.BaseUrl, path)
	data.url = url
	if len(data.Method) == 0 {
		data.Method = http.MethodGet
	}

	//
	data.Length = 0

	if len(data.ContentType) == 0 {
		data.ContentType = "json"
	}
	data.ContentLength = 0
	data.Content = nil

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

	var body io.Reader = nil
	//Post timeout
	if data.Method == http.MethodPost {
		timeout = 3.0 * 1000

		var payload = ""

		//
		switch data.Payload.(type) {
		case string:
			payload = data.Payload.(string)
		case int16:
		case int32:
		case int8:
			payload = fmt.Sprintf("%d", data.Payload)
		case nil:
			payload = ""
		default:
			bytes, err := json.Marshal(data.Payload)
			if err != nil {
				payload = fmt.Sprintf("%+v", data.Payload)
			} else {
				payload = string(bytes)
			}
		}

		if len(payload) > 0 {
			body = strings.NewReader(payload)
		}
	}

	var client = httpclient.Begin().WithOption(
		httpclient.OPT_DEBUG, true).WithOption(
		httpclient.OPT_USERAGENT, user_agent).WithOption(
		httpclient.OPT_CONNECTTIMEOUT_MS, int(timeout)).WithOption(
		httpclient.OPT_TIMEOUT_MS, 60*1000)

	if data.SkipVerify {
		client.WithOption(httpclient.OPT_UNSAFE_TLS, true)
	}

	client.WithHeader("Content-Type", "application/json;charset=utf-8")
	if data.HasStream {
		client.WithHeader("Accept", "text/event-stream")
	}

	if I.AdditionalHeaders != nil {
		for key, val := range I.AdditionalHeaders {
			client.WithHeader(key, val)
		}
	}
	if data.Headers != nil {
		for key, val := range data.Headers {
			client.WithHeader(key, val)
		}
	}

	// Additional parameters
	if params != nil {
		for key, val := range params {
			if !strings.Contains(url, "?") {
				url = url + "?"
			} else {
				url = url + "&"
			}
			url = url + key
			url = url + "="
			switch val.(type) {
			case int32:
			case int16:
			case int8:
				url = url + fmt.Sprintf("%d", val)
			case float32:
			case float64:
				url = url + fmt.Sprintf("%f", val)
			case bool:
				if val.(bool) {
					url = url + "true"
				} else {
					url = url + "false"
				}
			default:
				text, ok := val.(string)
				if(!ok) {
					text = ""
				}
				url = url + text
			}
		}
	}
	data.inner_url = url

	//
	I.Client = client
	I.Data = data

	//
	utils.Logger.Log("(API) Request (", data.Method, ") URL: ", data.url)

	var response *httpclient.Response
	var err error
	response, err = client.Do(data.Method, data.inner_url, nil, body)
	if err != nil {
		utils.Logger.LogError("(API) Request Error: ", err)

		data.ErrorCode = -1
		data.ErrorMessage = err.Error()

		if strings.Contains(err.Error(), "Client.Timeout") {
			data.ErrorMessage = fmt.Sprintf("Request timeout (%dms)", data.EndTime())
		}
		return nil
	}
	I.Response = response

	//
	var value string = ""
	value = response.Header.Get("Content-Length")
	if len(value) == 0 {
		value = "0"
	}
	content_length, _ := strconv.ParseInt(value, 10, 64)

	value = response.Header.Get("Content-Type")
	if len(value) > 0 && value == "text/event-stream" {
		data.HasStream = true
	}

	value = response.Header.Get("Transfer-Encoding")
	if len(value) > 0 && value == "chunked" {
		//data.ChunkedUsed = true
	}

	data.Length = int(content_length)
	data.ContentType = "json"
	data.Content = nil
	data.ContentLength = 0

	var result = -1
	if data.HasStream {
		result = I.HTTPReadableStream2(response, data)
	} else {
		result = I.HTTPReadable2(response, data)
	}

	if result < 0 {
		return nil
	}

	//
	if response.StatusCode != http.StatusOK {

		utils.Logger.LogError("(API) Response Body Failed: ",
			fmt.Sprintf("[(%d) Status:%s] ", response.StatusCode, response.Status))

		data.ErrorCode = response.StatusCode
		data.ErrorMessage = response.Status
		return nil
	}

	if data.ContentType != "json" {
		utils.Logger.LogError("(API) Response Body JSON Format Error: ", err)

		data.ErrorCode = -2
		data.ErrorMessage = "JSON Format error"
		return nil
	}

	data.ErrorCode = 0
	data.ErrorMessage = ""
	return data
}

func (I *HTTPClient2) HTTPReadable2(response *httpclient.Response, data *HTTPData2) int {

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
	data.Length = length

	data.ContentType = "json"
	data.Content = nil
	data.ContentLength = length

	var body any = nil
	if length > 0 {
		err = json.Unmarshal(bytes, &body)
		if err != nil {
			text := string(bytes)

			data.Content = text
			data.ContentType = "text"
			data.ContentLength = len(text)
		} else {
			data.Content = body
		}
	}
	return 0
}

func (I *HTTPClient2) HTTPReadableStreamChunkedData2(reader *io.Reader, chunks *[]byte) int {
	var result int = 1
	var bytes []byte = make([]byte, 1024) //make([]byte, 4096)
	count, err := (*reader).Read(bytes)
	if err != nil {
		if err == io.EOF {
			result = 0
		} else {
			result = -1
		}
	}

	if count > 0 {
		*chunks = append(*chunks, bytes[0:count]...)
	}

	// 如果数据还没有全部读取，则继续读取下一块数据
	if result < 0 {
		if strings.Contains(err.Error(), "Client.Timeout") {
			return -2
		}
		return -1
	}

	return count
}

func (I *HTTPClient2) HTTPParseStreamChunkedData2(data *[]byte, chunks *[][]byte) int {

	var buffer = *data
	for {
		l := len(buffer)
		pos := bytes.IndexByte(buffer, byte('\n'))
		if pos < 0 {
			break
		}

		pos++
		if pos+1 < l && buffer[pos+1] == '\n' {
			pos++
		}

		var chunk []byte = buffer[0:pos]
		*chunks = append(*chunks, chunk)

		if pos < l {
			buffer = buffer[pos:]
		} else {
			buffer = make([]byte, 0)
		}
	}

	*data = buffer
	var count = len(*chunks)
	return count
}

func (I *HTTPClient2) HTTPReadableStream2Async(response *httpclient.Response, data *HTTPData2) int {

	reader := io.Reader(response.Body)
	defer response.Body.Close()

	if data.Content == nil {
		data.ContentType = "binary"
		data.Content = []byte{}
		data.ContentLength = 0
	}

	//
	var buffer []byte
	var index int = 0
	var length int = 0

	//
	var chunk_data []byte
	var chunk_length int = 0

	//
	var count = 0
	for {

		count = I.HTTPReadableStreamChunkedData2(&reader, &chunk_data)
		if count <= 0 {
			break
		}

		var chunks [][]byte = [][]byte{}
		if I.HTTPParseStreamChunkedData2(&chunk_data, &chunks) > 0 {
			for i := 0; i < len(chunks); i++ {
				chunk_length = len(chunks[i])
				buffer = append(buffer, chunks[i]...)
				length += chunk_length

				if data.CallbackStream != nil {
					data.CallbackStream(index, &chunks[i], chunk_length, data)
				}

				index++
			}
		}
	}

	if count < 0 {
		data.ErrorCode = -2
		data.ErrorMessage = "Read stream chunked error."
		var elapsed_time = data.EndTime()
		if count == -2 {
			//nothing
			data.ErrorMessage = fmt.Sprintf("Read stream chunked timeout. (%dms)", elapsed_time)
		} else {
			utils.Logger.LogError("(API) Response Body Error: ", "Read stream chunked error.")
		}
		if data.CallbackStream != nil {
			if length > 0 {
				data.CallbackStream(-1, &buffer, length, data)
			} else {
				data.CallbackStream(-1, nil, 0, data)
			}
		}
		return -1
	} else if count == 0 {
		if data.CallbackStream != nil {
			data.CallbackStream(0, nil, 0, data)
		}
	}

	data.Content = append(data.Content.([]byte), buffer...)
	data.ContentLength += length

	return 0
}

func (I *HTTPClient2) HTTPReadableStream2(response *httpclient.Response, data *HTTPData2) int {

	go I.HTTPReadableStream2Async(response, data)

	return 0
}

func NewClient(url string, headers map[string]string) *HTTPClient2 {
	client := &HTTPClient2{
		BaseUrl:           url,
		AdditionalHeaders: headers,
	}
	return client
}
