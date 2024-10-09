package ollamaprovider

type olamaProviderConfig struct {
	baseURL string
	model   string
}

func NewOlamaProviderConfig(baseUrl, model string) olamaProviderConfig {
	return olamaProviderConfig{
		baseURL: baseUrl,
		model:   model,
	}
}

func (o olamaProviderConfig) BaseUrl() string {
	return o.baseURL
}

func (o olamaProviderConfig) WithBaseUrl(baseUrl string) olamaProviderConfig {
	o.baseURL = baseUrl
	return o
}

func (o olamaProviderConfig) Model() string {
	return o.model
}

func (o olamaProviderConfig) WithModel(model string) olamaProviderConfig {
	o.model = model
	return o
}
