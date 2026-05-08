package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"quinela/internal/database"
	"quinela/internal/models"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetGroups(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT t.id, t.name, t.short_code, t.iso2, t.group_id, g.name, t.is_host
		FROM teams t JOIN groups g ON t.group_id = g.id ORDER BY g.name, t.name
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	groupMap := make(map[string][]models.Team)
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.ShortCode, &t.ISO2, &t.GroupID, &t.GroupName, &t.IsHost); err != nil {
			continue
		}
		t.IsHost = t.IsHost || t.GroupName == "A" && t.Name == "Canada" || t.GroupName == "B" && t.Name == "United States" || t.GroupName == "A" && t.Name == "Mexico"
		groupMap[t.GroupName] = append(groupMap[t.GroupName], t)
	}
	c.JSON(200, groupMap)
}

func GetGroupStandings(c *gin.Context) {
	groupID := c.Param("id")
	rows, err := database.DB.Query(`
		SELECT t.id, t.name, t.iso2, t.is_host
		FROM teams t WHERE t.group_id = ? ORDER BY t.name
	`, groupID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type TeamStats struct {
		ID, PJ, G, E, P, GF, GC int
		Pts                      int
	}
	stats := make(map[uint]TeamStats)
	var teamIDs []uint

	for rows.Next() {
		var id uint
		var name, iso2 string
		var isHost bool
		if err := rows.Scan(&id, &name, &iso2, &isHost); err != nil {
			continue
		}
		teamIDs = append(teamIDs, id)
		stats[id] = TeamStats{ID: id}
	}

	mrows, _ := database.DB.Query(`
		SELECT local_team_id, visitor_team_id, local_score, visitor_score
		FROM matches WHERE group_id = ? AND status = 'finished' AND local_score IS NOT NULL
	`, groupID)
	defer mrows.Close()

	for mrows.Next() {
		var localID, visitorID uint
		var ls, vs sql.NullInt64
		if err := mrows.Scan(&localID, &visitorID, &ls, &vs); err != nil {
			continue
		}
		if !ls.Valid || !vs.Valid {
			continue
		}
		localScore := int(ls.Int64)
		visitorScore := int(vs.Int64)

		if s, ok := stats[localID]; ok {
			s.PJ++
			s.GF += localScore
			s.GC += visitorScore
			if localScore > visitorScore {
				s.G++
				s.Pts += 3
			} else if localScore == visitorScore {
				s.E++
				s.Pts++
			} else {
				s.P++
			}
			stats[localID] = s
		}
		if s, ok := stats[visitorID]; ok {
			s.PJ++
			s.GF += visitorScore
			s.GC += localScore
			if visitorScore > localScore {
				s.G++
				s.Pts += 3
			} else if visitorScore == localScore {
				s.E++
				s.Pts++
			} else {
				s.P++
			}
			stats[visitorID] = s
		}
	}

	type StandingRow struct {
		TeamID   uint
		TeamName string
		ISO2     string
		IsHost   bool
		PJ, G, E, P, GF, GC, DG, Pts int
		Position int
	}

	var standings []StandingRow
	for _, id := range teamIDs {
		s := stats[id]
		rows2, _ := database.DB.Query("SELECT name, iso2, is_host FROM teams WHERE id = ?", id)
		var name, iso2 string
		var isHost bool
		if rows2.Next() {
			rows2.Scan(&name, &iso2, &isHost)
		}
		rows2.Close()
		standings = append(standings, StandingRow{
			TeamID: s.ID, TeamName: name, ISO2: iso2, IsHost: isHost,
			PJ: s.PJ, G: s.G, E: s.E, P: s.P, GF: s.GF, GC: s.GC,
			DG: s.GF - s.GC, Pts: s.Pts,
		})
	}

	sort.Slice(standings, func(i, j int) bool {
		if standings[i].Pts != standings[j].Pts {
			return standings[i].Pts > standings[j].Pts
		}
		if standings[i].DG != standings[j].DG {
			return standings[i].DG > standings[j].DG
		}
		if standings[i].GF != standings[j].GF {
			return standings[i].GF > standings[j].GF
		}
		return standings[i].TeamName < standings[j].TeamName
	})

	for i := range standings {
		standings[i].Position = i + 1
	}

	c.JSON(200, standings)
}

func GetGroupMatches(c *gin.Context) {
	groupID := c.Param("id")
	rows, err := database.DB.Query(`
		SELECT m.id, m.match_number, m.stage, m.group_id, g.name,
			m.local_team_id, t1.name, t1.iso2,
			m.visitor_team_id, t2.name, t2.iso2,
			m.match_date, m.venue, m.city,
			m.local_score, m.visitor_score, m.status
		FROM matches m
		JOIN groups g ON m.group_id = g.id
		JOIN teams t1 ON m.local_team_id = t1.id
		JOIN teams t2 ON m.visitor_team_id = t2.id
		WHERE m.group_id = ? ORDER BY m.match_number
	`, groupID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		var gid uint
		var ls, vs sql.NullInt64
		if err := rows.Scan(&m.ID, &m.MatchNumber, &m.Stage, &gid, &m.GroupName,
			&m.LocalTeamID, &m.LocalTeam, &m.LocalISO2,
			&m.VisitorTeamID, &m.VisitorTeam, &m.VisitorISO2,
			&m.MatchDate, &m.Venue, &m.City,
			&ls, &vs, &m.Status); err != nil {
			continue
		}
		m.GroupID = &gid
		if ls.Valid {
			v := int(ls.Int64)
			m.LocalScore = &v
		}
		if vs.Valid {
			v := int(vs.Int64)
			m.VisitorScore = &v
		}
		matches = append(matches, m)
	}
	c.JSON(200, matches)
}

func GetMatches(c *gin.Context) {
	stage := c.DefaultQuery("stage", "")
	query := `
		SELECT m.id, m.match_number, m.stage, m.group_id, COALESCE(g.name,''),
			m.local_team_id, COALESCE(t1.name,''), COALESCE(t1.iso2,''),
			m.visitor_team_id, COALESCE(t2.name,''), COALESCE(t2.iso2,''),
			m.match_date, m.venue, m.city,
			m.local_score, m.visitor_score, m.status
		FROM matches m
		LEFT JOIN groups g ON m.group_id = g.id
		LEFT JOIN teams t1 ON m.local_team_id = t1.id
		LEFT JOIN teams t2 ON m.visitor_team_id = t2.id
	`
	var args []any
	if stage != "" {
		query += " WHERE m.stage = ?"
		args = append(args, stage)
	}
	query += " ORDER BY m.match_number"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		var gid sql.NullInt64
		var ls, vs sql.NullInt64
		if err := rows.Scan(&m.ID, &m.MatchNumber, &m.Stage, &gid, &m.GroupName,
			&m.LocalTeamID, &m.LocalTeam, &m.LocalISO2,
			&m.VisitorTeamID, &m.VisitorTeam, &m.VisitorISO2,
			&m.MatchDate, &m.Venue, &m.City,
			&ls, &vs, &m.Status); err != nil {
			continue
		}
		if gid.Valid {
			g := uint(gid.Int64)
			m.GroupID = &g
		}
		if ls.Valid {
			v := int(ls.Int64)
			m.LocalScore = &v
		}
		if vs.Valid {
			v := int(vs.Int64)
			m.VisitorScore = &v
		}
		matches = append(matches, m)
	}
	c.JSON(200, matches)
}

func GetScheduleByTeam(c *gin.Context) {
	iso2 := c.Param("iso2")
	rows, err := database.DB.Query(`
		SELECT m.id, m.match_number, m.stage, m.group_id, COALESCE(g.name,''),
			m.local_team_id, COALESCE(t1.name,''), COALESCE(t1.iso2,''),
			m.visitor_team_id, COALESCE(t2.name,''), COALESCE(t2.iso2,''),
			m.match_date, m.venue, m.city,
			m.local_score, m.visitor_score, m.status
		FROM matches m
		LEFT JOIN groups g ON m.group_id = g.id
		LEFT JOIN teams t1 ON m.local_team_id = t1.id
		LEFT JOIN teams t2 ON m.visitor_team_id = t2.id
		WHERE t1.iso2 = ? OR t2.iso2 = ?
		ORDER BY m.match_date
	`, iso2, iso2)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		var gid sql.NullInt64
		var ls, vs sql.NullInt64
		if err := rows.Scan(&m.ID, &m.MatchNumber, &m.Stage, &gid, &m.GroupName,
			&m.LocalTeamID, &m.LocalTeam, &m.LocalISO2,
			&m.VisitorTeamID, &m.VisitorTeam, &m.VisitorISO2,
			&m.MatchDate, &m.Venue, &m.City,
			&ls, &vs, &m.Status); err != nil {
			continue
		}
		if gid.Valid {
			g := uint(gid.Int64)
			m.GroupID = &g
		}
		if ls.Valid {
			v := int(ls.Int64)
			m.LocalScore = &v
		}
		if vs.Valid {
			v := int(vs.Int64)
			m.VisitorScore = &v
		}
		matches = append(matches, m)
	}
	c.JSON(200, matches)
}

func SaveResult(c *gin.Context) {
	var req models.SaveResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec(`
		UPDATE matches SET local_score = ?, visitor_score = ?, status = 'finished'
		WHERE id = ?
	`, req.LocalScore, req.VisitorScore, req.MatchID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	rows, _ := database.DB.Query(`
		SELECT p.id, p.user_id, p.match_id, p.local_score, p.visitor_score, m.local_score, m.visitor_score
		FROM predictions p JOIN matches m ON p.match_id = m.id WHERE p.match_id = ?
	`, req.MatchID)
	defer rows.Close()

	for rows.Next() {
		var predID, userID, matchID uint
		var predLS, predVS int
		var realLS, realVS sql.NullInt64
		if err := rows.Scan(&predID, &userID, &matchID, &predLS, &predVS, &realLS, &realVS); err != nil {
			continue
		}
		if !realLS.Valid || !realVS.Valid {
			continue
		}
		points := 0
		if int(realLS.Int64) == predLS && int(realVS.Int64) == predVS {
			points = 3
		} else if (predLS > predVS && int(realLS.Int64) > int(realVS.Int64)) ||
			(predLS < predVS && int(realLS.Int64) < int(realVS.Int64)) ||
			(predLS == predVS && int(realLS.Int64) == int(realVS.Int64)) {
			points = 1
		}
		database.DB.Exec("UPDATE predictions SET points = ? WHERE id = ?", points, predID)
	}

	c.JSON(200, gin.H{"message": "Result saved"})
}

func SavePrediction(c *gin.Context) {
	var pred models.Prediction
	if err := c.ShouldBindJSON(&pred); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec(`
		INSERT OR REPLACE INTO predictions (user_id, match_id, local_score, visitor_score)
		VALUES (?, ?, ?, ?)
	`, pred.UserID, pred.MatchID, pred.LocalScore, pred.VisitorScore)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Prediction saved"})
}

func GetStandings(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT u.id, u.username,
			COALESCE(SUM(p.points), 0) as total_points,
			COUNT(CASE WHEN p.points = 3 THEN 1 END) as exact_score,
			COUNT(CASE WHEN p.points = 1 THEN 1 END) as result_only
		FROM users u
		LEFT JOIN predictions p ON u.id = p.user_id
		GROUP BY u.id ORDER BY total_points DESC
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var standings []models.Standing
	for rows.Next() {
		var s models.Standing
		if err := rows.Scan(&s.UserID, &s.Username, &s.TotalPoints, &s.ExactScore, &s.ResultOnly); err != nil {
			continue
		}
		standings = append(standings, s)
	}
	c.JSON(200, standings)
}

func Register(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, password_hash, is_admin) VALUES (?, ?, 0)",
		req.Username, string(hash),
	)
	if err != nil {
		c.JSON(400, gin.H{"error": "Username already exists"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(200, gin.H{"id": id, "username": req.Username, "is_admin": false})
}

func Login(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	var pwdHash string
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, is_admin FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &pwdHash, &user.IsAdmin)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pwdHash), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	token := fmt.Sprintf("token_%d_%s", user.ID, user.Username)
	c.JSON(200, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func GetTeams(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT t.id, t.name, t.short_code, t.iso2, t.group_id, g.name, t.is_host
		FROM teams t JOIN groups g ON t.group_id = g.id ORDER BY g.name, t.name
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.ShortCode, &t.ISO2, &t.GroupID, &t.GroupName, &t.IsHost); err != nil {
			continue
		}
		teams = append(teams, t)
	}
	c.JSON(200, teams)
}

func GetPredictions(c *gin.Context) {
	userID := c.Param("userId")
	rows, err := database.DB.Query(`
		SELECT p.id, p.user_id, p.match_id, p.local_score, p.visitor_score, p.points
		FROM predictions p WHERE p.user_id = ?
	`, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var preds []models.Prediction
	for rows.Next() {
		var p models.Prediction
		if err := rows.Scan(&p.ID, &p.UserID, &p.MatchID, &p.LocalScore, &p.VisitorScore, &p.Points); err != nil {
			continue
		}
		preds = append(preds, p)
	}
	c.JSON(200, preds)
}

func GetResult(c *gin.Context) {
	matchID := c.Param("matchId")
	var m models.Match
	var ls, vs sql.NullInt64
	err := database.DB.QueryRow(`
		SELECT id, local_score, visitor_score FROM matches WHERE id = ?
	`, matchID).Scan(&m.ID, &ls, &vs)
	if err != nil {
		c.JSON(404, gin.H{"error": "Match not found"})
		return
	}
	if ls.Valid {
		v := int(ls.Int64)
		m.LocalScore = &v
	}
	if vs.Valid {
		v := int(vs.Int64)
		m.VisitorScore = &v
	}
	c.JSON(200, m)
}

func parseUint(s string) uint {
	if v, err := strconv.ParseUint(s, 10, 64); err == nil {
		return uint(v)
	}
	return 0
}