package generator

import (
	"strings"

	"fpk-compose-builder/internal/parser"
)

// ReplaceVariables replaces template variables in the given content
// Supported variables:
//   - ${SERVICE_NAME}: First service name
//   - ${CONTAINER_NAME}: First service container_name (or service name if not specified)
//   - ${FIRST_PORT}: First port of the first service (host port)
func ReplaceVariables(content string, vars parser.Variables) string {
	replacements := map[string]string{
		"${SERVICE_NAME}":   vars.ServiceName,
		"${CONTAINER_NAME}": vars.ContainerName,
		"${FIRST_PORT}":     vars.FirstPort,
	}

	result := content
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// ReplaceVariablesInMap replaces variables in all string values of a map
func ReplaceVariablesInMap(data map[string]interface{}, vars parser.Variables) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		switch v := value.(type) {
		case string:
			result[key] = ReplaceVariables(v, vars)
		case map[string]interface{}:
			result[key] = ReplaceVariablesInMap(v, vars)
		default:
			result[key] = value
		}
	}
	return result
}
