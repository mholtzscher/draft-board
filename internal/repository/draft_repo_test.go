package repository

import (
	"testing"

	"github.com/vibes/draft-board/internal/database"
	"github.com/vibes/draft-board/internal/models"
)

func TestDraftRepository_Create(t *testing.T) {
	db := database.NewTestDB(t)
	defer database.CloseTestDB(t, db)

	repo := NewDraftRepository(db)

	draft := &models.Draft{
		Name:          "Test League",
		NumTeams:      12,
		ScoringFormat: "PPR",
		DraftType:     "Redraft",
		QBSetting:     "1QB",
		SnakeDraft:    true,
		Status:        "setup",
		MaxRounds:     15,
		CommissionerID: "user123",
	}

	err := repo.Create(draft)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if draft.ID == 0 {
		t.Error("Create() did not set draft ID")
	}

	// Verify the draft was created
	retrieved, err := repo.GetByID(draft.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.Name != draft.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, draft.Name)
	}
	if retrieved.NumTeams != draft.NumTeams {
		t.Errorf("NumTeams = %v, want %v", retrieved.NumTeams, draft.NumTeams)
	}
	if retrieved.ScoringFormat != draft.ScoringFormat {
		t.Errorf("ScoringFormat = %v, want %v", retrieved.ScoringFormat, draft.ScoringFormat)
	}
	if retrieved.DraftType != draft.DraftType {
		t.Errorf("DraftType = %v, want %v", retrieved.DraftType, draft.DraftType)
	}
	if retrieved.SnakeDraft != draft.SnakeDraft {
		t.Errorf("SnakeDraft = %v, want %v", retrieved.SnakeDraft, draft.SnakeDraft)
	}
	if retrieved.Status != draft.Status {
		t.Errorf("Status = %v, want %v", retrieved.Status, draft.Status)
	}
	if retrieved.MaxRounds != draft.MaxRounds {
		t.Errorf("MaxRounds = %v, want %v", retrieved.MaxRounds, draft.MaxRounds)
	}
}

func TestDraftRepository_GetByID(t *testing.T) {
	db := database.NewTestDB(t)
	defer database.CloseTestDB(t, db)

	repo := NewDraftRepository(db)

	// Create a draft
	draft := &models.Draft{
		Name:          "Test League",
		NumTeams:      10,
		ScoringFormat: "Standard",
		DraftType:     "Dynasty",
		SnakeDraft:    false,
		Status:        "active",
		MaxRounds:     16,
	}

	err := repo.Create(draft)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Test successful retrieval
	retrieved, err := repo.GetByID(draft.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.ID != draft.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, draft.ID)
	}

	// Test non-existent draft
	_, err = repo.GetByID(99999)
	if err == nil {
		t.Error("GetByID() expected error for non-existent draft, got nil")
	}
}

func TestDraftRepository_Update(t *testing.T) {
	db := database.NewTestDB(t)
	defer database.CloseTestDB(t, db)

	repo := NewDraftRepository(db)

	// Create a draft
	draft := &models.Draft{
		Name:          "Original Name",
		NumTeams:      12,
		ScoringFormat: "PPR",
		DraftType:     "Redraft",
		SnakeDraft:    true,
		Status:        "setup",
		MaxRounds:     15,
	}

	err := repo.Create(draft)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the draft
	draft.Name = "Updated Name"
	draft.Status = "active"

	err = repo.Update(draft)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	retrieved, err := repo.GetByID(draft.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Name = %v, want %v", retrieved.Name, "Updated Name")
	}
	if retrieved.Status != "active" {
		t.Errorf("Status = %v, want %v", retrieved.Status, "active")
	}
}


func TestDraftRepository_Delete(t *testing.T) {
	db := database.NewTestDB(t)
	defer database.CloseTestDB(t, db)

	repo := NewDraftRepository(db)

	// Create a draft
	draft := &models.Draft{
		Name:          "Test League",
		NumTeams:      12,
		ScoringFormat: "PPR",
		DraftType:     "Redraft",
		Status:        "setup",
	}

	err := repo.Create(draft)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Delete the draft
	err = repo.Delete(draft.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's deleted
	_, err = repo.GetByID(draft.ID)
	if err == nil {
		t.Error("GetByID() expected error for deleted draft, got nil")
	}
}
