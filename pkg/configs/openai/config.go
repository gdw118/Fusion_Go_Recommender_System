package openai

import (
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

const (
	// BaseURL 是 OpenAI API 的基础 URL
	BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	
	// EmbeddingModel 是用于生成 embedding 的模型
	EmbeddingModel = "text-embedding-v3"
)

// getAPIKey 获取 OpenAI API Key
func getAPIKey() (string, error) {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("DASHSCOPE_API_KEY not set")
	}
	return apiKey, nil
}

// NewClient 创建一个新的 OpenAI 客户端
func NewClient() (*openai.Client, error) {
	apiKey, err := getAPIKey()
	if err != nil {
		return nil, err
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = BaseURL
	
	return openai.NewClientWithConfig(config), nil
} 