package service

import (
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/go-yaml/yaml"
)

// YAMLServiceRegistry implements the ServiceRegistry interface
type YAMLServiceRegistry struct {
	filename string
	services []*BackendService
	mutex    sync.RWMutex
}

// NewYAMLServiceRegistry creates a new YAMLServiceRegistry instance from a file
func NewYAMLServiceRegistry(filename string) (*YAMLServiceRegistry, error) {
	reg := &YAMLServiceRegistry{filename: filename}
	err := reg.ReadFromFile(filename)
	if err != nil {
		return nil, err
	}
	return reg, nil
}

// AddService adds a new backend service to the registry
func (r *YAMLServiceRegistry) AddService(service *BackendService) error {
	r.mutex.Lock()
	r.services = append(r.services, service)
	r.mutex.Unlock()
	return r.writeToFile(r.filename)
}

// UpdateService updates an existing backend service in the registry
func (r *YAMLServiceRegistry) UpdateService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for i, s := range r.services {
		if s.Name == service.Name {
			r.services[i] = service
			return r.writeToFile(r.filename)
		}
	}
	return errors.New("service not found")
}

// RemoveService removes a backend service from the registry
func (r *YAMLServiceRegistry) RemoveService(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for i, s := range r.services {
		if s.Name == name {
			r.services = append(r.services[:i], r.services[i+1:]...)
			return r.writeToFile(r.filename)
		}
	}
	return errors.New("service not found")
}

// GetServices returns a copy of the current list of backend services
func (r *YAMLServiceRegistry) GetServices() []*BackendService {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	services := make([]*BackendService, len(r.services))
	copy(services, r.services)
	return services
}


// ReadFromFile reads service data from a YAML file and updates the registry
func (r *YAMLServiceRegistry) ReadFromFile(filename string) error {
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

	r.services = services
	return nil

}

// WriteToFile writes the current registry data to a YAML file
func (r *YAMLServiceRegistry) writeToFile(filename string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	data, err := yaml.Marshal(r.services)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}
