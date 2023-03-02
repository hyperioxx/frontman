package frontman

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"plugin"
	"strings"
)

// Plugin is the interface that plugins must implement.
type Plugin interface {
    Name() string
    Init(bs *BackendServices) error
    Close() error
}

// loadPlugins loads all plugins in the specified directory.
func loadPlugins(dir string, bs *BackendServices) error {
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        return err
    }

    for _, file := range files {
        if !strings.HasSuffix(file.Name(), ".so") {
            continue
        }

        path := filepath.Join(dir, file.Name())
        p, err := plugin.Open(path)
        if err != nil {
            return err
        }

        sym, err := p.Lookup("myPlugin")
        if err != nil {
            return err
        }

        // Check if the symbol implements the Plugin interface.
        plugin, ok := sym.(Plugin)
        if !ok {
            return fmt.Errorf("%s does not implement the Plugin interface", path)
        }

        if err := plugin.Init(bs); err != nil {
            return err
        }

        log.Printf("Loaded plugin %s from %s\n", plugin.Name(), path)
    }

    return nil
}

