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
	IDX      uint32 `form:"idx" json:"idx"`
	Username string `form:"name" json:"name"`
	Password string `form:"pass" json:"pass"`
	// Timestamp
	Timestamp int64 `form:"timestamp" json:"timestamp"`
	// Device UID
	DeviceUID string `form:"device_uid" json:"device_uid"`
}

type TLoginResultData struct {
	IDX   uint32 `json:"idx"`
	Code  string `json:"auth_code"`
	Token string `json:"auth_token"`
	// Time
	Timestamp  int64 `json:"timestamp"`
	CreateTime string `json:"create_time"`
	// Device ID
	IPAddress string `json:"ip_address"`
	DeviceUID string `json:"device_uid"`
}

type TAuthResultData struct {
	IDX   uint32 `json:"idx"`
	Code  string `json:"auth_code"`
	Token string `json:"auth_token"`
	// Time
	Timestamp int64 `json:"timestamp"`
	// Device ID
	IPAddress string `json:"ip_address"`
	//DeviceUID string `json:"device_uid"`
}

var ErrorLoginFailed = errors.New("checking login information failed")
var ErrorLoginAccountInvalidate = errors.New("login account invalidate")

func HandleUserLogin(ctx *gin.Context) {
	result, handler := InitHandler(ctx, &HandlerOptions{HasAuthorization: false})
	if result < 0 {
		return
	}

	var err error = nil
	var login_data TLoginData = TLoginData{}
	if(handler.Method == http.MethodPost) {
		err = handler.GetData(&login_data)
	} else {
		err = handler.GetParamters(&login_data)
	}

	// 
	if err != nil {
		HandleResultFailed(ctx, -100, err.Error())
		return
	}

	login_data.IDX = login_data.IDX
	login_data.DeviceUID = strings.TrimSpace(login_data.DeviceUID)
	if !utils.CheckAccountIDX(fmt.Sprintf("%d", login_data.IDX), 6, 12) {
		HandleResultFailed(ctx, -101, ErrorLoginFailed.Error())
		return
	}

	//User login
	var result_data TLoginResultData = TLoginResultData{
		IDX:        login_data.IDX,
		Timestamp:int64(utils.GetTimeStamp64()),
		CreateTime: utils.DateFormat(time.Now(), 3),
	}

	result_data.DeviceUID = login_data.DeviceUID
	result_data.IPAddress = handler.AuthorizationData.IPAddress

	//
	result_data.Code = utils.GenerateCode(3)

	var token_rand = utils.GenerateCode(0)
	var token_text = fmt.Sprintf("%s_%d_%s", result_data.IDX, utils.GetTimeStamp64(), token_rand)
	result_data.Token = utils.SHA256(token_text)


	ctx.JSON(http.StatusOK, result_data)
}

func HandleUserAuth(ctx *gin.Context) {
	result, handler := InitHandler(ctx, &HandlerOptions{HasAuthorization: true})
	if result < 0 {
		return
	}

	var result_data TAuthResultData = TAuthResultData{
		IDX:       handler.AuthorizationData.IDX,
		Code:      handler.AuthorizationData.AuthCode,
		Token:     handler.AuthorizationData.AuthToken,
		Timestamp: int64(utils.GetTimeStamp64()),
	}
	result_data.IPAddress = handler.AuthorizationData.IPAddress

	ctx.JSON(http.StatusOK, result_data)
}
