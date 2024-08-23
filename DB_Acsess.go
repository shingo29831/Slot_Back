package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

// Database構造体を定義
type Database struct {
    DB *sql.DB
}

// Databaseの初期化メソッド（コンストラクタ風）
func NewDatabase(dsn string) (*Database, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    // 接続テスト
    err = db.Ping()
    if err != nil {
        return nil, err
    }

    return &Database{DB: db}, nil
}

// Database構造体に属するクエリメソッド
func (db *Database) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
    rows, err := db.DB.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    columns, err := rows.Columns()
    if err != nil {
        return nil, err
    }

    results := []map[string]interface{}{}
    values := make([]interface{}, len(columns))
    scanArgs := make([]interface{}, len(columns))

    for i := range values {
        scanArgs[i] = &values[i]
    }

    for rows.Next() {
        err = rows.Scan(scanArgs...)
        if err != nil {
            return nil, err
        }

        row := make(map[string]interface{})
        for i, col := range columns {
            var v interface{}
            switch values[i].(type) {
            case []byte:
                v = string(values[i].([]byte))
            default:
                v = values[i]
            }
            row[col] = v
        }
        results = append(results, row)
    }

    return results, nil
}

// Database構造体に属するExecメソッド
func (db *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
    result, err := db.DB.Exec(query, args...)
    if err != nil {
        return nil, err
    }
    return result, nil
}

// Database構造体に属するCloseメソッド
func (db *Database) Close() error {
    return db.DB.Close()
}

