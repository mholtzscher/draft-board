package repository

import (
	"testing"

	"github.com/vibes/draft-board/internal/database"
	"github.com/vibes/draft-board/internal/models"
)

func TestTeamRepository_Create(t *testing.T) {
	db := database.NewTestDB(t)
	defer database.CloseTestDB(t, db)

	// Create a draft first
	draftRepo := NewDraftRepository(db)
	draft := &models.Draft{
		Name:          "Test League",
		NumTeams:      12,
		ScoringFormat: "PPR",
		DraftType:     "Redraft",
		Status:        "setup",
	}
	err := draftRepo.Create(draft)
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}

	// Create a team
	teamRepo := NewTeamRepository(db)
	team := &models.Team{
		DraftID:       draft.ID,
		TeamName:      "Test Team",
		DraftPosition: 1,
	}

	err = teamRepo.Create(team)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if team.ID == 0 {
		t.Error("Create() did not set team ID")
	}

	// Verify the team was created
	retrieved, err := teamRepo.GetByID(team.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.TeamName != team.TeamName {
		t.Errorf("TeamName = %v, want %v", retrieved.TeamName, team.TeamName)
	}
	if retrieved.DraftPosition != team.DraftPosition {
		t.Errorf("DraftPosition = %v, want %v", retrieved.DraftPosition, team.DraftPosition)
	}
	if retrieved.DraftID != team.DraftID {
		t.Errorf("DraftID = %v, want %v", retrieved.DraftID, team.DraftID)
	}
}


func TestTeamRepository_Update(t *testing.T) {
	db := database.NewTestDB(t)
	defer database.CloseTestDB(t, db)

	// Create a draft
	draftRepo := NewDraftRepository(db)
	draft := &models.Draft{
		Name:          "Test League",
		NumTeams:      12,
		ScoringFormat: "PPR",
		DraftType:     "Redraft",
		Status:        "setup",
	}
	err := draftRepo.Create(draft)
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}

	// Create a team
	teamRepo := NewTeamRepository(db)
	team := &models.Team{
		DraftID:       draft.ID,
		TeamName:      "Original Name",
		DraftPosition: 1,
	}

	err = teamRepo.Create(team)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Update the team
	team.TeamName = "Updated Name"
	team.DraftPosition = 5

	err = teamRepo.Update(team)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	retrieved, err := teamRepo.GetByID(team.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if retrieved.TeamName != "Updated Name" {
		t.Errorf("TeamName = %v, want %v", retrieved.TeamName, "Updated Name")
	}
	if retrieved.DraftPosition != 5 {
		t.Errorf("DraftPosition = %v, want %v", retrieved.DraftPosition, 5)
	}
}
