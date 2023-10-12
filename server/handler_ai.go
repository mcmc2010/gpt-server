package server

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"mcmcx.com/gpt-server/httpx"
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

type OPEMAI_MODELS struct {
	Object string              `json:"object"`
	Data   []OPENAI_MODEL_ITEM `json:"data"`
}

var OPENAI_Models *OPEMAI_MODELS = nil

func (I *OPENAI_MODEL_ITEM) Name() string {
	return strings.ReplaceAll(I.ID, "-", " ")
}

func OpenAI_Init(models any) bool {
	if models == nil {
		return false
	}

	data, ok := models.(map[string]any)
	if !ok {
		return false
	}

	//
	bytes, err := json.Marshal(data)
	if err != nil {
		return false
	}

	var m OPEMAI_MODELS
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return false
	}
	OPENAI_Models = &m

	return true
}

func HandleOpenAIModels(ctx *gin.Context) {
	result, _ := InitHandler(ctx, &HandlerOptions{HasAuthorization: false})
	if result < 0 {
		return
	}

	if OPENAI_Models == nil {
		HandleResultFailed(ctx, -1, "Not found openai models")
		return
	}

	data := gin.H{
		"object": "list",
		"data":   OPENAI_Models.Data,
	}
	ctx.JSON(200, data)
}

// {
// 	"model": "gpt-3.5-turbo",
// 	// role (string) Required
// 	// The role of the author of this message. One of system, user, or assistant.
// 	"messages": messages,
// 	"max_tokens": this.TOKENS_MAX, //The maximum number of tokens to generate in the completion.
// 	"temperature": 1, //What sampling temperature to use, between 0 and 2.
// 	"top_p": 1,
// 	//"n": 1, //How many chat completion choices to generate for each input message.
// 	"presence_penalty": 0,
// 	"frequency_penalty": 0,
// 	"stream": stream,
// 	"stop": null,
// 	"user": id,
// }

func HandleOpenAICompletions(ctx *gin.Context) {
	result, handler := InitHandler(ctx, &HandlerOptions{PrintHeaders: true, DataType: "json", HasAuthorization: true})
	if result < 0 {
		return
	}

	body, ok := handler.Data.(map[string]any)
	if !ok {
		HandleResultFailed(ctx, -1, "Request payload data error.")
		return
	}

	body["max_tokens"] = 2048
	body["temperature"] = 1
	body["top_p"] = 1
	body["presence_penalty"] = 0
	body["frequency_penalty"] = 0
	body["stream"] = true
	body["stop"] = nil

	//
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	//context := ctx.Request.Context()
	ctx_context, ctx_cancel := context.WithCancel(ctx.Request.Context())

	var data *httpx.HTTPData2 = nil

	data = API_GPTCompletions2(body, func(index int, buffer *[]byte, length int, sender *httpx.HTTPData2) {

		if (sender.ErrorCode != httpx.HTTP_RESULT_OK) && (index > 0 || index == 0 && length > 0) {
			return
		}

		if ctx.IsAborted() || index < 0 {
			HandleResultFailed2(ctx, sender)
			ctx_cancel()
			return
		}

		if buffer != nil {
			_, err := ctx.Writer.Write(*buffer)
			if err != nil {
				HandleResultFailed(ctx, -10, "Write stream failed.")
				ctx_cancel()
				return
			}
		}

		// End
		if index == 0 && length == 0 {
			ctx.Writer.Flush()
			ctx_cancel()
		}
	})

	//ctx.String(http.StatusOK, "")
	select {
	case <-ctx_context.Done():
		if data.ErrorCode != httpx.HTTP_RESULT_OK {
			HandleResultFailed2(ctx, data)
			return
		}
		return
	}

}
