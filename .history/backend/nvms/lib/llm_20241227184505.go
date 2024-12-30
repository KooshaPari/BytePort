package lib

import (
	"fmt"
	"nvms/lib/providers/local"
	"nvms/lib/providers/openai"
	"nvms/models"
)

 
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

 
func RequestCompletion(prompt  string, config models.LLM) (string, error) {
    /*messages := []Message{
        {
            Role:    "user",
            Content: prompt,
        },
    }*/
    var response string; var err error
    reqBody := ChatRequest{
        //Model:    "gpt-4o",
        Model:    config.Providers[config.Provider].Modal,
        Prompt: prompt,
    }
    switch config.Provider {
        case("openAI"): 
        
            response, err  = openai.RequestChatCompletion(reqBody, config.Providers[config.Provider].APIKey)
        
        case("anthropic"): 
            fmt.Println("Anthropic is not Implemented Yet...")
        case("gemini"):
            fmt.Println("Gemini is not Implemented Yet...")
        case("local"):
            response, err  = local.RequestChatCompletion(reqBody, config.Providers[config.Provider].Modal)
        default :
            fmt.Println("Provider not found")
    }
    if err != nil {
        return "", fmt.Errorf("error sending request: %v", err)
    }
    return response,nil
}