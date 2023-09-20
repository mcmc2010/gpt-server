package main

import (
	"mcmcx.com/gpt-server/server"
	"mcmcx.com/gpt-server/utils"
)

func main() {
	logger := utils.NewLogger()
	logger.Init()

	server.API_HTTPRequest2("https://127.0.0.1:9443/ping", "", &server.API_HTTPData2{
		SkipVerify: true,
	})
}
