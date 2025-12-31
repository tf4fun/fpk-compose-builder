package parser

// XFnpack represents the x-fnpack extension field in docker-compose.yaml
// manifest is a YAML object that will be converted to key=value format
// Other fields are file paths with their content as multi-line text
type XFnpack struct {
	// Manifest contains app metadata as YAML object, converted to key=value format
	Manifest map[string]interface{} `yaml:"manifest,omitempty"`

	// Files contains all file paths and their content (multi-line text)
	// Key is the file path (e.g., "wizard/install", "app/ui/config", "config/custom")
	// Value is the file content as string
	Files map[string]string `yaml:"-"`

	// RawContent stores the raw x-fnpack content for file extraction
	RawContent map[string]interface{} `yaml:"-"`
}

// ComposeFile represents a docker-compose.yaml file with x-fnpack extension
type ComposeFile struct {
	// XFnpack contains the fnOS app configuration
	XFnpack XFnpack `yaml:"x-fnpack,omitempty"`

	// Services contains the docker service definitions
	Services map[string]Service `yaml:"services,omitempty"`

	// Networks contains network definitions
	Networks map[string]interface{} `yaml:"networks,omitempty"`

	// Volumes contains volume definitions
	Volumes map[string]interface{} `yaml:"volumes,omitempty"`
}

// Service represents a docker service definition
type Service struct {
	// Image is the docker image name
	Image string `yaml:"image,omitempty"`

	// ContainerName is the container name
	ContainerName string `yaml:"container_name,omitempty"`

	// Ports is the list of port mappings (e.g., "3000:3000")
	Ports []string `yaml:"ports,omitempty"`

	// Environment is the list of environment variables
	Environment []string `yaml:"environment,omitempty"`

	// Volumes is the list of volume mappings
	Volumes []string `yaml:"volumes,omitempty"`

	// Restart is the restart policy
	Restart string `yaml:"restart,omitempty"`

	// Networks is the list of networks
	Networks []string `yaml:"networks,omitempty"`

	// SecurityOpt is the list of security options
	SecurityOpt []string `yaml:"security_opt,omitempty"`

	// ShmSize is the shared memory size
	ShmSize string `yaml:"shm_size,omitempty"`

	// DependsOn is the list of service dependencies
	DependsOn []string `yaml:"depends_on,omitempty"`

	// Labels is the map of labels
	Labels map[string]string `yaml:"labels,omitempty"`

	// Command is the container command
	Command interface{} `yaml:"command,omitempty"`

	// Entrypoint is the container entrypoint
	Entrypoint interface{} `yaml:"entrypoint,omitempty"`

	// WorkingDir is the working directory
	WorkingDir string `yaml:"working_dir,omitempty"`

	// User is the user to run as
	User string `yaml:"user,omitempty"`

	// Privileged indicates if the container runs in privileged mode
	Privileged bool `yaml:"privileged,omitempty"`

	// CapAdd is the list of capabilities to add
	CapAdd []string `yaml:"cap_add,omitempty"`

	// CapDrop is the list of capabilities to drop
	CapDrop []string `yaml:"cap_drop,omitempty"`

	// Devices is the list of devices to map
	Devices []string `yaml:"devices,omitempty"`

	// ExtraHosts is the list of extra hosts
	ExtraHosts []string `yaml:"extra_hosts,omitempty"`

	// Logging is the logging configuration
	Logging interface{} `yaml:"logging,omitempty"`

	// Healthcheck is the health check configuration
	Healthcheck interface{} `yaml:"healthcheck,omitempty"`

	// Deploy is the deployment configuration
	Deploy interface{} `yaml:"deploy,omitempty"`
}

// Variables contains the extracted variables for template substitution
type Variables struct {
	// ServiceName is the name of the first service
	ServiceName string

	// ContainerName is the container_name of the first service
	// Falls back to ServiceName if not specified
	ContainerName string

	// FirstPort is the first port mapping of the first service
	// Extracted from the host port (e.g., "3000" from "3000:8080")
	FirstPort string

	// ImageOrg is the organization/user from the docker image name
	// Extracted from "org/image:tag" format, e.g., "lobehub" from "lobehub/lobe-chat"
	// Falls back to image name if no org (e.g., "alpine" from "alpine:latest")
	ImageOrg string

	// ImageName is the image name without org and tag
	// e.g., "lobe-chat" from "lobehub/lobe-chat:latest"
	ImageName string
}
