package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vibes/draft-board/internal/database"
	"github.com/vibes/draft-board/internal/handlers"
	"github.com/vibes/draft-board/internal/repository"
)

func main() {
	dbPath := database.GetDBPath()
	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize repositories
	draftRepo := repository.NewDraftRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	playerRepo := repository.NewPlayerRepository(db)
	pickRepo := repository.NewPickRepository(db)
	queueRepo := repository.NewQueueRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Initialize handlers
	h := handlers.NewHandler(draftRepo, teamRepo, playerRepo, pickRepo, queueRepo, auditRepo)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Static files
	workDir, _ := os.Getwd()
	filesDir := http.Dir(workDir + "/web/static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(filesDir)))

	// Routes
	r.Get("/", h.Home)
	r.Get("/draft/new", h.NewDraft)
	r.Post("/draft/create", h.CreateDraft)
	r.Get("/draft/{id}/setup", h.DraftSetup)
	r.Post("/draft/{id}/update", h.UpdateDraft)
	r.Post("/draft/{id}/start", h.StartDraft)
	r.Post("/draft/{id}/pause", h.PauseDraft)
	r.Post("/draft/{id}/resume", h.ResumeDraft)
	r.Post("/draft/{id}/complete", h.CompleteDraft)
	r.Delete("/draft/{id}", h.DeleteDraft)
	r.Get("/draft/{id}", h.GetDraftBoard)
	r.Get("/draft/{id}/big-board", h.GetBigBoard)
	r.Get("/draft/{id}/players", h.GetAvailablePlayers)
	r.Post("/draft/{id}/pick", h.MakePick)
	r.Post("/draft/{id}/undo", h.UndoPick)
	r.Post("/draft/{id}/trade", h.TradePick)
	r.Get("/draft/{id}/current", h.GetCurrentPick)
	r.Get("/draft/{id}/teams", h.GetTeams)
	r.Post("/draft/{id}/teams", h.CreateTeam)
	r.Put("/teams/{id}", h.UpdateTeam)
	r.Delete("/teams/{id}", h.DeleteTeam)
	r.Get("/draft/{id}/queue", h.GetQueue)
	r.Post("/draft/{id}/queue", h.AddToQueue)
	r.Delete("/draft/{id}/queue/{queueId}", h.RemoveFromQueue)
	r.Post("/players/custom", h.CreateCustomPlayer)

	// Stats routes
	r.Get("/draft/{id}/stats/franchise", h.GetFranchiseStats)
	r.Get("/draft/{id}/stats/position", h.GetDraftedByPosition)
	r.Get("/draft/{id}/stats/value-picks", h.GetValuePicks)

	// Export routes
	r.Get("/draft/{id}/export/csv", h.ExportCSV)
	r.Get("/draft/{id}/export/json", h.ExportJSON)

	// SSE route
	r.Get("/draft/{id}/stream", h.StreamUpdates)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Channel to track server errors
	serverErr := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		} else {
			serverErr <- nil
		}
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("Received signal: %v", sig)
	case err := <-serverErr:
		if err != nil {
			log.Printf("Server error: %v", err)
			os.Exit(1)
		}
		return
	}

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	shutdownErr := srv.Shutdown(ctx)
	if shutdownErr != nil {
		if errors.Is(shutdownErr, context.DeadlineExceeded) {
			log.Println("Shutdown timeout exceeded, forcing server close...")
			// Force close the server if graceful shutdown timed out
			if err := srv.Close(); err != nil && err != http.ErrServerClosed {
				log.Printf("Error force closing server: %v", err)
			}
		} else {
			log.Printf("Server shutdown error: %v", shutdownErr)
		}
	} else {
		log.Println("Server exited gracefully")
	}

	// Database will be closed by defer db.Close() above
	log.Println("Shutdown complete")
	// main() returns normally, which exits with code 0
}
