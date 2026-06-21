package models

type ModelProvider string

const (
	ProviderAnthropic   ModelProvider = "anthropic"
	ProviderOpenAI      ModelProvider = "openai"
	ProviderGemini      ModelProvider = "gemini"
	ProviderAzure       ModelProvider = "azure"
	ProviderBedrock     ModelProvider = "bedrock"
	ProviderOpenRouter  ModelProvider = "openrouter"
	ProviderGroq        ModelProvider = "groq"
	ProviderNvidiaNIM   ModelProvider = "nvidia-nim"
	ProviderKimi        ModelProvider = "kimi"
	ProviderGLM         ModelProvider = "glm"
	ProviderDeepSeek    ModelProvider = "deepseek"
	ProviderMistral     ModelProvider = "mistral"
	ProviderTogether    ModelProvider = "together"
	ProviderFireworks   ModelProvider = "fireworks"
	ProviderPerplexity  ModelProvider = "perplexity"
	ProviderAnyscale    ModelProvider = "anyscale"
	ProviderXAI         ModelProvider = "xai"
	ProviderCohere      ModelProvider = "cohere"
	ProviderVoyage      ModelProvider = "voyage"
	ProviderAI21        ModelProvider = "ai21"
)

func (p ModelProvider) DefaultBaseURL() string {
	switch p {
	case ProviderOpenAI:
		return "https://api.openai.com/v1"
	case ProviderOpenRouter:
		return "https://openrouter.ai/api/v1"
	case ProviderGroq:
		return "https://api.groq.com/openai/v1"
	case ProviderNvidiaNIM:
		return "https://integrate.api.nvidia.com/v1"
	case ProviderKimi:
		return "https://api.moonshot.cn/v1"
	case ProviderGLM:
		return "https://open.bigmodel.cn/api/paas/v4"
	case ProviderDeepSeek:
		return "https://api.deepseek.com/v1"
	case ProviderMistral:
		return "https://api.mistral.ai/v1"
	case ProviderTogether:
		return "https://api.together.xyz/v1"
	case ProviderFireworks:
		return "https://api.fireworks.ai/inference/v1"
	case ProviderPerplexity:
		return "https://api.perplexity.ai"
	case ProviderAnyscale:
		return "https://api.endpoints.anyscale.com/v1"
	case ProviderXAI:
		return "https://api.x.ai/v1"
	case ProviderCohere:
		return "https://api.cohere.com/v1"
	case ProviderVoyage:
		return "https://api.voyageai.com/v1"
	case ProviderAI21:
		return "https://api.ai21.com/studio/v1"
	default:
		return ""
	}
}

type Model struct {
	ID               string        `json:"id"`
	Provider         ModelProvider `json:"provider"`
	DisplayName      string        `json:"displayName"`
	ContextWindow    int64         `json:"contextWindow"`
	DefaultMaxTokens int64         `json:"defaultMaxTokens"`
	CostPer1MIn      float64       `json:"costPer1MIn"`
	CostPer1MOut     float64       `json:"costPer1MOut"`
	CanReason        bool          `json:"canReason"`
	SupportsImages   bool          `json:"supportsImages"`
}

var SupportedModels = []Model{
	{ID: "claude-sonnet-4-20250514", Provider: ProviderAnthropic, DisplayName: "Claude Sonnet 4", ContextWindow: 200000, DefaultMaxTokens: 8192, CostPer1MIn: 3.00, CostPer1MOut: 15.00, CanReason: true, SupportsImages: true},
	{ID: "claude-3.5-haiku-20241022", Provider: ProviderAnthropic, DisplayName: "Claude 3.5 Haiku", ContextWindow: 200000, DefaultMaxTokens: 8192, CostPer1MIn: 0.80, CostPer1MOut: 4.00, SupportsImages: true},
	{ID: "gpt-4o", Provider: ProviderOpenAI, DisplayName: "GPT-4o", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 2.50, CostPer1MOut: 10.00, SupportsImages: true},
	{ID: "gpt-4o-mini", Provider: ProviderOpenAI, DisplayName: "GPT-4o Mini", ContextWindow: 128000, DefaultMaxTokens: 16384, CostPer1MIn: 0.15, CostPer1MOut: 0.60, SupportsImages: true},
	{ID: "gemini-2.5-pro-exp-03-25", Provider: ProviderGemini, DisplayName: "Gemini 2.5 Pro", ContextWindow: 1048576, DefaultMaxTokens: 8192, CostPer1MIn: 1.25, CostPer1MOut: 10.00, CanReason: true, SupportsImages: true},
	{ID: "gemini-2.0-flash", Provider: ProviderGemini, DisplayName: "Gemini 2.0 Flash", ContextWindow: 1048576, DefaultMaxTokens: 8192, CostPer1MIn: 0.10, CostPer1MOut: 0.40, SupportsImages: true},
	{ID: "nvidia/nemotron-ultra", Provider: ProviderNvidiaNIM, DisplayName: "Nemotron Ultra", ContextWindow: 128000, DefaultMaxTokens: 8192, CostPer1MIn: 0.00, CostPer1MOut: 0.00},
	{ID: "nvidia/llama-3.1-nemotron-70b-instruct", Provider: ProviderNvidiaNIM, DisplayName: "Nemotron 70B", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 0.00, CostPer1MOut: 0.00},
	{ID: "moonshot-v1-8k", Provider: ProviderKimi, DisplayName: "Kimi Moonshot v1 8K", ContextWindow: 8000, DefaultMaxTokens: 4096, CostPer1MIn: 0.12, CostPer1MOut: 0.12},
	{ID: "moonshot-v1-32k", Provider: ProviderKimi, DisplayName: "Kimi Moonshot v1 32K", ContextWindow: 32000, DefaultMaxTokens: 8192, CostPer1MIn: 0.24, CostPer1MOut: 0.24},
	{ID: "moonshot-v1-128k", Provider: ProviderKimi, DisplayName: "Kimi Moonshot v1 128K", ContextWindow: 128000, DefaultMaxTokens: 16384, CostPer1MIn: 0.60, CostPer1MOut: 0.60},
	{ID: "glm-4-plus", Provider: ProviderGLM, DisplayName: "GLM-4 Plus", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 0.50, CostPer1MOut: 0.50},
	{ID: "glm-4-0520", Provider: ProviderGLM, DisplayName: "GLM-4 0520", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 0.10, CostPer1MOut: 0.10},
	{ID: "glm-4-air", Provider: ProviderGLM, DisplayName: "GLM-4 Air", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 0.05, CostPer1MOut: 0.05},
	{ID: "deepseek-chat", Provider: ProviderDeepSeek, DisplayName: "DeepSeek Chat", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 0.27, CostPer1MOut: 1.10},
	{ID: "deepseek-reasoner", Provider: ProviderDeepSeek, DisplayName: "DeepSeek Reasoner", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 0.55, CostPer1MOut: 2.19, CanReason: true},
	{ID: "mistral-large-2411", Provider: ProviderMistral, DisplayName: "Mistral Large", ContextWindow: 128000, DefaultMaxTokens: 8192, CostPer1MIn: 2.00, CostPer1MOut: 6.00},
	{ID: "mistral-small-2501", Provider: ProviderMistral, DisplayName: "Mistral Small", ContextWindow: 32000, DefaultMaxTokens: 4096, CostPer1MIn: 0.20, CostPer1MOut: 0.60},
	{ID: "mistral-saba-2502", Provider: ProviderMistral, DisplayName: "Mistral Saba", ContextWindow: 32000, DefaultMaxTokens: 4096, CostPer1MIn: 0.30, CostPer1MOut: 0.30},
	{ID: "codestral-2501", Provider: ProviderMistral, DisplayName: "Codestral", ContextWindow: 256000, DefaultMaxTokens: 8192, CostPer1MIn: 1.00, CostPer1MOut: 3.00},
	{ID: "Qwen/Qwen2.5-72B-Instruct-Turbo", Provider: ProviderTogether, DisplayName: "Qwen 2.5 72B", ContextWindow: 32000, DefaultMaxTokens: 4096, CostPer1MIn: 0.90, CostPer1MOut: 0.90},
	{ID: "meta-llama/Llama-3.3-70B-Instruct-Turbo", Provider: ProviderTogether, DisplayName: "Llama 3.3 70B", ContextWindow: 32000, DefaultMaxTokens: 4096, CostPer1MIn: 0.90, CostPer1MOut: 0.90},
	{ID: "accounts/fireworks/models/deepseek-v3", Provider: ProviderFireworks, DisplayName: "DeepSeek V3", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 0.90, CostPer1MOut: 0.90},
	{ID: "accounts/fireworks/models/llama-v3p3-70b-instruct", Provider: ProviderFireworks, DisplayName: "Llama 3.3 70B", ContextWindow: 128000, DefaultMaxTokens: 4096, CostPer1MIn: 0.90, CostPer1MOut: 0.90},
	{ID: "sonar-pro", Provider: ProviderPerplexity, DisplayName: "Sonar Pro", ContextWindow: 200000, DefaultMaxTokens: 4096, CostPer1MIn: 3.00, CostPer1MOut: 5.00},
	{ID: "sonar", Provider: ProviderPerplexity, DisplayName: "Sonar", ContextWindow: 200000, DefaultMaxTokens: 4096, CostPer1MIn: 1.00, CostPer1MOut: 1.00},
	{ID: "grok-2", Provider: ProviderXAI, DisplayName: "Grok 2", ContextWindow: 131072, DefaultMaxTokens: 4096, CostPer1MIn: 2.00, CostPer1MOut: 10.00},
	{ID: "grok-2-vision", Provider: ProviderXAI, DisplayName: "Grok 2 Vision", ContextWindow: 32768, DefaultMaxTokens: 4096, CostPer1MIn: 2.00, CostPer1MOut: 10.00, SupportsImages: true},
}

func ModelRegistry() []Model {
	return SupportedModels
}

func FindModel(id string) (Model, bool) {
	for _, m := range SupportedModels {
		if m.ID == id {
			return m, true
		}
	}
	return Model{}, false
}

func DefaultModelForProvider(provider ModelProvider) (Model, bool) {
	for _, m := range SupportedModels {
		if m.Provider == provider {
			return m, true
		}
	}
	return Model{}, false
}
