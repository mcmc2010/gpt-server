package server

import (
	"fmt"
	"net/http"
	"strings"
	
	"mcmcx.com/gpt-server/httpx"
	"mcmcx.com/gpt-server/utils"
)

//	{
//		"created":1651172509,
//		"id":"text-search-babbage-doc-001",
//		"object":"model",
//		"owned_by":"openai-dev",
//		"parent":null,
//		"permission":
//		[
//			{"allow_create_engine":false,"allow_fine_tuning":false,"allow_logprobs":true,"allow_sampling":true,"allow_search_indices":true,"allow_view":true,"created":1695933794,"group":null,"id":"modelperm-s9n5HnzbtVn7kNc5TIZWiCFS","is_blocking":false,"object":"model_permission","organization":"*"}
//		],
//		"root":"text-search-babbage-doc-001"
//	}
type OPENAI_MODEL_ITEM struct {
	Created int64  `json:"created"`
	ID      string `json:"id"`
	Object  string `json:"object"`
	Owned   string `json:"owned_by"`
	//Parent string `json:"parent"`
	Permission []any  `json:"permission"`
	Root       string `json:"root"`
}

func (I *OPENAI_MODEL_ITEM) Name() string {
	return strings.ReplaceAll(I.ID, "-", " ")
}

type OPEMAI_MODELS struct {
	Object string              `json:"object"`
	Data   []OPENAI_MODEL_ITEM `json:"data"`
}

func (I *OPEMAI_MODELS) Find(id string) *OPENAI_MODEL_ITEM {
	var item *OPENAI_MODEL_ITEM = nil
	var count = len(I.Data)
	for i := 0; i < count; i++ {
		var v = I.Data[i]
		if(v.ID == id) {
			item = &v
			break
		}
	}
	return item
}

var aiapi_client *httpx.HTTPClient2 = nil

func API_GPTInit(config Config) bool {

	var additional_headers map[string]string = map[string]string{}
	additional_headers["Openai-Organization"] = config.APIOrganization
	additional_headers["Authorization"] = fmt.Sprintf("Bearer %s", config.APIKey)

	aiapi_client = httpx.NewClient(config.APIUrl, additional_headers)
	if aiapi_client == nil {
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
		Payload: payload,
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
