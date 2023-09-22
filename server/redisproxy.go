package server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"mcmcx.com/gpt-server/database/redis"
	"mcmcx.com/gpt-server/utils"
)

const LOG_REDIS = "REDIS"

var redis_info database_redis.Config
var redis_instance *redis.Client = nil

func redis_loadconfig(filename string) bool {

	//
	bytes, err := os.ReadFile(filename)
	if err != nil {
		utils.Logger.LogError("[Load] Read file config.yaml error: ", err)
		return false
	}

	err = yaml.Unmarshal(bytes, &redis_info)
	if err != nil {
		utils.Logger.LogError("[Load] Parse config.yaml error: %s", err)
		return false
	}

	if redis_info.Port <= 0 {
		utils.Logger.LogError("[Load] Read redis info error")
		return false
	}

	if len(redis_info.TLSKey) > 0 && len(redis_info.TLSCrt) > 0 {
		redis_info.UseTLS = true
		utils.LogWithName(LOG_REDIS, "[Load] Redis Use TLS (OK)")
	}

	return true
}

func RedisRelease() int {
	return database_redis.Release()
}

func RedisInitialize(filename string) bool {

	//
	utils.LogAdd(utils.LogLevel_Info, LOG_REDIS, true, true)

	//
	if !redis_loadconfig(filename) {
		utils.Logger.LogError("[Load] redis information server error")
		return false
	}

	//
	var tls_config *tls.Config = nil
	if redis_info.UseTLS {
		pool := utils.LoadCertCAFromFile(redis_info.TLSCA)
		if pool == nil {
			utils.LogWithName(LOG_REDIS, "[Load] Load TLS Error : CA file (", redis_info.TLSCA, ")")
			return false
		}
		cert := utils.LoadCertFromFiles(redis_info.TLSCrt, redis_info.TLSKey)
		if cert == nil {
			utils.LogWithName(LOG_REDIS, "[Load] Load TLS Error : key file (", redis_info.TLSKey, "), crt file (", redis_info.TLSCrt, ")")
			return false
		}

		tls_config = &tls.Config{
			MinVersion: tls.VersionTLS12,
			//ServerName:         "localhost",
			InsecureSkipVerify: true,
			RootCAs:            pool,
			Certificates: []tls.Certificate{
				*cert,
			},
			ClientAuth: tls.RequireAndVerifyClientCert,
			//ClientCAs:  pool,
			VerifyConnection: func(cs tls.ConnectionState) error {
				if len(cs.PeerCertificates) == 0 {
					return errors.New("The peer certificates not found")
				}
				var cert = cs.PeerCertificates[0]
				_, err := cert.Verify(x509.VerifyOptions{
					DNSName: "",
					Roots:   pool,
				})
				if err != nil {
					return err
				}
				return nil
			},
		}

	}
	redis_info.TLSConfig = tls_config

	redis_instance = database_redis.NewAndInitialize(&redis_info)
	if redis_instance == nil {
		utils.Logger.LogError("[Load] Connect redis server error")
		return false
	}

	return true
}
