package database_redis

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port      int    `yaml:"redis_port" json:"redis_port"`
	Address   string `yaml:"redis_address" json:"redis_address"`
	Username  string `yaml:"redis_user" json:"redis_user"`
	Password  string `yaml:"redis_pass" json:"redis_pass"`
	UseTLS    bool
	TLSKey    string `yaml:"redis_tls_key" json:"redis_tls_key"`
	TLSCrt    string `yaml:"redis_tls_crt" json:"redis_tls_crt"`
	TLSCA     string `yaml:"redis_tls_ca" json:"redis_tls_ca"`
	TLSConfig *tls.Config
}

const KEEP_TIME = redis.KeepTTL

var _instance *redis.Client = nil
var _status = make(chan int, 0)
var _status_closing bool = false

func Instance() *redis.Client {
	return _instance
}

func Release() int {
	_status_closing = true

	//
	status := <-_status

	//
	println("[Redis] (work) released ")
	return status
}

// Only one instance
func NewAndInitialize(info *Config) *redis.Client {
	if _instance != nil {
		return _instance
	}

	//
	var address = fmt.Sprintf("%s:%d", info.Address, info.Port)

	//
	var options = redis.Options{
		// Connection
		Network:  "tcp",
		Addr:     address,
		Username: info.Username,
		Password: info.Password,
		DB:       0,
		//
		PoolSize: 15,
		//
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		//
		TLSConfig: info.TLSConfig,
	}
	_instance = redis.NewClient(&options)
	if _instance == nil {
		return nil
	}

	//
	redis_ping()

	//
	go redis_update()
	_status <- 1
	println("[Redis] (work) starting ")
	return _instance
}

func timestamp32() uint32 {
	return uint32(time.Now().Unix())
}

func timestamp64() uint64 {
	return uint64(math.Round(float64(time.Now().UnixMicro()) * 0.001))
}

func keep_time(value float32) time.Duration {
	var keep time.Duration = KEEP_TIME
	if value > 0 {
		keep = time.Duration(float64(value) * float64(time.Second))
	} else if value == 0.0 {
		keep = 0
		//Max date time : 9999 days
	} else if value == -1.0 {
		keep = time.Duration(9999 * 24 * 60 * 60 * time.Second)
	}
	return keep
}

func redis_update() {
	var tick = timestamp64()

	var status = <-_status
	println("[Redis] (update) starting ", status)

	var ping_time = 0
	var flag = 1
	for flag > 0 {
		var idle = int(timestamp64() - tick)
		tick = timestamp64()

		ping_time += idle
		if ping_time >= 30*1000 {
			redis_ping()
			ping_time = 0
		}

		time.Sleep(20 * time.Millisecond)
		if _status_closing {
			flag = 0
		}
	}

	println("[Redis] (update) ending")
	_status <- 0
}

func redis_ping() bool {
	var tick = timestamp64()
	var ctx = context.Background()
	result, err := _instance.Ping(ctx).Result()
	var consuming = fmt.Sprintf("%0.03f", float32(timestamp64()-tick)*0.001)
	if err != nil {
		println("[Redis] (ping) error: " + err.Error() + " [" + consuming + "s] (FAILED)")
		return false
	}
	println("[Redis] (ping) result: " + result + " [" + consuming + "s] (OK)")
	return true
}

func DelWithKey(key string) bool {
	var ctx = context.Background()
	err := _instance.Del(ctx, key).Err()
	if err != nil {
		return false
	}
	return true
}

func PushNumber(key string, value int64, keep float32) bool {
	var ctx = context.Background()
	err := _instance.SetEx(ctx, key, value, keep_time(keep)).Err()
	if err != nil {
		return false
	}
	return true
}

func GetNumber(key string) (int64, bool) {
	var ctx = context.Background()
	val, err := _instance.GetEx(ctx, key, 0).Int64()
	if err != nil {
		return 0, false
	}
	return val, true
}

func PushString(key string, value string, keep float32) bool {
	var ctx = context.Background()
	err := _instance.SetEx(ctx, key, value, keep_time(keep)).Err()
	if err != nil {
		return false
	}
	return true
}

func GetString(key string) (string, bool) {
	var ctx = context.Background()
	val, err := _instance.GetEx(ctx, key, 0).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

// HMSet is a deprecated version of HSet left for compatibility with Redis 3.
func PushFields(key string, values map[string]string) bool {
	var ctx = context.Background()
	err := _instance.HSet(ctx, key, values).Err()
	if err != nil {
		return false
	}
	return true
}

func GetFields(key string) (map[string]string, bool) {
	var ctx = context.Background()
	val, err := _instance.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

func PushJson[T any](key string, values *T, keep float32, encoding bool) bool {
	data, err := json.Marshal(values)
	if err != nil {
		return false
	}

	var text = string(data)
	if encoding {
		text = base64.StdEncoding.EncodeToString(data)
	}
	var ctx = context.Background()
	err = _instance.SetEx(ctx, key, text, keep_time(keep)).Err()
	if err != nil {
		return false
	}
	return true
}

func GetJson[T any](key string, value *T, encoding bool) bool {
	var ctx = context.Background()
	val, err := _instance.GetEx(ctx, key, 0).Result()
	if err != nil {
		return false
	}

	data := []byte(val)
	if encoding {
		data, err = base64.StdEncoding.DecodeString(val)
		if err != nil {
			return false
		}
	}

	err = json.Unmarshal(data, value)
	if err != nil {
		return false
	}
	return true
}

func PushData(key string, values []byte, keep float32) bool {
	var text = base64.StdEncoding.EncodeToString(values)
	var ctx = context.Background()
	err := _instance.SetEx(ctx, key, text, keep_time(keep)).Err()
	if err != nil {
		return false
	}
	return true
}

func GetData(key string) ([]byte, bool) {
	var ctx = context.Background()
	val, err := _instance.GetEx(ctx, key, 0).Result()
	if err != nil {
		return nil, false
	}

	data, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, false
	}
	return data, true
}
