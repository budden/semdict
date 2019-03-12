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
 salt text not null,
 registrationtimestamp timestamptz not null
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
 confirmationid text not null,
 registrationtimestamp timestamptz not null default current_timestamp
);

comment on table registrationattempt is 'registrationattempt gets a new record at each valid registration attempt. We keep registration attemps separated from users table for the case of registration flooding attack';
comment on column registrationattempt.confirmationid is 'confirmationid is random and as such it can be non-unique. In this case we ask user to re-register';

-- When a user is registrering with a both non-unique nickname and a non-unique E-mail, 
-- it is unspecified which unique constraint fires first (or I don't know).
-- It if fair to ask that nickname is checked first.
-- Experiment shows that the index first created is also first checked
-- But of course it is fragile (or, again, I don't know)
create unique index
 i_registrationattempt__nickname
 on registrationattempt(lower(nickname));

create unique index
 i_registrationattempt__registrationemail
 on registrationattempt(lower(registrationemail));

--- delete_expired_registrationattempts. 
--- We could run it from the process_registrationformsubmit, but in this case
--- a request to add a non-unique nickname would cause deletion and then rollback.
--- So we run this one in a separated transaction. But we use single goroutine for all
--- activity related to sd_users_db modifications, so calls to this one can't overlap
--- with other writes to the entire db.
create or replace function delete_expired_registrationattempts() 
returns void as $$
  declare 
    expiration_boundary timestamptz;
  begin
    select current_timestamp - interval '10' minute into expiration_boundary;
    -- raise info 'expiration_boundary = %', expiration_boundary;
    delete from registrationattempt where registrationtimestamp <= expiration_boundary;
  end
$$ language plpgsql;

--- nickname and password must be unique in the union of registrationattempt and sduser tables
--- use repeatable read transaction and/or single threaded registration processor
create or replace function process_registrationformsubmit(p_nickname text
  ,p_hash text
  ,p_salt text
  ,p_registrationemail text
  ,p_confirmationid text)
returns void as $$
 BEGIN
  if exists (select 1 from sduser ra where lower(ra.nickname)=lower(p_nickname)) THEN
    raise unique_violation using table = 'sduser', column = 'nickname', constraint = 'i_sduser_nickname';
  end if;
  if exists (select 1 from sduser ra where lower(ra.registrationemail)=lower(p_registrationemail)) THEN
    raise unique_violation using table = 'sduser', column = 'registrationemail', constraint = 'i_sduser_registrationemail';
  end if;
  insert into registrationattempt(nickname, hash, salt, registrationemail, confirmationid) 
    values (p_nickname, p_hash, p_salt, p_registrationemail, p_confirmationid);
 end;
$$ language plpgsql;

create or replace function process_registrationconfirmation(p_confirmationid text, p_nickname text)
returns void as $$
  declare v_id bigint := null;
  begin
    insert into sduser (nickname, registrationemail, hash, salt, registrationtimestamp)
     select nickname, registrationemail, hash, salt, registrationtimestamp from registrationattempt 
     where confirmationid = p_confirmationid and nickname = p_nickname returning id into v_id;
    if v_id is null THEN
     -- there is no no_data condition_name, so we resort
     -- to sqlstate
     raise exception sqlstate '02000' using message = 'registrationattempt not found';
    end if;
    -- we have a deadlock threat here in combination with the process_registrationformsubmit
    -- but we ensure at the application level that only one connection runs either of those procs
    -- simultaneously (all operations are protected with the mutex), so we don't care.
    -- But if we run those procs outside of our web app, deadlocks can occur in web app, so beware!
    delete from registrationattempt where id = v_id;
  end;
$$ language plpgsql;

