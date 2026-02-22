package alarm

import (
	"fmt"
	"os"
	"strings"
)

func ResolveAlarmConfigEnvVars(a *Alarm) error {
	if a == nil {
		return nil
	}

	resolvedConfig, err := resolveConfigMap(a.Config)
	if err != nil {
		return err
	}
	a.Config = resolvedConfig

	return nil
}

func resolveConfigMap(config map[string]any) (map[string]any, error) {
	if config == nil {
		return nil, nil
	}

	resolved := make(map[string]any, len(config))
	for key, value := range config {
		resolvedValue, err := resolveConfigValue(value)
		if err != nil {
			return nil, err
		}
		resolved[key] = resolvedValue
	}

	return resolved, nil
}

func resolveConfigValue(value any) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		return resolveConfigMap(typed)
	case []any:
		return resolveConfigSlice(typed)
	case string:
		return resolveStringEnvVars(typed)
	default:
		return value, nil
	}
}

func resolveConfigSlice(values []any) ([]any, error) {
	resolved := make([]any, len(values))
	for i, value := range values {
		resolvedValue, err := resolveConfigValue(value)
		if err != nil {
			return nil, err
		}
		resolved[i] = resolvedValue
	}
	return resolved, nil
}

func resolveStringEnvVars(input string) (string, error) {
	if input == "" {
		return input, nil
	}

	var builder strings.Builder
	builder.Grow(len(input))

	for i := 0; i < len(input); i++ {
		if input[i] != '$' || i+1 >= len(input) || input[i+1] != '{' {
			builder.WriteByte(input[i])
			continue
		}

		nameStart := i + 2
		nameEnd := strings.IndexByte(input[nameStart:], '}')
		if nameEnd == -1 {
			return "", fmt.Errorf("unterminated environment variable placeholder")
		}
		nameEnd += nameStart

		key := input[nameStart:nameEnd]
		if !isValidEnvVarName(key) {
			return "", fmt.Errorf("invalid environment variable name: %s", key)
		}

		value, ok := os.LookupEnv(key)
		if !ok {
			return "", fmt.Errorf("missing environment variable: %s", key)
		}
		builder.WriteString(value)
		i = nameEnd
	}

	return builder.String(), nil
}

func isValidEnvVarName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if !isEnvVarFirstChar(name[0]) {
		return false
	}
	for i := 1; i < len(name); i++ {
		if !isEnvVarChar(name[i]) {
			return false
		}
	}
	return true
}

func isEnvVarFirstChar(ch byte) bool {
	return ch == '_' || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func isEnvVarChar(ch byte) bool {
	return isEnvVarFirstChar(ch) || (ch >= '0' && ch <= '9')
}
