--/*
\connect sduser_db
\set ON_ERROR_STOP on
--*/ 


create or replace function test_fnsensorproposalforview() returns void
language plpgsql as $$
begin
 if not exists (select 1 from fnsenseorproposalforview(1,1,null,null) 
  where commonid = 1 and proposalid = 0 and senseid = 1) THEN
   raise exception 'test_fnsensorproposalforview failure 1'; end if; 
 if not exists (select 1 from fnsenseorproposalforview(1,null,null,1) 
  where commonid = 1 and proposalid = 0 and senseid = 1) THEN
   raise exception 'test_fnsensorproposalforview failure 2'; end if; 
 if not exists (select 1 from fnsenseorproposalforview(1,4,null,null) 
  where commonid = 4 and proposalid = 6 and senseid = 6) THEN
   raise exception 'test_fnsensorproposalforview failure 3'; end if; 
 if not exists (select 1 from fnsenseorproposalforview(1,null,6,null) 
  where commonid = 4 and proposalid = 6 and senseid = 6) THEN
   raise exception 'test_fnsensorproposalforview failure 4'; end if; 
 if not exists (select 1 from fnsenseorproposalforview(1,null,null,6) 
  where commonid = 4 and proposalid = 6 and senseid = 6) THEN
   raise exception 'test_fnsensorproposalforview failure 5'; end if; 
end;
$$;

create or replace function test_fnonepersonalsense() returns void
language plpgsql as $$
begin
  if not exists (select 1 from fnonepersonalsense(1,4)
    where r_commonid = 4 and coalesce(r_proposalid,0) = 6) then
      raise exception 'test_fnonepersonalsense failure 1'; end if;
  if not exists (select 1 from fnonepersonalsense(1,1)
    where r_commonid = 1 and coalesce(r_proposalid,0) = 0) then
      raise exception 'test_fnonepersonalsense failure 2'; end if; 
end;
$$;


create or replace function test_fncommonsenseandproposals() returns void
language plpgsql as $$
begin
 if (select count(1) from fncommonsenseandproposals(1,1)) <> 1 then
   raise exception 'test_fncommonsenseandproposals failure 1'; end if; 
 if (select count(1) from fncommonsenseandproposals(1,4)) <> 2 then
   raise exception 'test_fncommonsenseandproposals failure 1'; end if; 
end;
$$;


select test_fnsensorproposalforview();

select test_fnonepersonalsense();

select test_fncommonsenseandproposals();

\echo *** language_and_sense_tests.sql Done
