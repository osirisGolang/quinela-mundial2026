package models

type Team struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ShortCode string `json:"short_code"`
	ISO2      string `json:"iso2"`
	GroupID   uint   `json:"group_id"`
	GroupName string `json:"group_name"`
	IsHost    bool   `json:"is_host"`
}

type Group struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Match struct {
	ID            uint   `json:"id"`
	MatchNumber   int    `json:"match_number"`
	Stage         string `json:"stage"`
	GroupID       *uint  `json:"group_id"`
	GroupName     string `json:"group_name,omitempty"`
	LocalTeamID   uint   `json:"local_team_id"`
	LocalTeam     string `json:"local_team"`
	LocalISO2     string `json:"local_iso2"`
	VisitorTeamID uint   `json:"visitor_team_id"`
	VisitorTeam   string `json:"visitor_team"`
	VisitorISO2   string `json:"visitor_iso2"`
	MatchDate     string `json:"match_date"`
	Venue         string `json:"venue"`
	City          string `json:"city"`
	LocalScore    *int   `json:"local_score"`
	VisitorScore  *int   `json:"visitor_score"`
	Status        string `json:"status"`
}

type Prediction struct {
	ID           uint `json:"id"`
	UserID       uint `json:"user_id"`
	MatchID      uint `json:"match_id"`
	LocalScore   int  `json:"local_score"`
	VisitorScore int  `json:"visitor_score"`
	Points       int  `json:"points"`
}

type User struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	IsAdmin      bool   `json:"is_admin"`
}

type Standing struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	TotalPoints int    `json:"total_points"`
	ExactScore  int    `json:"exact_score"`
	ResultOnly  int    `json:"result_only"`
}

type GroupStanding struct {
	TeamID     uint   `json:"team_id"`
	TeamName   string `json:"team_name"`
	ISO2       string `json:"iso2"`
	IsHost     bool   `json:"is_host"`
	PJ         int    `json:"pj"`
	G          int    `json:"g"`
	E          int    `json:"e"`
	P          int    `json:"p"`
	GF         int    `json:"gf"`
	GC         int    `json:"gc"`
	DG         int    `json:"dg"`
	Pts        int    `json:"pts"`
	Position   int    `json:"position"`
}

type TeamStandings struct {
	TeamID     uint   `json:"team_id"`
	TeamName   string `json:"team_name"`
	ISO2       string `json:"iso2"`
	GroupName  string `json:"group_name"`
	IsHost     bool   `json:"is_host"`
	PJ         int    `json:"pj"`
	G          int    `json:"g"`
	E          int    `json:"e"`
	P          int    `json:"p"`
	GF         int    `json:"gf"`
	GC         int    `json:"gc"`
	DG         int    `json:"dg"`
	Pts        int    `json:"pts"`
	Position   int    `json:"position"`
}

type SaveResultRequest struct {
	MatchID      uint `json:"match_id"`
	LocalScore   int  `json:"local_score"`
	VisitorScore int  `json:"visitor_score"`
}

type UpdateAllResultsRequest struct {
	Results []SaveResultRequest `json:"results"`
}

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}