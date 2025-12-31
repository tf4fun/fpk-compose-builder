package generator

import (
	"encoding/json"

	"fpk-compose-builder/internal/parser"
)

// PrivilegeConfig represents the config/privilege JSON structure
type PrivilegeConfig struct {
	Defaults  PrivilegeDefaults `json:"defaults"`
	Username  string            `json:"username,omitempty"`
	Groupname string            `json:"groupname,omitempty"`
}

// PrivilegeDefaults represents the defaults section of privilege config
type PrivilegeDefaults struct {
	RunAs string `json:"run-as"`
}

// ResourceConfig represents the config/resource JSON structure
type ResourceConfig struct {
	DockerProject *DockerProjectConfig `json:"docker-project,omitempty"`
}

// DockerProjectConfig represents the docker-project section of resource config
type DockerProjectConfig struct {
	Projects []DockerProject `json:"projects"`
}

// DockerProject represents a single docker project entry
type DockerProject struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// GeneratePrivilege generates the config/privilege JSON content
// Default: run-as: package with username/groupname based on service name
func GeneratePrivilege(vars parser.Variables) (string, error) {
	config := PrivilegeConfig{
		Defaults: PrivilegeDefaults{
			RunAs: "package",
		},
		Username:  vars.ServiceName,
		Groupname: vars.ServiceName,
	}

	return marshalJSON(config)
}

// GenerateResource generates the config/resource JSON content
// Default: docker-project configuration pointing to the docker directory
func GenerateResource(vars parser.Variables) (string, error) {
	config := ResourceConfig{
		DockerProject: &DockerProjectConfig{
			Projects: []DockerProject{
				{
					Name: vars.ServiceName,
					Path: "docker",
				},
			},
		},
	}

	return marshalJSON(config)
}

// marshalJSON marshals data to indented JSON string
func marshalJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes) + "\n", nil
}
