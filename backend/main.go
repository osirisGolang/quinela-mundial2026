package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"quinela/internal/database"
	"quinela/internal/models"
	"regexp"
	"sort"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"golang.org/x/crypto/bcrypt"
)

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
		Pts                      int
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
		TeamName, ISO2                                  string
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

	_, err := database.DB.Exec(`
		INSERT OR REPLACE INTO predictions (user_id, match_id, local_score, visitor_score)
		VALUES (?, ?, ?, ?)
	`, pred.UserID, pred.MatchID, pred.LocalScore, pred.VisitorScore)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, 200, map[string]string{"message": "Prediction saved"})
}

func GetStandings(w http.ResponseWriter, r *http.Request) {
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
		respondJSON(w, 500, map[string]string{"error": err.Error()})
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
	respondJSON(w, 200, standings)
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondJSON(w, 500, map[string]string{"error": "Failed to hash password"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, password_hash, is_admin) VALUES (?, ?, 0)",
		req.Username, string(hash),
	)
	if err != nil {
		respondJSON(w, 400, map[string]string{"error": "Username already exists"})
		return
	}

	id, _ := result.LastInsertId()
	respondJSON(w, 200, map[string]any{"id": int(id), "username": req.Username, "is_admin": false})
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
			SaveResult(w, r)
		}
	}))

	mux.HandleFunc("/api/predictions", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			SavePrediction(w, r)
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
			r.URL.Path = "/api/groups/" + m[1] + "/" + m[2]
		}

		if strings.HasPrefix(r.URL.Path, "/api/schedule/") {
			GetScheduleByTeam(w, r)
			return
		}

		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))

	fmt.Printf("API server running on http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, mux)
}

func main() {
	database.Init("data/quinela.db")
	go runAPIServer("8080")

	wails.Run(&options.App{
		Title:  "Quinela Mundial 2026",
		Width:  1280,
		Height: 800,
	})
}