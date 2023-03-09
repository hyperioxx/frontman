package service

// MemoryServiceRegistry is an in-memory implementation of the ServiceRegistry interface
type MemoryServiceRegistry struct {
	Services map[string]*BackendService
}

// NewMemoryServiceRegistry creates a new MemoryServiceRegistry instance
func NewMemoryServiceRegistry() *MemoryServiceRegistry {
	return &MemoryServiceRegistry{
		Services: make(map[string]*BackendService),
	}
}

// GetService retrieves a backend service by name
func (r *MemoryServiceRegistry) GetService(name string) (*BackendService, error) {
	service, ok := r.Services[name]
	if !ok {
		return nil, ErrServiceNotFound{Name: name}
	}
	return service, nil
}

// GetServices retrieves all backend services
func (r *MemoryServiceRegistry) GetServices() []*BackendService {
	services := make([]*BackendService, 0, len(r.Services))
	for _, service := range r.Services {
		services = append(services, service)
	}
	return services
}

// AddService adds a new backend service
func (r *MemoryServiceRegistry) AddService(service *BackendService) error {
	if _, ok := r.Services[service.Name]; ok {
		return ErrServiceExists{Name: service.Name}
	}
	r.Services[service.Name] = service
	return nil
}

// UpdateService updates an existing backend service
func (r *MemoryServiceRegistry) UpdateService(service *BackendService) error {
	if _, ok := r.Services[service.Name]; !ok {
		return ErrServiceNotFound{Name: service.Name}
	}
	r.Services[service.Name] = service
	return nil
}

// RemoveService removes a backend service by name
func (r *MemoryServiceRegistry) RemoveService(name string) error {
	if _, ok := r.Services[name]; !ok {
		return ErrServiceNotFound{Name: name}
	}
	delete(r.Services, name)
	return nil
}
