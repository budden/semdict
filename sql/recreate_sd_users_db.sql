DROP DATABASE IF EXISTS sd_users_db;

CREATE DATABASE sd_users_db;

\connect sd_users_db;

CREATE SEQUENCE sequence_dict_user_id;

CREATE TABLE sd_user (
 id bigint DEFAULT nextval('public.sequence_dict_user_id') 
  NOT NULL primary key,
 nickname varchar(256) not null,
 registration_email text not null,
 hash text NOT NULL,
 salt text not null
);

create unique index 
 i_sd_user_registration_email 
 on sd_user(registration_email);

create unique index
 i_se_user_nickname
 on sd_user(nickname);

CREATE TABLE registration_attempt (
 id serial primary key,
 expiry timestamptz,
 registration_email text not null,
 nickname varchar(256) not null
);

create unique index
 i_registration_attempt__registration_email
 on registration_attempt(registration_email);

create unique index
 i_registration_attempt__nickname
 on registration_attempt(nickname);

