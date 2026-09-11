package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// preset 描述一个模型供应商的连接配置，修改 presets 即可增删供应商
type Preset struct {
	APIKey  string // API Key 的环境变量名（从 .env 读取）
	BaseURL string // OpenAI 兼容的 Base URL
	Model   string // 模型 ID
}

func Load() (map[string]Preset, error) {
	_ = godotenv.Load()

	deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	minimaxKey := os.Getenv("MINIMAX_API_KEY")

	if deepseekKey == "" && minimaxKey == "" {
		return nil, fmt.Errorf("no LLM API key configured")
	}

	presets := map[string]Preset{
		"deepseek": {
			APIKey:  "DEEPSEEK_API_KEY",
			BaseURL: "https://api.deepseek.com",
			Model:   "deepseek-v4-flash",
		},
		"minimax": {
			// 官方 OpenAI 兼容端点（海外: https://api.minimax.io/v1）
			APIKey:  "MINIMAX_API_KEY",
			BaseURL: "https://api.minimax.cn/v1",
			Model:   "MiniMax-M3",
		},
		"openai": {
			APIKey:  "OPENAI_API_KEY",
			BaseURL: "https://api.openai.com",
			Model:   "gpt-4o-mini",
		},
	}

	return presets, nil
}
