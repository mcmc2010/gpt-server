package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mcmcx.com/gpt-server/utils"
)

type TLoginData struct {
	IDX      string `form:"idx" json:"idx"`
	Username string `form:"name" json:"name"`
	Password string `form:"pass" json:"pass"`
	// Time
	DateTime string `form:"date_time" json:"date_time"`
	// Device ID
	DeviceID string `form:"device_id" json:"device_id"`
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

//
var ErrorLoginFailed = errors.New("checking login information failed")
var ErrorLoginAccountInvalidate = errors.New("login account invalidate")

//
func HandleUserLogin(ctx *gin.Context) {
	result, handler := InitHandler(ctx, &HandlerOptions{HasAuthorization: false})
	if result < 0 {
		return
	}

	var login_data TLoginData = TLoginData{}
	err := handler.GetParamters(&login_data)
	// err := handler.GetData(&login_data)
	if err != nil {
		HandleResultFailed(ctx, -100, err.Error())
		return
	}

	login_data.IDX = strings.TrimSpace(login_data.IDX)
	login_data.DeviceID = strings.TrimSpace(login_data.DeviceID)
	if !utils.CheckAccountIDX(login_data.IDX, 6, 12) {
		HandleResultFailed(ctx, -101, ErrorLoginFailed.Error())
		return
	}

	//User login
	var result_data TLoginResultData = TLoginResultData{
		IDX:      login_data.IDX,
		DateTime: utils.DateFormat(time.Now(), 3),
	}

	result_data.Code = utils.GenerateCode(3)
	result_data.DeviceID = login_data.DeviceID
	
	var token_rand = utils.GenerateCode(0)
	var token_text = fmt.Sprintf("%s_%d_%s", result_data.IDX, utils.GetTimeStamp64(), token_rand)
	result_data.Token = utils.SHA256(token_text)

	ctx.JSON(http.StatusOK, result_data)
}
