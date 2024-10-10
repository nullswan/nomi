package ollamaprovider

type olamaProviderConfig struct {
	baseURL string
	model   string
}

func NewOlamaProviderConfig(baseURL, model string) olamaProviderConfig {
	return olamaProviderConfig{
		baseURL: baseURL,
		model:   model,
	}
}

func (o olamaProviderConfig) BaseURL() string {
	return o.baseURL
}

func (o olamaProviderConfig) WithBaseURL(baseURL string) olamaProviderConfig {
	o.baseURL = baseURL
	return o
}

func (o olamaProviderConfig) Model() string {
	return o.model
}

func (o olamaProviderConfig) WithModel(model string) olamaProviderConfig {
	o.model = model
	return o
}
