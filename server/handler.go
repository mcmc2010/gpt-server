package server

import (
	"encoding/json"
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
}

func HandleResultFailed(ctx *gin.Context, code int, message string) {
	result := gin.H{
		"error_code":    code,
		"error_message": message,
	}
	if !ctx.IsAborted() {
		ctx.JSON(http.StatusOK, result)
	}
}

func HandleResultFailed2(ctx *gin.Context, data *API_HTTPData) {
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
		if !ctx.Writer.Written() {
			ctx.JSON(http.StatusOK, gin.H{
				"error_code":    data.ErrorCode,
				"error_message": data.ErrorMessage,
				"error":         nil,
			})
		}
	}
}

func HandleOpenAIModels(ctx *gin.Context) {
	data := API_GPTModels()
	if data.ErrorCode != API_HTTP_RESULT_OK {
		HandleResultFailed2(ctx, data)
		return
	}
	ctx.JSON(200, data.Data())
}

func HandleOpenAICompletions(ctx *gin.Context) {
	var bytes []byte = make([]byte, ctx.Request.ContentLength)
	length, err := ctx.Request.Body.Read(bytes)
	if length == 0 || err != nil {
		HandleResultFailed(ctx, -9, "Your HTTP request body is null or not in JSON format. Rejected your request.")
		return
	}
	defer ctx.Request.Body.Close()

	var body any
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		HandleResultFailed(ctx, -9, "Your HTTP request body is null or not in JSON format. Rejected your request.")
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	var data *API_HTTPData
	data = API_GPTCompletions(body, func(index int, buffer *[]byte, length int) {

		if index >= 0 {
			_, err := ctx.Writer.Write(*buffer)
			if err != nil {
				HandleResultFailed(ctx, -10, "Write stream failed.")
				return
			}

			ctx.Writer.Flush()
		} else {
			HandleResultFailed2(ctx, data)
			return
		}
	})
	if data.ErrorCode != API_HTTP_RESULT_OK {
		HandleResultFailed2(ctx, data)
		return
	}

	//ctx.String(http.StatusOK, "")
	<-ctx.Done()
}
