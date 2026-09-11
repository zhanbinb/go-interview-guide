package llm

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func NewClient(apiKey string, baseURL string) openai.Client {
	return openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
}
