package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	database_redis "mcmcx.com/gpt-server/database/redis"
	"mcmcx.com/gpt-server/utils"
)

type TLoginData struct {
	IDX      utils.TIDX `form:"idx" json:"idx"`
	Username string     `form:"name" json:"name"`
	Password string     `form:"pass" json:"pass"`
	// Timestamp
	Timestamp int64 `form:"timestamp" json:"timestamp"` //client timestamp
	// Device UID
	DeviceUID string `form:"device_uid" json:"device_uid"`
}

type TLoginResultData struct {
	IDX   utils.TIDX `json:"idx"`
	Code  string     `json:"auth_code"`
	Token string     `json:"auth_token"`
	// Time
	Timestamp  int64  `json:"timestamp"` //server timestamp
	CreateTime string `json:"create_time"`
	// Device ID
	IPAddress string `json:"ip_address"`
	DeviceUID string `json:"device_uid"`
}

type DBLoginData struct {
	IDX utils.TIDX `json:"idx"`
	//
	Code  string `json:"auth_code"`
	Token string `json:"auth_token"`
	//
	CreateTime string `json:"create_time"`
	// Device ID
	IPAddress string `json:"ip_address"`
	DeviceUID string `json:"device_uid"`
}

type DBLoginDataSet struct {
	IDX       utils.TIDX             `json:"idx"`
	Timestamp int64                  `json:"timestamp"`
	List      map[string]DBLoginData `json:"list"`
}

type TUserAuthData struct {
	IDX  utils.TIDX `form:"idx" json:"idx"`
	Code string     `json:"code"`
	// Timestamp
	Timestamp int64 `form:"timestamp" json:"timestamp"` //client timestamp
	// Device UID
	DeviceUID string `form:"device_uid" json:"device_uid"`
}

type TUserAuthResultData struct {
	IDX   utils.TIDX `json:"idx"`
	Code  string     `json:"auth_code"`
	Token string     `json:"auth_token"`
	// Time
	Timestamp int64 `json:"timestamp"`
	// Device ID
	IPAddress string `json:"ip_address"`
	//DeviceUID string `json:"device_uid"`
}

var ErrorLoginFailed = errors.New("checking login information failed")
var ErrorLoginAccountInvalidate = errors.New("login account invalidate")
var ErrorLoginError = errors.New("login internal error")

func HandleUserLogin(ctx *gin.Context) {
	result, handler := InitHandler(ctx, &HandlerOptions{HasAuthorization: false})
	if result < 0 {
		return
	}

	var err error = nil
	var login_data TLoginData = TLoginData{}
	if handler.Method == http.MethodPost {
		err = handler.GetData(&login_data)
	} else {
		err = handler.GetParamters(&login_data)
	}

	//
	if err != nil {
		HandleResultFailed(ctx, -100, err.Error())
		return
	}

	login_data.DeviceUID = strings.TrimSpace(login_data.DeviceUID)
	if !utils.CheckAccountIDX(login_data.IDX, 6, 12) {
		HandleResultFailed(ctx, -101, ErrorLoginFailed.Error())
		return
	}

	//User login
	var result_data TLoginResultData = TLoginResultData{
		IDX:        login_data.IDX,
		Timestamp:  int64(utils.GetTimeStamp64()),
		CreateTime: utils.DateFormat(time.Now(), 3),
	}

	result_data.DeviceUID = login_data.DeviceUID
	result_data.IPAddress = handler.RemoteAddress

	//
	result_data.Code = utils.GenerateCode(3)

	var token_rand = utils.GenerateCode(0)
	var token_text = fmt.Sprintf("%d_%d_%s", result_data.IDX, utils.GetTimeStamp64(), token_rand)
	result_data.Token = utils.SHA256(token_text)

	//
	if db_login_data_add(&result_data) == nil {
		HandleResultFailed(ctx, -102, ErrorLoginError.Error())
		return
	}

	utils.Logger.LogWarning("[Login] Username:", login_data.Username,
		" Result (", result_data.Code, ", ", result_data.Token, ")",
		" IPAddress (", result_data.IPAddress, ",", result_data.DeviceUID, ")")

	ctx.JSON(http.StatusOK, result_data)
}

func db_login_data_add(data *TLoginResultData) *DBLoginDataSet {

	var db_id = fmt.Sprintf("login_user_%d", data.IDX)
	var db_login_set DBLoginDataSet
	if !database_redis.GetJson(db_id, &db_login_set, false) {
		db_login_set = DBLoginDataSet{
			IDX:       data.IDX,
			Timestamp: int64(utils.GetTimeStamp64()),
			List:      map[string]DBLoginData{},
		}
	}

	var val DBLoginData
	var key = data.DeviceUID
	val, ok := db_login_set.List[key]
	if !ok {
		val = DBLoginData{
			IDX:        data.IDX,
			CreateTime: utils.DateFormat(time.Now(), 3),
			DeviceUID:  data.DeviceUID,
		}
	}

	val.Code = data.Code
	val.Token = data.Token
	val.IPAddress = data.IPAddress

	db_login_set.List[key] = val

	// Save login data
	if !database_redis.PushJson[DBLoginDataSet](db_id, &db_login_set, database_redis.KEEP_TIME, false) {
		return nil
	}

	// generate authorization data
	var auth_data DBAuthorizationData = DBAuthorizationData{
		IDX:        val.IDX,
		Code:       val.Code,
		Token:      val.Token,
		CreateTime: val.CreateTime,
		DeviceUID:  val.DeviceUID,
		IPAddress:  val.IPAddress,
	}
	auth_data.AuthTime = utils.DateFormat(time.Now(), 3)
	auth_data.AuthCount = 1

	var db_auth_id = fmt.Sprintf("auth_user_%d_%s", data.IDX, auth_data.Token)
	if !database_redis.PushJson[DBAuthorizationData](db_auth_id, &auth_data, database_redis.KEEP_TIME, false) {
		return nil
	}

	return &db_login_set
}

func HandleUserAuth(ctx *gin.Context) {
	result, handler := InitHandler(ctx, &HandlerOptions{HasAuthorization: true})
	if result < 0 {
		return
	}

	//
	var err error = nil
	var auth_data TUserAuthData = TUserAuthData{}
	if handler.Method == http.MethodPost {
		err = handler.GetData(&auth_data)
	} else {
		err = handler.GetParamters(&auth_data)
	}

	//
	if err != nil {
		HandleResultFailed(ctx, -100, err.Error())
		return
	}

	auth_data.Code = strings.TrimSpace(auth_data.Code)
	if auth_data.Code != handler.AuthorizationData.AuthCode {
		HandleResultFailed(ctx, -101, errors.New("user authorization failed").Error())
		return
	}

	utils.Logger.LogWarning("[Auth] IDX:", auth_data.IDX,
		" Result (OK)",
		" IPAddress (", handler.AuthorizationData.IPAddress, ",", handler.AuthorizationData.DeviceUID, ")")

	var result_data TUserAuthResultData = TUserAuthResultData{
		IDX:       handler.AuthorizationData.IDX,
		Code:      handler.AuthorizationData.AuthCode,
		Token:     handler.AuthorizationData.AuthToken,
		Timestamp: int64(utils.GetTimeStamp64()),
	}
	result_data.IPAddress = handler.AuthorizationData.IPAddress

	ctx.JSON(http.StatusOK, result_data)
}
