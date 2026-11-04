package config

// DocsConfig is the canonical, zero-environment configuration used to
// generate the committed OpenAPI document deterministically.
func DocsConfig() *Config {
	return &Config{OpenAPI: OpenAPI{
		Enabled:     true,
		Title:       "Mint Alternative API",
		Description: "API documentation for Mint Alternative",
		Version:     "1.0.0",
		Servers:     []string{"https://localhost:3000"},
		License: &License{
			Name: "MIT",
			URL:  "https://opensource.org/license/mit",
		},
	}}
}
