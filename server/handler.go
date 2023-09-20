package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mcmcx.com/gpt-server/utils"
)

func HandlePing(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message":  "pong",
		"time_utc": utils.DateFormat(time.Now().UTC(), 9),
		"time_now": utils.DateFormat(time.Now(), -1),
	})
}

func HandleResultError(ctx *gin.Context, code int, message string) {
	result := gin.H{
		"error_code":    code,
		"error_message": message,
	}
	if !ctx.IsAborted() {
		ctx.JSON(http.StatusBadRequest, result)
	}
	ctx.Abort()
}

func HandleResultFailed(ctx *gin.Context, code int, message string) {
	result := gin.H{
		"error_code":    code,
		"error_message": message,
	}
	if !ctx.IsAborted() {
		ctx.JSON(http.StatusOK, result)
	}
	ctx.Abort()
}

func HandleResultFailed2(ctx *gin.Context, data *API_HTTPData2) {
	if data == nil {
		return
	}

	if data.ErrorCode == http.StatusUnauthorized {
		result := data.Data().(map[string]any)
		if result == nil {
			result = map[string]any{
				"error": nil,
			}
		}
		result["error_code"] = data.ErrorCode
		result["error_message"] = data.ErrorMessage

		if !ctx.IsAborted() {
			ctx.JSON(http.StatusOK, result)
		}
	} else {
		if !ctx.IsAborted() {
			ctx.JSON(http.StatusOK, gin.H{
				"error_code":    data.ErrorCode,
				"error_message": data.ErrorMessage,
				"error":         nil,
			})
		}
	}

	ctx.Abort()
}

func HandleOpenAIModels(ctx *gin.Context) {
	data := API_GPTModels2()
	if data.ErrorCode != API_HTTP_RESULT_OK {
		HandleResultFailed2(ctx, data)
		return
	}
	ctx.JSON(200, data.Data())
}

func HandleOpenAICompletions(ctx *gin.Context) {
	for k, v := range ctx.Request.Header {
		utils.Logger.Log(k, v)
	}

	var content_length int = int(ctx.Request.ContentLength);
	var bytes []byte = make([]byte, content_length)
	length, err := ctx.Request.Body.Read(bytes)
	if(err != nil && err != io.EOF || length < content_length) {
		utils.Logger.LogWarning("request context length:", length, ", Error:", err)
		HandleResultFailed(ctx, -9, "Your HTTP request body is null. Rejected your request.")
		return
	}

	defer ctx.Request.Body.Close()

	var body any
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		utils.Logger.LogWarning("request context length:", length, ", Error:", err)
		HandleResultFailed(ctx, -9, "Your HTTP request body not in JSON format. Rejected your request.")
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	//context := ctx.Request.Context()
	ctx_context, ctx_cancel := context.WithCancel(ctx.Request.Context())

	var data *API_HTTPData2 = nil

	data = API_GPTCompletions2(body, func(index int, buffer *[]byte, length int, sender *API_HTTPData2) {

		if ctx.IsAborted() || sender.ErrorCode != API_HTTP_RESULT_OK {
			ctx_cancel()
			return
		}

		if index < 0 {
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

	if data.ErrorCode != API_HTTP_RESULT_OK {
		HandleResultFailed2(ctx, data)
		return
	}

	//ctx.String(http.StatusOK, "")
	select {
	case <-ctx_context.Done():
		return
	}
}
