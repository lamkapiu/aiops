package utils

import (
	"context"  // 用于上下文管理，通常在 Go 中用于管理请求的生命周期。
	"errors"
	"os"

	"github.com/sashabaranov/go-openai"  // Go 语言的 OpenAI API 客户端库
)

type OpenAI struct {
	Client *openai.Client
	ctx    context.Context
}

func NewOpenAIClient() (*OpenAI, error) {
	apiKey := os.Getenv("OPENAI_API_KEY_2")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY_2 environment variable is not set")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://vip.apiyi.com/v1"
	client := openai.NewClientWithConfig(config)

	ctx := context.Background()

	return &OpenAI{
		Client: client,
		ctx:    ctx,
	}, nil
}

func (o *OpenAI) SendMessage(prompt string, content string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: prompt,
			},
			{
				Role:    "user",
				Content: content,
			},
		},
	}
    
	// 向 openai 发送消息
	resp, err := o.Client.CreateChatCompletion(o.ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
