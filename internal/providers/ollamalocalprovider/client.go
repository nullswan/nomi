package olamalocalprovider

type olamaProviderConfig struct {
	baseUrl string
	model   string
}

func NewOlamaProviderConfig(baseUrl, model string) olamaProviderConfig {
	return olamaProviderConfig{
		baseUrl: baseUrl,
		model:   model,
	}
}

func (o olamaProviderConfig) BaseUrl() string {
	return o.baseUrl
}

func (o olamaProviderConfig) WithBaseUrl(baseUrl string) olamaProviderConfig {
	o.baseUrl = baseUrl
	return o
}

func (o olamaProviderConfig) Model() string {
	return o.model
}

func (o olamaProviderConfig) WithModel(model string) olamaProviderConfig {
	o.model = model
	return o
}
