package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

/**
*   アカウント作成手順
*   1. 情報受信
*   2. ユーザーネームが誰かと被っていないかを確かめる
*   3. 被っていた場合、エラー、再送要求送信
*   4. 被っていない場合、DBに情報登録、通ったことを送信
*
*   フロントエンドではJSを利用して作成するつもりです
 */

//外部接続潰し、正しい端末以外からの通信を破棄する
var Authentication_Key = "aaa"
var ACCOUNT_TABLE = "account_system:acpassword@(%)/account_server"

type USER_JSON struct {
	Key 	 string `json:"key"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}


//アカウント登録（サーバー方面）　完成
//とりあえずだけど、Webから情報を受け取って、アカウントを作成するだけ
//まだ、ログイン、ログアウトはできてないし、ゲスト用も作れてないけど、それは後々

//追記、Dos等への対策が一切ないため、何かはしなければならない


func Create_User_Handle(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/Create_User", http.StatusSeeOther)
		return
	}
	var create_js struct{
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := data2json(r, &create_js)
	if err != nil{
		http.Error(w, "jsonの形が異なるか、送信されていません",400)
		return
	}
	db, err := NewDatabase(ACCOUNT_TABLE)
	if err != nil {
		http.Error(w, "{\"result\":\"Error\",\"message\":\"データベース接続でエラーが発生しました\"}", 500)
		return
	}
	//USERにprimary keyを指定しているので、エラーが起きたら、すでにその名前があると認識させ、もう一度と返します
	//ここバグの温床になる気がするで、元気があったら改善する
	
	_ ,err = db.Exec("Insert into Account_table(username, usertype,password,money,TOKEN) values (?,1,?,0, NULL)",create_js.Username,create_js.Password)
	if err != nil{
		w.WriteHeader(200)
		w.Write([]byte("{\"result\":\"username_already_exists\",\"message\":\"ユーザーネームがダブっているので変更してください\"}"))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("{\"result\":\"succsess\"}"))
}
func MakeRandomStr(digit uint32) (string) {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    // 乱数を生成
    b := make([]byte, digit)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    for i, v := range b {
        b[i] = letters[int(v)%len(letters)]
    }
    return string(b)
}
//ゲストアカウント作成、ゲストアカウント作成が作成されたら、同時にトークン生成もして返送する
func Create_guest_user(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/Create_User", http.StatusSeeOther)
		return
	}
	//ゲストログインするのがうちかを認証する
	var ca_js struct{
		Key string `json:"key"`
	}
	err := data2json(r, &ca_js)
	if err != nil{
		http.Error(w, "jsonの形が異なるか、送信されていません",400)
		return
	}
	if ca_js.Key != Authentication_Key {
		http.Error(w,"誰だ貴様",400)
		fmt.Printf("多分攻撃うけてる")
		return
	}
	var answer struct{
		Result string 	`json:"result"`
		Username string	`json:"username"`
		Password string `json:"password"`
		Token string	`json:"token"`
	}
	
	answer.Username = string(append_byte([]byte(time.Now().GoString()), []byte(MakeRandomStr(10))))
	answer.Password = MakeRandomStr(255)
	db, err := NewDatabase(ACCOUNT_TABLE)
	if err != nil {
		http.Error(w, fmt.Sprintf("データベースに接続できませんでした:%s",err.Error()), 500)
		return
	}
	answer.Token = MakeRandomStr(255)
	_ ,err = db.Exec("Insert into Account_table(username, usertype,password,money,TOKEN) values (?,2,?,0,?)",answer.Username,answer.Password,answer.Token)
	if err != nil{
		w.WriteHeader(200)
		w.Write([]byte("{\"result\":\"username_already_exists\"}"))
		return
	}
	answer.Result = "succsess"
	data,err := json.Marshal(answer)
	if err != nil{
		http.Error(w, fmt.Sprintf("不明なエラー:%s", err.Error()),500)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

func User_Login(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/Create_User", http.StatusSeeOther)
		return
	}
	var login_user struct{
		Key      string `json:"key"`
		Username string `json:"usernamr"`
		Password string `json:"password"`
	}
	if err := data2json(r, &login_user); err != nil{
		http.Error(w, err.Error(), 400)
	}
	if login_user.Key != Authentication_Key {
		http.Error(w,"誰だ貴様",400)
		fmt.Printf("多分攻撃うけてる")
		return
	}
	password := fmt.Sprintf("%x", md5.Sum([]byte(login_user.Password)))
	db, err :=  NewDatabase(ACCOUNT_TABLE)
    if err != nil {
		http.Error(w,fmt.Sprintf("dbが開けませんでした:%s",err.Error()),500)
    }
    defer db.Close()
    query := "SELECT COUNT(*) FROM Account_table WHERE username = ? AND password = ?"
    var count int
    err = db.QueryRow(query, login_user.Username, password).Scan(&count)
    if err != nil {
		http.Error(w, fmt.Sprintf("クエリエラー:%s",err.Error()),500)
		return
    }
	if count == 1 {
		var ans struct{
			Result string `json:"result"`
			Username string `json:"username"`
			Password string `json:"password"`
			Token string `json:"token"`
		}
		ans.Username = login_user.Username
		ans.Password = login_user.Password
		ans.Token = MakeRandomStr(128)
		ans.Result = "succsess"
		_, err = db.Exec("UPDATE Account_table SET TOKEN = ? WHERE username = ? AND password = ?", ans.Token, ans.Username, ans.Password)
		if err != nil {
			http.Error(w, fmt.Sprintf("クエリエラー:%s",err.Error()), 500)
			return
		}
		data,err := json.Marshal(ans)
		if err != nil{
			http.Error(w, fmt.Sprintf("不明なエラー:%s", err.Error()),500)
			return
		}
		w.WriteHeader(200)
		w.Write(data)
	} else if count == 0{
		w.WriteHeader(200)
		w.Write([]byte("{\"result\":\"Account does not exist\"}"))
	}else {
		w.WriteHeader(200)
		w.Write([]byte("{\"result\":\"Account_ERROR\",\"message\":\"管理者に問い合わせてください\"}"))
	}
}

func User_Logout(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/Create_User", http.StatusSeeOther)
		return
	}
	var user  USER_JSON
	err := data2json(r, &user)
	if err != nil{
		http.Error(w, "jsonの形が異なるか、送信されていません",400)
		return
	}
	if user.Key != Authentication_Key {
		http.Error(w, "エラー",400)
		return
	}
	password := fmt.Sprintf("%x", md5.Sum([]byte(user.Password)))
	db, err :=  NewDatabase(ACCOUNT_TABLE)
    if err != nil {
		http.Error(w,fmt.Sprintf("dbが開けませんでした:%s",err.Error()),500)
    }
    defer db.Close()
    query := "SELECT COUNT(*) FROM Account_table WHERE username = ? AND password = ? AND TOKEN = ?"
    var count int
    err = db.QueryRow(query, user.Username, password, user.Token).Scan(&count)
    if err != nil {
		http.Error(w, fmt.Sprintf("クエリエラー:%s",err.Error()),500)
		return
    }
	if count > 0 {
		query = "select usertype from Account_table where username = ? AND password = ? AND TOKEN = ?"
		err = db.QueryRow(query, user.Username, password, user.Token).Scan(&count)
		if err != nil{
			http.Error(w, "Error",500)
			return
		}
		if count == 2{
			query = "delete from Account_table WHERE username = ? AND password = ? AND TOKEN = ?"
		}else {
			query = "UPDATE Account_table SET token = NULL  WHERE username = ? AND password = ? AND TOKEN = ?"
		}
		_, err = db.Exec(query, user.Username, password, user.Token)
		if err != nil {
			http.Error(w, fmt.Sprintf("クエリエラー:%s",err.Error()), 500)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("{\"result\":\"succsess\",\"message\":\"ログアウトしました\"}"))
	}else{
		w.WriteHeader(200)
		w.Write([]byte("{\"result\":\"Account_ERROR\",\"message\":\"管理者に問い合わせてください\"}"))
	}
}