package server

import (
	"strings"

	"mcmcx.com/gpt-server/httpx"
	"mcmcx.com/gpt-server/utils"
)

var ipapi_base_url string = "http://ip-api.com/json/"
var ipapi_client *httpx.HTTPClient2 = nil

func API_IPInit() bool {

	//var additional_headers map[string]string = map[string]string{}

	ipapi_client = httpx.NewClient(ipapi_base_url, nil)
	if ipapi_client == nil {
		return false
	}

	return true
}

// IPv4/IPv6
// Localized city, regionName and country can be requested by setting the GET parameter lang to one of the following:
//
//	{
//		"query": "100.177.20.1",
//		"status": "success",
//		"country": "United States",
//		"countryCode": "US",
//		"regionName": "Illinois",
//		"city": "Chicago",
//		"district": "",
//		"zip": "60666",
//		"timezone": "America/Chicago",
//		"org": "T-Mobile USA, Inc.",
//		"mobile": true
//	 }
func API_IPGet(address string) map[string]any {
	address = strings.TrimSpace(address)
	if len(address) == 0 {
		return nil
	}

	data := httpx.HTTPData2{
		SkipVerify: true,
	}

	var IsIPv6 bool = false
	if strings.ContainsAny(address, ":") {
		IsIPv6 = true
	}

	var params map[string]any = map[string]any{
		"lang":   "en",
		"fields": "status,message,country,countryCode,regionName,city,district,zip,timezone,org,mobile,query",
	}

	path := address
	ipapi_client.HTTPRequest2(path, params, &data)

	utils.Logger.Log("(API) Get IP :", address, " (Time: ", data.EndTime(), "ms)")

	if data.ErrorCode != httpx.HTTP_RESULT_OK {
		return nil
	}

	result, ok := data.Data().(map[string]any)
	if !ok || result["status"] == "fail" {
		if result["message"] == "private range" {
			result["status"] = "success"
			result["type"] = "private"
		} else {
			return nil
		}
	}
	_, ok = result["type"]
	if(!ok) {
		result["type"] = "public"
	}

	result["ipv6"] = IsIPv6

	//
	return result
}
