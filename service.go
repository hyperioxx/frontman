package frontman

import (
	"database/sql"
	"sync"
	"time"
)

// BackendService holds the details of a backend service
type BackendService struct {
    Name          string
    URL           string
    HealthCheck   string
    RetryAttempts int
    Timeout       time.Duration
    MaxIdleConns  int
    MaxIdleTime   time.Duration
}


// BackendServices holds a list of backend services
type BackendServices struct {
    db       *sql.DB
    services []*BackendService
    sync.RWMutex
}

// NewBackendServices creates a new BackendServices instance with a database connection
func NewBackendServices(db *sql.DB) (*BackendServices, error) {
    bs := &BackendServices{db: db}
    if err := bs.ensureTableExists(); err != nil {
        return nil, err
    }
    if err := bs.loadServices(); err != nil {
        return nil, err
    }
    return bs, nil
}

// ensureTableExists creates the services table if it does not exist
func (bs *BackendServices) ensureTableExists() error {
    _, err := bs.db.Exec(`
        CREATE TABLE IF NOT EXISTS services (
            name TEXT PRIMARY KEY,
            url TEXT NOT NULL
        );
    `)
    return err
}

// loadServices retrieves the list of backend services from the database
func (bs *BackendServices) loadServices() error {
    rows, err := bs.db.Query("SELECT name, url FROM services")
    if err != nil {
        return err
    }
    defer rows.Close()
    for rows.Next() {
        var service BackendService
        if err := rows.Scan(&service.Name, &service.URL); err != nil {
            return err
        }
        bs.services = append(bs.services, &service)
    }
    return rows.Err()
}

// AddService adds a new backend service to the database and the list
func (bs *BackendServices) AddService(service *BackendService) error {
    _, err := bs.db.Exec("INSERT INTO services(name, url) VALUES($1, $2)", service.Name, service.URL)
    if err != nil {
        return err
    }
    bs.Lock()
    defer bs.Unlock()
    bs.services = append(bs.services, service)
    return nil
}

// RemoveService removes a backend service from the database and the list
func (bs *BackendServices) RemoveService(name string) error {
    _, err := bs.db.Exec("DELETE FROM services WHERE name = $1", name)
    if err != nil {
        return err
    }
    bs.Lock()
    defer bs.Unlock()
    for i, service := range bs.services {
        if service.Name == name {
            bs.services = append(bs.services[:i], bs.services[i+1:]...)
            break
        }
    }
    return nil
}

// GetServices returns a copy of the current list of backend services
func (bs *BackendServices) GetServices() []*BackendService {
    bs.RLock()
    defer bs.RUnlock()
    services := make([]*BackendService, len(bs.services))
    copy(services, bs.services)
    return services
}



