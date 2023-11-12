package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/maps"
	database_redis "mcmcx.com/gpt-server/database/redis"
	"mcmcx.com/gpt-server/utils"
)

type HandlerOptions struct {
	//
	HasAuthorization bool

	//
	DataType string

	//
	PrintHeaders   bool
	PrintUserAgent bool
}

type Handler struct {

	//
	Method        string
	Context       *gin.Context
	Timestamp     int64
	RemoteAddress string
	UserAgent     string
	Headers       map[string][]string
	//
	ContentLength int
	Data          any
	DataType      string
	Length        int

	//
	Error error

	//
	AuthorizationData *TAuthorizationData
}

// API: Authorization
type TAuthorizationData struct {
	IDX       utils.TIDX `form:"idx" json:"idx"`
	AuthCode  string     `form:"auth_code" json:"auth_code"`
	AuthToken string     `form:"auth_token" json:"auth_token"`
	//
	AuthTime    string
	IPAddress   string
	IPLocalized string
	DeviceUID   string
}

type DBAuthorizationData struct {
	IDX   utils.TIDX `json:"idx"`
	Code  string     `json:"auth_code"`
	Token string     `json:"auth_token"`
	//
	CreateTime string `json:"create_time"`
	AuthTime   string `json:"auth_time"`
	AuthCount  int    `json:"auth_count"`
	// Device ID
	IPAddress   string `json:"ip_address"`
	IPLocalized string `json:"ip_localized"`
	DeviceUID   string `json:"device_uid"`
}

func (I *Handler) TimeStamp() uint32 {
	return uint32(time.Now().Unix())
}

func (I *Handler) TimeStamp64() uint64 {
	return uint64(math.Round(float64(time.Now().UnixMicro()) * 0.001))
}

func (I *Handler) GetHeader(key string, def string) string {
	if I.Headers == nil {
		return def
	}
	var value = def
	for k, v := range I.Headers {
		if strings.ToLower(k) == strings.TrimSpace(strings.ToLower(key)) {
			if len(v) > 0 {
				value = v[0]
			}
			break
		}
	}
	return value
}

func (I *Handler) GetParamters(paramters any) error {
	return I.Context.ShouldBindQuery(paramters)
}

func (I *Handler) GetData(data any) error {
	if I.DataType == "json" {
		bytes, err := json.Marshal(I.Data)
		if err != nil {
			return err
		}
		return json.Unmarshal(bytes, data)
	}

	return errors.New("Data not support format")
}

func (I *Handler) InitData() error {

	//
	I.ContentLength = 0
	I.Data = nil
	I.DataType = "none"
	I.Length = 0

	var encoding = I.GetHeader("Accept-Encoding", "")
	if len(encoding) > 0 {
		encoding = strings.ToLower(strings.TrimSpace(encoding))
		var values = strings.Split(encoding, ",")
		if len(values) > 0 {
			encoding = strings.TrimSpace(values[0])
		}
	}

	//
	defer I.Context.Request.Body.Close()

	//POST method read all payload
	if I.Context.Request.Method != http.MethodPost {
		return nil
	}

	I.ContentLength = int(I.Context.Request.ContentLength)

	//var err error = nil
	var length = 0
	//var buffer []byte = make([]byte, 0)
	buffer, err := io.ReadAll(I.Context.Request.Body)
	if err != nil && err != io.EOF {
		return err
	}
	length = len(buffer)
	// for length < I.ContentLength {
	// 	var temp []byte = make([]byte, 1024)
	// 	count, err := I.Context.Request.Body.Read(temp)
	// 	if err != nil && err != io.EOF {
	// 		return err
	// 	}
	// 	if(count == 0) {
	// 		break;
	// 	}

	// 	length += count
	// 	buffer = append(buffer, temp[0:count]...)
	// }
	//buffer = buffer[0:length]

	// Incomplete data received due to network issues.
	if err == io.EOF && I.ContentLength > length {
		return err
	} else if I.ContentLength > length && err == nil && encoding == "gzip" {
		utils.Logger.Log("Compress Buffer: ", utils.BinaryToHexString(buffer, 16))
		reader, err := gzip.NewReader(bytes.NewReader(buffer))
		if err != nil {
			utils.Logger.LogWarning("Buffer length:", length, " GZIP Uncompress error:", err.Error())
			return err
		}

		buffer = make([]byte, I.ContentLength+2)
		length, err = reader.Read(buffer)
		if err != nil {
			utils.Logger.LogWarning("Buffer length:", length, " GZIP Uncompress error:", err.Error())
			return err
		}

		buffer = buffer[0:length]
		reader.Close()
	}

	I.Length = length
	I.Data = []byte{}
	I.DataType = "binary"

	if I.Length > 0 {
		I.Data = buffer

		if I.Length < I.ContentLength {
			utils.Logger.LogWarning("Request context length:", I.ContentLength, " Receive length:", I.Length)
		}

		//Parse data
		var payload any
		err = json.Unmarshal(buffer, &payload)
		if err == nil {
			I.Data = payload
			I.DataType = "json"
		}
	}
	return nil
}

func (I *Handler) Init(ctx *gin.Context, options *HandlerOptions) int {
	I.Context = ctx
	I.Method = I.Context.Request.Method
	I.Timestamp = int64(I.TimeStamp64())
	I.AuthorizationData = nil

	//
	I.RemoteAddress = ctx.RemoteIP()
	if len(ctx.ClientIP()) > 0 {
		I.RemoteAddress = ctx.ClientIP()
	}
	ip := net.ParseIP(I.RemoteAddress)
	ip = ip.To4()
	if ip != nil {
		I.RemoteAddress = ip.String()
	}

	//
	I.UserAgent = ctx.Request.UserAgent()
	//
	I.Headers = map[string][]string{}
	maps.Copy(I.Headers, ctx.Request.Header)

	var authorization_data TAuthorizationData = TAuthorizationData{}
	var authorization_text = I.GetHeader("Authorization", "")
	if len(authorization_text) == 0 {
		err := I.GetParamters(&authorization_data)
		if err != nil {
			I.Error = err
		} else {
			authorization_text = fmt.Sprintf("%d-%s", authorization_data.IDX, authorization_data.AuthToken)
		}
	}

	// Default response headers
	I.Context.Header("Content-Type", "application/json;charset=utf-8")

	//
	I.Error = I.InitData()

	//
	if options != nil && len(options.DataType) > 0 && options.DataType != I.DataType {
		I.Error = errors.New("request body not (" + options.DataType + ") support format")
	}

	if options != nil && options.PrintHeaders {
		I.PrintHeaders()
	}
	if options != nil && options.PrintUserAgent {
		I.PrintUserAgent()
	}

	//
	if I.Error == nil && options != nil && options.HasAuthorization {
		_, I.Error = I.Authorization(authorization_text, &authorization_data)
		if I.Error == nil {
			I.AuthorizationData = &authorization_data
		}
	}

	if I.Error != nil {
		return -1
	}
	return 0
}

func (I *Handler) ResultFailed(code int, message string) int {

	if I.Context.IsAborted() {
		return -1
	}

	result := gin.H{
		"error_code":    code,
		"error_message": message,
	}

	I.Context.JSON(http.StatusOK, result)
	return 0
}

func (I *Handler) ResultError(code int, message string) int {

	if I.Context.IsAborted() {
		return -1
	}

	result := gin.H{
		"error_code":    code,
		"error_message": message,
	}

	I.Context.JSON(http.StatusBadRequest, result)
	return 0
}

func (I *Handler) PrintHeaders() {
	utils.Logger.Log("Request Headers :")
	for k, v := range I.Headers {
		utils.Logger.Log(k, v)
	}
}

func (I *Handler) PrintUserAgent() {
	utils.Logger.Log("UserAgent :", I.UserAgent)
}

func (I *Handler) Authorization(text string, data *TAuthorizationData) (int, error) {
	if len(text) == 0 {
		return -10, errors.New("not authorization data")
	}

	var values []string = strings.Split(text, "-")
	if len(values) == 0 {
		return -11, errors.New("authorization data invalidate")
	}

	idx, err := strconv.ParseInt(strings.TrimSpace(values[0]), 10, 32)
	if err != nil {
		return -11, errors.New("authorization data invalidate")
	}

	data.IDX = utils.TIDX(idx)
	if len(values) >= 2 {
		data.AuthToken = strings.TrimSpace(values[1])
	}
	if !utils.CheckAccountIDX(data.IDX, 6, 12) || !utils.CheckToken(data.AuthToken) {
		return -11, errors.New("authorization data invalidate")
	}

	//
	data.AuthTime = utils.DateFormat(time.Now(), 3)
	data.IPAddress = I.RemoteAddress
	data.IPLocalized = IPLocalized(I.RemoteAddress).Localize()
	result, db_data := db_auth_data_verfiy(data)
	if result < 0 {
		return -12, errors.New("authorization data expiration or expiration")
	}

	//
	data.DeviceUID = db_data.DeviceUID

	//
	if result >= 1 {
		data.AuthCode = db_data.Code
	}

	//
	return 0, nil
}

func db_auth_data_verfiy(data *TAuthorizationData) (int, *DBAuthorizationData) {
	var db_id = fmt.Sprintf("auth_user_%d_%s", data.IDX, data.AuthToken)
	var db_auth_data DBAuthorizationData
	if !database_redis.GetJson(db_id, &db_auth_data, false) || db_auth_data.IDX != data.IDX {
		return -1, nil
	}

	var result = 0
	if len(data.AuthToken) > 0 && db_auth_data.Token == data.AuthToken {
		result = 1
	}
	if len(data.AuthCode) > 0 && db_auth_data.Code == data.AuthCode {
		result = 2
	}

	//IPAddress, Device
	//None

	db_auth_data.AuthCount++
	db_auth_data.AuthTime = utils.DateFormat(time.Now(), 3)
	if !database_redis.PushJson[DBAuthorizationData](db_id, &db_auth_data, database_redis.KEEP_TIME, false) {
		return -2, nil
	}
	return result, &db_auth_data
}

func InitHandler(ctx *gin.Context, options *HandlerOptions) (int, *Handler) {
	handler := &Handler{}
	result := handler.Init(ctx, options)
	if result < 0 {
		handler.ResultError(-1, "Init handler error: "+handler.Error.Error())
		return result, handler
	}
	return 0, handler
}
