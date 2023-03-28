package service

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

// YAMLServiceRegistry implements the ServiceRegistry interface
type YAMLServiceRegistry struct {
	baseRegistry
	filename string
}

// NewYAMLServiceRegistry creates a new YAMLServiceRegistry instance from a file
func NewYAMLServiceRegistry(filename string) (*YAMLServiceRegistry, error) {
	reg := &YAMLServiceRegistry{filename: filename}
	err := reg.readFromFile(filename)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

// AddService adds a new backend service to the registry
func (r *YAMLServiceRegistry) AddService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	err := r.addService(service, func() error {
		return r.writeToFile(r.filename)
	})

	return err
}

// UpdateService updates an existing backend service in the registry
func (r *YAMLServiceRegistry) UpdateService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	err := r.updateService(service, func() error {
		return r.writeToFile(r.filename)
	})

	return err
}

// RemoveService removes a backend service from the registry
func (r *YAMLServiceRegistry) RemoveService(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	err := r.removeService(name, func() error {
		return r.writeToFile(r.filename)
	})

	return err
}

// GetServices returns a copy of the current list of backend services
func (r *YAMLServiceRegistry) GetServices() []*BackendService {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.baseRegistry.getServices()
}

// readFromFile reads service data from a YAML file and updates the registry
func (r *YAMLServiceRegistry) readFromFile(filename string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// Create an empty file if it doesn't exist
		err = ioutil.WriteFile(filename, []byte{}, 0644)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var services []*BackendService
	err = yaml.Unmarshal(data, &services)
	if err != nil {
		return err
	}
	for _, service := range services {
		service.Init()
	}

	r.services = services

	return nil

}

// WriteToFile writes the current registry data to a YAML file
func (r *YAMLServiceRegistry) writeToFile(filename string) error {
	data, err := yaml.Marshal(r.services)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}
