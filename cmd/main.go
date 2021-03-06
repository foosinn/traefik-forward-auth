package main

import (
	"net/http"

	"github.com/gorilla/sessions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	internal "github.com/mesosphere/traefik-forward-auth/internal"
	logger "github.com/mesosphere/traefik-forward-auth/internal/log"
)

// Main
func main() {
	// Parse options
	config := internal.NewGlobalConfig()

	// Setup logger
	log := logger.NewDefaultLogger(config.LogLevel, config.LogFormat)

	// Perform config validation
	config.Validate()

	// Query the OIDC provider
	config.SetOidcProvider()

	// Get clientset for Authorizers
	var clientset kubernetes.Interface
	if config.EnableRBAC {
		icc, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("error getting in cluster configuration for RBAC client: %v", err)
		}
		clientset, err = kubernetes.NewForConfig(icc)
		if err != nil {
			log.Fatalf("error getting kubernetes client: %v", err)
		}
	} else {
		clientset = nil
	}

	// Build server
	server := internal.NewServer(sessions.NewCookieStore([]byte(config.SessionKey)), clientset)

	// Attach router to default server
	http.HandleFunc("/", server.RootHandler)

	// Start
	log.Debugf("Starting with options: %s", config)
	log.Info("Listening on :4181")
	log.Info(http.ListenAndServe(":4181", nil))
}
