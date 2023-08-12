package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"mcmcx.com/gpt-server/utils"
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
	url     string
	Headers []string

	//
	Body   any
	Length int

	//
	ErrorCode    int
	ErrorMessage string
}

// HTTP Request
func API_HTTPRequest(base_url string, path string, data *API_HTTPData) *API_HTTPData {
	if data == nil {
		data = &API_HTTPData{
			url:     "",
			Headers: nil,
		}
	}

	url, _ := url.JoinPath(base_url, path)
	data.url = url
	data.Body = nil
	data.Length = 0
	data.ErrorCode = -1
	data.ErrorMessage = "<null>"

	//
	var client = &http.Client{}
	utils.Logger.Log("(API) Request URL: ", data.url)

	//
	request, err := http.NewRequest("Get", url, nil)
	if err != nil {
		utils.Logger.LogError("(API) Create Request Error: ", err)
		return nil
	}
	request.Header.Set("Content-Type", "application/json;charset=utf-8")

	count := len(data.Headers)
	if count%2 > 0 {
		utils.Logger.LogError("(API) Create Request Error: Add header error.")
		return nil
	}
	for i := 0; i < count; i += 2 {
		request.Header.Set(data.Headers[i+0], data.Headers[i+1])
	}

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
	if response.StatusCode != 200 {
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

// OpenAI API : Models
func API_GPTModels(config Config) []*ChatGPTModel {

	var models []*ChatGPTModel

	data := API_HTTPData{
		Headers: []string{
			//"Openai-Organization", config.APIOrganization,
			//"Authorization", fmt.Sprintf("Bearer %s", config.APIKey),
		},
	}
	//API_HTTPRequest(config.APIUrl, "/v1/models", &data)
	API_HTTPRequest("https://www.baidu.com", "", &data)

	return models
}

func API_GPTCompletions() {

}
