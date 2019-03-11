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

-- https://stackoverflow.com/a/9808332/9469533 - it is considered safe to lowercase an E-mail
create unique index 
 i_sduser_registrationemail 
 on sduser(lower(registrationemail));

create unique index
 i_sduser_nickname
 on sduser(lower(nickname));

CREATE TABLE registrationattempt (
 id serial primary key,
 nickname varchar(256) not null,
 registrationemail text not null,
 hash text NOT NULL,
 salt text not null,
 expiry timestamptz
);

create unique index
 i_registrationattempt__registrationemail
 on registrationattempt(lower(registrationemail));

create unique index
 i_registrationattempt__nickname
 on registrationattempt(lower(nickname));

--- nickname and password must be unique in the union of registrationattempt and sduser tables
--- use repeatable read transaction and/or single threaded registration processor
create or replace function process_registrationformsubmit(p_nickname text, p_hash text, p_salt text, p_registrationemail text)
returns void as $$
 BEGIN
  if exists (select 1 from sduser ra where lower(ra.nickname)=lower(p_nickname)) THEN
   raise unique_violation using table = 'sduser', column = 'nickname', constraint = 'i_sduser_nickname';
  end if;
  if exists (select 1 from sduser ra where lower(ra.registrationemail)=lower(p_registrationemail)) THEN
   raise unique_violation using table = 'sduser', column = 'registrationemail', constraint = 'i_sduser_registrationemail';
  end if;
  insert into registrationattempt(nickname, hash, salt, registrationemail) 
   values (p_nickname, p_hash, p_salt, p_registrationemail);
 end;
$$ language plpgsql;
