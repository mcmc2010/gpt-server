package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
