package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	graphqlhandler "github.com/venue-master/platform/services/api-gateway/internal/graphql"
	"github.com/venue-master/platform/services/api-gateway/internal/services"

	"github.com/venue-master/platform/internal/server"
	"github.com/venue-master/platform/lib/jwtutil"
)

func main() {
	srv, err := server.New("api-gateway")
	if err != nil {
		log.Fatalf("failed to init server: %v", err)
	}

	jwtManager := jwtutil.NewManager(srv.Config.JWT)
	clients := selectServiceClients()

	handler, err := graphqlhandler.New(clients, jwtManager, srv.Logger)
	if err != nil {
		log.Fatalf("failed to init graphql handler: %v", err)
	}

	handler.Register(srv.Engine)

	if err := srv.Run(); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func selectServiceClients() *services.ServiceClients {
	userURL := os.Getenv("USER_SERVICE_URL")
	bookingURL := os.Getenv("BOOKING_SERVICE_URL")
	if strings.EqualFold(os.Getenv("USE_MOCK_SERVICES"), "true") || userURL == "" || bookingURL == "" {
		return services.NewMockClients()
	}
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	return services.NewHTTPClients(httpClient, userURL, bookingURL)
}
