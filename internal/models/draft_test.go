package models

import (
	"testing"
)

func TestDraft_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "active status",
			status: "active",
			want:   true,
		},
		{
			name:   "paused status",
			status: "paused",
			want:   false,
		},
		{
			name:   "completed status",
			status: "completed",
			want:   false,
		},
		{
			name:   "draft status",
			status: "draft",
			want:   false,
		},
		{
			name:   "empty status",
			status: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Draft{Status: tt.status}
			if got := d.IsActive(); got != tt.want {
				t.Errorf("Draft.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDraft_IsCompleted(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		completed bool
		want      bool
	}{
		{
			name:      "status completed",
			status:    "completed",
			completed: false,
			want:      true,
		},
		{
			name:      "completed flag true",
			status:    "active",
			completed: true,
			want:      true,
		},
		{
			name:      "both completed",
			status:    "completed",
			completed: true,
			want:      true,
		},
		{
			name:      "neither completed",
			status:    "active",
			completed: false,
			want:      false,
		},
		{
			name:      "paused not completed",
			status:    "paused",
			completed: false,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Draft{
				Status:    tt.status,
				Completed: tt.completed,
			}
			if got := d.IsCompleted(); got != tt.want {
				t.Errorf("Draft.IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDraft_IsPaused(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "paused status",
			status: "paused",
			want:   true,
		},
		{
			name:   "active status",
			status: "active",
			want:   false,
		},
		{
			name:   "completed status",
			status: "completed",
			want:   false,
		},
		{
			name:   "draft status",
			status: "draft",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Draft{Status: tt.status}
			if got := d.IsPaused(); got != tt.want {
				t.Errorf("Draft.IsPaused() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDraft_CanMakePicks(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "active - can make picks",
			status: "active",
			want:   true,
		},
		{
			name:   "paused - cannot make picks",
			status: "paused",
			want:   false,
		},
		{
			name:   "completed - cannot make picks",
			status: "completed",
			want:   false,
		},
		{
			name:   "draft - cannot make picks",
			status: "draft",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Draft{Status: tt.status}
			if got := d.CanMakePicks(); got != tt.want {
				t.Errorf("Draft.CanMakePicks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDraft_CheckDraftCompletion(t *testing.T) {
	tests := []struct {
		name      string
		draft     *Draft
		pickCount int
		want      bool
	}{
		{
			name: "not complete - picks remaining",
			draft: &Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "active",
			},
			pickCount: 100,
			want:      false,
		},
		{
			name: "complete - exact pick count",
			draft: &Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "active",
			},
			pickCount: 180,
			want:      true,
		},
		{
			name: "complete - exceeded pick count",
			draft: &Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "active",
			},
			pickCount: 200,
			want:      true,
		},
		{
			name: "not complete - status completed but max rounds not met",
			draft: &Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "completed",
			},
			pickCount: 50,
			want:      false, // MaxRounds check happens first
		},
		{
			name: "not complete - no max rounds set",
			draft: &Draft{
				NumTeams:  10,
				MaxRounds: 0,
				Status:    "active",
			},
			pickCount: 100,
			want:      false,
		},
		{
			name: "complete - status completed, no max rounds",
			draft: &Draft{
				NumTeams:  10,
				MaxRounds: 0,
				Status:    "completed",
			},
			pickCount: 50,
			want:      true,
		},
		{
			name: "boundary - one pick away from completion",
			draft: &Draft{
				NumTeams:  10,
				MaxRounds: 16,
				Status:    "active",
			},
			pickCount: 159,
			want:      false,
		},
		{
			name: "boundary - exactly at completion",
			draft: &Draft{
				NumTeams:  10,
				MaxRounds: 16,
				Status:    "active",
			},
			pickCount: 160,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.draft.CheckDraftCompletion(tt.pickCount)
			if got != tt.want {
				t.Errorf("Draft.CheckDraftCompletion() = %v, want %v", got, tt.want)
			}
		})
	}
}
