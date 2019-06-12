--/*
\connect sduser_db
\set ON_ERROR_STOP on
--*/ 

create or replace function get_language_slug(p_languageid int) returns text
 language plpgsql strict as $$
 declare v_result text;
 declare v_len_limit int;
  begin
  
  v_len_limit = 256;
  with recursive r as 
  (select id, parentid, cast(slug as text) from tlanguage
  where id = p_languageid 
  union 
  select r.id, tl.parentid, r.slug || '/' || tl.slug from r 
  left join tlanguage tl on tl.id = r.parentid 
  where tl.id is not null 
    or r.slug is null -- this should never happen as slug is not null, but just in case
    or length(r.slug) > v_len_limit -- guard against an unlimited recursion 
  )

  select slug from r 
  where parentid is null 
  into v_result;

  if length(v_result) > v_len_limit then
    v_result = 'bad slug for languageid='||p_languageid;
  end if;

  return v_result;
  end;
$$;

-- see also vsense_wide
/* there are four interpreation of sense id:
 - senseid == tsense.id, regardless of if sense is a common sense or a proposal
 - tsense.originid is a common sense for change and delete proposals
 - commonid is an id of common sense, if this sense is common or a change or delete proposal
 - proposalid is an id, if this sense is a proposal
 */
create or replace view vsense as select s.*
  ,coalesce(case when s.ownerid is not null then s.originid else cast(s.id as bigint) end,0) as commonid
  ,coalesce(case when s.ownerid is not null then cast(s.id as bigint) else null end,0) as proposalid
  ,cast(s.id as bigint) as senseid
  from tsense s;

-- see also vsense
create or replace view vsense_wide as select s.*
  ,coalesce(case when s.ownerid is not null then s.originid else cast(s.id as bigint) end,0) as commonid
  ,coalesce(case when s.ownerid is not null then cast(s.id as bigint) else null end,0) as proposalid
  ,cast(s.id as bigint) as senseid
  ,u.nickname as sdusernickname
  -- FIXME suboptimal!
  ,get_language_slug(s.languageid) as languageslug
  from tsense s left join sduser u on s.ownerid=u.id;



-- fnPersonalSenses returns all personal senses for the user. If the user is 0 or null,
-- then common senses are returned as well as unparallel personal
-- to copy-paste or complicate this one to have a good select plan for searches.
create or replace function fnpersonalsenses(p_sduserid bigint) 
  returns table(r_commonid bigint, r_proposalid bigint, r_proposalstatus enum_proposalstatus, r_countofproposals bigint, r_addedbyme bool)
  language plpgsql as $$
  begin
  if coalesce(p_sduserid, 0) = 0 then
    return query(
      select cast(orig.id as bigint) as r_commonid
      ,cast(null as bigint) as r_proposalid
      ,'n/a' as r_proposalstatus
      ,(select count(1) from tsense varic where varic.originid = orig.id) as r_countofproposals
      ,false as r_addedbyme
      from tsense orig where orig.originid is null and orig.ownerid is null); 
  else
    return query(
      select cast(orig.id as bigint) as r_commonid
      ,cast(vari.id as bigint) as r_proposalid
      ,vari.proposalstatus as r_proposalstatus
      ,(select count(1) from tsense varic where varic.originid = orig.id) as r_countofproposals
      ,case when orig.ownerid = p_sduserid then true else false end as r_addedbyme
      from tsense orig 
      left join tsense vari on orig.id = vari.originid and vari.ownerid = p_sduserid 
      where orig.originid is null); end if; end;
$$;


-- fnOnePersonalSense returns a personal or common sense for the specific sense id
create or replace function fnonepersonalsense(p_sduserid bigint, p_commonid bigint) 
  returns table(r_commonid bigint, r_proposalid bigint, r_senseid bigint)
  language plpgsql as $$
  begin
  return query(
    select cast(orig.id as bigint) as r_commonid
      ,cast(vari.id as bigint) as r_proposalid
      ,cast(coalesce(vari.id, orig.id) as bigint) as r_senseid
    from tsense orig 
    left join tsense vari on orig.id = vari.originid and vari.ownerid = p_sduserid 
    where orig.id = p_commonid and orig.originid is null); end;
$$;

-- fnSavePersonalSense saves the sense. p_evenifidentical must be false for now
-- Use cases:
/* commonid is not null, proposalid is null:
    We are adding proposal to the existing sense
   commonid is not null, proposalid is not null
    We are updating a pre-existing proposal */
create or replace function fnsavepersonalsense(
    p_sduserid bigint, p_commonid bigint, p_proposalid bigint
    ,p_proposalstatus enum_proposalstatus, p_phrase text, p_word text, p_evenifidentical bool)
  returns table (r_proposalid bigint)
  language plpgsql as $$
  declare v_deleted bool;
  declare update_count int;
  declare v_commonid bigint;
  declare v_proposalid bigint;
  begin
  p_proposalid = coalesce(p_proposalid,0);
  p_commonid = coalesce(p_commonid,0);
  if coalesce(p_proposalstatus,'n/a') = 'n/a' then
    raise exception 'proposal status must be not null, not "n/a"'; end if;
  if p_evenifidentical then
    raise exception 'invalid parameter p_evenifidentical'; end if;
  if p_proposalid <> 0 then
    select originid, deleted 
      from tsense where id = p_proposalid 
      into v_commonid, v_deleted;
    if coalesce(v_commonid, 0) <> p_commonid then
      raise exception 'origin mismatch'; end if;
    if exists (select 1 from tsense where 
        id = v_commonid 
        and word = p_word 
        and phrase = p_phrase 
        and deleted = v_deleted) then
    -- nothing differs from the official version, delete our proposal
      delete from tsense where id = p_proposalid;
      return query(select true); return; end if;
    v_proposalid = p_proposalid;
  else -- hence p_proposalid=0
    select ensuresenseproposal(p_sduserid, p_commonid) into v_proposalid; end if;
  
  update tsense set 
    proposalstatus = p_proposalstatus
    ,phrase = p_phrase
    ,word = p_word
    where id = v_proposalid;

  get diagnostics update_count = row_count;
  if update_count != 1 then
    raise exception 'expected to update just one record, which didn''t hapen'; end if;
  return query(select v_proposalid); return; end;
$$;

-- EnsureSenseProposal ensures that a user has his own proposal of a sense. One should not
-- make a proposal of user's unparallel sense.
create or replace function ensuresenseproposal(p_sduserid bigint, p_commonid bigint)
returns table (proposalid bigint) 
language plpgsql as $$
  declare r_proposalid bigint;
  declare v_ownerid bigint;
  declare v_row_count int;
  begin
    lock table themutex;
    if coalesce(p_commonid,0) = 0 then
      raise exception 'p_commonid must be specified'; end if;
    if coalesce(p_sduserid,0) = 0 then
      raise exception 'p_sduserid must be specified'; end if;
    select ownerid from tsense where id = p_commonid into v_ownerid;
    get diagnostics v_row_count = row_count;
    if v_row_count != 1 then
      raise exception 'Common sense not found'; end if;
    if nullif(v_ownerid,0) is not null then
      raise exception 
      'You can''t make a proposal of user''s new sense, until it is accepted to the language'; end if;
    select min(id) from tsense 
      where originid = p_commonid and ownerid = p_sduserid
      into r_proposalid;
    if r_proposalid is not null then 
      return query (select r_proposalid); 
      return; end if;
    insert into tsense (languageid, phrase, word, originid, ownerid)
      select languageid, phrase, word, id, p_sduserid 
      from tsense where id = p_commonid returning id into r_proposalid;
    if r_proposalid is null then
      raise exception 
        'something went wrong, sense cloning failed'; 
    end if;
  return query (select r_proposalid);
  end;
$$;

-- this is a mess...
select ensuresenseproposal(1,4);
update tsense set proposalstatus = 'draft', phrase = 'updated sense' where id=5;

-- end of mess
-- FIXME - addition is shown incorrectly
create or replace function explainSenseEssenseVsProposals(
    p_sduserid bigint, p_commonid bigint, p_proposalid bigint, p_ownerid bigint, p_deleted bool) 
  returns
  table (commonorproposal varchar(128), whos varchar(512), kindofchange varchar(128))
  language plpgsql CALLED ON NULL INPUT as $$
  declare r_commonorproposal varchar(128);
  declare r_whos varchar(512);
  declare r_kindofchange varchar(128);
begin
  r_commonorproposal = case
    when coalesce(p_proposalid,0) = 0 then 'common' 
    else 'proposal' end;
  r_whos = case 
    when coalesce(p_ownerid,0) = 0 then '' -- common - irrelevant
    when p_sduserid = p_ownerid then '<my>' 
    else 
      coalesce((select nickname from sduser where id = p_ownerid)
        ,'owner not found') end;
  r_kindofchange = case
    when coalesce(p_ownerid,0)=0 then '' -- common - irrelevant
    when coalesce(p_commonid,0)=0 then 'addition'
    when p_deleted then 'deletion'
    else 'change' end;
  return query(select r_commonorproposal, r_whos, r_kindofchange); end;
$$;

create or replace function explainCommonAndMine(
    p_sduserid bigint, p_commonid bigint, p_proposalid bigint, p_ownerid bigint, p_deleted bool)
  returns
  table (iscommon bool, ismine bool)
  language plpgsql CALLED ON NULL INPUT as $$
begin
  return query(select
  	case when coalesce(p_proposalid,0) = 0 then true 
      else false end as iscommon
  	,case when coalesce(p_sduserid,0) = 0 then false 
      when coalesce(p_ownerid,0) = p_sduserid then true 
      else false end as ismine); end;
$$;

create or replace function fncommonsenseandproposals(p_sduserid bigint, p_commonid bigint) 
  returns table (commonid bigint
  ,proposalid bigint
  ,senseid bigint
  ,proposalstatus enum_proposalstatus
  ,phrase text
  ,word varchar(512)
  ,deleted bool
  ,ownerid bigint
  ,sdusernickname varchar(128)
  ,languageslug text
  ,commonorproposal varchar(128)
  ,whos varchar(512)
  ,kindofchange varchar(128)
  ,iscommon bool
  ,ismine bool
  ) language plpgsql as $$ 
begin
return query(
  select vari.commonid, vari.proposalid, vari.senseid
    ,vari.proposalstatus
  	,vari.phrase, vari.word, vari.deleted, vari.ownerid, vari.sdusernickname, vari.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,vari.deleted)).*
    ,(explainCommonAndMine(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,vari.deleted)).*
  	from vsense_wide as vari where vari.originid = p_commonid 
	union all 
  	select s.commonid, s.proposalid, s.senseid
    ,cast('n/a' as enum_proposalstatus)
  	,s.phrase, s.word, s.deleted, cast(0 as bigint) as ownerid, '<common>' as sdusernickname, s.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.deleted)).*
    ,(explainCommonAndMine(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.deleted)).*
  	from vsense_wide s where id = p_commonid
	order by iscommon desc, ismine desc); end;
$$;

-- fnProposalAndCommonSenseForProposalAcceptOrReject
create or replace function fnproposalandcommonsenseforproposalacceptorreject(p_sduserid bigint, p_proposalid bigint)
  returns table (commonid bigint
  ,proposalid bigint
  ,senseid bigint
  ,proposalstatus enum_proposalstatus
  ,phrase text
  ,word varchar(512)
  ,deleted bool
  ,ownerid bigint
  ,sdusernickname varchar(128)
  ,languageslug text
  ,commonorproposal varchar(128)
  ,whos varchar(512)
  ,kindofchange varchar(128)
  ,iscommon bool
  ,ismine bool
  ) language plpgsql as $$ 
declare v_commonid bigint;
begin
select vari.commonid from vsense_wide as vari where vari.proposalid = p_proposalid into v_commonid;
return query(
  select vari.commonid, vari.proposalid, vari.senseid
    ,vari.proposalstatus
  	,vari.phrase, vari.word, vari.deleted, vari.ownerid, vari.sdusernickname, vari.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,vari.deleted)).*
    ,(explainCommonAndMine(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,vari.deleted)).*
  	from vsense_wide as vari where vari.proposalid = p_proposalid
	union all 
  	select s.commonid, s.proposalid, s.senseid
    ,cast('n/a' as enum_proposalstatus)
  	,s.phrase, s.word, s.deleted, cast(0 as bigint) as ownerid, '<common>' as sdusernickname, s.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.deleted)).*
    ,(explainCommonAndMine(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.deleted)).*
  	from vsense_wide s where id = v_commonid
	order by iscommon desc); end;
$$;



create or replace function fnlanguageproposals(p_sduserid bigint, p_commonid bigint) 
  returns table (commonid bigint
  ,proposalid bigint
  ,senseid bigint
  ,proposalstatus enum_proposalstatus
  ,phrasecommon text
  ,word varchar(512)
  ,deleted bool
  ,ownerid bigint
  ,sdusernickname varchar(128)
  ,languageslug text
  ,commonorproposal varchar(128)
  ,whos varchar(512)
  ,kindofchange varchar(128)
  ,iscommon bool
  ,ismine bool
  ) language plpgsql as $$ 
begin
return query(
  select vari.commonid, vari.proposalid, vari.senseid
    ,vari.proposalstatus
  	,vari.phrase, vari.word, vari.deleted, vari.ownerid, vari.sdusernickname, vari.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,vari.deleted)).*
    ,(explainCommonAndMine(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,vari.deleted)).*
  	from vsense_wide as vari where vari.originid = p_commonid 
	union all 
  	select s.commonid, s.proposalid, s.senseid
    ,cast('n/a' as enum_proposalstatus)
  	,s.phrase, s.word, s.deleted, cast(0 as bigint) as ownerid, '<common>' as sdusernickname, s.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.deleted)).*
    ,(explainCommonAndMine(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.deleted)).*
  	from vsense_wide s where id = p_commonid
	order by iscommon desc, ismine desc); end;
$$;


create or replace function fnsenseorproposalforview(p_sduserid bigint
  ,p_commonid bigint
  ,p_proposalid bigint
  ,p_senseid bigint)
returns table (commonid bigint
  ,proposalid bigint
  ,senseid bigint
  ,proposalstatus enum_proposalstatus
  ,phrase text
  ,word varchar(512)
  ,deleted bool
  ,sdusernickname varchar(256)
  ,languageslug text
  ,commonorproposal varchar(128)
  ,whos varchar(512)
  ,kindofchange varchar(128)
  )
language plpgsql as $$
  declare someid bigint;
  begin
  if coalesce(p_commonid,0) <> 0 then
    return query(
      select s.commonid, s.proposalid, s.senseid
        ,s.proposalstatus
        ,s.phrase, s.word, s.deleted 
        ,s.sdusernickname
        ,s.languageslug
        ,(explainSenseEssenseVsProposals(p_sduserid, s.commonid, s.proposalid, s.ownerid, s.deleted)).*
	      from fnonepersonalsense(p_sduserid, p_commonid) ops
  		  left join vsense_wide as s on s.id = ops.r_senseid
        limit 1);
  else
    someid = coalesce(nullif(p_proposalid,0),p_senseid);
    return query(
      select s.commonid, s.proposalid, s.senseid
        ,s.proposalstatus
        ,s.phrase, s.word, s.deleted 
        ,s.sdusernickname
        ,s.languageslug
        ,(explainSenseEssenseVsProposals(p_sduserid, s.commonid, s.proposalid, s.ownerid, s.deleted)).*
  	    from vsense_wide as s where s.id = someid
			  limit 1); end if; end;
$$;

-- tests
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
  where commonid = 4 and proposalid = 5 and senseid = 5) THEN
   raise exception 'test_fnsensorproposalforview failure 3'; end if; 
 if not exists (select 1 from fnsenseorproposalforview(1,null,5,null) 
  where commonid = 4 and proposalid = 5 and senseid = 5) THEN
   raise exception 'test_fnsensorproposalforview failure 4'; end if; 
 if not exists (select 1 from fnsenseorproposalforview(1,null,null,5) 
  where commonid = 4 and proposalid = 5 and senseid = 5) THEN
   raise exception 'test_fnsensorproposalforview failure 5'; end if; 
end;
$$;

create or replace function test_fnonepersonalsense() returns void
language plpgsql as $$
begin
  if not exists (select 1 from fnonepersonalsense(1,4)
    where r_commonid = 4 and coalesce(r_proposalid,0) = 5) then
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

\echo *** language_and_sense.sql Done
