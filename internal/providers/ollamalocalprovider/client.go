package olamalocalprovider

type olamaProviderConfig struct {
	model   string
	baseUrl string
}

func NewOlamaProviderConfig(model, baseUrl string) olamaProviderConfig {
	return olamaProviderConfig{
		model:   model,
		baseUrl: baseUrl,
	}
}

func (o olamaProviderConfig) Model() string {
	return o.model
}

func (o olamaProviderConfig) WithModel(model string) olamaProviderConfig {
	o.model = model
	return o
}

func (o olamaProviderConfig) BaseUrl() string {
	return o.baseUrl
}

func (o olamaProviderConfig) WithBaseUrl(baseUrl string) olamaProviderConfig {
	o.baseUrl = baseUrl
	return o
}
