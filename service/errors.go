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
