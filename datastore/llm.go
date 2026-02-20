package datastore

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

func GetLLM(ctx context.Context) (llms.Model, error) {
	switch MapConf.LLMProvider {
	case "ollama":
		baseURL := "http://localhost:11434"
		if MapConf.LLMBaseURL != "" {
			baseURL = MapConf.LLMBaseURL
		}
		return ollama.New(
			ollama.WithModel(MapConf.LLMModel),
			ollama.WithServerURL(baseURL),
		)
	case "gemini", "googleai":
		opts := []googleai.Option{
			googleai.WithAPIKey(MapConf.LLMAPIKey),
		}
		if MapConf.LLMModel != "" {
			opts = append(opts, googleai.WithDefaultModel(MapConf.LLMModel))
		}
		return googleai.New(ctx, opts...)
	case "openai":
		opts := []openai.Option{}
		if MapConf.LLMModel != "" {
			opts = append(opts, openai.WithModel(MapConf.LLMModel))
		}
		if MapConf.LLMAPIKey != "" {
			opts = append(opts, openai.WithToken(MapConf.LLMAPIKey))
		}
		if MapConf.LLMBaseURL != "" {
			opts = append(opts, openai.WithBaseURL(MapConf.LLMBaseURL))
		}
		return openai.New(opts...)
	case "anthropic", "claude":
		opts := []anthropic.Option{}
		if MapConf.LLMModel != "" {
			opts = append(opts, anthropic.WithModel(MapConf.LLMModel))
		}
		if MapConf.LLMAPIKey != "" {
			opts = append(opts, anthropic.WithToken(MapConf.LLMAPIKey))
		}
		if MapConf.LLMBaseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(MapConf.LLMBaseURL))
		}
		return anthropic.New(opts...)
	}
	return nil, fmt.Errorf("llm provider not found")
}
