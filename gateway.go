package frontman

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hyperioxx/frontman/config"
	"github.com/hyperioxx/frontman/plugins"
	"github.com/hyperioxx/frontman/service"
)

// Gateway contains the backend services and the router
type Gateway struct {
	router          *mux.Router
	service         *mux.Router
	backendServices service.ServiceRegistry
	conf            *config.Config
}

func NewServicesRouter(backendServices service.ServiceRegistry) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/services", getServicesHandler(backendServices)).Methods("GET")
	router.HandleFunc("/api/services", addServiceHandler(backendServices)).Methods("POST")
	router.HandleFunc("/api/services/{name}", removeServiceHandler(backendServices)).Methods("DELETE")
	router.HandleFunc("/api/services/{name}", updateServiceHandler(backendServices)).Methods("PUT")
	router.HandleFunc("/api/health", getHealthHandler(backendServices)).Methods("GET")

	return router
}

// NewGateway creates a new Gateway instance with a Redis client connection factory
func NewGateway(conf *config.Config) (*Gateway, error) {

	// Retrieve the Redis client connection from the factory
	ctx := context.Background()

	// Create a new BackendServices instance
	backendServices, err := service.NewServiceRegistry(ctx, conf.GlobalConfig.ServiceType, conf)
	if err != nil {
		return nil, err
	}

	servicesRouter := NewServicesRouter(backendServices)

	// Create a new router instance
	proxyRouter := mux.NewRouter()

	// Load plugins
	var plug []plugins.FrontmanPlugin

	if conf.PluginConfig.Enabled {
		plug, err = plugins.LoadPlugins(conf.PluginConfig.Order)
		if err != nil {
			return nil, err
		}

	}

	proxyRouter.HandleFunc("/{proxyPath:.+}", gatewayHandler(backendServices, plug, conf)).Methods("GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
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
		router:          proxyRouter,
		service:         servicesRouter,
		backendServices: backendServices,
		conf:            conf,
	}, nil
}

func (gw *Gateway) Start() error {
	apiAddr := gw.conf.APIConfig.Addr
	if apiAddr == "" {
		apiAddr = "0.0.0.0:8080"
	}
	gatewayAddr := gw.conf.GatewayConfig.Addr
	if gatewayAddr == "" {
		gatewayAddr = "0.0.0.0:8000"
	}

	var apiHandler http.Handler
	var gatewayHandler http.Handler

	if gw.conf.APIConfig.SSL.Enabled {
		apiHandler = gw.service
		cert, err := loadCert(gw.conf.APIConfig.SSL.Cert, gw.conf.APIConfig.SSL.Key)
		if err != nil {
			return err
		}
		apiServer := createServer(apiAddr, apiHandler, &cert)
		log.Println("Starting Frontman Gateway with SSL...")
		go startServer(apiServer)
	} else {
		apiHandler = gw.service
		api := createServer(apiAddr, apiHandler, nil)
		log.Println("Starting Frontman Gateway...")
		go startServer(api)
	}

	if gw.conf.GatewayConfig.SSL.Enabled {
		gatewayHandler = gw.router
		cert, err := loadCert(gw.conf.GatewayConfig.SSL.Cert, gw.conf.GatewayConfig.SSL.Key)
		if err != nil {
			return err
		}
		gatewayServer := createServer(gatewayAddr, gatewayHandler, &cert)
		log.Println("Starting Gateway with SSL...")
		startServer(gatewayServer)
	} else {
		gatewayHandler = gw.router
		gateway := createServer(gatewayAddr, gatewayHandler, nil)
		log.Println("Starting Gateway...")
		startServer(gateway)
	}

	return nil
}

func loadCert(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
		return tls.Certificate{}, err
	}
	return cert, nil
}

func createServer(addr string, handler http.Handler, cert *tls.Certificate) *http.Server {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	if cert != nil {
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{*cert},
		}
	}
	return server
}

func startServer(server *http.Server) {
	if server.TLSConfig != nil {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("Failed to start server with TLS: %v", err)
		}
	} else {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server without TLS: %v", err)
		}
	}
}
