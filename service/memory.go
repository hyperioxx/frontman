package service

// MemoryServiceRegistry is an in-memory implementation of the ServiceRegistry interface
type MemoryServiceRegistry struct {
	*baseRegistry
	Services map[string]*BackendService
}

// NewMemoryServiceRegistry creates a new MemoryServiceRegistry instance
func NewMemoryServiceRegistry(br *baseRegistry) *MemoryServiceRegistry {
	return &MemoryServiceRegistry{
		baseRegistry: br,
		Services:     make(map[string]*BackendService),
	}
}

// AddService adds a new backend service
func (r *MemoryServiceRegistry) AddService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.addService(service, func() error {
		r.Services[service.Name] = service
		return nil
	})
}

// UpdateService updates an existing backend service
func (r *MemoryServiceRegistry) UpdateService(service *BackendService) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.updateService(service, func() error {
		r.Services[service.Name] = service
		return nil
	})
}

// RemoveService removes a backend service by name
func (r *MemoryServiceRegistry) RemoveService(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.removeService(name, func() error {
		delete(r.Services, name)
		return nil
	})
}

// GetServices retrieves all backend services
func (r *MemoryServiceRegistry) GetServices() []*BackendService {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.getServices()
}
