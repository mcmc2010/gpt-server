package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/maps"
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
	Context       *gin.Context
	Timestamp     uint32
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
}

// API: Authorization
type TAuthorizationData struct {
	IDX       string `form:"idx"`
	AuthCode  string `form:"auth_code"`
	AuthToken string `form:"auth_token"`
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

func (I *Handler) Init(ctx *gin.Context, options *HandlerOptions) int {
	I.Context = ctx
	I.Timestamp = uint32(I.TimeStamp64())

	//
	I.RemoteAddress = ctx.RemoteIP()
	if len(ctx.ClientIP()) > 0 {
		I.RemoteAddress = ctx.ClientIP()
	}
	//
	I.UserAgent = ctx.Request.UserAgent()
	//
	I.Headers = map[string][]string{}
	maps.Copy(I.Headers, ctx.Request.Header)

	var authorization_data TAuthorizationData = TAuthorizationData{}
	var authorization_text = I.GetHeader("Authorization", "")
	if len(authorization_text) == 0 {
		err := ctx.ShouldBind(&authorization_data)
		if err != nil {
			I.Error = err
		} else {
			authorization_text = fmt.Sprintf("%s-%s", authorization_data.IDX, authorization_data.AuthToken)
		}
	}

	// Default response headers
	I.Context.Header("Content-Type", "application/json;charset=utf-8")

	//
	I.ContentLength = 0
	I.Data = nil
	I.DataType = "none"
	I.Length = 0
	defer ctx.Request.Body.Close()

	//POST method read all payload
	if I.Context.Request.Method == http.MethodPost {
		I.ContentLength = int(ctx.Request.ContentLength)
		var bytes []byte = make([]byte, I.ContentLength)
		length, err := ctx.Request.Body.Read(bytes)
		if err != nil && err != io.EOF {
			I.Error = err
		} else {
			I.Length = length
			I.Data = []byte{}
			I.DataType = "binary"
			if length > 0 {
				I.Data = bytes
			}
		}

		if I.Length > 0 {
			if I.Length < I.ContentLength {
				utils.Logger.LogWarning("Request context length:", I.ContentLength, "Receive length:", I.Length)
			}

			//Parse data
			var payload any
			err = json.Unmarshal(bytes, &payload)
			if err == nil {
				I.Data = payload
				I.DataType = "json"
			}
		}
	}

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
		return -10, errors.New("Not Authorization Data")
	}

	var values []string = strings.Split(text, "-")
	if len(values) == 0 {
		return -11, errors.New("Authorization Data Invalidate")
	}

	if len(values) >= 2 {
		data.IDX = strings.TrimSpace(values[0])
		data.AuthToken = strings.TrimSpace(values[1])
	} else {
		data.IDX = strings.TrimSpace(values[0])
	}

	if !utils.CheckAccountIDX(data.IDX, 6, 12) {
		return -11, errors.New("Authorization Data Invalidate")
	}

	return 0, nil
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
