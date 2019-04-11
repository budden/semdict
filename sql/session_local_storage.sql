
--/*
\connect sduser_db
\set ON_ERROR_STOP on
--*/ 

-- Origin: 
-- https://www.depesz.com/2009/08/20/getting-session-variables-without-touching-postgresql-conf/#comment-27980

CREATE OR REPLACE FUNCTION set_session_var(name varchar(128), val integer) RETURNS integer
LANGUAGE pltcl AS $_$
global sess
if {[string is alnum $1]} {
 set sess($1) $2;
} else {
 error "Name of a session variable must contain alphanumeric characters only"
}
$_$
;

CREATE OR REPLACE FUNCTION get_session_var(varchar(128)) RETURNS integer
LANGUAGE pltcl AS $_$
global sess
if {[string is alnum $1]} {
 return $sess($1)
} else {
 error "Name of a session variable must contain alphanumeric characters only"
}
$_$
;

\echo *** session_local_storage.sql Done
