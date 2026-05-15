import sqlite3
import pandas as pd
import os

# Ruta a la base de datos SQLite
DB_PATH = r'C:\Datos\Categorias\Desarrollo\quinela-mundial2026\build\bin\data\quinela.db'
# Ruta de salida del archivo Excel
XLSX_PATH = r'C:\Datos\Categorias\Desarrollo\quinela-mundial2026\build\bin\data\quinela_export.xlsx'

def export_sqlite_to_excel(db_path, xlsx_path):
    # Conectar a la base de datos
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # Obtener todas las tablas
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table';")
    tables = [row[0] for row in cursor.fetchall()]

    # Crear un archivo Excel con cada tabla en una hoja
    with pd.ExcelWriter(xlsx_path, engine='openpyxl') as writer:
        for table in tables:
            df = pd.read_sql_query(f'SELECT * FROM {table}', conn)
            df.to_excel(writer, sheet_name=table, index=False)
    conn.close()
    print(f"Exportación completada: {xlsx_path}")

if __name__ == "__main__":
    export_sqlite_to_excel(DB_PATH, XLSX_PATH)
