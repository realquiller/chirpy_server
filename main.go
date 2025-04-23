package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/realquiller/chirpy_server/internal/database"
	"github.com/realquiller/chirpy_server/internal/handlers"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := handlers.ApiConfig{}

	apiCfg.Platform = os.Getenv("PLATFORM")

	apiCfg.DbQueries = dbQueries

	apiCfg.Secret = os.Getenv("SECRET")
	apiCfg.PolkaKey = os.Getenv("POLKA_KEY")

	// Health check
	mux.HandleFunc("GET /api/healthz", handlers.ReadinessHandler)

	// App handler
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(
		http.StripPrefix("/app/", http.FileServer(http.Dir("./app/"))),
	))

	// Metrics handler
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)

	// Reset handler
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetHandler)

	// NewUser handler
	mux.HandleFunc("POST /api/users", apiCfg.NewUserHandler)

	// Chirp handler
	mux.HandleFunc("POST /api/chirps", apiCfg.ChirpHandler)

	// GetChirps handler
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirpsHandler)

	// GetChirp handler
	mux.HandleFunc("GET /api/chirps/{chirpid}", apiCfg.GetChirpHandler)

	// Login handler
	mux.HandleFunc("POST /api/login", apiCfg.LoginHandler)

	// Refresh handler
	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshHandler)

	// Revoke handler
	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeHandler)

	// UpdateUser handler
	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUserHandler)

	// DeleteChirp handler
	mux.HandleFunc("DELETE /api/chirps/{chirpid}", apiCfg.DeleteChirpHandler)

	// WebhookUpgradeUser handler
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.WebhookUpgradeUserHandler)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(server.ListenAndServe())
}

// go build -o out && ./out

// postgres: sudo -u postgres psql
// connection string: psql "postgres://postgres:postgres@localhost:5432/chirpy"
// goose: goose postgres postgres://postgres:postgres@localhost:5432/chirpy up
