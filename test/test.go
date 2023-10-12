package main

import (
	"mcmcx.com/gpt-server/httpx"
	"mcmcx.com/gpt-server/utils"
)

func main() {
	logger := utils.NewLogger()
	logger.Init()

	client := httpx.NewClient("https://127.0.0.1:9443", nil)
	client.HTTPRequest2("/ping", nil, &httpx.HTTPData2{
		SkipVerify: true,
	})
}
