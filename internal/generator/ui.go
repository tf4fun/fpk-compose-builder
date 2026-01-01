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
// Creates a default configuration based on appname and port
// Note: fnOS requires entry keys to use appname as prefix
func GenerateDefaultUIConfig(vars parser.Variables, appname string) (string, error) {
	appEntry := appname + ".Application"

	config := map[string]interface{}{
		".url": map[string]interface{}{
			appEntry: UIConfigEntry{
				Title:    appname,
				Icon:     "images/icon-{0}.png",
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
