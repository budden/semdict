
--/*
\connect sduser_db
\set ON_ERROR_STOP on
drop table if exists tprivilegekind cascade;
drop table if exists tuserprivilege cascade;
drop table if exists tuserlanguageprivilege cascade;
--*/ 

-- FIXME perlanguage2 can help for translations
create table tprivilegekind (
  id int primary key,
  name varchar(128) not null,
  perlanguage bool not null
);

insert into tprivilegekind (id, name, perlanguage)
values
 (1,'Login',false)
 ,(2,'Manage access',false)
 ,(3,'Edit language attributes',false)
 ,(4,'Accept/decline change requests',true);

create table tuserprivilege (
  id serial primary key,
  sduserid bigint not null references sduser on delete cascade,
  privilegekindid int not null references tprivilegekind
);

create table tuserlanguageprivilege (
 id serial primary key,
 sduserid int not null references sduser on delete cascade,
 privilegekindid int not null references tprivilegekind,
 languageid int not null references tlanguage on delete cascade
);

insert into tuserprivilege (sduserid, privilegekindid)
 values
 (1,1)
 ,(1,2); 

insert into tuserlanguageprivilege (sduserid, privilegekindid, languageid)
 values
 (1,3,1)
 ,(1,3,2)
 ,(1,4,1)
 ,(1,4,2);

create or replace function isuserhaveprivilege(p_sduserid bigint, p_privilegekindid int)
returns table (result bool) 
language plpgsql strict as $$
 BEGIN
  if exists (select 1 from tuserprivilege 
    where sduserid = p_sduserid and privilegekindid = p_privilegekindid) THEN
    return query(select true);
  ELSE
    return query(select false);
  END if; END;
$$;


create or replace function grantuserprivilege(p_sduserid bigint, p_privilegekindid int) returns void
language plpgsql strict as $$
 begin
  if not (select isuserhaveprivilege(p_sduserid, p_privilegekindid)) then
    insert into tuserprivilege (sduserid, privilegekindid) values (p_sduserid, p_privilegekindid); end if; end;
$$;


create or replace function isuserhavelanguageprivilege(p_sduserid bigint, p_privilegekindid int, p_languageid int)
returns table (result bool)
language plpgsql strict as $$
  BEGIN
  if exists (select 1 from tuserlanguageprivilege ulp
    where 
    ulp.sduserid = p_sduserid 
    and ulp.privilegekindid = p_privilegekindid 
    and ulp.languageid = p_languageid) THEN
    return query(select true);
  ELSE
    return query(select false);
  end if; end;
$$;

-- tests
create or replace function test_privilege() returns void
language plpgsql strict as $$
begin
 if not exists (select result from isuserhaveprivilege(1,1) where result = true) THEN
   raise exception 'Default user has no access to the database';
 end if;
end;
$$;

select test_privilege();



\echo *** privilege.sql Done
