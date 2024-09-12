package main

import (
	"net/http"
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

type Create_Account struct{
	Username string `json:"username"`
	Password string `json:"password"`
}

//アカウント登録（サーバー方面）　完成
//とりあえずだけど、Webから情報を受け取って、アカウントを作成するだけ
//まだ、ログイン、ログアウトはできてないし、ゲスト用も作れてないけど、それは後々
func Create_User_Handle(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST"{
        http.Redirect(w, r, "/Create_User", http.StatusSeeOther)
		return
	}
	var ca_js Create_Account
	err := data2json(r, &ca_js)
	if err != nil{
		http.Error(w, "jsonの形が異なるか、送信されていません",400)
		return
	}
	db, err := NewDatabase("account_system:acpassword@(localhost)/account_server")
	if err != nil {
		http.Error(w, "データベースに接続できませんでした", 500)
		return
	}
	//USERにprimary keyを指定しているので、エラーが起きたら、すでにその名前があると認識させ、もう一度と返します
	//ここバグの温床になる気がするで、元気があったら改善する
	
	_ ,err = db.Exec("Insert into Account_table(username, password,money,TOKEN) values (?,?,0, NULL)",ca_js.Username,ca_js.Password)
	if err != nil{
		w.WriteHeader(200)
		w.Write([]byte("{\"result\":\"username_already_exists\"}"))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("{\"result\":\"succsess\"}"))
	
}
