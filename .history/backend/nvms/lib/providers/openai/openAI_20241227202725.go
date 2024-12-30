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

 

type FormatSchema struct {
    Type       string      `json:"type"`
    JsonSchema interface{} `json:"json_schema"`
}
type OAIMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type OAIChatRequest struct {
    Model    string    `json:"model"`
    Messages []OAIMessage `json:"messages"`
    ResponseFormat FormatSchema `json:"response_format"`
}
type OAIChoice struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
   
}

type OAIChatResponse struct {
    Choices []OAIChoice `json:"choices"`
}

func RequestChatCompletion(reqBody lib.ChatRequest, key string, modal string) (string, error) {
    var objStruct interface{}
    if err := json.Unmarshal([]byte(reqBody.ObjStruct), &objStruct); err != nil {
        fmt.Printf("error decoding objStruct: %v\n", err)
        return "", fmt.Errorf("error decoding objStruct: %v", err)
    }
	reqBase := OAIChatRequest{
		Model:    modal,
		Messages: []OAIMessage{
			{
				Role:    "user",
				Content: reqBody.Prompt,} },
        ResponseFormat:   FormatSchema{Type: "json_schema", JsonSchema: objStruct},}
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