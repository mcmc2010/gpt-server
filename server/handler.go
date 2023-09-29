package server

import (
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
		"ip_address": ctx.ClientIP(),
		"ip_remote_address": ctx.RemoteIP(),
	})
}
