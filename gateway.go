package frontman

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Gateway contains the backend services and the router
type Gateway struct {
    router          *mux.Router
    backendServices *BackendServices
}

// NewDB creates a new *sql.DB instance given a database URI
func NewDB(uri string) (*sql.DB, error) {
    db, err := sql.Open("postgres", uri)
    if err != nil {
        return nil, err
    }
    return db, nil
}

// NewGateway creates a new Gateway instance with a database connection factory
func NewGateway(dbFactory func() (*sql.DB, error)) (*Gateway, error) {
    // Retrieve the database connection from the factory
    db, err := dbFactory()
    if err != nil {
        return nil, err
    }

    // Create a new router instance
    r := mux.NewRouter()

    // Create a new BackendServices instance
    backendServices, err := NewBackendServices(db)
    if err != nil {
        return nil, err
    }

    // Define your API endpoints and services using the router
    r.HandleFunc("/api/services", getServicesHandler(backendServices)).Methods("GET")
    r.HandleFunc("/api/services", addServiceHandler(backendServices)).Methods("POST")
    r.HandleFunc("/api/services/{name}", removeServiceHandler(backendServices)).Methods("DELETE")
	r.HandleFunc("/api/services/{name}", updateServiceHandler(backendServices)).Methods("PUT")
	r.HandleFunc("/api/health", getHealthHandler(backendServices)).Methods("GET")
    r.HandleFunc("/{proxyPath:.+}", reverseProxyHandler(backendServices)).Methods("GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
        vars := mux.Vars(r)
        proxyPath := vars["proxyPath"]
        for _, prefix := range []string{"/api/"} {
            if strings.HasPrefix(proxyPath, prefix) {
                return false
            }
        }
        return true
    })

    // Create the Gateway instance
    return &Gateway{
        router:          r,
        backendServices: backendServices,
    }, nil
}


// Start starts the server
func (gw *Gateway) Start() error {
    // Start the server
    log.Println("Starting Frontman Gateway...")
    return http.ListenAndServe(":8080", gw.router)
}


