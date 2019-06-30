--/*
\connect sduser_db
\set ON_ERROR_STOP on
--*/ 

-- create function queue_email()
-- create function peek_email()
-- create function deque_email()

/* create or replace function foo(id int) returns table (jd int)
language plpgsql as $$ 
 declare s tsense;
 declare msg text;
 begin 
 select * from tsense t into s limit 1;
 select t.id,t.word from tsense t into s.id, s.word limit 1;
 msg = format('s.id = %s, s.word = %s, s.phrase = %s'
   ,s.id, s.word, s.phrase);
 raise exception using message = msg;
 end; $$; 

select foo(1); */

\echo *** email.sql Done
