package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

var account_db *sql.DB
var err error
var Authentication_Key = "aaa"
var ACCOUNT_TABLE = os.Getenv("ACCOUNT_SERVER")

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
func init_account_db(){
	account_db, err = NewDatabase(ACCOUNT_TABLE)
	if err != nil {
		log.Fatal(err)
	}
}

// Authentication check method
func checkAuthenticationKey(user user_auth) error {
    if user.Key != Authentication_Key {
        return fmt.Errorf("認証キーが無効です")
    }
    return nil
}

func createToken(user user_auth)(string ,error){
	token := MakeRandomStr(128)
	query := `
		Insert Into session_tokens (time, TOKEN,table_id, username) values (?,?,?,?)
	`
	_, err := account_db.Exec(query, time.Now().Format("2006-01-02 15:04:05"), token, user.Table,user.Username)
	if err != nil {
		error_print("クリエエラー%v",err)
		return "", err
	}
	return token, nil
}

// Query user token in DB method
func checkUserToken(user user_auth) (bool, error) {
    query := `
		SELECT COUNT(*) FROM Account_table WHERE TOKEN = ? AND username = ? AND password = ?
	`    
	var count int
    err := account_db.QueryRow(query, user.Token, user.Username,
				fmt.Sprintf("%x", sha256.Sum256([]byte(user.Password)))).Scan(&count)
    if err != nil {
        return false, fmt.Errorf("クエリエラー")
    }
    return count == 1, nil
}

// Create a unified message response
func Message(result, message string, user *user_auth) user_result {
    fmt.Printf("%s: %s\n", result, message)
    if user == nil {
        return user_result{
            Result:   result,
            Message:  message,
            Username: "",
            Password: "",
            Token:    "",
            Table:    "",
            Money:    0,
        }
    }
    return user_result{
        Result:   result,
        Message:  message,
        Username: user.Username,
        Password: user.Password,
        Token:    user.Token,
        Table:    user.Table,
        Money:    user.Money,
    }
}

// Unified error response
func ErrorResponse(message string, user *user_auth, w http.ResponseWriter) {
    resp, err := json.Marshal(Message("Error", message, user))
    if err != nil {
        log.Fatal(err)
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    w.Write(resp)

	error_print(message)
}

// User authentication process refactored
func userAuthentication(user user_auth) (bool, error) {
    if err := checkAuthenticationKey(user); err != nil {
        return false, err
    }

    // Check if token is valid
    return checkUserToken(user)
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
	err := json.NewDecoder(r.Body).Decode(&create_js)
	if err != nil{
		http.Error(w, "Bad Request",http.StatusBadRequest)
		error_print("jsonえらー")
		return
	}
	hashPassword := fmt.Sprintf("%x",sha256.Sum256([]byte(create_js.Password)))
	//USERにprimary keyを指定しているので、エラーが起きたら、すでにその名前があると認識させ、もう一度と返します
	//ここバグの温床になる気がするで、元気があったら改善する
	
	_ ,err = account_db.Exec("Insert into Account_table(username, usertype,password,money,TOKEN) values (?,1,?,0, NULL)",create_js.Username,hashPassword)
	if err != nil{
		resp, err := json.Marshal(Message("username_already_exists","ユーザーネームが既に存在しています",&create_js))
		if err != nil {
			log.Fatal(err)
		}
		error_print("create_user: %s", err.Error())
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	} else {
		resp, err := json.Marshal(Message("success","ユーザー登録が完了しました♡",&create_js))
		if err != nil {
			log.Fatal(err)
		}
		log_print("Usercreate: %s", create_js.Username)
		w.WriteHeader(http.StatusOK)
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
	err := json.NewDecoder(r.Body).Decode(&ca_js)
	if err != nil{
		ErrorResponse(err.Error(),nil,w)
		return
	}
	if ca_js.Key != Authentication_Key {
		ErrorResponse("テーブル認証に失敗しました",&ca_js,w)
		return
	}
	
	var answer user_result
	answer.Username = MakeRandomStr(32)
	answer.Password = MakeRandomStr(32)
	answer.Table = ca_js.Table

	ca_js.Username = answer.Username
	ca_js.Password = answer.Password
	answer.Token,_ = createToken(ca_js)
	_ ,err = account_db.Exec("Insert into Account_table(username, usertype,password,money,table_id,TOKEN) values (?,2,?,0,?,?)",answer.Username,fmt.Sprintf("%x", sha256.Sum256([]byte(answer.Password))),answer.Table,answer.Token)
	if err != nil{
		fmt.Println(err)
		resp, err := json.Marshal(Message("username_already_exists","ユーザーが既に存在しています",&ca_js))
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}


//ログイン　ログインをしたらトークンを生成して、やり取りができるようになる
func User_Login(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	var login_user user_auth
	if err = json.NewDecoder(r.Body).Decode(&login_user); err != nil{
		ErrorResponse(err.Error(),nil,w)
		return
	}
	if err = checkAuthenticationKey(login_user); err != nil {
		ErrorResponse("データベース接続に失敗しました",&login_user, w)
		return
	}
	password := fmt.Sprintf("%x", sha256.Sum256([]byte(login_user.Password)))
	query := "SELECT COUNT(*) FROM Account_table WHERE username = ? AND password = ?"
    var count int
    err = account_db.QueryRow(query, login_user.Username, password).Scan(&count)
    if err != nil {
		ErrorResponse(err.Error(),&login_user, w)
		return
    }
	if count == 1 {
		var ans user_result
		ans.Username = login_user.Username
		ans.Password = login_user.Password
		ans.Table = login_user.Table
		ans.Token, _ = createToken(login_user)
		ans.Result = "success"
		_ ,err := account_db.Exec("UPDATE Account_table SET TOKEN = ?,table_id = ? WHERE username = ? AND password = ?",
				 ans.Token, ans.Table,ans.Username, password)
		if err != nil {
			ErrorResponse(err.Error(), &login_user, w)
			return
		}
		
		data,err := json.Marshal(ans)
		if err != nil{
			ErrorResponse(err.Error(), &login_user, w)
			return
		}
		log_print("ユーザーログイン ID:%s, TABLE:%s", ans.Username, ans.Table)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	} else if count == 0{
		resp , err := json.Marshal(Message("Account does not exist","アカウント認証に失敗しました",&login_user))
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}else {
		error_print("認証エラー ID:%s TABLE:%s",login_user.Username, login_user.Table)
		resp , err := json.Marshal(Message("Error","アカウントエラーが発生しました。管理者にお問い合わせください",&login_user))
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}


//ログアウト用、トークンを破棄する




func GET_SET_MONEY(w http.ResponseWriter, r *http.Request, userres func(*user_auth)(user_result)){
	var user user_auth
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil{
		ErrorResponse(err.Error(),nil,w)
		return
	}
	if tmp, err := userAuthentication(user); !tmp || err != nil{
		ErrorResponse("認証失敗",nil, w)
		return
	} 
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(userres(&user)); err != nil{
		http.Error(w, "Server Error", http.StatusInternalServerError)
	}
}


//ユーザーの現在金額取得用のプログラム
func GET_USER_MONEY(w http.ResponseWriter, r *http.Request){
	GET_SET_MONEY(w,r,func (user *user_auth)(user_result){
		password := fmt.Sprintf("%x", sha256.Sum256([]byte(user.Password)))
		var money int
		query := "select money from Account_table WHERE username = ? AND password = ? AND TOKEN = ?"
		err := account_db.QueryRow(query, user.Username, password, user.Token).Scan(&money)
		if err != nil {
			return Message("Error", "クリエでエラーが発生しました",user)
		}
		user.Money = money
		return Message("success", "完了",user)
	})
}

func get_session_id(user *user_auth)(int, error){
	query := `
		select id from session_tokens
		where TOKEN = ? AND username = ? AND table_id = ?
	`
	var id int
	if err := account_db.QueryRow(query, user.Token, user.Username, user.Table).Scan(&id); err != nil{
		return -1, err
	}
	return id , nil
}

//ユーザーの現在金額更新用のプログラム
//社長、ここのことで合ってる？
func UPDATE_USER_MONEY(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	GET_SET_MONEY(w,r,func (user * user_auth)(user_result){
		query := `
			select money from Account_table WHERE username = ? AND TOKEN = ?
		`
		now := 0
		if err := account_db.QueryRow(query, user.Username,user.Token).Scan(&now); err != nil{
			error_print("クリエエラー(UPDATE_USER_MONEY):%v",err)
			return Message("Error","クエリエラー１",user)
		}
		
		query = `
			Insert Into slot_result_table(time, money, fluctuation, type, session_id, user, table_id)
			values(?,?,?,?,?,?,?)
		`
		sessionid , _ := get_session_id(user)
		if _, err := account_db.Exec(query, time.Now().Format("2006-01-02 15:04:05"), user.Money, user.Money - now, 1, sessionid, user.Username,user.Table); err != nil{
			return Message("Error", "クリエエラー２",user)
		}
		query = "update Account_table set money = ?  WHERE username = ? AND password = ? AND TOKEN = ?"
		_ ,err := account_db.Exec(query, user.Money ,user.Username, fmt.Sprintf("%x",sha256.Sum256([]byte(user.Password))), user.Token)
		if err != nil {
			return Message("Error", "クリエでエラーが発生しました",user)
		}
		return Message("success", "完了",user)
	})
}

