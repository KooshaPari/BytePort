package providers

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Prompt string `json:"prompt"`
    ObjStruct struct {
}

type Choice struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
}
