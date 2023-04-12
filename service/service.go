package service

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/Frontman-Labs/frontman/auth"
	"github.com/Frontman-Labs/frontman/config"
	"github.com/Frontman-Labs/frontman/loadbalancer"
	"github.com/Frontman-Labs/frontman/oauth"
)

// BackendService holds the details of a backend service
type BackendService struct {
	Name               string             `json:"name" yaml:"name"`
	Scheme             string             `json:"scheme" yaml:"scheme"`
	UpstreamTargets    []string           `json:"upstreamTargets" yaml:"upstreamTargets"`
	Path               string             `json:"path,omitempty" yaml:"path,omitempty"`
	Domain             string             `json:"domain" yaml:"domain"`
	HealthCheck        string             `json:"healthCheck" yaml:"healthCheck"`
	RetryAttempts      int                `json:"retryAttempts,omitempty" yaml:"retryAttempts,omitempty"`
	Timeout            time.Duration      `json:"timeout" yaml:"timeout"`
	MaxIdleConns       int                `json:"maxIdleConns,omitempty" yaml:"maxIdleConns,omitempty"`
	MaxIdleTime        time.Duration      `json:"maxIdleTime" yaml:"maxIdleTime"`
	StripPath          bool               `json:"stripPath,omitempty" yaml:"stripPath,omitempty"`
	AuthConfig         *config.AuthConfig `json:"auth,omitempty" yaml:"auth,omitempty"`
	LoadBalancerPolicy LoadBalancerPolicy `json:"loadBalancerPolicy,omitempty" yaml:"loadBalancerPolicy,omitempty"`
	RewriteMatch       string             `json:"rewriteMatch,omitempty" yaml:"rewriteMatch,omitempty"`
	RewriteReplace     string             `json:"rewriteReplace,omitempty" yaml:"rewriteReplace,omitempty"`

	httpClient           *http.Client
	compiledRewriteMatch *regexp.Regexp
	loadBalancer         loadbalancer.LoadBalancer
	provider             oauth.OAuthProvider
	tokenValidator       *auth.TokenValidator
}

type LoadBalancerPolicy struct {
	Type    string        `json:"type" yaml:"type"`
	Options PolicyOptions `json:"options,omitempty" yaml:"options,omitempty"`
}

type PolicyOptions struct {
	Weights []int `json:"weights,omitempty" yaml:"weights,omitempty"`
}

// GetHealthCheck performs a health check on the backend service and returns true if it is healthy.
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

func (bs *BackendService) setTokenValidator() {
	if bs.AuthConfig == nil {
		return
	}

	validator, err := auth.GetTokenValidator(*bs.AuthConfig)
	if err != nil {
		log.Printf("Error adding auth to backend service: %s: %s", bs.Name, err.Error())
	} else {
		bs.tokenValidator = &validator
	}
}

func (bs *BackendService) GetTokenValidator() auth.TokenValidator {
	if bs.AuthConfig != nil && bs.tokenValidator == nil {
		// Token validator has not been instantiated for this backend service
		// Instantiating here to avoid having to call setTokenValidator on each update/add
		bs.setTokenValidator()
	}
	return *bs.tokenValidator
}

func (bs *BackendService) GetUserDataHeader() string {
	if bs.AuthConfig.UserDataHeader != "" {
		return bs.AuthConfig.UserDataHeader
	}
	return "user"
}

func (bs *BackendService) GetLoadBalancer() loadbalancer.LoadBalancer {
	return bs.loadBalancer
}

// GetCompiledRewriteMatch returns the compiled rewrite match regular expression for the backend service.
func (bs *BackendService) GetCompiledRewriteMatch() *regexp.Regexp {
	return bs.compiledRewriteMatch
}

func (bs *BackendService) GetHttpClient() *http.Client {
	return bs.httpClient
}

func (bs *BackendService) setLoadBalancer() {
	switch bs.LoadBalancerPolicy.Type {
	case loadbalancer.Random:
		bs.loadBalancer = loadbalancer.NewRandomLoadBalancer()
	case loadbalancer.RoundRobin:
		bs.loadBalancer = loadbalancer.NewRoundRobinLoadBalancer()
	case loadbalancer.WeightedRoundRobin:
		bs.loadBalancer = loadbalancer.NewWRoundRobinLoadBalancer(bs.LoadBalancerPolicy.Options.Weights)
	case loadbalancer.LeastConnection:
		bs.loadBalancer = loadbalancer.NewLeastConnLoadBalancer(bs.UpstreamTargets, nil)
	case loadbalancer.WeightedLeastConnection:
		bs.loadBalancer = loadbalancer.NewLeastConnLoadBalancer(bs.UpstreamTargets, bs.LoadBalancerPolicy.Options.Weights)
	default:
		bs.loadBalancer = loadbalancer.NewRoundRobinLoadBalancer()
	}
}

// CompilePath compiles the rewrite match regular expression for the backend service and
// stores it in the compiledRewriteMatch field. If there's an error while compiling,
// the error is returned.
func (bs *BackendService) compilePath() {
	if bs.RewriteMatch == "" || bs.RewriteReplace == "" {
		return
	}

	compiled, err := regexp.Compile(bs.RewriteMatch)
	if err != nil {
		return
	}

	bs.compiledRewriteMatch = compiled
}

func (bs *BackendService) setHttpClient() {
	transport := &http.Transport{
		MaxIdleConns:        bs.MaxIdleConns,
		IdleConnTimeout:     bs.MaxIdleTime * time.Second,
		TLSHandshakeTimeout: bs.Timeout * time.Second,
	}

	bs.httpClient = &http.Client{Transport: transport}
}

func (bs *BackendService) Init() {
	bs.setTokenValidator()
	bs.setLoadBalancer()
	bs.setHttpClient()
	bs.compilePath()
}
