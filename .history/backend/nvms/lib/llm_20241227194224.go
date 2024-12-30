package lib

import (
	"fmt"
	lib "nvms/lib/providers"
	"nvms/lib/providers/local"
	"nvms/lib/providers/openai"
	"nvms/models"
)

 
 
 const (
    ProviderOpenAI    = "openai"
    ProviderAnthropic = "anthropic"
    ProviderGemini    = "gemini"
    ProviderLocal     = "local"
)
func RequestCompletion(prompt  string, config models.LLM) (string, error) {
 
    var response string; var err error
    reqBody := lib.ChatRequest{
        //Model:    "gpt-4o",
        Model:    config.Providers[config.Provider].Modal,
        Prompt: prompt,
    }
 
    switch config.Provider {
        case ProviderOpenAI: 
            fmt.Println("OpenAI")
            response, err  = openai.RequestChatCompletion(reqBody, config.Providers[config.Provider].APIKey,config.Providers[config.Provider].Modal)
        
        case ProviderAnthropic: 
            fmt.Println("Anthropic is not Implemented Yet...")
        case ProviderGemini:
            fmt.Println("Gemini is not Implemented Yet...")
        case ProviderLocal:
            response, err  = local.RequestCompletion(reqBody, config.Providers[config.Provider].Modal)
        default :
            fmt.Println("Provider not found")
            return "", fmt.Errorf("Provider not found: %v", err)
    }
    if err != nil {
        return "", fmt.Errorf("error sending request: %v", err)
    }
    return response,nil
}