package converter

type ConversionConfig map[string]interface{}

func MergeConfigs(configs ...ConversionConfig) ConversionConfig {
	result := make(ConversionConfig)
	for _, config := range configs {
		for key, value := range config {
			result[key] = value
		}
	}
	return result
}
