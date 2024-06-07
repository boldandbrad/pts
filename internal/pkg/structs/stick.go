package structs

type Stick struct {
	Name                  string
	GamesPlayed           int
	PlateAppearances      int
	Singles               int
	Doubles               int
	Triples               int
	HomeRuns              int
	RunsScored            int
	RunsBattedIn          int
	StolenBases           int
	Walks                 int
	HitByPitches          int
	SacBunts              int
	StrikeOuts            int
	GroundIntoDoublePlays int
	Errors                int
}

func (s Stick) PositivePoints() int {
	return s.Singles + (2 * s.Doubles) + (3 * s.Triples) + (4 * s.HomeRuns) + s.RunsScored + s.RunsBattedIn + s.StolenBases + s.Walks + s.HitByPitches + s.SacBunts
}

func (s Stick) NegativePoints() int {
	return s.StrikeOuts + (2 * s.GroundIntoDoublePlays) + s.Errors
}
