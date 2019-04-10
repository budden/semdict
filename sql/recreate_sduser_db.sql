DROP DATABASE IF EXISTS sduser_db;
CREATE DATABASE sduser_db;
\connect sduser_db
/* END_CREATE - keep this line intact. It is used to make the test db */

\set ON_ERROR_STOP on
\set thisdir `echo "$GOPATH/src/github.com/budden/semdict/sql"`
\i :thisdir/mutex.sql
\i :thisdir/user_registration_session.sql
\i :thisdir/language_and_sense.sql

create table tprivilegekind (
  id int primary key,
  name varchar(128),
  perlanguage bool
);

-- insert into tprivilegekind 

create table tuserprivelege (
  id serial primary key,
  privilegekindid int not null references tprivilegekind
);


-- keep this one the last statement!
create view marker_of_script_success as select current_timestamp;

\echo *** recreate_sduser_db.sql Done