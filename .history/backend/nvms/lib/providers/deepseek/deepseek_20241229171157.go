package deepseek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	lib "nvms/lib/providers"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)
const deepseekEndpoint = "https://api.deepseek.ai/v1/chat/completions"

  
type DSMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type DSChatRequest struct {
    Model    string    `json:"model"`
    Messages []DSMessage `json:"messages"`
    //ResponseFormat FormatSchema `json:"response_format"`
}
type DSChoice struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
   
}

type DSChatResponse struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Choices []struct {
        Index   int       `json:"index"`
        Message DSMessage `json:"message"`
    } `json:"choices"`
}
func RequestChatCompletion(reqBody lib.ChatRequest, key string, modal string) (string, error) {
    var objStruct interface{}
    if err := json.Unmarshal([]byte(reqBody.ObjStruct), &objStruct); err != nil {
        fmt.Printf("error decoding objStruct: %v\n", err)
        return "", fmt.Errorf("error decoding objStruct: %v", err)
    }
	reqBase := DSChatRequest{
		Model:    modal,
		Messages: []DSMessage{
			{
				Role:    "user",
				Content: reqBody.Prompt,} },
        /*ResponseFormat: FormatSchema{
            Type: "json_object",
            JsonSchema: ResponseSchema{
                Type:       "object",
                Properties: make(map[string]interface{}),
                Required:   []string{},
                Schema:    "http://json-schema.org/draft-07/schema#",
            },
        },*/}
    jsonBody, err := json.Marshal(reqBase)
    if err != nil {
        return "", fmt.Errorf("error marshaling request: %v", err)
    }

    req, err := http.NewRequest("POST",deepsEndpoint, bytes.NewBuffer(jsonBody))
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

    var response DSChatResponse
    fmt.Println("Response: ", resp)
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", fmt.Errorf("error decoding response: %v", err)
    }
    fmt.Println("Response: ", response)

    if len(response.Choices) == 0 {
        return "", fmt.Errorf("no response choices returned")
    }

    return response.Choices[0].Message.Content, nil
}