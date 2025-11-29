package util

import (
	"database/sql"
)

// RowsToMaps consumes sql.Rows and returns column names plus JSON-friendly row maps.
func RowsToMaps(rows *sql.Rows) ([]string, []map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	data := make([]map[string]any, 0)
	for rows.Next() {
		// values represents the actual values
		values := make([]any, len(columns))
		// array of ptrs used so rows.Scan can be utilized
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, nil, err
		}

		rowMap := make(map[string]any, len(columns))
		for i, col := range columns {
			switch v := values[i].(type) {
			case []byte:
				rowMap[col] = string(v)
			default:
				rowMap[col] = v
			}
		}
		data = append(data, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return columns, data, nil
}
