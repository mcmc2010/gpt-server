package server

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"mcmcx.com/gpt-server/httpx"
	"mcmcx.com/gpt-server/utils"
)

var OPENAI_Models *OPEMAI_MODELS = nil

// /
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

// https://platform.openai.com/docs/models/continuous-model-upgrades
// gpt-3.5-turbo  | Currently points to gpt-3.5-turbo-0613. | 4,096 tokens	| Up to Sep 2021
// gpt-4	      | Currently points to gpt-4-0613.         | 8,192 tokens	| Up to Sep 2021
// gpt-4-32k	  | Currently points to gpt-4-32k-0613.     | 32,768 tokens	| Up to Sep 2021
// curl https://api.openai.com/v1/chat/completions \
//   -H "Content-Type: application/json" \
//   -H "Authorization: Bearer $OPENAI_API_KEY" \
//   -d '{
//     "model": "gpt-3.5-turbo",
// 	   // role (string) Required
// 	   // The role of the author of this message. One of system, user, or assistant.
//     "messages": [
//       {
//         "role": "system",
//         "content": "You are a helpful assistant."
//       },
//       {
//         "role": "user",
//         "content": "Who won the world series in 2020?"
//       }
//     ]
// 	   "max_tokens": TOKENS_MAX, //The maximum number of tokens to generate in the completion.
// 	   "temperature": 1, //What sampling temperature to use, between 0 and 2.
// 	   "top_p": 1,
// 	   //"n": 1, //How many chat completion choices to generate for each input message.
// 	   "presence_penalty": 0,
// 	   "frequency_penalty": 0,
// 	   "stream": stream,
// 	   "stop": null,
// 	   "user": id,
//   }'
// curl https://api.openai.com/v1/images/generations \
//   -H "Content-Type: application/json" \
//   -H "Authorization: Bearer $OPENAI_API_KEY" \
//   -d '{
//     "model": "dall-e-3",
//     "prompt": "a white siamese cat",
//     "n": 1,
//     "size": "1024x1024"
//   }'

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

	//
	body["temperature"] = 1
	body["top_p"] = 1
	body["presence_penalty"] = 0
	body["frequency_penalty"] = 0
	body["stream"] = true
	body["stop"] = nil

	id, ok := body["user"].(string)
	if(!ok) {
		id = "id0000"
	}
	id = strings.ToLower(strings.TrimSpace(id))

	// Checking models
	// ChatGPT-3 : 4096 tokens
	body["max_tokens"] = 2048
	model_id, ok := body["model"].(string)
	if !ok {
		model_id = "gpt-3.5-turbo"
	}

	model_id = strings.ToLower(strings.TrimSpace(model_id))
	var item = OPENAI_Models.Find(model_id)
	if(item == nil) {
		model_id = "gpt-3.5-turbo"
		item = OPENAI_Models.Find(model_id)
	}

	// ChatGPT-4 : 8192 tokens
	if strings.Contains(model_id, "gpt-4") {
		body["max_tokens"] = 4096
	}

	body["model"] = model_id
	utils.Logger.Log("[AI] Completions (Model:", item.ID, ", ID:", id, ")")

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
