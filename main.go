package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"mcmcx.com/gpt-server/server"
	"mcmcx.com/gpt-server/utils"
)

func main() {
	logger := utils.NewLogger()
	logger.Init()

	//
	bytes, err := os.ReadFile("config.yaml")
	if err != nil {
		logger.LogError("read file config.yaml error: ", err)
		return
	}

	var config server.Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		logger.LogError("parse config.yaml error: %s", err)
		return
	}

	//ChatGPT
	var models = server.API_GPTModels(config)
	if len(models) == 0 {

	}

	//
	logger.Log("GPT service loading ...")
	var service *server.Server = server.InitServer(config, gin.DebugMode)
	if service == nil {
		logger.LogError("[Server] Error: ", "init service error.")
		return
	}

	logger.Log("GPT service starting ...")
	service.StartHTTPServer()
	service.StartHTTPSServer()

	//select {}
	sigs := make(chan os.Signal, 1)
	//signal.Ignore(os.Interrupt)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	println("Signal -> %+v", sig)

	//
	println("Exiting ...")
	return
}
