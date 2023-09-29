package server

import (
	"context"

	"github.com/gin-gonic/gin"
)

func HandleOpenAIModels(ctx *gin.Context) {
	result, _ := InitHandler(ctx, &HandlerOptions{HasAuthorization: false})
	if result < 0 {
		return
	}

	data := API_GPTModels2()
	if data.ErrorCode != API_HTTP_RESULT_OK {
		HandleResultFailed2(ctx, data)
		return
	}
	ctx.JSON(200, data.Data())
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
	if(!ok) {
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

	var data *API_HTTPData2 = nil

	data = API_GPTCompletions2(body, func(index int, buffer *[]byte, length int, sender *API_HTTPData2) {

		if (sender.ErrorCode != API_HTTP_RESULT_OK) && (index > 0 || index == 0 && length > 0) {
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
		if data.ErrorCode != API_HTTP_RESULT_OK {
			HandleResultFailed2(ctx, data)
			return
		}
		return
	}

}
