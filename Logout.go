package main

import (
	"io"
	"net/http"
	"os"
	"encoding/json"
	"fmt"
)

type Logout_user struct {
	TableId string `json:"tableId"`
	Name 	 string `json:"username"`
}

var Logout_user_Array Array

func Logout_page(w http.ResponseWriter, r *http.Request){
	file, err := os.Open("./web/Logout_req.html")
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func style_css(w http.ResponseWriter, r *http.Request){
	file, err := os.Open("./web/styles.css")
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func approve_logout(w http.ResponseWriter, r *http.Request){
	query := `
		SELECT COUNT(*) FROM Account_table 	WHERE table_id = ? AND username = ?
	`
	if r.Method != http.MethodPost {
		http.Error(w, "BadRequest",http.StatusBadRequest)
		return
	}
	var user Logout_user
	if err := json.NewDecoder(r.Body).Decode(&user);err != nil {
		http.Error(w, "BadRequest", http.StatusBadRequest)
		return
	}
	Logout_user_Array.del_data(user)
	count := -1
	if err = account_db.QueryRow(query, user.TableId, user.Name).Scan(&count); err != nil || count != 1{
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		error_print("認証エラー%s, クエリ結果:%d",err, count)
		return
	}
	query = `
		UPDATE Account_table SET token = NULL,table_id = NULL 
		WHERE table_id = ? AND username = ? 
	`
	
	if _, err = account_db.Exec(query, user.TableId,user.Name); err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		error_print("クリエエラー%s", err.Error())
		return
	}

	log_print("ユーザーログアウト ID:%s, TABLE:%s", user.Name, user.TableId)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,"ALL DONE")
}


func User_Logout(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	var user user_auth
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil{
		ErrorResponse(err.Error(),nil,w)
		return
	}
	logout_user := Logout_user{
		Name: user.Username,
		TableId: user.Table,
	}
	Logout_user_Array.append(logout_user)
	if err = json.NewEncoder(w).Encode(Message("success","ログアウト待機中です", &user)); err != nil{
		http.Error(w,"InternalServerError",http.StatusInternalServerError)
	}
}

func Token_exists(w http.ResponseWriter, r *http.Request){
	query :=`
		SELECT COUNT(*) from Account_table 
		WHERE table_id = ? AND token = ? AND username = ?
	`
	if r.Method != "POST"{
        http.Redirect(w, r, "/create_User", http.StatusSeeOther)
		return
	}
	var user user_auth
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil{
		ErrorResponse(err.Error(),nil,w)
		return
	}
	count := -1
	if err = account_db.QueryRow(query, user.Table, user.Token,user.Username).Scan(&count); err != nil || count == -1  {
		ErrorResponse("クエリエラー",&user, w)
		return
	}
	if count == 0{
		if err = json.NewEncoder(w).Encode(Message("success","ログアウトに成功しました", &user)); err != nil{
			http.Error(w,"InternalServerError",http.StatusInternalServerError)
		}
	}else if count == 1{
		if err = json.NewEncoder(w).Encode(Message("success","ログアウト待機中です", &user)); err != nil{
			http.Error(w,"InternalServerError",http.StatusInternalServerError)
		}
	}else{
		if err = json.NewEncoder(w).Encode(Message("Error","エラーメッセージ　管理者に問い合わせてください", &user)); err != nil{
			http.Error(w,"InternalServerError",http.StatusInternalServerError)
		}
		error_print("認証エラー username:%s", user.Username)
	}
	
}

func logout_requests(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Redirect(w,r,"/create_User",http.StatusSeeOther)
		return
	}
	var ans struct{
		Result string `json:"result"`
		User []Logout_user `json:"user"`
	}
	if Logout_user_Array.size == 0{
		ans.Result = "Nothing"
	} else {
		ans.Result = "Exists"
	}
	for _, v := range Logout_user_Array.data {
		user_tmp, ok := v.(Logout_user);
		if  ok{
			ans.User = append(ans.User, user_tmp)
		}else{
			http.Error(w,"InternalServerError",http.StatusInternalServerError)
			return
		}
	}
	
	if err := json.NewEncoder(w).Encode(ans); err != nil{
		http.Error(w, "InternalServerError",http.StatusInternalServerError)
	}
}