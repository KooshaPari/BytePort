package lib

import (
	"fmt"
	"nvms/models"

	local "nvms/lib/providers/local"
	openAI "nvms/lib/providers/openAI"
)

// ProviderEndpoint maps providers to their API endpoints
var ProviderEndpoint = map[string]string{
    "openAI":    "https://api.openai.com/v1/chat/completions",
    "anthropic": "https://api.anthropic.com/v1/chat/completions",
    "gemini":    "https://api.gemini.com/v1/chat/completions",
    "local":    "https://api.ollama.com/v1/chat/completions",
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Prompt string `json:"prompt"`
}

type Choice struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
}

type ChatResponse struct {
    Choices []Choice `json:"choices"`
}

func RequestCompletion(prompt  string, config models.LLM) (string, error) {
    /*messages := []Message{
        {
            Role:    "user",
            Content: prompt,
        },
    }*/
    var response ChatResponse
    reqBody := ChatRequest{
        //Model:    "gpt-4o",
        Model:    config.Providers[config.Provider].Modal,
        Prompt: prompt,
    }
    switch config.Provider {
        case("openAI"): 
        
            response, err := openAI.RequestChatCompletion(reqBody, config.Providers[config.Provider].APIKey)
        
        case("anthropic"): 
            fmt.Println("Anthropic is not Implemented Yet...")
        case("gemini"):
            fmt.Println("Gemini is not Implemented Yet...")
        case("local"):
            response, err := local.RequestChatCompletion(reqBody, config.Providers[config.Provider].Modal)
            
}
    return response,nil
}