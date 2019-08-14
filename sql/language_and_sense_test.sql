--/*
\connect sduser_db
\set ON_ERROR_STOP on
--*/ 


/* create or replace function test_fncommonsenseandproposals() returns void
language plpgsql as $$
begin
 if (select count(1) from fncommonsenseandproposals(1,1)) <> 1 then
   raise exception 'test_fncommonsenseandproposals failure 1'; end if; 
 if (select count(1) from fncommonsenseandproposals(1,4)) <> 2 then
   raise exception 'test_fncommonsenseandproposals failure 1'; end if; 
end;
$$;

select test_fncommonsenseandproposals(); */

\echo *** language_and_sense_tests.sql Done
