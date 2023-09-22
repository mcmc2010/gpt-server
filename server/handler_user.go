package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TLoginData struct {
	IDX      string `json:"idx"`
	Username string `json:"name"`
	Password string `json:"pass"`
	// Time
	DateTime string `json:"date_time"`
	// Device ID
	DeviceID string `json:"device_id"`
}

type TLoginResultData struct {
	IDX   string `json:"idx"`
	Code  string `json:"auth_code"`
	Token string `json:"auth_token"`
	// Time
	DateTime string `json:"date_time"`
	// Device ID
	DeviceID string `json:"device_id"`
}

func HandleUserLogin(ctx *gin.Context) {
	result, handler := InitHandler(ctx, &HandlerOptions{HasAuthorization: false})
	if result < 0 {
		return
	}

	var login_data TLoginData = TLoginData{}
	err := handler.GetData(&login_data)
	if err != nil {
		HandleResultFailed(ctx, -100, err.Error())
		return
	}

	//User login
	var result_data TLoginResultData = TLoginResultData{
		IDX: login_data.IDX,
	}

	ctx.JSON(http.StatusOK, result_data)
}
