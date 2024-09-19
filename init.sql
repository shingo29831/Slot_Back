Create user if not exists 'logsystem'@'localhost' Identified by 'logsyspassword';
Create Database if not exists log_server;
use log_server
create table if not exists Log_table(
    time DATETIME,
    level Integer,
    location varchar(30),
    message varchar(256)
);
Grant All Privileges on log_server.* to 'logsystem'@'localhost';
FLUSH PRIVILEGES;


create user if not exists 'account_system'@'localhost' identified by 'acpassword';
Create database if not exists account_server;
use account_server;
create table if not exists Account_table(
    username varchar(256) Primary key,
    usertype Integer,
    password varchar(256),
    money Integer,
    TOKEN varchar(256)
);
Grant All Privileges on account_server.* to 'account_system'@'localhost';
FLUSH PRIVILEGES;
