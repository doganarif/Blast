package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/doganarif/blast/internal/ca"
	"github.com/doganarif/blast/internal/cert"
)

// Server represents the proxy server
type Server struct {
	rootCA *ca.CA
	routes map[string]string // domain -> localhost:port
	certs  map[string]tls.Certificate
	mu     sync.RWMutex
	server *http.Server
}

// NewServer creates a new proxy server
func NewServer(rootCA *ca.CA) *Server {
	return &Server{
		rootCA: rootCA,
		routes: make(map[string]string),
		certs:  make(map[string]tls.Certificate),
	}
}

// AddRoute adds a new route mapping
func (s *Server) AddRoute(domain, localPort string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate certificate for the domain
	tlsCert, err := cert.GenerateCertificate(s.rootCA, domain)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	s.routes[domain] = "localhost:" + localPort
	s.certs[domain] = tlsCert

	return nil
}

// RemoveRoute removes a route mapping
func (s *Server) RemoveRoute(domain string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.routes, domain)
	delete(s.certs, domain)
}

// ClearRoutes clears all route mappings
func (s *Server) ClearRoutes() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.routes = make(map[string]string)
	s.certs = make(map[string]tls.Certificate)
}

// Start starts the proxy server on port 443
func (s *Server) Start() error {
	// Create TLS config with dynamic certificate selection
	tlsConfig := &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			s.mu.RLock()
			defer s.mu.RUnlock()

			cert, ok := s.certs[hello.ServerName]
			if !ok {
				return nil, fmt.Errorf("no certificate for %s", hello.ServerName)
			}
			return &cert, nil
		},
	}

	// Create HTTP server
	handler := http.HandlerFunc(s.handleRequest)
	s.server = &http.Server{
		Addr:      ":443",
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	return s.server.ListenAndServeTLS("", "")
}

// handleRequest handles incoming HTTP requests
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	target, ok := s.routes[r.Host]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "No route configured for "+r.Host, http.StatusNotFound)
		return
	}

	// Parse target URL
	targetURL, err := url.Parse("http://" + target)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify request
	r.URL.Host = targetURL.Host
	r.URL.Scheme = targetURL.Scheme
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("X-Forwarded-Proto", "https")

	// Serve the request
	proxy.ServeHTTP(w, r)
}

// Stop stops the proxy server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}
