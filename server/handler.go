package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mcmcx.com/gpt-server/utils"
)

func HandlePing(ctx *gin.Context) {
	//
	ctx.JSON(http.StatusOK, gin.H{
		"message":  "pong",
		"time_utc": utils.DateFormat(time.Now().UTC(), 9),
		"time_now": utils.DateFormat(time.Now(), -1),
	})
}

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

func HandleOpenAICompletions(ctx *gin.Context) {
	result, handler := InitHandler(ctx, &HandlerOptions{PrintHeaders: true, DataType: "json", HasAuthorization: true})
	if result < 0 {
		return
	}

	var body = handler.Data

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
