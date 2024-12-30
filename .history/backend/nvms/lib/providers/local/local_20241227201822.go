package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	lib "nvms/lib/providers"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)
const localEndpoint = "http://localhost:11434/api/generate"

 
 
 type LocalChatRequest struct {
    Model    string `json:"model"`
    Prompt   string `json:"prompt"`
    Stream
 
type LocalChatResponse struct {
    Response string `json:"response"`
}

func RequestCompletion(reqBody lib.ChatRequest,modal string) (string, error) {
 
    jsonBody, err := json.Marshal(...reqBody,stream=false)
    if err != nil {
        fmt.Println("error marshaling request: %v", err)
        return "", fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST", localEndpoint, bytes.NewBuffer(jsonBody))
    if err != nil {
        fmt.Println("error creating request: %v", err)
        return "", fmt.Errorf("error creating request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")
    

    resp, err := spinhttp.Send(req)
    if err != nil {
        fmt.Printf("error sending request: %v\n", err)
        return "", fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    var response LocalChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        fmt.Printf("error decoding response: %v\n", err)
        return "", fmt.Errorf("error decoding response: %v", err)
    }
 

    return response.Response, nil
}