package converter

import (
	"fmt"
	"strings"
)

type ConversionConfig map[string]string

type ConversionFormats map[string]ConversionConfig

func Formats(str string) []string {
	return strings.Split(str, ",")
}

func ParseFormats(formatsStr string) (ConversionFormats, error) {
	entries := Formats(formatsStr)
	if len(entries) == 0 {
		return nil, fmt.Errorf("no conversion formats specified")
	}
	convFormats := make(ConversionFormats)
	for _, entry := range entries {
		format, conf, err := ParseFormat(entry)
		if err != nil {
			return nil, err
		}
		convFormats[format] = conf
	}
	return convFormats, nil
}

func ParseFormat(formatStr string) (string, ConversionConfig, error) {
	formatParts := strings.Split(formatStr, "?")
	if len(formatParts) == 0 {
		return "", nil, fmt.Errorf("invalid conversion format '%s'", formatStr)
	}
	if len(formatParts) == 1 {
		return formatParts[0], nil, nil
	}
	if len(formatParts) > 2 {
		return "", nil, fmt.Errorf("invalid conversion format '%s'", formatStr)
	}
	return formatParts[0], ParseParams(formatParts[1]), nil
}

func ParseParams(paramsStr string) ConversionConfig {
	params := strings.Split(paramsStr, "&")
	if len(params) == 0 {
		return nil
	}
	config := make(ConversionConfig)
	for _, param := range params {
		parsedParam := strings.Split(param, "=")
		if len(parsedParam) == 1 {
			config[parsedParam[0]] = ""
		} else if len(parsedParam) == 2 {
			config[parsedParam[0]] = parsedParam[1]
		}
	}
	return config
}

func MergeConfigs(configs ...ConversionConfig) ConversionConfig {
	result := make(ConversionConfig)
	for _, config := range configs {
		for key, value := range config {
			result[key] = value
		}
	}
	return result
}
