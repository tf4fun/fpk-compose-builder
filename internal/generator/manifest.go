package generator

import (
	"fmt"
	"sort"
	"strings"

	"fpk-compose-builder/internal/parser"
)

// ManifestDefaults contains default values for manifest fields
var ManifestDefaults = map[string]string{
	"arch":            "x86_64",
	"source":          "thirdparty",
	"desktop_uidir":   "ui",
	"version":         "1.0.0",
}

// ManifestFieldOrder defines the order of fields in the manifest file
var ManifestFieldOrder = []string{
	"appname",
	"version",
	"display_name",
	"desc",
	"arch",
	"source",
	"maintainer",
	"maintainer_url",
	"distributor",
	"distributor_url",
	"os_min_ver",
	"beta",
	"reloadui",
	"desktop_uidir",
	"desktop_applaunchname",
	"changelog",
	"ctl_stop",
	"checkport",
	"install_type",
	"service_port",
}

// GenerateManifest generates manifest content in key=value format from YAML object
// It applies default values for missing fields and replaces variables
func GenerateManifest(manifest map[string]interface{}, vars parser.Variables) string {
	// Create a working copy with defaults applied
	result := make(map[string]string)

	// Apply defaults first
	for key, defaultValue := range ManifestDefaults {
		result[key] = defaultValue
	}

	// Apply variable-based defaults
	if vars.ServiceName != "" {
		result["appname"] = vars.ServiceName
		result["display_name"] = vars.ServiceName
		result["desktop_applaunchname"] = vars.ServiceName + ".Application"
	}

	// Apply image-based defaults for maintainer/distributor and desc
	if vars.ImageOrg != "" {
		result["maintainer"] = vars.ImageOrg
		result["distributor"] = vars.ImageOrg
	}

	// Use image name as default desc
	if vars.ImageName != "" {
		result["desc"] = vars.ImageName
	} else if vars.ServiceName != "" {
		result["desc"] = vars.ServiceName
	}

	// Use first port as default service_port
	if vars.FirstPort != "" {
		result["service_port"] = vars.FirstPort
	}

	// Override with provided manifest values
	if manifest != nil {
		for key, value := range manifest {
			strValue := formatManifestValue(value)
			// Replace variables in the value
			strValue = ReplaceVariables(strValue, vars)
			result[key] = strValue
		}
	}

	// Generate desktop_applaunchname if not specified but appname is
	if _, hasDesktopAppname := manifest["desktop_applaunchname"]; !hasDesktopAppname {
		// Also check for legacy desktop_appname
		if _, hasLegacy := manifest["desktop_appname"]; !hasLegacy {
			if appname, ok := result["appname"]; ok && appname != "" {
				result["desktop_applaunchname"] = appname + ".Application"
			}
		}
	}

	// Build output in defined order
	var lines []string
	addedKeys := make(map[string]bool)

	// Add fields in defined order first
	for _, key := range ManifestFieldOrder {
		if value, ok := result[key]; ok && value != "" {
			lines = append(lines, formatManifestLine(key, value))
			addedKeys[key] = true
		}
	}

	// Add any remaining fields not in the predefined order (sorted alphabetically)
	var remainingKeys []string
	for key := range result {
		if !addedKeys[key] && result[key] != "" {
			remainingKeys = append(remainingKeys, key)
		}
	}
	sort.Strings(remainingKeys)

	for _, key := range remainingKeys {
		lines = append(lines, formatManifestLine(key, result[key]))
	}

	return strings.Join(lines, "\n") + "\n"
}

// formatManifestValue converts various types to string for manifest
func formatManifestValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "yes"
		}
		return "no"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatManifestLine formats a key-value pair for the manifest file
// Uses padding to align values for readability
func formatManifestLine(key, value string) string {
	// Pad key to 16 characters for alignment (like the reference manifest)
	if len(key) < 16 {
		return fmt.Sprintf("%-16s= %s", key, value)
	}
	return fmt.Sprintf("%s = %s", key, value)
}

// GetManifestAppname extracts the appname from manifest or returns default
func GetManifestAppname(manifest map[string]interface{}, vars parser.Variables) string {
	if manifest != nil {
		if appname, ok := manifest["appname"]; ok {
			return formatManifestValue(appname)
		}
	}
	return vars.ServiceName
}

// GetManifestVersion extracts the version from manifest or returns default
func GetManifestVersion(manifest map[string]interface{}) string {
	if manifest != nil {
		if version, ok := manifest["version"]; ok {
			return formatManifestValue(version)
		}
	}
	return ManifestDefaults["version"]
}
