package validation

type Config struct {
	EnableLogging             bool
	SanitizeSensitiveData     bool
	AdditionalSensitiveFields []string
	LogSuccessfulValidations  bool
}

func DefaultConfig() *Config {
	return &Config{
		EnableLogging:             true,
		SanitizeSensitiveData:     true,
		AdditionalSensitiveFields: []string{},
		LogSuccessfulValidations:  false,
	}
}
