Create user if not exists 'logsystem'@'%' Identified by 'logsyspassword';
Create Database if not exists log_server;
use log_server
create table if not exists Log_table(
    time DATETIME,
    level Integer,
    location varchar(30),
    message varchar(256)
);
Grant All Privileges on log_server.* to 'logsystem'@'%';
FLUSH PRIVILEGES;


create user if not exists 'account_system'@'%' identified by 'xM7B)NY-eexsJm';
Create database if not exists account_server;
use account_server;
create table if not exists Account_table(
    username varchar(256) Primary key,
    usertype Integer,
    password varchar(256),
    money Integer,
    table_id varchar(128),
    TOKEN varchar(256)
);
Grant All Privileges on account_server.* to 'account_system'@'%';
FLUSH PRIVILEGES;

create table if not exists table_table(
    table_id    varchar(128) unique NOT NULL,
    probability Integer ,
    table_hash varchar(256) 
);

create table if not exists slot_result_table(
    time TIMESTAMP,
    money Integer,
    table_id varchar(256)
);