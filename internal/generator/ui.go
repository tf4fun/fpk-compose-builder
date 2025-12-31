package generator

import (
	"fpk-compose-builder/internal/parser"
)

// UIConfigEntry represents a single UI entry configuration
type UIConfigEntry struct {
	Title    string `json:"title"`
	Icon     string `json:"icon"`
	Type     string `json:"type"`
	Protocol string `json:"protocol,omitempty"`
	Port     string `json:"port,omitempty"`
	URL      string `json:"url"`
	AllUsers bool   `json:"allUsers"`
}

// GenerateDefaultUIConfig generates the default app/ui/config JSON content
// Creates a default configuration based on service name and port
func GenerateDefaultUIConfig(vars parser.Variables) (string, error) {
	appEntry := vars.ServiceName + ".Application"

	config := map[string]interface{}{
		".url": map[string]interface{}{
			appEntry: UIConfigEntry{
				Title:    vars.ServiceName,
				Icon:     "images/icon_{0}.png",
				Type:     "url",
				Protocol: "http",
				Port:     vars.FirstPort,
				URL:      "/",
				AllUsers: true,
			},
		},
	}

	return marshalJSON(config)
}
