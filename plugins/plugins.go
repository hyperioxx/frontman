package plugins

import (
	"fmt"

	"net/http"

	"plugin"


	"github.com/hyperioxx/frontman/config"
	"github.com/hyperioxx/frontman/service"
)

// FrontmanPlugin is an interface for creating plugins for Frontman.
type FrontmanPlugin interface {
    // Name returns the name of the plugin.
    Name() string

    // PreRequest is called before sending the request to the target service.
    // The method takes in the original request, a ServiceRegistry, and a Config.
    // An error is returned if the plugin encounters any issues.
    PreRequest(*http.Request, service.ServiceRegistry, *config.Config) error

    // PostResponse is called after receiving the response from the target service.
    // The method takes in the response, a ServiceRegistry, and a Config.
    // An error is returned if the plugin encounters any issues.
    PostResponse(*http.Response, service.ServiceRegistry, *config.Config) error

    // Close is called when the plugin is being shut down.
    // An error is returned if the plugin encounters any issues.
    Close() error
}



// LoadPlugins loads the plugins in the specified order and returns a slice of FrontmanPlugin instances.
func LoadPlugins(pluginPaths []string) ([]FrontmanPlugin, error) {
	plugins := make([]FrontmanPlugin, 0)

	// Iterate through each plugin file path
	for _, path := range pluginPaths {
		// Load the plugin
		p, err := plugin.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin %s: %v", path, err)
		}

		// Get the symbol for the FrontmanPlugin instance
		sym, err := p.Lookup("FrontmanPlugin")
		if err != nil {
			return nil, fmt.Errorf("failed to get symbol for plugin %s: %v", path, err)
		}

		// Check that the symbol is of the correct type
		frontmanPlugin, ok := sym.(FrontmanPlugin)
		if !ok {
			return nil, fmt.Errorf("symbol for plugin %s is not of type FrontmanPlugin", path)
		}

		// Add the plugin to the slice of plugins
		plugins = append(plugins, frontmanPlugin)
	}

	return plugins, nil
}
