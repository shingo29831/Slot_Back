-- メモ用:現在までのMYSQLの設定をここに書き連ねてあります。全部Mysqlに打ち込めば、動くはずです

Create user 'logsystem'@'%' Identified by 'logsyspassword';
Create Database log_server;
use log_server
create table Log_table(
    time DATETIME,
    level Integer,
    location varchar(30),
    message varchar(256)
);
Grant All Privileges on log_server.* to 'logsystem'@'%';
FLUSH PRIVILEGES;


create user 'account_system'@'%' identified by 'acpassword';
Create database account_server;
use account_server;
create table Account_table(
    username varchar(256) Primary key,
    usertype Integer,
    password varchar(256),
    money Integer,
    TOKEN varchar(256)
);
Grant All Privileges on account_server.* to 'account_system'@'%';
FLUSH PRIVILEGES;
