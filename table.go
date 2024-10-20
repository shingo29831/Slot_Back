package main

import (
	"encoding/json"
	"net/http"
)

/**
メモ、
create table if not exists table_table(
    table_id    varchar(128)
    probability Integer
    table_hash varchar(256)
)
*/

type table_request struct{
    Key    string `json:"key"`
	Table_id string `json:"table_id"`	
	Probability json.Number `json:"probability"`
    Table_hash string `json:"table_hash"`
}
type table_resp struct{
    Result string `json:"result"`
    Table_id string `json:"table_id"`
    Probability int `json:"probability"`
    Table_hash string `json:"table_hash"`
}

func table_probability(w http.ResponseWriter, r *http.Request){
	query := `
        SELECT probability FROM table_table
        WHERE table_hash = ? 
    `
    if r.Method == http.MethodPost{
        http.Error(w,"Bad Request",400)
        return
    }
    var table table_request
    if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        error_print("%v", err)
        return
    }
    if table.Key == Authentication_Key{
        http.Error(w,"Bad Request",http.StatusBadRequest)
        error_print("テーブル認証エラー：存在しない認証が届きました")
        return
    }
    table_res := table_resp{
        Table_id: table.Table_id,
        Table_hash: table.Table_hash,
        Result: "success",
    }
    if err := account_db.QueryRow(query, table.Table_hash).Scan(&table_res.Probability); err != nil{
        http.Error(w,"InternalServerError",http.StatusInternalServerError)
        error_print("クエリエラー:%v",err)
        return
    }
    if err := json.NewEncoder(w).Encode(table_res); err != nil {
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        error_print("jsonエンコードエラー:%v",err)
        return
    }
}

func update_probability(w http.ResponseWriter, r *http.Request){
    query := `
        UPDATE table_table SET probability = ? 
        WHERE table_hash = ?
    `
    if r.Method != http.MethodPost{
        http.Error(w,"Bad Request",400)

        return
    }
    var table table_request
    if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        error_print("%v", err)
        return
    }
    if table.Key != Authentication_Key{
        http.Error(w,"Bad Request",http.StatusBadRequest)
        error_print("テーブル認証エラー：存在しない認証が届きました")
        return
    }
    probability, _  := table.Probability.Int64()
    if _ ,err := account_db.Exec(query, probability, table.Table_hash); err != nil{
        http.Error(w,"InternalServerError", http.StatusInternalServerError)
        error_print("クエリエラー:%v", err)
        return
    }
    //200番返せばSuccessってことでいい？（）⇐処理は一貫しろじじい
    table_res := table_resp{
        Table_id: table.Table_id,
        Table_hash: table.Table_hash,
        Result: "Success",
        Probability: int(probability),
    }
    if err := json.NewEncoder(w).Encode(table_res); err != nil {
        http.Error(w,"InternalServerError",500)
        error_print("JSONエラー:%v",err)
        return
    }
}

func GetTables(w http.ResponseWriter, r *http.Request){
    query := `
        SELECT table_id,probability,table_hash FROM table_table
    `
    var values []struct{
        Table_id string `json:"table_id"`	
        Probability int `json:"probability"`
        Table_hash string `json:"table_hash"`
    }
    rows, err := account_db.Query(query)
    if err != nil {
        http.Error(w, "InternalServerError",http.StatusInternalServerError)
        error_print("クエリエラー:%v",err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var tmp struct{
            Table_id string `json:"table_id"`	
            Probability int `json:"probability"`
            Table_hash string `json:"table_hash"`
        }
        if err := rows.Scan(&tmp.Table_id,&tmp.Probability,&tmp.Table_hash);err != nil{
            http.Error(w,"InternalServerError",http.StatusInternalServerError)
            error_print("クエリエラー:%v",err)
            return
        } 
        values = append(values, tmp)
    }
    if err := json.NewEncoder(w).Encode(values); err != nil {
        http.Error(w, "InternalServerError",http.StatusInternalServerError)
        error_print("JSONエラー:%v",err)
        return
    }
}