package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	lib "nvms/lib/providers"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)
const oaiEndpoint = "https://api.openai.com/v1/chat/completions"

 
type OAIMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type OAIChatRequest struct {
    Model    string    `json:"model"`
    Messages []OAIMessage `json:"messages"`
}
type OAIChoice struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
}

type OAIChatResponse struct {
    Choices []OAIChoice `json:"choices"`
}

func RequestCompletion(reqBody lib.ChatRequest, key string, modal string) (string, error) {
	reqBase := OAIChatRequest{
		Model:    modal,
		Messages: []OAIMessage{
			{
				Role:    "user",
				Content: reqBody.Prompt,}}};
    jsonBody, err := json.Marshal(reqBase)
    if err != nil {
        return "", fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST", oaiEndpoint, bytes.NewBuffer(jsonBody))
    if err != nil {
        return "", fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+key)

    resp, err := spinhttp.Send(req)
    if err != nil {
        return "", fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    var response OAIChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", fmt.Errorf("error decoding response: %v", err)
    }

    if len(response.Choices) == 0 {
        return "", fmt.Errorf("no response choices returned")
    }

    return response.Choices[0].Message.Content, nil
}