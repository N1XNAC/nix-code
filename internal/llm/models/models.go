package models

type ModelProvider string

const (
	ProviderAnthropic ModelProvider = "anthropic"
	ProviderOpenAI    ModelProvider = "openai"
	ProviderGemini    ModelProvider = "gemini"
	ProviderAzure     ModelProvider = "azure"
	ProviderBedrock   ModelProvider = "bedrock"
	ProviderOpenRouter ModelProvider = "openrouter"
	ProviderGroq      ModelProvider = "groq"
)

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
