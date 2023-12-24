package server

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	database_level "mcmcx.com/gpt-server/database/level"
	"mcmcx.com/gpt-server/httpx"
)

type IPLocalizedData struct {
	IPType    string `json:"type"`
	IPAddress string `json:"address"`
	Country   string `json:"country"`
	Region    string `json:"region"`
	City      string `json:"city"`
	District  string `json:"district"`
	Zip       int    `json:"zip"`
	Status    string `json:"status"`
}

// / ?.Localize() return ""
func (I *IPLocalizedData) Localize() string {
	if I == nil {
		return ""
	}

	if len(I.Country) == 0 && len(I.Region) == 0 && len(I.City) == 0 {
		return ""
	}

	text := I.Country
	if len(I.Region) > 0 {
		text = text + fmt.Sprintf(",%s", I.Region)
	}
	if len(I.City) > 0 {
		text = text + fmt.Sprintf(",%s", I.City)
	}
	return text
}

// / ?.FullLocalize() return ""
func (I *IPLocalizedData) FullLocalize() string {
	if I == nil {
		return ""
	}

	if len(I.Country) == 0 && len(I.Region) == 0 && len(I.City) == 0 {
		return ""
	}

	text := I.Country
	if len(I.Region) > 0 {
		text = text + fmt.Sprintf(",%s", I.Region)
	}
	if len(I.City) > 0 {
		text = text + fmt.Sprintf(",%s", I.City)
	}

	if len(I.District) > 0 {
		text = text + fmt.Sprintf(" %s", I.District)
	}
	if I.Zip > 0 {
		text = text + fmt.Sprintf(" (%d)", I.Zip)
	}
	return text
}

var ipapi_base_url string = "http://ip-api.com/json/"
var ipapi_client *httpx.HTTPClient2 = nil
var ipapi_db *database_level.LevelDB = nil

func API_IPInit() bool {

	//var additional_headers map[string]string = map[string]string{}

	ipapi_client = httpx.NewClient(ipapi_base_url, nil)
	if ipapi_client == nil {
		return false
	}

	return true
}

func IPLocalized(address string) *IPLocalizedData {

	//
	var data *IPLocalizedData = IPDB2Get(address)
	if data != nil {
		return data
	}

	var result = API_IPGet(address)
	if result != nil {
		ipi, ok := IPSet2DB(result)
		if !ok {
			fmt.Printf("[WARNING] (IPAPI) DB save %s failure\n", result["query"])
		}

		bytes, err := json.Marshal(ipi)
		if err != nil {
			fmt.Printf("[WARNING] (IPAPI) DB save %s failure\n", result["query"])
			return nil
		}

		var v IPLocalizedData
		err = json.Unmarshal(bytes, &v)
		if err != nil {
			fmt.Printf("[WARNING] (IPAPI) DB save %s failure\n", result["query"])
			return nil
		}
		data = &v
		return data
	}

	return nil
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
	if data.ErrorCode != httpx.HTTP_RESULT_OK {
		return nil
	}

	result, ok := data.Data().(map[string]any)
	if !ok || result["status"] == "fail" {
		if result["message"] == "private range" {
			result["status"] = "success"
			result["type"] = "private"
		} else if result["message"] == "reserved range" {
			result["status"] = "success"
			result["type"] = "reserved"
		} else {
			return nil
		}
	}
	_, ok = result["type"]
	if !ok {
		result["type"] = "public"
	}

	result["ipv6"] = IsIPv6

	fmt.Printf("(IPAPI) Get IP : %s %s (Time: %d ms)\n", address, result["country"], data.EndTime())

	//
	return result // ipi.(map[string]any)
}

func IPDBAddress(address string) string {
	ip := net.ParseIP(address)
	if ip == nil {
		return ""
	}

	//IPv4
	ip4 := ip.To4()
	if ip4 == nil {
		return ""
	}

	//IPv6 not support
	ip6 := ip.To16()
	if ip6 != nil {
		fmt.Printf("[WARNING] (IPAPI) Not support %s IPv6:%s\n", ip.String(), ip6.String())
	}

	//IPv4
	address = ip4.String()
	// IP address 192.168.0.1, mask 255.255.255.0
	// or 192.168.0.1/24
	vs := strings.Split(address, ".")
	address = vs[0] + "." + vs[1] + "." + vs[2] + "." + "0"

	//
	return address
}

func IPDB2Get(address string) *IPLocalizedData {
	ip_address := IPDBAddress(address)
	if len(ip_address) == 0 {
		return nil
	}

	if ipapi_db == nil {
		ipapi_db = database_level.NewAndInitialize("./ipv4.db")
	}

	data, err := ipapi_db.Get(ip_address)
	if err != nil {
		return nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var v IPLocalizedData
	err = json.Unmarshal(bytes, &v)
	if err != nil {
		return nil
	}
	return &v
}

func IPSet2DB(ipi map[string]any) (any, bool) {
	ret, ok := ipi["status"]
	if ok && ret == "fail" {
		return nil, false
	}

	ret, ok = ipi["type"]
	if !ok {
		return nil, false
	}
	ip_type := ret.(string)
	ip_type = strings.ToLower(strings.TrimSpace(ip_type))

	ret, ok = ipi["query"]
	if !ok {
		return nil, false
	}
	ip_address := ret.(string)

	// Not support IPv6
	ret, ok = ipi["ipv6"]
	if !ok {
		return nil, false
	}
	ip_v6 := ret.(bool)
	if ip_v6 {
		return nil, false
	}

	ip_address = IPDBAddress(ip_address)
	if len(ip_address) == 0 {
		return nil, false
	}

	if ipapi_db == nil {
		ipapi_db = database_level.NewAndInitialize("./ipv4.db")
	}

	//
	lcountry, ok := ipi["country"].(string)
	if !ok || len(lcountry) == 0 {
		lcountry = ""
	}

	//
	lregion, ok := ipi["regionName"].(string)
	if !ok || len(lregion) == 0 {
		lregion, ok := ipi["region"].(string)
		if !ok || len(lregion) == 0 {
			lregion = ""
		}
	}

	lcity, ok := ipi["city"].(string)
	if !ok || len(lcity) == 0 {
		lcity = ""
	}

	ldistrict, ok := ipi["district"].(string)
	if !ok || len(ldistrict) == 0 {
		ldistrict = ""
	}

	lzip, ok := ipi["zip"].(int)
	if !ok {
		lzip = 0
	}

	status := "reserved"
	if ip_type == "private" {
		status = "local"
	} else if ip_type == "reserved" {
		//nothing
	} else {
		status = ""
	}

	v := map[string]any{
		"type":     ip_type,
		"address":  ip_address,
		"country":  lcountry,
		"region":   lregion,
		"city":     lcity,
		"district": ldistrict,
		"zip":      lzip,
		"status":   status,
	}
	ipapi_db.Set(ip_address, v)
	return v, true
}
