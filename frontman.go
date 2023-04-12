package frontman

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Frontman-Labs/frontman/api"
	"github.com/julienschmidt/httprouter"
	"net/http"

	"github.com/Frontman-Labs/frontman/config"
	"github.com/Frontman-Labs/frontman/gateway"
	"github.com/Frontman-Labs/frontman/log"
	"github.com/Frontman-Labs/frontman/plugins"
	"github.com/Frontman-Labs/frontman/service"
	"github.com/Frontman-Labs/frontman/ssl"
)

// Frontman contains the backend services and the router
type Frontman struct {
	router          *gateway.APIGateway
	service         *httprouter.Router
	backendServices service.ServiceRegistry
	conf            *config.Config
	log             log.Logger
}

// NewFrontman creates a new Frontman instance with a Redis client connection factory
func NewFrontman(conf *config.Config, log log.Logger) (*Frontman, error) {
	ctx := context.Background()

	// Create a new serviceRegistry instance
	serviceRegistry, err := service.NewServiceRegistry(ctx, conf.GlobalConfig.ServiceType, conf)
	if err != nil {
		return nil, err
	}

	// Create management API router
	servicesRouter := api.NewServicesRouter(serviceRegistry)

	// Load plugins
	var plug []plugins.FrontmanPlugin

	if conf.PluginConfig.Enabled {
		plug, err = plugins.LoadPlugins(conf.PluginConfig.Order)
		if err != nil {
			return nil, err
		}

	}

	// Create new APIGateway instance
	apiGateway := gateway.NewAPIGateway(serviceRegistry, plug, conf, log)

	// Create the Frontman instance
	return &Frontman{
		router:          apiGateway,
		service:         servicesRouter,
		backendServices: serviceRegistry,
		conf:            conf,
		log:             log,
	}, nil
}

func (gw *Frontman) Start() error {
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

	apiHandler = gw.service
	var apicert *tls.Certificate
	if gw.conf.APIConfig.SSL.Enabled {
		cert, err := ssl.LoadCert(gw.conf.APIConfig.SSL.Cert, gw.conf.APIConfig.SSL.Key)
		if err != nil {
			return err
		}
		apicert = cert
	}
	apiHandler = gw.service
	api := createServer(apiAddr, apiHandler, apicert)
	go func() {
		if err := startServer(api); err != nil {
			gw.log.Fatal(err)
		}
	}()
	gw.log.WithFields(log.InfoLevel, fmt.Sprintf("Started Frontman API on %s", apiAddr), log.Bool("tls_enabled", gw.conf.APIConfig.SSL.Enabled))

	var gwcert *tls.Certificate
	gatewayHandler = gw.router
	if gw.conf.GatewayConfig.SSL.Enabled {
		cert, err := ssl.LoadCert(gw.conf.GatewayConfig.SSL.Cert, gw.conf.GatewayConfig.SSL.Key)
		if err != nil {
			return err
		}
		gwcert = cert
		// Redirect HTTP traffic to HTTPS
		httpAddr := "0.0.0.0:80"
		httpRedirect := createRedirectServer(httpAddr, gatewayAddr)
		gw.log.Infof("Started HTTP redirect server on %s", httpAddr)
		go func() {
			if err := startServer(httpRedirect); err != nil {
				gw.log.Fatal(err)
			}
		}()
	}
	gatewayHandler = gw.router
	gateway := createServer(gatewayAddr, gatewayHandler, gwcert)
	gw.log.WithFields(log.InfoLevel, fmt.Sprintf("Started Frontman Frontman on %s", gatewayAddr), log.Bool("tls_enabled", gw.conf.GatewayConfig.SSL.Enabled))
	if err := startServer(gateway); err != nil {
		return err
	}

	return nil
}

func createRedirectServer(addr string, redirectAddr string) *http.Server {
	redirect := func(w http.ResponseWriter, req *http.Request) {
		httpsURL := "https://" + req.Host + req.URL.Path
		http.Redirect(w, req, httpsURL, http.StatusMovedPermanently)
	}
	return &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(redirect),
	}
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

func startServer(server *http.Server) error {
	if server.TLSConfig != nil {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			return fmt.Errorf("Failed to start server with TLS: %w", err)
		}
	} else {
		if err := server.ListenAndServe(); err != nil {
			return fmt.Errorf("Failed to start server without TLS: %w", err)
		}
	}
	return nil
}
