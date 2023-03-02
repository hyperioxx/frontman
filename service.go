package frontman

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"
)

// BackendService holds the details of a backend service
type BackendService struct {
    Name          string        `json:"name"`
    Scheme        string        `json:"scheme"`
    URL           string        `json:"url"`
    Path          string        `json:"path"`
    Domain        string        `json:"domain"`
    HealthCheck   string        `json:"healthCheck"`
    RetryAttempts int           `json:"retryAttempts"`
    Timeout       time.Duration `json:"timeout"`
    MaxIdleConns  int           `json:"maxIdleConns"`
    MaxIdleTime   time.Duration `json:"maxIdleTime"`
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
            scheme TEXT NOT NULL,
            url TEXT NOT NULL,
            path TEXT NOT NULL,
            domain TEXT,
            health_check TEXT,
            retry_attempts INT,
            timeout INT,
            max_idle_conns INT,
            max_idle_time INT
        );
    `)
    return err
}


// loadServices retrieves the list of backend services from the database
func (bs *BackendServices) loadServices() error {
    rows, err := bs.db.Query("SELECT name, scheme, url, path, domain, health_check, retry_attempts, timeout, max_idle_conns, max_idle_time FROM services")
    if err != nil {
        return err
    }
    defer rows.Close()
    for rows.Next() {
        var service BackendService
        if err := rows.Scan(&service.Name, &service.Scheme, &service.URL, &service.Path, &service.Domain, &service.HealthCheck, &service.RetryAttempts, &service.Timeout, &service.MaxIdleConns, &service.MaxIdleTime); err != nil {
            return err
        }
        bs.services = append(bs.services, &service)
    }
    return rows.Err()
}


// AddService adds a new backend service to the database and the list
func (bs *BackendServices) AddService(service *BackendService) error {
    tx, err := bs.db.Begin()
    if err != nil {
        log.Println("Error beginning transaction:", err)
        return err
    }

    _, err = tx.Exec(`
        INSERT INTO services(name, scheme, url, path, domain, health_check, retry_attempts, timeout, max_idle_conns, max_idle_time)
        VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, service.Name, service.Scheme, service.URL, service.Path, service.Domain, service.HealthCheck, service.RetryAttempts, service.Timeout, service.MaxIdleConns, service.MaxIdleTime)
    if err != nil {
        log.Println("Error executing insert statement:", err)
        return err
    }

    if err := tx.Commit(); err != nil {
        log.Println("Error committing transaction:", err)
        return err
    }

    bs.Lock()
    defer bs.Unlock()
    bs.services = append(bs.services, service)

    return nil
}

// UpdateService updates an existing backend service in the database and the list
func (bs *BackendServices) UpdateService(service *BackendService) error {
    tx, err := bs.db.Begin()
    if err != nil {
        log.Println("Error beginning transaction:", err)
        return err
    }

    _, err = tx.Exec(`
        UPDATE services SET scheme = $1, url = $2, path = $3, domain = $4, health_check = $5, retry_attempts = $6, timeout = $7, max_idle_conns = $8, max_idle_time = $9 WHERE name = $10
    `, service.Scheme, service.URL, service.Path, service.Domain, service.HealthCheck, service.RetryAttempts, service.Timeout, service.MaxIdleConns, service.MaxIdleTime, service.Name)
    if err != nil {
        log.Println("Error executing update statement:", err)
        return err
    }

    if err := tx.Commit(); err != nil {
        log.Println("Error committing transaction:", err)
        return err
    }

    bs.Lock()
    defer bs.Unlock()
    for i, s := range bs.services {
        if s.Name == service.Name {
            bs.services[i] = service
            break
        }
    }

    return nil
}

// HealthCheck performs a health check on the backend service and returns true if it is healthy.
func (bs *BackendService) GetHealthCheck() bool {
    resp, err := http.Get(bs.HealthCheck)
    if err != nil {
        log.Printf("Error performing health check for service %s: %s", bs.Name, err.Error())
        return false
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
        return true
    }

    log.Printf("Service %s health check failed with status code %d", bs.Name, resp.StatusCode)
    return false
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



