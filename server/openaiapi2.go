package server

import (
	"fmt"
	"net/http"

	"mcmcx.com/gpt-server/httpx"
	"mcmcx.com/gpt-server/utils"
)

var aiapi_client *httpx.HTTPClient2 = nil
func API_GPTInit(config Config) bool {

	var additional_headers map[string]string = map[string]string{}
	additional_headers["Openai-Organization"] = config.APIOrganization
	additional_headers["Authorization"] = fmt.Sprintf("Bearer %s", config.APIKey)

	aiapi_client = httpx.NewClient(config.APIUrl, additional_headers)
	if(aiapi_client == nil) {
		return false
	}

	return true
}

// OpenAI API : Models
// curl https://api.openai.com/v1/models \
// -H "Authorization: Bearer $OPENAI_API_KEY" \
// -H "OpenAI-Organization: YOUR_ORG_ID"
func API_GPTModels2() *httpx.HTTPData2 {

	//var models []ChatGPTModel

	data := httpx.HTTPData2{
		SkipVerify: true,
		//Headers:    http_additional_headers,
		//Body: ChatGPTModel{},
	}
	//data.Get(&models)

	aiapi_client.HTTPRequest2("/v1/models", nil, &data)
	return &data
}

func API_GPTCompletions2(payload any, ondata func(int, *[]byte, int, *httpx.HTTPData2)) *httpx.HTTPData2 {
	data := httpx.HTTPData2{
		Method:     http.MethodPost,
		SkipVerify: true,
		//Headers:    http_additional_headers,
		Payload:    payload,
		//
		HasStream: true,
		CallbackStream: func(index int, buffer *[]byte, length int, sender *httpx.HTTPData2) {
			if ondata != nil {
				ondata(index, buffer, length, sender)
			}
		},
	}

	aiapi_client.HTTPRequest2("/v1/chat/completions", nil, &data)

	utils.Logger.Log("(API) Request GPTCompletions (Time: ", data.EndTime(), "ms)")
	return &data
}
