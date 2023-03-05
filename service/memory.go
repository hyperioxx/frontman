package service

// MemoryServiceRegistry is an in-memory implementation of the ServiceRegistry interface
type MemoryServiceRegistry struct {
	services map[string]*BackendService
}

// NewMemoryServiceRegistry creates a new MemoryServiceRegistry instance
func NewMemoryServiceRegistry() *MemoryServiceRegistry {
	return &MemoryServiceRegistry{
		services: make(map[string]*BackendService),
	}
}

// GetService retrieves a backend service by name
func (r *MemoryServiceRegistry) GetService(name string) (*BackendService, error) {
	service, ok := r.services[name]
	if !ok {
		return nil, ErrServiceNotFound{Name: name}
	}
	return service, nil
}

// GetServices retrieves all backend services
func (r *MemoryServiceRegistry) GetServices() []*BackendService {
	services := make([]*BackendService, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}
	return services
}

// AddService adds a new backend service
func (r *MemoryServiceRegistry) AddService(service *BackendService) error {
	if _, ok := r.services[service.Name]; ok {
		return ErrServiceExists{Name: service.Name}
	}
	r.services[service.Name] = service
	return nil
}

// UpdateService updates an existing backend service
func (r *MemoryServiceRegistry) UpdateService(service *BackendService) error {
	if _, ok := r.services[service.Name]; !ok {
		return ErrServiceNotFound{Name: service.Name}
	}
	r.services[service.Name] = service
	return nil
}

// RemoveService removes a backend service by name
func (r *MemoryServiceRegistry) RemoveService(name string) error {
	if _, ok := r.services[name]; !ok {
		return ErrServiceNotFound{Name: name}
	}
	delete(r.services, name)
	return nil
}
