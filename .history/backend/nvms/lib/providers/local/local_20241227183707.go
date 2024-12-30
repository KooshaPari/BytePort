package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"nvms/lib/providers"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)
const localEndpoint = "https://api.ollama.com/v1/chat/completions"

 
type LocalMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type LocalChatRequest struct {
    Model    string    `json:"model"`
    Messages []LocalMessage `json:"messages"`
}
type LocalChoice struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
}

type LocalChatResponse struct {
    Choices []LocalChoice `json:"choices"`
}

func RequestCompletion(reqBody lib.ChatRequest, key string, modal string) (string, error) {
	reqBase := LocalChatRequest{
		Model:    modal,
		Messages: []LocalMessage{
			{
				Role:    "user",
				Content: reqBody.Prompt,}}};
    jsonBody, err := json.Marshal(reqBase)
    if err != nil {
        return "", fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST", localEndpoint, bytes.NewBuffer(jsonBody))
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

    var response LocalChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", fmt.Errorf("error decoding response: %v", err)
    }

    if len(response.Choices) == 0 {
        return "", fmt.Errorf("no response choices returned")
    }

    return response.Choices[0].Message.Content, nil
}