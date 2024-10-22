package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func total_Query()([]byte, error){
	//間違っているかもしれないし、しないかもしれない　
	query := `
SELECT session_id,
    SUM(CASE WHEN type = 0 THEN fluctuation ELSE 0 END) AS seles,
    MAX(CASE WHEN type = 1 THEN money END) AS last_amount 
FROM slot_result_table t1
WHERE t1.type = 0 OR (t1.type = 1 AND t1.time = (
    SELECT MAX(t2.time) FROM slot_result_table t2
    WHERE t2.type = 1 AND t1.session_id = t2.session_id
))
GROUP BY session_id 
ORDER BY session_id

	`
	rows, err := account_db.Query(query)
	if err != nil{
		return nil ,fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}
	results := make([]map[string]interface{},0)
	for rows.Next(){
		values := make([]interface{}, len(columns))
		valuePointers := make([]interface{}, len(columns))
		for i := range values {
			valuePointers[i] = &values[i]
		}
		if err := rows.Scan(valuePointers...);err != nil{
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		rowMap := make(map[string]interface{})
	
		for i, col := range columns {
			if values[i] != nil {
				// []uint8 の場合は string に変換
				if byteValue, ok := values[i].([]uint8); ok {
					rowMap[col] = string(byteValue)
				} else {
					rowMap[col] = values[i] // それ以外はそのまま
				}
			} else {
				rowMap[col] = nil
			}
		}
		results = append(results, rowMap)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	return jsonData, nil
}

func totals(w http.ResponseWriter, r *http.Request){
	//認証	しろ
	session, _ := store.Get(r, "auth-session")

	// 認証されていない場合、ログインページにリダイレクト
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	buf, err := total_Query()
	log_print("%s",string(buf))
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		error_print("クリエエラー%v",err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}