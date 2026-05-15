package main

import (
	"database/sql"
	"fmt"
	"log"

	excelize "github.com/xuri/excelize/v2"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "build/bin/data/quinela.db"
	xlsxPath := "build/bin/data/quinela_export.xlsx"

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error abriendo la base de datos: %v", err)
	}
	defer db.Close()

	// Obtener todas las tablas
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';")
	if err != nil {
		log.Fatalf("Error obteniendo tablas: %v", err)
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			log.Fatalf("Error leyendo nombre de tabla: %v", err)
		}
		tables = append(tables, table)
	}

	f := excelize.NewFile()
	firstSheet := true

	for _, table := range tables {
		dataRows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			log.Printf("Error leyendo datos de %s: %v", table, err)
			continue
		}
		cols, _ := dataRows.Columns()
		values := make([]interface{}, len(cols))
		scanArgs := make([]interface{}, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		sheetName := table
		if firstSheet {
			f.SetSheetName("Sheet1", sheetName)
			firstSheet = false
		} else {
			f.NewSheet(sheetName)
		}

		// Escribir encabezados
		for i, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, col)
		}

		rowIdx := 2
		for dataRows.Next() {
			if err := dataRows.Scan(scanArgs...); err != nil {
				log.Printf("Error escaneando fila en %s: %v", table, err)
				continue
			}
			for i, val := range values {
				cell, _ := excelize.CoordinatesToCellName(i+1, rowIdx)
				f.SetCellValue(sheetName, cell, val)
			}
			rowIdx++
		}
		dataRows.Close()
	}

	if err := f.SaveAs(xlsxPath); err != nil {
		log.Fatalf("Error guardando archivo Excel: %v", err)
	}
	fmt.Printf("Exportación completada: %s\n", xlsxPath)
}
