package service

import "fmt"

type ErrServiceExists struct {
	Name string
}

func (e ErrServiceExists) Error() string {
	return fmt.Sprintf("service with name '%s' already exists", e.Name)
}

type ErrServiceNotFound struct {
	Name string
}

func (e ErrServiceNotFound) Error() string {
	return fmt.Sprintf("service with name '%s' not found", e.Name)
}

type ErrUnsupportedServiceType struct {
	serviceType string
}

func (e ErrUnsupportedServiceType) Error() string {
	return fmt.Sprintf("unsupported service type: %s", e.serviceType)
}
