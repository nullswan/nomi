package openrouterprovider

type openRouterProviderConfig struct {
	apiKey string
	model  string
}

const baseURL = "https://openrouter.ai/api/v1"

func NewORProviderConfig(
	apiKey, model string,
) openRouterProviderConfig {
	return openRouterProviderConfig{
		apiKey: apiKey,
		model:  model,
	}
}

func (o openRouterProviderConfig) APIKey() string {
	return o.apiKey
}

func (o openRouterProviderConfig) WithAPIKey(
	apiKey string,
) openRouterProviderConfig {
	o.apiKey = apiKey
	return o
}

func (o openRouterProviderConfig) Model() string {
	return o.model
}

func (o openRouterProviderConfig) WithModel(
	model string,
) openRouterProviderConfig {
	o.model = model
	return o
}
