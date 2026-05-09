package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	if dir := filepath.Dir(dbPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating database directory: %w", err)
		}
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	if err = createTables(); err != nil {
		return fmt.Errorf("error creating tables: %w", err)
	}

	if err = seedData(); err != nil {
		return fmt.Errorf("error seeding data: %w", err)
	}

	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS teams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			short_code TEXT NOT NULL,
			iso2 TEXT NOT NULL,
			group_id INTEGER NOT NULL,
			is_host INTEGER DEFAULT 0,
			FOREIGN KEY (group_id) REFERENCES groups(id)
		)`,
		`CREATE TABLE IF NOT EXISTS matches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			match_number INTEGER NOT NULL,
			stage TEXT NOT NULL,
			group_id INTEGER,
			local_team_id INTEGER NOT NULL,
			visitor_team_id INTEGER NOT NULL,
			match_date TEXT NOT NULL,
			venue TEXT NOT NULL,
			city TEXT NOT NULL,
			local_score INTEGER,
			visitor_score INTEGER,
			status TEXT DEFAULT 'pending',
			FOREIGN KEY (group_id) REFERENCES groups(id),
			FOREIGN KEY (local_team_id) REFERENCES teams(id),
			FOREIGN KEY (visitor_team_id) REFERENCES teams(id)
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			is_admin INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS predictions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			match_id INTEGER NOT NULL,
			local_score INTEGER NOT NULL,
			visitor_score INTEGER NOT NULL,
			points INTEGER DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (match_id) REFERENCES matches(id),
			UNIQUE(user_id, match_id)
		)`,
		`CREATE TABLE IF NOT EXISTS prediction_locks (
			user_id INTEGER PRIMARY KEY,
			locked INTEGER DEFAULT 0,
			locked_at TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
	}

	return nil
}

func seedData() error {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM groups").Scan(&count)
	if count > 0 {
		return nil
	}

	groups := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}
	for _, g := range groups {
		DB.Exec("INSERT INTO groups (name) VALUES (?)", g)
	}

	// Equipos reales FIFA Mundial 2026 (datos oficiales)
	// Grupos A-L con 4 equipos cada uno, IDs 1-48
	teams := []struct {
		Name    string
		Code    string
		ISO2    string
		GroupID int
		IsHost  bool
	}{
		// Grupo A (id=1)
		{"México", "MEX", "mx", 1, true},
		{"Sudáfrica", "RSA", "za", 1, false},
		{"Corea del Sur", "KOR", "kr", 1, false},
		{"República Checa", "CZE", "cz", 1, false},
		// Grupo B (id=2)
		{"Canadá", "CAN", "ca", 2, true},
		{"Bosnia y Herzegovina", "BIH", "ba", 2, false},
		{"Catar", "QAT", "qa", 2, false},
		{"Suiza", "SUI", "ch", 2, false},
		// Grupo C (id=3)
		{"Brasil", "BRA", "br", 3, false},
		{"Marruecos", "MAR", "ma", 3, false},
		{"Haití", "HAI", "ht", 3, false},
		{"Escocia", "SCO", "gb-sct", 3, false},
		// Grupo D (id=4)
		{"Estados Unidos", "USA", "us", 4, true},
		{"Paraguay", "PAR", "py", 4, false},
		{"Australia", "AUS", "au", 4, false},
		{"Türkiye", "TUR", "tr", 4, false},
		// Grupo E (id=5)
		{"Alemania", "GER", "de", 5, false},
		{"Curazao", "CUW", "cw", 5, false},
		{"Costa de Marfil", "CIV", "ci", 5, false},
		{"Ecuador", "ECU", "ec", 5, false},
		// Grupo F (id=6)
		{"Países Bajos", "NED", "nl", 6, false},
		{"Japón", "JPN", "jp", 6, false},
		{"Suecia", "SWE", "se", 6, false},
		{"Túnez", "TUN", "tn", 6, false},
		// Grupo G (id=7)
		{"Bélgica", "BEL", "be", 7, false},
		{"Egipto", "EGY", "eg", 7, false},
		{"Irán", "IRN", "ir", 7, false},
		{"Nueva Zelanda", "NZL", "nz", 7, false},
		// Grupo H (id=8)
		{"España", "ESP", "es", 8, false},
		{"Cabo Verde", "CPV", "cv", 8, false},
		{"Arabia Saudita", "KSA", "sa", 8, false},
		{"Uruguay", "URU", "uy", 8, false},
		// Grupo I (id=9)
		{"Francia", "FRA", "fr", 9, false},
		{"Senegal", "SEN", "sn", 9, false},
		{"Irak", "IRQ", "iq", 9, false},
		{"Noruega", "NOR", "no", 9, false},
		// Grupo J (id=10)
		{"Argentina", "ARG", "ar", 10, false},
		{"Argelia", "ALG", "dz", 10, false},
		{"Austria", "AUT", "at", 10, false},
		{"Jordania", "JOR", "jo", 10, false},
		// Grupo K (id=11)
		{"Portugal", "POR", "pt", 11, false},
		{"RD Congo", "COD", "cd", 11, false},
		{"Uzbekistán", "UZB", "uz", 11, false},
		{"Colombia", "COL", "co", 11, false},
		// Grupo L (id=12)
		{"Inglaterra", "ENG", "gb-eng", 12, false},
		{"Croacia", "CRO", "hr", 12, false},
		{"Ghana", "GHA", "gh", 12, false},
		{"Panamá", "PAN", "pa", 12, false},
	}

	for _, t := range teams {
		DB.Exec("INSERT INTO teams (name, short_code, iso2, group_id, is_host) VALUES (?, ?, ?, ?, ?)", t.Name, t.Code, t.ISO2, t.GroupID, t.IsHost)
	}

	// 72 partidos de fase de grupos — datos oficiales FIFA Mundial 2026
	// Team IDs: A(1-4), B(5-8), C(9-12), D(13-16), E(17-20), F(21-24),
	//           G(25-28), H(29-32), I(33-36), J(37-40), K(41-44), L(45-48)
	type matchSeed struct {
		Num, Group, Local, Visitor int
		Date, Venue, City          string
	}

	matches := []matchSeed{
		// ── Jornada 1 ──────────────────────────────────────────────────────
		{1, 1, 1, 2, "2026-06-11 13:00", "Estadio Azteca", "Ciudad de México"},
		{2, 1, 3, 4, "2026-06-11 20:00", "Estadio Akron", "Zapopan"},
		{3, 2, 5, 6, "2026-06-12 15:00", "BMO Field", "Toronto"},
		{4, 4, 13, 14, "2026-06-12 18:00", "SoFi Stadium", "Inglewood"},
		{5, 3, 11, 12, "2026-06-13 21:00", "Gillette Stadium", "Foxborough"},
		{6, 4, 15, 16, "2026-06-13 21:00", "BC Place", "Vancouver"},
		{7, 3, 9, 10, "2026-06-13 18:00", "MetLife Stadium", "East Rutherford"},
		{8, 2, 7, 8, "2026-06-13 12:00", "Levi's Stadium", "Santa Clara"},
		{9, 5, 19, 20, "2026-06-14 19:00", "Lincoln Financial Field", "Filadelfia"},
		{10, 5, 17, 18, "2026-06-14 12:00", "NRG Stadium", "Houston"},
		{11, 6, 21, 22, "2026-06-14 15:00", "AT&T Stadium", "Arlington"},
		{12, 6, 23, 24, "2026-06-14 20:00", "Estadio BBVA", "Guadalupe"},
		{13, 8, 31, 32, "2026-06-15 18:00", "Hard Rock Stadium", "Miami Gardens"},
		{14, 8, 29, 30, "2026-06-15 12:00", "Mercedes-Benz Stadium", "Atlanta"},
		{15, 7, 27, 28, "2026-06-15 18:00", "SoFi Stadium", "Inglewood"},
		{16, 7, 25, 26, "2026-06-15 12:00", "Lumen Field", "Seattle"},
		{17, 9, 33, 34, "2026-06-16 20:00", "Arrowhead Stadium", "Kansas City"},
		{18, 10, 39, 40, "2026-06-16 21:00", "Levi's Stadium", "Santa Clara"},
		{19, 10, 37, 38, "2026-06-16 20:00", "Arrowhead Stadium", "Kansas City"},
		{20, 9, 35, 36, "2026-06-16 21:00", "Levi's Stadium", "Santa Clara"},
		{21, 12, 47, 48, "2026-06-17 19:00", "BMO Field", "Toronto"},
		{22, 12, 45, 46, "2026-06-17 15:00", "AT&T Stadium", "Arlington"},
		{23, 11, 41, 42, "2026-06-17 12:00", "NRG Stadium", "Houston"},
		{24, 11, 43, 44, "2026-06-17 20:00", "Estadio Azteca", "Ciudad de México"},
		// ── Jornada 2 ──────────────────────────────────────────────────────
		{25, 1, 4, 2, "2026-06-18 12:00", "Mercedes-Benz Stadium", "Atlanta"},
		{26, 2, 8, 6, "2026-06-18 12:00", "SoFi Stadium", "Inglewood"},
		{27, 2, 5, 7, "2026-06-18 15:00", "BC Place", "Vancouver"},
		{28, 1, 1, 3, "2026-06-18 19:00", "Estadio Akron", "Zapopan"},
		{29, 3, 9, 11, "2026-06-19 20:30", "Lincoln Financial Field", "Filadelfia"},
		{30, 3, 12, 10, "2026-06-19 18:00", "Gillette Stadium", "Foxborough"},
		{31, 4, 16, 14, "2026-06-19 20:00", "Levi's Stadium", "Santa Clara"},
		{32, 4, 13, 15, "2026-06-19 12:00", "Lumen Field", "Seattle"},
		{33, 5, 17, 19, "2026-06-20 16:00", "BMO Field", "Toronto"},
		{34, 5, 20, 18, "2026-06-20 19:00", "Arrowhead Stadium", "Kansas City"},
		{35, 6, 21, 23, "2026-06-20 12:00", "NRG Stadium", "Houston"},
		{36, 6, 24, 22, "2026-06-20 22:00", "Estadio BBVA", "Guadalupe"},
		{37, 8, 32, 30, "2026-06-21 18:00", "Hard Rock Stadium", "Miami Gardens"},
		{38, 8, 29, 31, "2026-06-21 12:00", "Mercedes-Benz Stadium", "Atlanta"},
		{39, 7, 25, 27, "2026-06-21 12:00", "SoFi Stadium", "Inglewood"},
		{40, 7, 28, 26, "2026-06-21 18:00", "BC Place", "Vancouver"},
		{41, 9, 36, 34, "2026-06-21 18:00", "Hard Rock Stadium", "Miami Gardens"},
		{42, 9, 33, 35, "2026-06-21 20:00", "Hard Rock Stadium", "Miami Gardens"},
		{43, 10, 37, 39, "2026-06-22 12:00", "AT&T Stadium", "Arlington"},
		{44, 10, 40, 38, "2026-06-22 20:00", "Levi's Stadium", "Santa Clara"},
		{45, 12, 45, 47, "2026-06-23 16:00", "Gillette Stadium", "Foxborough"},
		{46, 12, 48, 46, "2026-06-23 19:00", "BMO Field", "Toronto"},
		{47, 11, 41, 43, "2026-06-23 12:00", "NRG Stadium", "Houston"},
		{48, 11, 44, 42, "2026-06-23 20:00", "Estadio Akron", "Zapopan"},
		// ── Jornada 3 ──────────────────────────────────────────────────────
		{49, 3, 12, 9, "2026-06-24 18:00", "Hard Rock Stadium", "Miami Gardens"},
		{50, 3, 10, 11, "2026-06-24 18:00", "Mercedes-Benz Stadium", "Atlanta"},
		{51, 2, 8, 5, "2026-06-24 12:00", "BC Place", "Vancouver"},
		{52, 2, 6, 7, "2026-06-24 12:00", "Lumen Field", "Seattle"},
		{53, 1, 4, 1, "2026-06-24 19:00", "Estadio Azteca", "Ciudad de México"},
		{54, 1, 2, 3, "2026-06-24 19:00", "Estadio BBVA", "Guadalupe"},
		{55, 5, 18, 19, "2026-06-25 16:00", "Lincoln Financial Field", "Filadelfia"},
		{56, 5, 20, 17, "2026-06-25 16:00", "MetLife Stadium", "East Rutherford"},
		{57, 6, 22, 23, "2026-06-25 18:00", "AT&T Stadium", "Arlington"},
		{58, 6, 24, 21, "2026-06-25 18:00", "Arrowhead Stadium", "Kansas City"},
		{59, 4, 16, 13, "2026-06-25 19:00", "SoFi Stadium", "Inglewood"},
		{60, 4, 14, 15, "2026-06-25 19:00", "Levi's Stadium", "Santa Clara"},
		{61, 9, 36, 33, "2026-06-26 20:00", "Hard Rock Stadium", "Miami Gardens"},
		{62, 9, 34, 35, "2026-06-26 20:00", "Hard Rock Stadium", "Miami Gardens"},
		{63, 7, 26, 27, "2026-06-26 20:00", "Lumen Field", "Seattle"},
		{64, 7, 28, 25, "2026-06-26 20:00", "BC Place", "Vancouver"},
		{65, 8, 30, 31, "2026-06-26 19:00", "NRG Stadium", "Houston"},
		{66, 8, 32, 29, "2026-06-26 18:00", "Estadio Akron", "Zapopan"},
		{67, 12, 48, 45, "2026-06-27 17:00", "MetLife Stadium", "East Rutherford"},
		{68, 12, 46, 47, "2026-06-27 17:00", "Lincoln Financial Field", "Filadelfia"},
		{69, 10, 38, 39, "2026-06-27 21:00", "Arrowhead Stadium", "Kansas City"},
		{70, 10, 40, 37, "2026-06-27 21:00", "AT&T Stadium", "Arlington"},
		{71, 11, 44, 41, "2026-06-27 19:30", "Hard Rock Stadium", "Miami Gardens"},
		{72, 11, 42, 43, "2026-06-27 19:30", "Mercedes-Benz Stadium", "Atlanta"},
	}

	for _, m := range matches {
		DB.Exec(`INSERT INTO matches (match_number, stage, group_id, local_team_id, visitor_team_id, match_date, venue, city, status)
			VALUES (?, 'group', ?, ?, ?, ?, ?, ?, 'pending')`,
			m.Num, m.Group, m.Local, m.Visitor, m.Date, m.Venue, m.City)
	}

	return nil
}
