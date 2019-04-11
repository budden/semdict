DROP DATABASE IF EXISTS sduser_db;
CREATE DATABASE sduser_db;
\connect sduser_db
/* END_CREATE - keep this line intact. It is used to make the test db */

\set ON_ERROR_STOP on
\set thisdir `echo "$GOPATH/src/github.com/budden/semdict/sql"`

create language pltcl;

\i :thisdir/session_local_storage.sql
\i :thisdir/forward_declarations.sql
\i :thisdir/mutex.sql
\i :thisdir/user_registration_session.sql
\i :thisdir/language_and_sense.sql
\i :thisdir/privilege.sql

\echo *** recreate_sduser_db.sql Done
