package main

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
 
)

func requestFilledTemplate(prompt string, key string) (string, error) {
	client := openai.NewClient(
		option.WithAPIKey(key),  
	)
	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			 openai.UserMessage(prompt),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		panic(err.Error())
		return "", err
	}
	println(chatCompletion.Choices[0].Message.Content)
	return chatCompletion.Choices[0].Message.Content,nil
}
