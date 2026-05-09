package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"quinela/internal/database"
	"quinela/internal/models"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	excelize "github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

//go:embed all:frontend/dist
var assets embed.FS

func enableCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func respondJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func GetGroups(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
		SELECT t.id, t.name, t.short_code, t.iso2, t.group_id, g.name, t.is_host
		FROM teams t JOIN groups g ON t.group_id = g.id ORDER BY g.name, t.name
	`)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	groupMap := make(map[string][]models.Team)
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.ShortCode, &t.ISO2, &t.GroupID, &t.GroupName, &t.IsHost); err != nil {
			continue
		}
		groupMap[t.GroupName] = append(groupMap[t.GroupName], t)
	}
	respondJSON(w, 200, groupMap)
}

func GetGroupStandings(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	groupID := parts[len(parts)-2]

	rows, err := database.DB.Query(`
		SELECT t.id, t.name, t.iso2, t.is_host
		FROM teams t WHERE t.group_id = ? ORDER BY t.name
	`, groupID)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	type TeamStats struct {
		ID, PJ, G, E, P, GF, GC int
		Pts                     int
	}
	stats := make(map[int]TeamStats)
	var teamIDs []int

	for rows.Next() {
		var id int
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
		var localID, visitorID int
		var ls, vs int
		if err := mrows.Scan(&localID, &visitorID, &ls, &vs); err != nil {
			continue
		}
		if s, ok := stats[localID]; ok {
			s.PJ++
			s.GF += ls
			s.GC += vs
			if ls > vs {
				s.G++
				s.Pts += 3
			} else if ls == vs {
				s.E++
				s.Pts++
			} else {
				s.P++
			}
			stats[localID] = s
		}
		if s, ok := stats[visitorID]; ok {
			s.PJ++
			s.GF += vs
			s.GC += ls
			if vs > ls {
				s.G++
				s.Pts += 3
			} else if vs == ls {
				s.E++
				s.Pts++
			} else {
				s.P++
			}
			stats[visitorID] = s
		}
	}

	type StandingRow struct {
		TeamID, PJ, G, E, P, GF, GC, DG, Pts, Position int
		TeamName, ISO2                                 string
		IsHost                                         bool
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
			PJ: s.PJ, G: s.G, E: s.E, P: s.P,
			GF: s.GF, GC: s.GC, DG: s.GF - s.GC, Pts: s.Pts,
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

	respondJSON(w, 200, standings)
}

func GetGroupMatches(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	groupID := parts[len(parts)-2]

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
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		var gid int
		var ls, vs *int
		if err := rows.Scan(&m.ID, &m.MatchNumber, &m.Stage, &gid, &m.GroupName,
			&m.LocalTeamID, &m.LocalTeam, &m.LocalISO2,
			&m.VisitorTeamID, &m.VisitorTeam, &m.VisitorISO2,
			&m.MatchDate, &m.Venue, &m.City,
			&ls, &vs, &m.Status); err != nil {
			continue
		}
		gidU := uint(gid)
		m.GroupID = &gidU
		m.LocalScore = ls
		m.VisitorScore = vs
		matches = append(matches, m)
	}
	respondJSON(w, 200, matches)
}

func GetMatches(w http.ResponseWriter, r *http.Request) {
	stage := r.URL.Query().Get("stage")
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
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		var gid *int
		var ls, vs *int
		if err := rows.Scan(&m.ID, &m.MatchNumber, &m.Stage, &gid, &m.GroupName,
			&m.LocalTeamID, &m.LocalTeam, &m.LocalISO2,
			&m.VisitorTeamID, &m.VisitorTeam, &m.VisitorISO2,
			&m.MatchDate, &m.Venue, &m.City,
			&ls, &vs, &m.Status); err != nil {
			continue
		}
		if gid != nil {
			g := uint(*gid)
			m.GroupID = &g
		}
		// Asignar goles al struct (antes se escaneaban en vars locales y se perdían)
		m.LocalScore = ls
		m.VisitorScore = vs
		matches = append(matches, m)
	}
	respondJSON(w, 200, matches)
}

func GetScheduleByTeam(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	iso2 := parts[len(parts)-1]

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
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		var gid *int
		var ls, vs *int
		if err := rows.Scan(&m.ID, &m.MatchNumber, &m.Stage, &gid, &m.GroupName,
			&m.LocalTeamID, &m.LocalTeam, &m.LocalISO2,
			&m.VisitorTeamID, &m.VisitorTeam, &m.VisitorISO2,
			&m.MatchDate, &m.Venue, &m.City,
			&ls, &vs, &m.Status); err != nil {
			continue
		}
		if gid != nil {
			g := uint(*gid)
			m.GroupID = &g
		}
		m.LocalScore = ls
		m.VisitorScore = vs
		matches = append(matches, m)
	}
	respondJSON(w, 200, matches)
}

func SaveResult(w http.ResponseWriter, r *http.Request) {
	var req models.SaveResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec(`
		UPDATE matches SET local_score = ?, visitor_score = ?, status = 'finished'
		WHERE id = ?
	`, req.LocalScore, req.VisitorScore, req.MatchID)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	rows, _ := database.DB.Query(`
		SELECT p.id, p.user_id, p.match_id, p.local_score, p.visitor_score, m.local_score, m.visitor_score
		FROM predictions p JOIN matches m ON p.match_id = m.id WHERE p.match_id = ?
	`, req.MatchID)
	defer rows.Close()

	for rows.Next() {
		var predID, userID, matchID uint
		var predLS, predVS, realLS, realVS int
		if err := rows.Scan(&predID, &userID, &matchID, &predLS, &predVS, &realLS, &realVS); err != nil {
			continue
		}
		points := 0
		if realLS == predLS && realVS == predVS {
			points = 3
		} else if (predLS > predVS && realLS > realVS) ||
			(predLS < predVS && realLS < realVS) ||
			(predLS == predVS && realLS == realVS) {
			points = 1
		}
		database.DB.Exec("UPDATE predictions SET points = ? WHERE id = ?", points, predID)
	}

	respondJSON(w, 200, map[string]string{"message": "Result saved"})
}

func SavePrediction(w http.ResponseWriter, r *http.Request) {
	var pred models.Prediction
	if err := json.NewDecoder(r.Body).Decode(&pred); err != nil {
		respondJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	// Check if user has locked their predictions
	var locked int
	database.DB.QueryRow("SELECT locked FROM prediction_locks WHERE user_id = ?", pred.UserID).Scan(&locked)
	if locked == 1 {
		respondJSON(w, 403, map[string]string{"error": "Los pronósticos están cerrados y no pueden modificarse"})
		return
	}

	_, err := database.DB.Exec(`
		INSERT OR REPLACE INTO predictions (user_id, match_id, local_score, visitor_score)
		VALUES (?, ?, ?, ?)
	`, pred.UserID, pred.MatchID, pred.LocalScore, pred.VisitorScore)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	// Si el partido ya tiene resultado, recalcular los puntos inmediatamente
	var realLS, realVS int
	var status string
	dbErr := database.DB.QueryRow("SELECT COALESCE(local_score,0), COALESCE(visitor_score,0), status FROM matches WHERE id = ?", pred.MatchID).Scan(&realLS, &realVS, &status)
	if dbErr == nil && status == "finished" {
		points := 0
		if realLS == pred.LocalScore && realVS == pred.VisitorScore {
			points = 3
		} else if (pred.LocalScore > pred.VisitorScore && realLS > realVS) ||
			(pred.LocalScore < pred.VisitorScore && realLS < realVS) ||
			(pred.LocalScore == pred.VisitorScore && realLS == realVS) {
			points = 1
		}
		database.DB.Exec("UPDATE predictions SET points = ? WHERE user_id = ? AND match_id = ?", points, pred.UserID, pred.MatchID)
	}

	respondJSON(w, 200, map[string]string{"message": "Prediction saved"})
}

func GetAllGroupMatchesWithPredictions(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	userID := parts[len(parts)-1]

	rows, err := database.DB.Query(`
		SELECT m.id, m.match_number, m.group_id, g.name,
			m.local_team_id, t1.name, t1.iso2,
			m.visitor_team_id, t2.name, t2.iso2,
			m.match_date, m.venue, m.city,
			m.local_score, m.visitor_score, m.status,
			p.id, p.local_score, p.visitor_score, p.points
		FROM matches m
		JOIN groups g ON m.group_id = g.id
		JOIN teams t1 ON m.local_team_id = t1.id
		JOIN teams t2 ON m.visitor_team_id = t2.id
		LEFT JOIN predictions p ON p.match_id = m.id AND p.user_id = ?
		WHERE m.stage = 'group'
		ORDER BY m.group_id, m.match_number
	`, userID)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	type MatchWithPrediction struct {
		MatchID          uint   `json:"match_id"`
		MatchNumber      int    `json:"match_number"`
		GroupID          int    `json:"group_id"`
		GroupName        string `json:"group_name"`
		LocalTeamID      uint   `json:"local_team_id"`
		LocalTeam        string `json:"local_team"`
		LocalISO2        string `json:"local_iso2"`
		VisitorTeamID    uint   `json:"visitor_team_id"`
		VisitorTeam      string `json:"visitor_team"`
		VisitorISO2      string `json:"visitor_iso2"`
		MatchDate        string `json:"match_date"`
		Venue            string `json:"venue"`
		City             string `json:"city"`
		RealLocalScore   *int   `json:"real_local_score"`
		RealVisitorScore *int   `json:"real_visitor_score"`
		MatchStatus      string `json:"match_status"`
		PredID           *int   `json:"pred_id"`
		PredLocalScore   *int   `json:"pred_local_score"`
		PredVisitorScore *int   `json:"pred_visitor_score"`
		Points           *int   `json:"points"`
	}

	var result []MatchWithPrediction
	for rows.Next() {
		var m MatchWithPrediction
		var rls, rvs, predID, pls, pvs, pts *int
		if err := rows.Scan(
			&m.MatchID, &m.MatchNumber, &m.GroupID, &m.GroupName,
			&m.LocalTeamID, &m.LocalTeam, &m.LocalISO2,
			&m.VisitorTeamID, &m.VisitorTeam, &m.VisitorISO2,
			&m.MatchDate, &m.Venue, &m.City,
			&rls, &rvs, &m.MatchStatus,
			&predID, &pls, &pvs, &pts,
		); err != nil {
			continue
		}
		m.RealLocalScore = rls
		m.RealVisitorScore = rvs
		m.PredID = predID
		m.PredLocalScore = pls
		m.PredVisitorScore = pvs
		m.Points = pts
		result = append(result, m)
	}
	respondJSON(w, 200, result)
}

func GetPredictionLockStatus(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	userID := parts[len(parts)-1]

	var locked int
	var lockedAt *string
	err := database.DB.QueryRow("SELECT locked, locked_at FROM prediction_locks WHERE user_id = ?", userID).Scan(&locked, &lockedAt)
	if err != nil {
		// No record means not locked
		respondJSON(w, 200, map[string]any{"locked": false, "locked_at": nil})
		return
	}
	respondJSON(w, 200, map[string]any{"locked": locked == 1, "locked_at": lockedAt})
}

func LockUserPredictions(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uint `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec(`
		INSERT INTO prediction_locks (user_id, locked, locked_at)
		VALUES (?, 1, ?)
		ON CONFLICT(user_id) DO UPDATE SET locked=1, locked_at=?
	`, req.UserID, now, now)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, 200, map[string]string{"message": "Predictions locked successfully"})
}

func GetStandings(w http.ResponseWriter, r *http.Request) {
	// Calcula puntos dinámicamente comparando pronósticos con resultados reales.
	// Solo cuenta partidos ya terminados (status='finished') para el cálculo.
	rows, err := database.DB.Query(`
		SELECT
			u.id,
			u.username,
			COALESCE(SUM(
				CASE
					WHEN m.status = 'finished' AND p.local_score = m.local_score AND p.visitor_score = m.visitor_score THEN 3
					WHEN m.status = 'finished' AND (
						(p.local_score > p.visitor_score AND m.local_score > m.visitor_score) OR
						(p.local_score < p.visitor_score AND m.local_score < m.visitor_score) OR
						(p.local_score = p.visitor_score AND m.local_score = m.visitor_score)
					) THEN 1
					ELSE 0
				END
			), 0) AS total_points,
			COUNT(CASE
				WHEN m.status = 'finished' AND p.local_score = m.local_score AND p.visitor_score = m.visitor_score THEN 1
			 END) AS exact_score,
			COUNT(CASE
				WHEN m.status = 'finished'
					AND NOT (p.local_score = m.local_score AND p.visitor_score = m.visitor_score)
					AND (
						(p.local_score > p.visitor_score AND m.local_score > m.visitor_score) OR
						(p.local_score < p.visitor_score AND m.local_score < m.visitor_score) OR
						(p.local_score = p.visitor_score AND m.local_score = m.visitor_score)
					) THEN 1
			 END) AS result_only,
			COUNT(CASE
				WHEN m.status = 'finished'
					AND NOT (
						(p.local_score = m.local_score AND p.visitor_score = m.visitor_score) OR
						(p.local_score > p.visitor_score AND m.local_score > m.visitor_score) OR
						(p.local_score < p.visitor_score AND m.local_score < m.visitor_score) OR
						(p.local_score = p.visitor_score AND m.local_score = m.visitor_score)
					) THEN 1
			 END) AS wrong,
			COUNT(p.id) AS pred_count
		FROM users u
		LEFT JOIN predictions p ON u.id = p.user_id
		LEFT JOIN matches m ON p.match_id = m.id
		GROUP BY u.id
		ORDER BY total_points DESC, exact_score DESC, result_only DESC, u.username ASC
	`)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var standings []models.Standing
	for rows.Next() {
		var s models.Standing
		if err := rows.Scan(&s.UserID, &s.Username, &s.TotalPoints, &s.ExactScore, &s.ResultOnly, &s.Wrong, &s.PredCount); err != nil {
			continue
		}
		standings = append(standings, s)
	}
	respondJSON(w, 200, standings)
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	if req.Username == "" || req.Password == "" {
		respondJSON(w, 400, map[string]string{"error": "Usuario y contraseña son requeridos"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": "Failed to hash password"})
		return
	}

	// El primer usuario registrado es administrador; los siguientes son usuarios normales
	var userCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	isAdmin := 0
	if userCount == 0 {
		isAdmin = 1
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, password_hash, is_admin) VALUES (?, ?, ?)",
		req.Username, string(hash), isAdmin,
	)
	if err != nil {
		respondJSON(w, 400, map[string]string{"error": "El nombre de usuario ya existe"})
		return
	}

	id, _ := result.LastInsertId()
	respondJSON(w, 200, map[string]any{"id": int(id), "username": req.Username, "is_admin": isAdmin == 1})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	var user models.User
	var pwdHash string
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, is_admin FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &pwdHash, &user.IsAdmin)
	if err != nil {
		respondJSON(w, 401, map[string]string{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pwdHash), []byte(req.Password)); err != nil {
		respondJSON(w, 401, map[string]string{"error": "Invalid credentials"})
		return
	}

	token := fmt.Sprintf("token_%d_%s_%d", user.ID, user.Username, time.Now().Unix())
	respondJSON(w, 200, models.AuthResponse{Token: token, User: user})
}

func GetUserPredictions(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	userID := parts[len(parts)-1]

	rows, err := database.DB.Query(`
		SELECT p.id, p.user_id, p.match_id, p.local_score, p.visitor_score, p.points,
			m.match_number, m.stage, m.match_date,
			COALESCE(t1.name,'') as local_team, COALESCE(t1.iso2,'') as local_iso2,
			COALESCE(t2.name,'') as visitor_team, COALESCE(t2.iso2,'') as visitor_iso2,
			m.local_score as real_ls, m.visitor_score as real_vs, m.status
		FROM predictions p
		JOIN matches m ON p.match_id = m.id
		LEFT JOIN teams t1 ON m.local_team_id = t1.id
		LEFT JOIN teams t2 ON m.visitor_team_id = t2.id
		WHERE p.user_id = ?
		ORDER BY m.match_number
	`, userID)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	type PredictionRow struct {
		ID           uint   `json:"id"`
		UserID       uint   `json:"user_id"`
		MatchID      uint   `json:"match_id"`
		LocalScore   int    `json:"local_score"`
		VisitorScore int    `json:"visitor_score"`
		Points       int    `json:"points"`
		MatchNumber  int    `json:"match_number"`
		Stage        string `json:"stage"`
		MatchDate    string `json:"match_date"`
		LocalTeam    string `json:"local_team"`
		LocalISO2    string `json:"local_iso2"`
		VisitorTeam  string `json:"visitor_team"`
		VisitorISO2  string `json:"visitor_iso2"`
		RealLS       *int   `json:"real_local_score"`
		RealVS       *int   `json:"real_visitor_score"`
		Status       string `json:"status"`
	}

	var predictions []PredictionRow
	for rows.Next() {
		var p PredictionRow
		if err := rows.Scan(&p.ID, &p.UserID, &p.MatchID, &p.LocalScore, &p.VisitorScore, &p.Points,
			&p.MatchNumber, &p.Stage, &p.MatchDate,
			&p.LocalTeam, &p.LocalISO2, &p.VisitorTeam, &p.VisitorISO2,
			&p.RealLS, &p.RealVS, &p.Status); err != nil {
			continue
		}
		predictions = append(predictions, p)
	}
	respondJSON(w, 200, predictions)
}

// ExportPredictionsExcel generates an Excel file that the user can fill in offline
// and later import back via ImportPredictionsExcel.
func ExportPredictionsExcel(w http.ResponseWriter, r *http.Request) {
	// Extract userID and username from query params
	userID := r.URL.Query().Get("user_id")
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "usuario"
	}

	// Fetch all group-stage matches WITH existing predictions for this user
	rows, err := database.DB.Query(`
		SELECT m.id, m.match_number, g.name,
			t1.name, t1.iso2,
			t2.name, t2.iso2,
			m.match_date, m.city,
			COALESCE(p.local_score, -1), COALESCE(p.visitor_score, -1)
		FROM matches m
		JOIN groups g ON m.group_id = g.id
		JOIN teams t1 ON m.local_team_id = t1.id
		JOIN teams t2 ON m.visitor_team_id = t2.id
		LEFT JOIN predictions p ON p.match_id = m.id AND p.user_id = ?
		WHERE m.stage = 'group'
		ORDER BY m.group_id, m.match_number
	`, userID)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	type mRow struct {
		ID, Num                 int
		Group, Local, LocalISO2 string
		Visitor, VisitorISO2    string
		Date, City              string
		PredLocal, PredVisitor  int // -1 = not set
	}
	var data []mRow
	for rows.Next() {
		var m mRow
		if err := rows.Scan(&m.ID, &m.Num, &m.Group,
			&m.Local, &m.LocalISO2, &m.Visitor, &m.VisitorISO2,
			&m.Date, &m.City, &m.PredLocal, &m.PredVisitor); err != nil {
			continue
		}
		data = append(data, m)
	}

	// Build Excel
	f := excelize.NewFile()
	sh := "Pronósticos"
	f.SetSheetName("Sheet1", sh)

	// --- Styles ---
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"1976D2"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"0D47A1"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "FFFFFF", Style: 1}, {Type: "right", Color: "FFFFFF", Style: 1}},
	})
	editableStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 13, Color: "0D47A1"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"E3F2FD"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "1976D2", Style: 2}, {Type: "right", Color: "1976D2", Style: 2}, {Type: "top", Color: "1976D2", Style: 2}, {Type: "bottom", Color: "1976D2", Style: 2}},
	})
	sepStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 13},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	teamStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    []excelize.Border{{Type: "bottom", Color: "CCCCCC", Style: 1}},
	})
	groupStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"1565C0"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	idStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "CCCCCC", Size: 8},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F5F5F5"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Row 1: Title
	f.MergeCell(sh, "A1", "I1")
	f.SetCellValue(sh, "A1", fmt.Sprintf("Quinela Mundial 2026 — Pronósticos de %s", strings.ToUpper(username)))
	f.SetCellStyle(sh, "A1", "I1", titleStyle)
	f.SetRowHeight(sh, 1, 28)

	// Row 2: Instructions
	f.MergeCell(sh, "A2", "I2")
	f.SetCellValue(sh, "A2", "⚠ Completa las columnas F (Goles Local) y G (Goles Visitante) y luego importa este archivo en la aplicación.")
	instructStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Italic: true, Color: "555555"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FFF9C4"}},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
	})
	f.SetCellStyle(sh, "A2", "I2", instructStyle)
	f.SetRowHeight(sh, 2, 36)

	// Row 3: Header
	headers := []string{"ID", "#", "Grupo", "Fecha", "Equipo Local", "GOLES LOCAL", "GOLES VISITANTE", "Equipo Visitante", "Ciudad"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		f.SetCellValue(sh, cell, h)
	}
	f.SetCellStyle(sh, "A3", "I3", headerStyle)
	f.SetRowHeight(sh, 3, 20)

	// Column widths
	f.SetColWidth(sh, "A", "A", 6)  // ID
	f.SetColWidth(sh, "B", "B", 6)  // #
	f.SetColWidth(sh, "C", "C", 7)  // Grupo
	f.SetColWidth(sh, "D", "D", 20) // Fecha
	f.SetColWidth(sh, "E", "E", 22) // Local
	f.SetColWidth(sh, "F", "F", 14) // G.Local
	f.SetColWidth(sh, "G", "G", 14) // G.Visitor
	f.SetColWidth(sh, "H", "H", 22) // Visitante
	f.SetColWidth(sh, "I", "I", 18) // Ciudad

	// Data rows
	currentGroup := ""
	row := 4 // start after 3 header rows
	for _, m := range data {
		// Group separator row
		if m.Group != currentGroup {
			currentGroup = m.Group
			f.MergeCell(sh, fmt.Sprintf("A%d", row), fmt.Sprintf("I%d", row))
			f.SetCellValue(sh, fmt.Sprintf("A%d", row), fmt.Sprintf("▶  GRUPO %s", m.Group))
			f.SetCellStyle(sh, fmt.Sprintf("A%d", row), fmt.Sprintf("I%d", row), groupStyle)
			f.SetRowHeight(sh, row, 18)
			row++
		}

		// match_id (col A)
		f.SetCellValue(sh, fmt.Sprintf("A%d", row), m.ID)
		f.SetCellStyle(sh, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), idStyle)
		// match number (col B)
		f.SetCellValue(sh, fmt.Sprintf("B%d", row), m.Num)
		f.SetCellStyle(sh, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), teamStyle)
		// group (col C)
		f.SetCellValue(sh, fmt.Sprintf("C%d", row), m.Group)
		f.SetCellStyle(sh, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), teamStyle)
		// date (col D)
		f.SetCellValue(sh, fmt.Sprintf("D%d", row), m.Date)
		f.SetCellStyle(sh, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), teamStyle)
		// local team (col E)
		f.SetCellValue(sh, fmt.Sprintf("E%d", row), m.Local)
		f.SetCellStyle(sh, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), teamStyle)
		// goles local (col F) — editable
		if m.PredLocal >= 0 {
			f.SetCellValue(sh, fmt.Sprintf("F%d", row), m.PredLocal)
		}
		f.SetCellStyle(sh, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), editableStyle)
		// goles visitante (col G) — editable
		if m.PredVisitor >= 0 {
			f.SetCellValue(sh, fmt.Sprintf("G%d", row), m.PredVisitor)
		}
		f.SetCellStyle(sh, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), editableStyle)
		// separator label between scores
		f.SetCellStyle(sh, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), editableStyle)
		_ = sepStyle
		// visitor team (col H)
		f.SetCellValue(sh, fmt.Sprintf("H%d", row), m.Visitor)
		f.SetCellStyle(sh, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), teamStyle)
		// city (col I)
		f.SetCellValue(sh, fmt.Sprintf("I%d", row), m.City)
		f.SetCellStyle(sh, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), teamStyle)

		f.SetRowHeight(sh, row, 20)
		row++
	}

	// Freeze header rows
	f.SetPanes(sh, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      3,
		TopLeftCell: "A4",
		ActivePane:  "bottomLeft",
	})

	// Write response
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"quinela_mundial2026_%s.xlsx\"", strings.ToLower(username)))
	f.Write(w)
}

// ImportPredictionsExcel reads an uploaded Excel file and saves the predictions
// for the given user. Columns: A=match_id, F=goles_local, G=goles_visitante
func ImportPredictionsExcel(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB max

	file, _, err := r.FormFile("file")
	if err != nil {
		respondJSON(w, 400, map[string]string{"error": "No se recibió ningún archivo"})
		return
	}
	defer file.Close()

	userIDStr := r.FormValue("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		respondJSON(w, 400, map[string]string{"error": "user_id inválido"})
		return
	}

	// Check lock
	var locked int
	database.DB.QueryRow("SELECT locked FROM prediction_locks WHERE user_id = ?", userID).Scan(&locked)
	if locked == 1 {
		respondJSON(w, 403, map[string]string{"error": "Los pronósticos están cerrados. No se puede importar."})
		return
	}

	xf, err := excelize.OpenReader(file)
	if err != nil {
		respondJSON(w, 400, map[string]string{"error": "Archivo Excel inválido: " + err.Error()})
		return
	}

	// Find sheet
	sheets := xf.GetSheetList()
	if len(sheets) == 0 {
		respondJSON(w, 400, map[string]string{"error": "El archivo no tiene hojas"})
		return
	}
	sh := sheets[0]

	rows, err := xf.GetRows(sh)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	saved := 0
	skipped := 0
	var importErrors []string

	for rowIdx, row := range rows {
		if rowIdx < 3 {
			continue // skip title, instructions, header
		}
		if len(row) < 7 {
			continue
		}

		// Col A = match_id (index 0)
		matchIDStr := strings.TrimSpace(row[0])
		if matchIDStr == "" {
			continue // group separator row
		}
		matchID, err := strconv.Atoi(matchIDStr)
		if err != nil || matchID <= 0 {
			continue // group separator or invalid row
		}

		// Col F = goles local (index 5)
		lsStr := strings.TrimSpace(row[5])
		// Col G = goles visitante (index 6)
		vsStr := ""
		if len(row) > 6 {
			vsStr = strings.TrimSpace(row[6])
		}

		if lsStr == "" || vsStr == "" {
			skipped++
			continue // user didn't fill this row
		}

		ls, errL := strconv.Atoi(lsStr)
		vs, errV := strconv.Atoi(vsStr)
		if errL != nil || errV != nil || ls < 0 || vs < 0 {
			importErrors = append(importErrors, fmt.Sprintf("Fila %d: valor inválido", rowIdx+1))
			continue
		}

		_, dbErr := database.DB.Exec(`
			INSERT INTO predictions (user_id, match_id, local_score, visitor_score)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id, match_id) DO UPDATE SET local_score=excluded.local_score, visitor_score=excluded.visitor_score
		`, userID, matchID, ls, vs)
		if dbErr != nil {
			importErrors = append(importErrors, fmt.Sprintf("Partido %d: error DB", matchID))
		} else {
			saved++
		}
	}

	respondJSON(w, 200, map[string]any{
		"saved":   saved,
		"skipped": skipped,
		"errors":  importErrors,
		"message": fmt.Sprintf("%d pronósticos importados correctamente", saved),
	})
}

// ─── User management (admin only) ────────────────────────────────────────────

type UserAdmin struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	IsAdmin   bool   `json:"is_admin"`
	PredCount int    `json:"pred_count"`
	Locked    bool   `json:"locked"`
	Points    int    `json:"points"`
}

func requireAdmin(r *http.Request) bool {
	// Simple token validation: token format is "token_{id}_{username}_{ts}"
	// The admin flag is carried in the token after resolution.
	// We verify by looking up the user from the Authorization header token.
	auth := r.Header.Get("Authorization")
	if auth == "" {
		auth = r.URL.Query().Get("token")
	}
	if !strings.HasPrefix(auth, "token_") {
		return false
	}
	// token_<id>_<username>_<ts>
	parts := strings.Split(auth, "_")
	if len(parts) < 3 {
		return false
	}
	userID := parts[1]
	var isAdmin int
	database.DB.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
	return isAdmin == 1
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(r) {
		respondJSON(w, 403, map[string]string{"error": "Acceso denegado: se requiere administrador"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT u.id, u.username, u.is_admin,
			COALESCE(COUNT(p.id),0) as pred_count,
			COALESCE(SUM(p.points),0) as points,
			COALESCE(pl.locked,0) as locked
		FROM users u
		LEFT JOIN predictions p ON p.user_id = u.id
		LEFT JOIN prediction_locks pl ON pl.user_id = u.id
		GROUP BY u.id ORDER BY u.username
	`)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []UserAdmin
	for rows.Next() {
		var u UserAdmin
		var admin, locked int
		if err := rows.Scan(&u.ID, &u.Username, &admin, &u.PredCount, &u.Points, &locked); err != nil {
			continue
		}
		u.IsAdmin = admin == 1
		u.Locked = locked == 1
		users = append(users, u)
	}
	respondJSON(w, 200, users)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(r) {
		respondJSON(w, 403, map[string]string{"error": "Acceso denegado"})
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	userID := parts[len(parts)-1]

	var req struct {
		Username string `json:"username"`
		IsAdmin  *bool  `json:"is_admin"`
		Password string `json:"password"` // optional: reset password
		Locked   *bool  `json:"locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	if req.Username != "" {
		if _, err := database.DB.Exec("UPDATE users SET username = ? WHERE id = ?", req.Username, userID); err != nil {
			respondJSON(w, 500, map[string]string{"error": "Error actualizando username: " + err.Error()})
			return
		}
	}

	if req.IsAdmin != nil {
		adminVal := 0
		if *req.IsAdmin {
			adminVal = 1
		}
		if _, err := database.DB.Exec("UPDATE users SET is_admin = ? WHERE id = ?", adminVal, userID); err != nil {
			respondJSON(w, 500, map[string]string{"error": "Error actualizando rol: " + err.Error()})
			return
		}
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			respondJSON(w, 500, map[string]string{"error": "Error generando hash"})
			return
		}
		if _, err := database.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hash), userID); err != nil {
			respondJSON(w, 500, map[string]string{"error": "Error actualizando contraseña"})
			return
		}
	}

	if req.Locked != nil {
		if *req.Locked {
			now := time.Now().Format("2006-01-02 15:04:05")
			database.DB.Exec(`INSERT INTO prediction_locks (user_id,locked,locked_at) VALUES (?,1,?)
				ON CONFLICT(user_id) DO UPDATE SET locked=1,locked_at=?`, userID, now, now)
		} else {
			database.DB.Exec("UPDATE prediction_locks SET locked=0 WHERE user_id=?", userID)
		}
	}

	respondJSON(w, 200, map[string]string{"message": "Usuario actualizado correctamente"})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(r) {
		respondJSON(w, 403, map[string]string{"error": "Acceso denegado"})
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	userID := parts[len(parts)-1]

	// Prevent deleting the last admin
	var adminID string
	database.DB.QueryRow("SELECT id FROM users WHERE is_admin=1 AND id != ?", userID).Scan(&adminID)
	var isAdmin int
	database.DB.QueryRow("SELECT is_admin FROM users WHERE id=?", userID).Scan(&isAdmin)
	if isAdmin == 1 && adminID == "" {
		respondJSON(w, 400, map[string]string{"error": "No se puede eliminar el único administrador"})
		return
	}

	database.DB.Exec("DELETE FROM prediction_locks WHERE user_id=?", userID)
	database.DB.Exec("DELETE FROM predictions WHERE user_id=?", userID)
	if _, err := database.DB.Exec("DELETE FROM users WHERE id=?", userID); err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, 200, map[string]string{"message": "Usuario eliminado"})
}

func runAPIServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/groups", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/groups" && r.Method == http.MethodGet {
			GetGroups(w, r)
		}
	}))

	mux.HandleFunc("/api/matches", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			GetMatches(w, r)
		}
	}))

	mux.HandleFunc("/api/standings", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			GetStandings(w, r)
		}
	}))

	mux.HandleFunc("/api/results", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if !requireAdmin(r) {
				respondJSON(w, 403, map[string]string{"error": "Acceso denegado: solo administradores pueden ingresar resultados"})
				return
			}
			SaveResult(w, r)
		}
	}))

	mux.HandleFunc("/api/predictions", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			SavePrediction(w, r)
		}
	}))

	mux.HandleFunc("/api/predictions/lock", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			LockUserPredictions(w, r)
		}
	}))

	mux.HandleFunc("/api/predictions/export", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ExportPredictionsExcel(w, r)
		}
	}))

	mux.HandleFunc("/api/predictions/import", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ImportPredictionsExcel(w, r)
		}
	}))

	mux.HandleFunc("/api/users", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetUsers(w, r)
		case http.MethodPost:
			Register(w, r) // create user (admin may create too)
		}
	}))

	mux.HandleFunc("/api/auth/register", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			Register(w, r)
		}
	}))

	mux.HandleFunc("/api/auth/login", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			Login(w, r)
		}
	}))

	re := regexp.MustCompile(`^/api/groups/(\d+)/(standings|matches)$`)

	mux.HandleFunc("/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if m := re.FindStringSubmatch(r.URL.Path); len(m) > 0 {
			switch m[2] {
			case "standings":
				GetGroupStandings(w, r)
			case "matches":
				GetGroupMatches(w, r)
			}
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/schedule/") {
			GetScheduleByTeam(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/predictions/user/") {
			GetAllGroupMatchesWithPredictions(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/predictions/lock/") && r.Method == http.MethodGet {
			GetPredictionLockStatus(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/predictions/") {
			GetUserPredictions(w, r)
			return
		}

		reUser := regexp.MustCompile(`^/api/users/(\d+)$`)
		if m := reUser.FindStringSubmatch(r.URL.Path); len(m) > 0 {
			switch r.Method {
			case http.MethodPut:
				UpdateUser(w, r)
			case http.MethodDelete:
				DeleteUser(w, r)
			}
			return
		}

		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))

	fmt.Printf("API server running on http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, mux)
}

func main() {
	if err := database.Init("data/quinela.db"); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	go runAPIServer("8080")

	app := NewApp()
	err := wails.Run(&options.App{
		Title:  "Quinela Mundial 2026",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		fmt.Printf("Error running app: %v\n", err)
	}
}
