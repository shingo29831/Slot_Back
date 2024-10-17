package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"encoding/json"
    _ "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

//ここからログ鯖関連
//Log一つに対しての構造体
//Levelがことの重大度
//locationがその台
//messageが本文
type send_logs struct {
	Level   int    `json:"level"`
	Location string`json:"location"`	
	Message string `json:"message"`
}
//送るJSON本体
type send_logs_file struct{
	Logs []send_logs `json:"logs"`
}

type Log struct {
    Time     string `json:"time"`
    Level    int    `json:"level"`
    Location string `json:"location"`
    Message  string `json:"message"`
}

//Logの重大度をstringに変換するメゾット
//func level_int2str(lv int)string{
//	switch lv {
//		case 0:return "succsess"
//		case 1:return "note"
//		case 2:return "warning"
//		case 3:return "error"
//		default :return "????"
//	}
//}


//複数ログ用ハンドル,主にこいつを利用してほしい

func Log_accsess(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var filterCondition struct {
		Location  string    `json:"location"`  // Locationを追加
		Level     *int      `json:"level"`
		StartTime time.Time `json:"startTime"`
		EndTime   time.Time `json:"endTime"`
	}
	if err := json.NewDecoder(r.Body).Decode(&filterCondition); err != nil{
		http.Error(w, "Bad request", 400)
		fmt.Printf("%v\n",err)
		return
	}
	query := `
		SELECT time, level, location, message FROM Log_table WHERE 1=1 
	`
	// Locationのフィルタリング
	if filterCondition.Location != ""{
		query += fmt.Sprintf(" AND location = '%s'", filterCondition.Location)
	}
	// 重要度のフィルタリング
	if filterCondition.Level != nil  {
		query += fmt.Sprintf(" AND level = %d", filterCondition.Level)
	}
	// 開始日と終了日のフィルタリング
	if !filterCondition.StartTime.IsZero() {
		query += fmt.Sprintf(" AND time >= '%s'", filterCondition.StartTime.String())
	}
	if !filterCondition.EndTime.IsZero()  {
		query += fmt.Sprintf(" AND time <= '%s'",filterCondition.EndTime.String())
	}
	query += " ORDER BY time DESC LIMIT 100"
	db, err := NewDatabase("logsystem:logsyspassword@tcp(localhost:3306)/log_server")
	if err != nil {
		http.Error(w,"サーバーエラー",500)
		return
	}
	rows, err := db.Query(query)
    if err != nil {
        http.Error(w, "データベースからのデータ取得に失敗しました : "+ err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var logs []Log

    // クエリの結果を読み込んでログ構造体にマッピング
    for rows.Next() {
        var logEntry Log
        var logTime string

        if err := rows.Scan(&logTime, &logEntry.Level, &logEntry.Location, &logEntry.Message); err != nil {
            http.Error(w, "データの読み込みに失敗しました"+ err.Error(), http.StatusInternalServerError)
            return
        }
		parsedTime, err := time.Parse("2006-01-02 15:04:05", logTime)
		if err != nil {
			http.Error(w, "時間のパースに失敗しました", http.StatusInternalServerError)
    		return
        }
        // 時刻をRFC3339形式にフォーマット
        logEntry.Time = parsedTime.Format(time.RFC3339)
        logs = append(logs, logEntry)
    }

    // JSONとして返す
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(logs); err != nil {
        http.Error(w, "JSONエンコードに失敗しました", http.StatusInternalServerError)
    }
}

func Log_ALL_recive(w http.ResponseWriter, r *http.Request){
    //POSTじゃないなら弾く
	if r.Method != "POST" {
		http.Error(w, "権限がありません\n",http.StatusForbidden)	
		return
	}
	var logs  send_logs_file
    //ボディ読み取り,JSONに変換
	if err := json.NewDecoder(r.Body).Decode(&logs);err != nil {
		http.Error(w, "こっちの定義通り送ってくれ\n",400)
		return
	}
    //ここ好きポイント
    //goを利用して、DB書き込みと、レスポンス送信を並行処理で行う
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		db, err:= NewDatabase(/*仮*/"logsystem:logsyspassword@tcp(localhost:3306)/log_server")
		if err != nil{
			fmt.Printf("Error:%s\n",err)
			wg.Done()
			return
		}
		//エラー処理を書かなきゃいけない　あとでやる
		defer db.Close()
		for _, v := range logs.Logs {
			_, err = db.Exec("insert into Log_table (time,level,location,message) values (?,?,?,?)",time.Now().Format("2006-01-02T15:04:05Z07:00"),v.Level,v.Location,v.Message)
			if err != nil {
				log.Fatal(err)
			}
		}
		wg.Done()
	}()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ALL DONE"))
	wg.Wait()
}

//一件のログ用のハンドル
//上とほぼほぼ同じ内容(メソッドにまとめたほうがいいかな？)
func Log_recive(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != "POST" {
		http.Error(w, "権限がありません\n",http.StatusForbidden)	
		return
	}
	logs := send_logs{}
	if err := json.NewDecoder(r.Body).Decode(&logs);err != nil {
		http.Error(w, "こっちの定義通り送ってくれ\n",http.StatusForbidden)
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func ()  {
		
		db, err:= NewDatabase("logsystem:logsyspassword@tcp(localhost:3306)/log_server")
		if err != nil{
			fmt.Printf("Error:%s\n",err)
			wg.Done()
			return
		}
		//エラー処理を書かなきゃいけない　あとでやる
		defer db.Close()
		_, err = db.Exec("insert into Log_table (time,level,location,message) values (?,?,?,?)",time.Now().Format("2006-01-02T15:04:05Z07:00"),logs.Level,logs.Location,logs.Message)
		if err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ALL DONE"))
	wg.Wait()
}

func log_print(message string, arg... any){
	server_log(0,message,arg...)
}

func error_print(message string, arg... any){
	server_log(2,message,arg...)
}

func server_log(level int, message string, arg... any){
	var log1 Log
	log1.Message = fmt.Sprintf(message, arg...)
	log1.Level = level
	log1.Time = time.Now().Format("2006-01-02T15:04:05Z07:00")
	log1.Location = "server"
	db, err:= NewDatabase("logsystem:logsyspassword@tcp(localhost:3306)/log_server")
	if err != nil{
		fmt.Printf("Error:%s\n",err)
		return
	}
	//エラー処理を書かなきゃいけない　あとでやる
	defer db.Close()
	_, err = db.Exec("insert into Log_table (time,level,location,message) values (?,?,?,?)",log1.Time,log1.Level,log1.Location,log1.Message)
	if err != nil {
		log.Fatal(err)	
	}

}