Create user if not exists 'logsystem'@'%' Identified by 'logsyspassword';
Create Database if not exists log_server;
use log_server;
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
    table_id    varchar(128) Primary key,
    probability Integer,
    table_hash varchar(256)
);
Insert Into table_table(table_id, probability, table_hash)values('table_120', 1, '21382817893728');


create table if not exists session_tokens(
    time TIMESTAMP,
    TOKEN varchar(256),
    username varchar(256),
    table_id varchar(128),
    id Integer AUTO_INCREMENT Primary key,
    FOREIGN KEY (table_id) REFERENCES table_table(table_id)
);

create table if not exists slot_result_table(
    time TIMESTAMP,
    money Integer,
    fluctuation Integer,
    type Integer,
    session_id Integer REFERENCES session_tokens(id),
    user varchar(256),
    table_id varchar(128)
);

create table if not exists slot_result_type(
    type Integer Primary Key,
    name varchar(256)
);

Insert Into slot_result_type values(0, '入金');
Insert Into slot_result_type values(1, '残金');
Insert Into slot_result_type values(2, '出金');

CREATE VIEW sessions (session_id, seles, balance)
AS 
SELECT session_id,
    SUM(CASE WHEN type = 0 THEN fluctuation ELSE 0 END) AS seles,
    MAX(CASE WHEN type = 1 THEN money END) AS last_amount 
FROM slot_result_table t1
WHERE t1.type = 0 OR (t1.type = 1 AND t1.time = (
    SELECT MAX(t2.time) FROM slot_result_table t2
    WHERE t2.type = 1 AND t1.session_id = t2.session_id
))
GROUP BY session_id 
ORDER BY session_id;