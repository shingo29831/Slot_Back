package main

import (
	"encoding/json"
	"io"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

var validUsername = "admin"
var validPassword = "password"

type pay_struct struct{
	TableId string `json:"tableId"`
	DepositAmount int `json:"depositAmount"`
	WithdrawalAmoun int `json:"withdrawalAmoun"`
}

func loginPage(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet {
		file, err := os.Open("./web/login_root.html")
		if err != nil {
			http.Error(w,"サーバーエラー", 500)
			return
		}
		defer file.Close()

		buf, err := io.ReadAll(file)
		if err != nil {
			http.Error(w,"サーバーエラー", 500)
			return
		}
		
		w.WriteHeader(200)
		w.Write(buf)
	}else if r.Method == http.MethodPost {

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == validUsername && password == validPassword {
			// セッションを取得
			session, _ := store.Get(r, "auth-session")

			// 認証情報をセッションに保存
			session.Values["authenticated"] = true
			session.Save(r, w)

			// ダッシュボードにリダイレクト
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		} else {
			// 認証失敗
			fmt.Fprintf(w, "Invalid credentials. Please try again.")
		}
	}
}

func pay_root(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		http.Error(w,"権限がありません", http.StatusForbidden)
		return
	}
	session, _ := store.Get(r, "auth-session")

	// 認証されていない場合、ログインページにリダイレクト
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	file, err := os.Open("./web/pay_root.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
	w.Write(buf)
}

func dashboardPage(w http.ResponseWriter, r *http.Request) {
	// セッションを取得
	session, _ := store.Get(r, "auth-session")

	// 認証されていない場合、ログインページにリダイレクト
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	file, err := os.Open("./web/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
	w.Write(buf)
}

func logout(w http.ResponseWriter, r *http.Request) {
	// セッションを取得
	session, _ := store.Get(r, "auth-session")

	// 認証情報をクリア
	session.Values["authenticated"] = false
	session.Save(r, w)

	// ログインページにリダイレクト
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func submit_transaction(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w,"権限がありません", http.StatusForbidden)
		return
	}
	var req pay_struct 
	err := data2json(r, &req)
	if err != nil{
		Error_res(err.Error(),nil,w)
		return
	}
	db, err := NewDatabase("")
	if err != nil {
		Error_res(err.Error(),nil,w)
		return
	}
	
	_ ,err = db.Exec("update Account_table set money = SUM(money, ?) where table = ?",req.WithdrawalAmoun + req.DepositAmount, req.TableId)
	if err != nil {
		Error_res(err.Error(),nil,w)
		return
	}
	resp, _:= json.Marshal(Message("succsess","ALL DONE",nil))
	w.WriteHeader(200)
	w.Write(resp)
}