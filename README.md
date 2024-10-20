# Slot_Back

## Log用やり取りJson
```json
{
    "level":0,
    "location":"",
    "message"
}

{
    "logs": [
        {
            "level":0,
            "location":"",
            "message"
        },{
            "level":0,
            "location":"",
            "message"
        },{
            "level":0,
            "location":"",
            "message"
        },{
            "level":0,
            "location":"",
            "message"
        },{
            "level":0,
            "location":"",
            "message"
        },{
            "level":0,
            "location":"",
            "message"
        }
    ]
}
```

## Account用やり取りJson
認証用
```json
{
    "Key"     :"",
    "username":"",
    "password":"",
    "token"   :"",
    "table"   :"",
    "money"   :00
}
```

応答用
```json
{
    "result"  :"",
    "message" :"",
    "username":"",
    "password":"",
    "token"   :"",
    "table"   :"",
    "money"   :00
}
```

## Table用やり取りJSON
table_hash実装したほうがいいのかわからん
（とりあえず、テーブル認識用に入れとく）
```json
{
    "key":"aaa",
	"table_id" :"",
	"probability" :0,
    "table_hash":""
}
```


# 管理側セキュリティガバガバ問題

想像以上にこれやばいかもしれないので、修正必須