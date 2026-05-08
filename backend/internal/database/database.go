package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
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

	teams := []struct {
		Name      string
		Code      string
		ISO2      string
		GroupID   int
		IsHost    bool
	}{
		{"Canada", "CAN", "ca", 1, true}, {"Mexico", "MEX", "mx", 1, true}, {"Argentina", "ARG", "ar", 1, false}, {"Peru", "PER", "pe", 1, false},
		{"United States", "USA", "us", 2, true}, {"Brazil", "BRA", "br", 2, false}, {"Uruguay", "URU", "uy", 2, false}, {"Colombia", "COL", "co", 2, false},
		{"Germany", "GER", "de", 3, false}, {"Spain", "ESP", "es", 3, false}, {"France", "FRA", "fr", 3, false}, {"Italy", "ITA", "it", 3, false},
		{"England", "ENG", "gb-eng", 4, false}, {"Portugal", "POR", "pt", 4, false}, {"Netherlands", "NED", "nl", 4, false}, {"Belgium", "BEL", "be", 4, false},
		{"Japan", "JPN", "jp", 5, false}, {"South Korea", "KOR", "kr", 5, false}, {"Iran", "IRN", "ir", 5, false}, {"Australia", "AUS", "au", 5, false},
		{"Saudi Arabia", "KSA", "sa", 6, false}, {"Qatar", "QAT", "qa", 6, false}, {"UAE", "UAE", "ae", 6, false}, {"Iraq", "IRQ", "iq", 6, false},
		{"Poland", "POL", "pl", 7, false}, {"Switzerland", "SUI", "ch", 7, false}, {"Croatia", "CRO", "hr", 7, false}, {"Denmark", "DEN", "dk", 7, false},
		{"Sweden", "SWE", "se", 8, false}, {"Norway", "NOR", "no", 8, false}, {"Austria", "AUT", "at", 8, false}, {"Czech Republic", "CZE", "cz", 8, false},
		{"Nigeria", "NGA", "ng", 9, false}, {"Cameroon", "CMR", "cm", 9, false}, {"Senegal", "SEN", "sn", 9, false}, {"Ghana", "GHA", "gh", 9, false},
		{"Egypt", "EGY", "eg", 10, false}, {"Morocco", "MAR", "ma", 10, false}, {"Algeria", "ALG", "dz", 10, false}, {"Tunisia", "TUN", "tn", 10, false},
		{"New Zealand", "NZL", "nz", 11, false}, {"Papua New Guinea", "PNG", "pg", 11, false}, {"Solomon Islands", "SOL", "sb", 11, false}, {"Fiji", "FIJ", "fj", 11, false},
		{"China", "CHN", "cn", 12, false}, {"India", "IND", "in", 12, false}, {"Indonesia", "IDN", "id", 12, false}, {"Malaysia", "MAS", "my", 12, false},
	}

	for _, t := range teams {
		DB.Exec("INSERT INTO teams (name, short_code, iso2, group_id, is_host) VALUES (?, ?, ?, ?, ?)", t.Name, t.Code, t.ISO2, t.GroupID, t.IsHost)
	}

	loc, _ := time.LoadLocation("America/Mexico_City")

	matches := []struct {
		Num         int
		Stage       string
		GroupID     int
		LocalTeamID int
		VisitorID   int
		Date        string
	}{
		{1, "group", 1, 1, 3, time.Date(2026, 6, 11, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {2, "group", 1, 2, 4, time.Date(2026, 6, 11, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{3, "group", 2, 5, 7, time.Date(2026, 6, 12, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {4, "group", 2, 6, 8, time.Date(2026, 6, 12, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{5, "group", 3, 9, 11, time.Date(2026, 6, 13, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {6, "group", 3, 10, 12, time.Date(2026, 6, 13, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{7, "group", 4, 13, 15, time.Date(2026, 6, 14, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {8, "group", 4, 14, 16, time.Date(2026, 6, 14, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{9, "group", 5, 17, 19, time.Date(2026, 6, 15, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {10, "group", 5, 18, 20, time.Date(2026, 6, 15, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{11, "group", 6, 21, 23, time.Date(2026, 6, 16, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {12, "group", 6, 22, 24, time.Date(2026, 6, 16, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{13, "group", 1, 1, 2, time.Date(2026, 6, 19, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {14, "group", 1, 4, 3, time.Date(2026, 6, 19, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{15, "group", 2, 6, 7, time.Date(2026, 6, 20, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {16, "group", 2, 8, 5, time.Date(2026, 6, 20, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{17, "group", 3, 9, 10, time.Date(2026, 6, 21, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {18, "group", 3, 12, 11, time.Date(2026, 6, 21, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{19, "group", 4, 13, 14, time.Date(2026, 6, 22, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {20, "group", 4, 16, 15, time.Date(2026, 6, 22, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{21, "group", 5, 17, 18, time.Date(2026, 6, 23, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {22, "group", 5, 20, 19, time.Date(2026, 6, 23, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{23, "group", 6, 21, 22, time.Date(2026, 6, 24, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {24, "group", 6, 24, 23, time.Date(2026, 6, 24, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{25, "group", 1, 3, 2, time.Date(2026, 6, 26, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {26, "group", 1, 1, 4, time.Date(2026, 6, 26, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{27, "group", 2, 7, 6, time.Date(2026, 6, 27, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {28, "group", 2, 5, 8, time.Date(2026, 6, 27, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{29, "group", 3, 11, 10, time.Date(2026, 6, 28, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {30, "group", 3, 9, 12, time.Date(2026, 6, 28, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{31, "group", 4, 15, 14, time.Date(2026, 6, 29, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {32, "group", 4, 13, 16, time.Date(2026, 6, 29, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{33, "group", 7, 25, 27, time.Date(2026, 6, 17, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {34, "group", 7, 26, 28, time.Date(2026, 6, 17, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{35, "group", 8, 29, 31, time.Date(2026, 6, 18, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {36, "group", 8, 30, 32, time.Date(2026, 6, 18, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{37, "group", 9, 33, 35, time.Date(2026, 6, 19, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {38, "group", 9, 34, 36, time.Date(2026, 6, 19, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{39, "group", 10, 37, 39, time.Date(2026, 6, 20, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {40, "group", 10, 38, 40, time.Date(2026, 6, 20, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{41, "group", 11, 41, 43, time.Date(2026, 6, 21, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {42, "group", 11, 42, 44, time.Date(2026, 6, 21, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{43, "group", 12, 45, 47, time.Date(2026, 6, 22, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {44, "group", 12, 46, 48, time.Date(2026, 6, 22, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{45, "group", 7, 27, 26, time.Date(2026, 6, 25, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {46, "group", 7, 28, 25, time.Date(2026, 6, 25, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{47, "group", 8, 31, 30, time.Date(2026, 6, 26, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {48, "group", 8, 32, 29, time.Date(2026, 6, 26, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{49, "group", 9, 35, 34, time.Date(2026, 6, 27, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {50, "group", 9, 36, 33, time.Date(2026, 6, 27, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{51, "group", 10, 39, 38, time.Date(2026, 6, 28, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {52, "group", 10, 40, 37, time.Date(2026, 6, 28, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{53, "group", 11, 43, 42, time.Date(2026, 6, 29, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {54, "group", 11, 44, 41, time.Date(2026, 6, 29, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{55, "group", 12, 47, 46, time.Date(2026, 6, 30, 14, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {56, "group", 12, 48, 45, time.Date(2026, 6, 30, 20, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{57, "group", 7, 25, 28, time.Date(2026, 7, 1, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {58, "group", 7, 27, 26, time.Date(2026, 7, 1, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{59, "group", 8, 29, 32, time.Date(2026, 7, 2, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {60, "group", 8, 31, 30, time.Date(2026, 7, 2, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{61, "group", 9, 33, 36, time.Date(2026, 7, 3, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {62, "group", 9, 35, 34, time.Date(2026, 7, 3, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{63, "group", 10, 37, 40, time.Date(2026, 7, 4, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {64, "group", 10, 39, 38, time.Date(2026, 7, 4, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{65, "group", 11, 41, 44, time.Date(2026, 7, 5, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {66, "group", 11, 43, 42, time.Date(2026, 7, 5, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
		{67, "group", 12, 45, 48, time.Date(2026, 7, 6, 12, 0, 0, 0, loc).Format("2006-01-02 15:04")}, {68, "group", 12, 47, 46, time.Date(2026, 7, 6, 18, 0, 0, 0, loc).Format("2006-01-02 15:04")},
	}

	venues := map[int][2]string{
		1: {"Estadio Azteca", "Ciudad de Mexico"}, 2: {"NRG Stadium", "Houston"}, 3: {"Rose Bowl", "Los Angeles"}, 4: {"MetLife Stadium", "Nueva York"},
		5: {"Estadio Akron", "Guadalajara"}, 6: {"BMO Field", "Toronto"}, 7: {"Levi's Stadium", "San Francisco"}, 8: {"Lumen Field", "Seattle"},
		9: {"AT&T Stadium", "Dallas"}, 10: {"SoFi Stadium", "Los Angeles"}, 11: {"GEODIS Park", "Nashville"}, 12: {"Mercedes-Benz Stadium", "Atlanta"},
		13: {"Red Bull Arena", "Nueva Jersey"}, 14: {"Lincoln Financial Field", "Philadelphia"}, 15: {"Hard Rock Stadium", "Miami"}, 16: {"Gillette Stadium", "Boston"},
		17: {"Arrowhead Stadium", "Kansas City"}, 18: {"Sports Authority Field", "Denver"}, 19: {"Allegiant Stadium", "Las Vegas"}, 20: {"State Farm Stadium", "Glendale"},
		21: {"Emirates Stadium", "San Nicolas"}, 22: {"BC Place", "Vancouver"}, 23: {"Stade Montreal", "Montreal"}, 24: {"SAP Stadium", "San Jose"},
		25: {"Estadio BBVA", "Monterrey"}, 26: {"Estadio Akron", "Guadalajara"}, 27: {"Estadio Azteca", "Ciudad de Mexico"}, 28: {"NRG Stadium", "Houston"},
		29: {"Rose Bowl", "Los Angeles"}, 30: {"MetLife Stadium", "Nueva York"}, 31: {"AT&T Stadium", "Dallas"}, 32: {"SoFi Stadium", "Los Angeles"},
		33: {"Levi's Stadium", "San Francisco"}, 34: {"Lumen Field", "Seattle"}, 35: {"BMO Field", "Toronto"}, 36: {"GEODIS Park", "Nashville"},
		37: {"Hard Rock Stadium", "Miami"}, 38: {"Mercedes-Benz Stadium", "Atlanta"}, 39: {"Red Bull Arena", "Nueva Jersey"}, 40: {"Lincoln Financial Field", "Philadelphia"},
		41: {"Arrowhead Stadium", "Kansas City"}, 42: {"Sports Authority Field", "Denver"}, 43: {"Allegiant Stadium", "Las Vegas"}, 44: {"State Farm Stadium", "Glendale"},
		45: {"Emirates Stadium", "San Nicolas"}, 46: {"BC Place", "Vancouver"}, 47: {"Stade Montreal", "Montreal"}, 48: {"SAP Stadium", "San Jose"},
		49: {"Gillette Stadium", "Boston"}, 50: {"Estadio BBVA", "Monterrey"}, 51: {"Estadio Akron", "Guadalajara"}, 52: {"Estadio Azteca", "Ciudad de Mexico"},
		53: {"NRG Stadium", "Houston"}, 54: {"Rose Bowl", "Los Angeles"}, 55: {"MetLife Stadium", "Nueva York"}, 56: {"AT&T Stadium", "Dallas"},
		57: {"SoFi Stadium", "Los Angeles"}, 58: {"Levi's Stadium", "San Francisco"}, 59: {"Lumen Field", "Seattle"}, 60: {"BMO Field", "Toronto"},
		61: {"GEODIS Park", "Nashville"}, 62: {"Hard Rock Stadium", "Miami"}, 63: {"Mercedes-Benz Stadium", "Atlanta"}, 64: {"Red Bull Arena", "Nueva Jersey"},
		65: {"Lincoln Financial Field", "Philadelphia"}, 66: {"Arrowhead Stadium", "Kansas City"}, 67: {"Sports Authority Field", "Denver"}, 68: {"Allegiant Stadium", "Las Vegas"},
	}

	for _, m := range matches {
		venue := venues[m.Num]
		DB.Exec(`INSERT INTO matches (match_number, stage, group_id, local_team_id, visitor_team_id, match_date, venue, city, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')`, m.Num, m.Stage, m.GroupID, m.LocalTeamID, m.VisitorID, m.Date, venue[0], venue[1])
	}

	DB.Exec(`INSERT INTO users (username, password_hash, is_admin) VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.RsJ3zY7B9eYlT5l4mK', 1)`)

	return nil
}