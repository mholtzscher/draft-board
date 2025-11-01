package models

import "time"

type Player struct {
	ID            int       `db:"id"`
	Name          string    `db:"name"`
	Team          string    `db:"team"`
	Position      string    `db:"position"`
	ByeWeek       *int      `db:"bye_week"`
	DynastyRank   *int      `db:"dynasty_rank"`
	SFRank        *int      `db:"sf_rank"`
	StdRank       *int      `db:"std_rank"`
	HalfPPRRank   *int      `db:"half_ppr_rank"`
	PPRRank       *int      `db:"ppr_rank"`
	IsCustom      bool      `db:"is_custom"`
	CreatedAt     time.Time `db:"created_at"`
}

func (p *Player) GetADPRank(draftType, scoringFormat string) *int {
	if draftType == "Dynasty" {
		return p.DynastyRank
	}
	
	switch scoringFormat {
	case "PPR":
		return p.PPRRank
	case "Half-PPR":
		return p.HalfPPRRank
	default:
		return p.StdRank
	}
}

