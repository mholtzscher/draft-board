package main

import (
	"log"
	"net/http"
	"os"

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
	defer db.Close()

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

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

