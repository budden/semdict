DROP DATABASE IF EXISTS sd_db;

CREATE DATABASE sd_db;

\connect sd_db;

CREATE SEQUENCE sequence_sduser2_id;

CREATE TABLE sduser2 (
 id integer NOT NULL primary key,
 nickname varchar(256) not null,
 registertimestamp timestamptz not null default current_timestamp
);

comment on table sduser2 is 'sduser2 is a user of an application. We also have sduser table in sdusers_db where password is stored, hence the suffix 2. We want to make a dump of this db publically available so we only include id and nickname here.';

comment on column sduser2.id is 'sduser2 is related to sduser via id';

create unique index
 i_sduser2_nickname
 on sduser2(lower(nickname));

