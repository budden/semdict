DROP DATABASE IF EXISTS sdusers_db;

CREATE DATABASE sdusers_db;

\connect sdusers_db;

CREATE SEQUENCE sequence_sduser_id;

CREATE TABLE sduser (
 id bigint DEFAULT nextval('public.sequence_sduser_id') 
  NOT NULL primary key,
 nickname varchar(256) not null,
 registrationemail text not null,
 hash text NOT NULL,
 salt text not null
);

create unique index 
 i_sduser_registrationemail 
 on sduser(registrationemail);

create unique index
 i_se_user_nickname
 on sduser(nickname);

CREATE TABLE registrationattempt (
 id serial primary key,
 expiry timestamptz,
 registrationemail text not null,
 nickname varchar(256) not null
);

create unique index
 i_registrationattempt__registrationemail
 on registrationattempt(registrationemail);

create unique index
 i_registrationattempt__nickname
 on registrationattempt(nickname);

