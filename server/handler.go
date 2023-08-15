package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"mcmcx.com/gpt-server/utils"
)

func HandlePing(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":  "pong",
		"time_utc": utils.DateFormat(time.Now().UTC(), 9),
		"time_now": utils.DateFormat(time.Now(), -1),
	})
}

func HandleResultFailed(ctx *gin.Context, data *API_HTTPData) {
	if data.ErrorCode == http.StatusUnauthorized {
		result := data.Data().(map[string]any)
		if result == nil {
			result = map[string]any{
				"error": nil,
			}
		}
		result["error_code"] = data.ErrorCode
		result["error_message"] = data.ErrorMessage

		ctx.JSON(200, result)
	} else {
		ctx.JSON(200, gin.H{
			"error_code":    data.ErrorCode,
			"error_message": data.ErrorMessage,
			"error":         nil,
		})
	}
}

func HandleOpenAIModels(ctx *gin.Context) {
	data := API_GPTModels()
	if data.ErrorCode != API_HTTP_RESULT_OK {
		HandleResultFailed(ctx, data)
		return
	}
	ctx.JSON(200, data.Data())
}

func HandleOpenAICompletions(ctx *gin.Context) {
	data := API_GPTCompletions()
	if data.ErrorCode != API_HTTP_RESULT_OK {
		HandleResultFailed(ctx, data)
		return
	}

	ctx.JSON(200, data.Data())
}
