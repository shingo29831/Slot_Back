package main

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
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
var ACCOUNT_TABLE = "account_system:xM7B)NY-eexsJm@tcp(localhost:3306)/account_server"

type user_auth struct {
	Key 	 string `json:"key"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Table 	 string `json:"table"`
	Money 	 int	`json:"money"`
}

type user_result struct {
	Result 	 string `json:"result"`
	Message  string `json:"message"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Table 	 string `json:"table"`
	Money 	 int	`json:"money"`
}

func user_(db *sql.DB, user user_auth)(bool, error){
	if user.Key != Authentication_Key {
		return false, fmt.Errorf("テーブル認証に失敗しました")
	}
    defer db.Close()
    query := "SELECT COUNT(*) FROM Account_table WHERE TOKEN = ?"
    var count int
    err := db.QueryRow(query, user.Token).Scan(&count)
    if err != nil {
		return false, fmt.Errorf("クリエエラー")
    }
	if count == 1{
		return true, nil
	} else if count == 0{
		return false, nil
	}else {
		return false, fmt.Errorf("認証エラー")
	}
}

func Message(result, message string,user *user_auth)user_result{
	fmt.Printf("%s: %s\n",result,message)
	if user == nil{
		return user_result{
			Result: result,
			Message:message,
			Username: "",
			Password: "",
			Token: "",
			Table: "",
			Money: 0,
		}
	}
	return user_result{
		Result: result,
		Message:message,
		Username: user.Username,
		Password: user.Password,
		Token: user.Token,
		Table: user.Table,
		Money: user.Money,
	}
}

func Error_res(message string, user *user_auth, w http.ResponseWriter){
	resp, err:= json.Marshal(Message("Error", message, user))
	if err != nil{
		log.Fatal(err)
	}
	w.WriteHeader(200)
	w.Write(resp)
	error_print(message)
}

//アカウント登録（サーバー方面）　完成
//とりあえずだけど、Webから情報を受け取って、アカウントを作成するだけ
//まだ、ログイン、ログアウトはできてないし、ゲスト用も作れてないけど、それは後々

//追記、Dos等への対策が一切ないため、何かはしなければならない

func create_User_Handle(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	var create_js user_auth
	err := data2json(r, &create_js)
	if err != nil{
		http.Error(w, "jsonの形が異なるか、送信されていません",200)
		error_print("jsonえらー")
		return
	}
	db, err := NewDatabase(ACCOUNT_TABLE)
	if err != nil {
		Error_res("データベース接続でエラーが発生しました",nil,w)
		error_print("DBエラー")
		return
	}
	//USERにprimary keyを指定しているので、エラーが起きたら、すでにその名前があると認識させ、もう一度と返します
	//ここバグの温床になる気がするで、元気があったら改善する
	
	_ ,err = db.Exec("Insert into Account_table(username, usertype,password,money,TOKEN) values (?,1,?,0, NULL)",create_js.Username,create_js.Password)
	if err != nil{
		resp, err := json.Marshal(Message("username_already_exists","ユーザーネームが既に存在しています",&create_js))
		if err != nil {
			log.Fatal(err)
		}
		error_print("create_user: %s", err.Error())
		w.WriteHeader(200)
		w.Write(resp)
	} else {
		resp, err := json.Marshal(Message("success","ユーザー登録が完了しました♡",&create_js))
		if err != nil {
			log.Fatal(err)
		}
		log_print("Usercreate: %s", create_js.Username)
		w.WriteHeader(200)
		w.Write(resp)
	}
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
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	//ゲストログインするのがうちかを認証する
	var ca_js user_auth
	err := data2json(r, &ca_js)
	if err != nil{
		Error_res(err.Error(),nil,w)
		return
	}
	if ca_js.Key != Authentication_Key {
		Error_res("テーブル認証に失敗しました",&ca_js,w)
		return
	}
	var answer user_result
	answer.Username = MakeRandomStr(32)
	answer.Password = MakeRandomStr(32)
	answer.Table = ca_js.Table
	db, err := NewDatabase(ACCOUNT_TABLE)
	if err != nil {
		Error_res("データベース接続に失敗しました",&ca_js, w)
		return
	}
	answer.Token = MakeRandomStr(128)
	_ ,err = db.Exec("Insert into Account_table(username, usertype,password,money,table_id,TOKEN) values (?,2,?,0,?,?)",answer.Username,fmt.Sprintf("%x", md5.Sum([]byte(answer.Password))),answer.Table,answer.Token)
	if err != nil{
		fmt.Println(err)
		resp, err := json.Marshal(Message("username_already_exists","ユーザーが既に存在しています",&ca_js))
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(200)
		w.Write(resp)
		return
	}
	answer.Result = "success"
	answer.Message = "ゲストアカウントの作成に成功しました"
	answer.Money = 0
	log_print("ゲストログイン ID:%s, TABLE:%s",answer.Username,answer.Table)
	data, err:= json.Marshal(answer)
	if err != nil{
		log.Fatal(err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}


//ログイン　ログインをしたらトークンを生成して、やり取りができるようになる
func User_Login(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	var login_user user_auth
	if err := data2json(r, &login_user); err != nil{
		Error_res(err.Error(),nil,w)
		return
	}
	if login_user.Key != Authentication_Key {
		Error_res("データベース接続に失敗しました",&login_user, w)
		return
	}
	password := fmt.Sprintf("%x", md5.Sum([]byte(login_user.Password)))
	db, err :=  NewDatabase(ACCOUNT_TABLE)
    if err != nil {
		Error_res(err.Error(), &login_user,w)
    }
    defer db.Close()
    query := "SELECT COUNT(*) FROM Account_table WHERE username = ? AND password = ?"
    var count int
    err = db.QueryRow(query, login_user.Username, password).Scan(&count)
    if err != nil {
		Error_res(err.Error(),&login_user, w)
		return
    }
	if count == 1 {
		var ans user_result
		ans.Username = login_user.Username
		ans.Password = login_user.Password
		ans.Table = login_user.Table
		ans.Token = MakeRandomStr(128)
		ans.Result = "success"
		_ ,err := db.Exec("UPDATE Account_table SET TOKEN = ?,table_id = ? WHERE username = ? AND password = ?",
				 ans.Token, ans.Table,ans.Username, password)
		if err != nil {
			Error_res(err.Error(), &login_user, w)
			return
		}
		
		data,err := json.Marshal(ans)
		if err != nil{
			Error_res(err.Error(), &login_user, w)
			return
		}
		log_print("ユーザーログイン ID:%s, TABLE:%s", ans.Username, ans.Table)
		w.WriteHeader(200)
		w.Write(data)
	} else if count == 0{
		resp , err := json.Marshal(Message("Account does not exist","アカウント認証に失敗しました",&login_user))
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(200)
		w.Write(resp)
	}else {
		error_print("認証エラー ID:%s TABLE:%s",login_user.Username, login_user.Table)
		resp , err := json.Marshal(Message("Error","アカウントエラーが発生しました。管理者にお問い合わせください",&login_user))
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(200)
		w.Write(resp)
	}
}


//ログアウト用、トークンを破棄する
func User_Logout(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	var user user_auth
	if err := data2json(r, &user); err != nil{
		Error_res(err.Error(),nil,w)
		return
	}
	db, err := NewDatabase(ACCOUNT_TABLE)
	if err != nil {
		Error_res(err.Error(), nil, w)
		return
	}

	if tmp, err := user_(db, user); tmp || err == nil {
		password := fmt.Sprintf("%x", md5.Sum([]byte(user.Password)))
		query := "UPDATE Account_table SET token = NULL,table_id = NULL WHERE TOKEN = ?"
		_, err := db.Exec(query, user.Username, password, user.Token)
		if err != nil {
			Error_res(fmt.Sprintf("クエリエラー:%s",err.Error()),&user,w)
			return
		}
		resp, err := json.Marshal(Message("success","ログアウトに成功しました",&user))
		if err != nil {
			log.Fatal(err)
		}
		log_print("ユーザーログアウト ID:%s, TABLE:%s", user.Username, user.Table)
		w.WriteHeader(200)
		w.Write(resp)
	}else{
		Error_res("認証エラーが発生しました、管理者に教えてください",&user,w)
	}
}

func get_user_money(user *user_auth)(user_result){
	db, err := NewDatabase(ACCOUNT_TABLE)
	if err != nil{
		return Message("Error",err.Error(),nil)
	}
	password := fmt.Sprintf("%x", md5.Sum([]byte(user.Password)))
    query := "SELECT COUNT(*) FROM Account_table WHERE username = ? AND password = ? AND TOKEN = ?"
    var count int
    err = db.QueryRow(query, user.Username, password, user.Token).Scan(&count)
	if err != nil {
		return Message("Error", "クリエでエラーが発生しました",user)
	}
	if count == 1{
		var money int
		query := "select money from Account_table WHERE username = ? AND password = ? AND TOKEN = ?"
		err = db.QueryRow(query, user.Username, password, user.Token).Scan(&money)
		if err != nil {
			return Message("Error", "クリエでエラーが発生しました",user)
		}
		user.Money = money
	}else{
		return Message("Error", "認証でエラーが発生しました",user)
	}
	return Message("success", "完了",user)
}

func update_user_money(user * user_auth)(user_result){
	db, err := NewDatabase(ACCOUNT_TABLE)
	if err != nil{
		return Message("Error",err.Error(),nil)
	}
	password := fmt.Sprintf("%x", md5.Sum([]byte(user.Password)))
    query := "SELECT COUNT(*) FROM Account_table WHERE username = ? AND password = ? AND TOKEN = ?"
    var count int
    err = db.QueryRow(query, user.Username, password, user.Token).Scan(&count)
	if err != nil {
		return Message("Error", "クリエでエラーが発生しました",user)
	}
	if count == 1{
		query := "update Account_table set money = ?  WHERE username = ? AND password = ? AND TOKEN = ?"
		_ ,err = db.Exec(query, user.Money ,user.Username, password, user.Token)
		if err != nil {
			return Message("Error", "クリエでエラーが発生しました",user)
		}
	}else{
		return Message("Error", "認証でエラーが発生しました",user)
	}
	return Message("success", "完了",user)
}

func GET_SET_MONEY(w http.ResponseWriter, r *http.Request, userres func(*user_auth)(user_result)){
	if r.Method != "POST"{
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	var user user_auth
	if err := data2json(r, &user); err != nil{
		Error_res(err.Error(),nil,w)
		return
	}
	if user.Key != Authentication_Key {
		Error_res("データベース接続に失敗しました",&user, w)
		return
	}
	resp, err := json.Marshal(userres(&user))
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(200)
	w.Write(resp)
}


//ユーザーの現在金額取得用のプログラム
func GET_USER_MONEY(w http.ResponseWriter, r *http.Request){
	GET_SET_MONEY(w,r,get_user_money)
}

//ユーザーの現在金額更新用のプログラム
func UPDATE_USER_MONEY(w http.ResponseWriter, r *http.Request){
	GET_SET_MONEY(w,r,update_user_money)
}