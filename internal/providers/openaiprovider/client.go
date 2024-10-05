package openaiprovider

type oaiProviderConfig struct {
	apiKey string
	model  string
}

func NewOAIProviderConfig(apiKey, model string) oaiProviderConfig {
	return oaiProviderConfig{
		apiKey: apiKey,
		model:  model,
	}
}

func (o oaiProviderConfig) APIKey() string {
	return o.apiKey
}

func (o oaiProviderConfig) WithAPIKey(apiKey string) oaiProviderConfig {
	o.apiKey = apiKey
	return o
}

func (o oaiProviderConfig) Model() string {
	return o.model
}

func (o oaiProviderConfig) WithModel(model string) oaiProviderConfig {
	o.model = model
	return o
}
