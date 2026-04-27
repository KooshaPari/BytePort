package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"nvms/models"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)


const (
    OpenAI models.AIProvider = iota
    Anthropic
    Gemini
    Ollama
)

// ProviderEndpoint maps providers to their API endpoints
var ProviderEndpoint = map[string]string{
   "openAI":    "https://api.openai.com/v1/chat/completions",
    Anthropic: "https://api.anthropic.com/v1/chat/completions",
    Gemini:    "https://api.gemini.com/v1/chat/completions",
    Ollama:    "https://api.ollama.com/v1/chat/completions",
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
}

type Choice struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
}

type ChatResponse struct {
    Choices []Choice `json:"choices"`
}

func RequestChatCompletion(prompt, apiKey string) (string, error) {
    messages := []Message{
        {
            Role:    "user",
            Content: prompt,
        },
    }

    reqBody := ChatRequest{
        Model:    "gpt-4o",
        Messages: messages,
    }

    jsonBody, err := json.Marshal(reqBody)
    if err != nil {
        return "", fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST", ProviderEndpoint["OpenAI"], bytes.NewBuffer(jsonBody))
    if err != nil {
        return "", fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)

    resp, err := spinhttp.Send(req)
    if err != nil {
        return "", fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    var response ChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", fmt.Errorf("error decoding response: %v", err)
    }

    if len(response.Choices) == 0 {
        return "", fmt.Errorf("no response choices returned")
    }

    return response.Choices[0].Message.Content, nil
}