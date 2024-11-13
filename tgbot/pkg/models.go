// models.go
package pkg

type Player struct {
	Name       string
	TelegramID int64
	Password   string
	Team       *Team
	IsOwner    bool
}

func NewPlayer(name string, telegramID int64, password string) *Player {
	return &Player{Name: name, TelegramID: telegramID, Password: password}
}

type Team struct {
	Name    string
	Owner   *Player
	Captain *Player
	Players []*Player
}

func NewTeam(name string, owner *Player) *Team {
	team := &Team{Name: name, Owner: owner, Captain: owner}
	team.Players = append(team.Players, owner)
	return team
}

func (t *Team) AddPlayer(player *Player) {
	t.Players = append(t.Players, player)
	player.Team = t
}

func (t *Team) SetCaptain(player *Player) {
	t.Captain = player
}

type Match struct {
	TeamA     *Team
	TeamB     *Team
	MatchTime string
}

func NewMatch(teamA, teamB *Team, matchTime string) *Match {
	return &Match{TeamA: teamA, TeamB: teamB, MatchTime: matchTime}
}
