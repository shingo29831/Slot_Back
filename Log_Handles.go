package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

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

func Log_ALL_recive(w http.ResponseWriter, r *http.Request){
    //POSTじゃないなら弾く
	if r.Method != "POST" {
		Error_serve(403, "権限がありません\n",w, r)	
		return
	}
	var logs  send_logs_file
    //ボディ読み取り,JSONに変換
	err := data2json(r, &logs)
	if err != nil {
		Error_serve(403, "こっちの定義通り送ってくれ\n",w,r)
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
	w.WriteHeader(200)
	w.Write([]byte("ALL DONE"))
	wg.Wait()
}

//一件のログ用のハンドル
//上とほぼほぼ同じ内容(メソッドにまとめたほうがいいかな？)
func Log_recive(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != "POST" {
		Error_serve(403, "権限がありません\n",w, r)	
		return
	}
	logs := send_logs{}
	err := data2json(r, &logs)
	if err != nil {
		Error_serve(403, "こっちの定義通り送ってくれ\n",w,r)
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
	w.WriteHeader(200)
	w.Write([]byte("ALL DONE"))
	wg.Wait()
}