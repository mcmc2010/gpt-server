package server

import (
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

func HandleOpenAIModels(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "hello",
	})
}

func HandleOpenAICompletions(ctx *gin.Context) {

}
