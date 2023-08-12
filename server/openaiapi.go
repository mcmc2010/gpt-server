package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"mcmcx.com/gpt-server/utils"
)

type ChatGPTPermission struct {
	AllowCreateEngine  bool `json:"allow_create_engine"`
	AllowFineTuning    bool `json:"allow_fine_tuning"`
	AllowLogprobs      bool `json:"allow_logprobs"`
	AllowSampling      bool `json:"allow_sampling"`
	AllowSearchIndices bool `json:"allow_search_indices"`
	AllowView          bool `json:"allow_view"`
	//
	Created      uint32 `json:"created"`
	ID           string `json:"id" validate:"required"`
	Object       string `json:"object"`
	IsBlocking   bool   `json:"is_blocking"`
	Organization string `json:"organization"`
}

type ChatGPTModel struct {
	ID         string              `json:"id" validate:"required"`
	Created    uint32              `json:"created"`
	Object     string              `json:"object"`
	Permission []ChatGPTPermission `json:"permission"`
	Root       string              `json:"root"`
	Owned      string              `json:"owned_by"`
}

type ChatGPTRequest struct {
	Prompt   string `json:"prompt"`
	MaxTurns int    `json:"max_turns"`
}

type ChatGPTResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func API_GPTModels(config Config) []*ChatGPTModel {
	var models []*ChatGPTModel
	url, _ := url.JoinPath(config.APIUrl, "/v1/models")

	//
	var client = &http.Client{}
	request, err := http.NewRequest("Get", url, nil)
	if err != nil {
		utils.Logger.LogError("(API) Create Request Error: ", err)
		return nil
	}
	request.Header.Set("Openai-Organization", config.APIOrganization)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	response, err := client.Do(request)
	if err != nil {
		utils.Logger.LogError("(API) Request Error: ", err)
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		utils.Logger.LogError("(API) Request ", "(Status:", response.StatusCode, ") Error: ", response.Status)
		return nil
	}

	var bytes []byte
	length, err := response.Body.Read(bytes)
	if err != nil {
		utils.Logger.LogError("(API) Response Body Error: ", err)
		return nil
	}

	if length > 0 {
		err = json.Unmarshal(bytes, &models)
		if err != nil {
			utils.Logger.LogError("(API) Response Body Error: ", err)
			return nil
		}
	} else {
		utils.Logger.LogError("(API) Response Body : (null)")
		return nil
	}

	return models
}

func API_GPTCompletions() {

}
