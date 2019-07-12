--/*
\connect sduser_db
\set ON_ERROR_STOP on
--*/ 


-- fnOnePersonalSense returns a user's proposal for the specific common sense if there is one
-- and the common sense otherwise
create or replace function fnonepersonalsense(p_sduserid bigint, p_commonid bigint) 
  returns table(r_commonid bigint, r_proposalid bigint, r_senseid bigint, r_deletionproposed bool)
  language plpgsql as $$
  begin
  return query(
    select cast(orig.id as bigint) as r_commonid
      ,cast(vari.id as bigint) as r_proposalid
      ,cast(coalesce(vari.id, orig.id) as bigint) as r_senseid
      ,coalesce(vari.deletionproposed, false) as r_deletionproposed
    from tsense orig 
    left join tsense vari 
    on orig.id = vari.originid and vari.ownerid = p_sduserid and not vari.phantom
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
    ,p_proposalstatus enum_proposalstatus, p_phrase text, p_word text
    ,p_deletionproposed bool)
  returns table (r_proposalid bigint)
  language plpgsql as $$
  declare v_phantom bool;
  declare update_count int;
  declare v_commonid bigint;
  declare v_proposalid bigint;
  begin
  p_proposalid = coalesce(p_proposalid,0);
  p_commonid = coalesce(p_commonid,0);
  if coalesce(p_proposalstatus,'n/a') = 'n/a' then
    raise exception 'proposal status must be not null, not "n/a"'; end if;
  if p_commonid = 0 and p_deletionproposed then
    raise exception 'you''re suggesting to delete a non-existent sense'; end if;
  if p_proposalid <> 0 then
    select originid from tsense where id = p_proposalid 
      into v_commonid;
    if coalesce(v_commonid, 0) <> p_commonid then
      raise exception 'origin mismatch'; end if;
    if exists (select 1 from tsense where 
        id = v_commonid 
        and word = p_word 
        and phrase = p_phrase 
        and phantom = p_deletionproposed) then
      raise exception 'You suggest no change to the common sense. Can''t save'; end if;
    v_proposalid = p_proposalid;
  else -- hence p_proposalid=0
    select ensuresenseproposal(p_sduserid, p_commonid) into v_proposalid; end if;
  
  update tsense set 
    proposalstatus = p_proposalstatus
    ,phrase = p_phrase
    ,word = p_word
    ,deletionproposed = p_deletionproposed
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
    p_sduserid bigint, p_commonid bigint, p_proposalid bigint, p_ownerid bigint, p_phantom bool, p_deletionproposed bool) 
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
    when p_phantom then 'deleted (phantom)'
    when p_deletionproposed then 'deletion'
    else 'change' end;
  return query(select r_commonorproposal, r_whos, r_kindofchange); end;
$$;

create or replace function explainCommonAndMine(
    p_sduserid bigint, p_commonid bigint, p_proposalid bigint, p_ownerid bigint, p_phantom bool)
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
  ,phantom bool
  ,deletionproposed bool
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
  	,vari.phrase, vari.word, vari.phantom, vari.deletionproposed 
    ,vari.ownerid, vari.sdusernickname, vari.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,false,vari.deletionproposed)).*
    ,(explainCommonAndMine(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,vari.phantom)).*
  	from vsense_wide as vari where vari.originid = p_commonid and not vari.phantom
	union all 
  	select s.commonid, s.proposalid, s.senseid
    ,cast('n/a' as enum_proposalstatus)
  	,s.phrase, s.word, s.phantom, false as deletionproposed
    ,cast(0 as bigint) as ownerid, '<common>' as sdusernickname, s.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.phantom,false)).*
    ,(explainCommonAndMine(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.phantom)).*
  	from vsense_wide s where id = p_commonid
	order by iscommon desc, ismine desc); end;
$$;



create or replace function fnlanguageproposals(p_sduserid bigint, p_commonid bigint) 
  returns table (commonid bigint
  ,proposalid bigint
  ,senseid bigint
  ,proposalstatus enum_proposalstatus
  ,phrasecommon text
  ,word varchar(512)
  ,phantom bool 
  ,deletionproposed bool
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
  	,vari.phrase, vari.word, false as phantom, vari.deletionproposed 
    ,vari.ownerid, vari.sdusernickname, vari.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,false,vari.deletionproposed)).*
    ,(explainCommonAndMine(p_sduserid,vari.commonid,vari.proposalid,vari.ownerid,vari.phantom)).*
  	from vsense_wide as vari where vari.originid = p_commonid and not vari.phantom 
	union all 
  	select s.commonid, s.proposalid, s.senseid
    ,cast('n/a' as enum_proposalstatus)
  	,s.phrase, s.word, s.phantom, false as deletionproposed
    ,cast(0 as bigint) as ownerid, '<common>' as sdusernickname, s.languageslug
  	,(explainSenseEssenseVsProposals(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.phantom,false)).*
    ,(explainCommonAndMine(p_sduserid,s.commonid,s.proposalid,s.ownerid,s.phantom)).*
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
  ,phantom bool -- if we return a proposal, it is taken from a proposal, not from the origin!
  ,deletionproposed bool
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
      -- ops is a proposal or a common sense. s is the same
      select s.commonid, s.proposalid, s.senseid
        ,s.proposalstatus
        ,s.phrase, s.word, s.phantom, s.deletionproposed
        ,s.sdusernickname
        ,s.languageslug
        ,(explainSenseEssenseVsProposals(p_sduserid, s.commonid, s.proposalid, s.ownerid, s.phantom, ops.r_deletionproposed)).*
	      from fnonepersonalsense(p_sduserid, p_commonid) ops
        -- actually it is an inner join to the same record
  		  left join vsense_wide as s on s.id = ops.r_senseid
        limit 1);
  else
    -- again it can be a proposal or a common sense
    someid = coalesce(nullif(p_proposalid,0),p_senseid);
    return query(
      select s.commonid, s.proposalid, s.senseid
        ,s.proposalstatus
        ,s.phrase, s.word, s.phantom, s.deletionproposed
        ,s.sdusernickname
        ,s.languageslug
        ,(explainSenseEssenseVsProposals(p_sduserid, s.commonid, s.proposalid, s.ownerid, s.phantom, s.deletionproposed)).*
  	    from vsense_wide as s where s.id = someid 
			  limit 1); end if; end;
$$;

\echo *** language_and_sense.sql Done
