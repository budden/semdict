--/*
\connect sduser_db
\set ON_ERROR_STOP on
--*/ 

-- fnSaveSense saves the sense. p_evenifidentical must be false for now
-- Use cases:
create or replace function fnsavesense(
    p_sduserid bigint, p_senseid bigint, p_oword text, p_theme text,
    p_phrase text, p_ownerid bigint
    )
  returns table (r_senseid bigint)
  language plpgsql as $$
  declare update_count int;
  begin
  if coalesce(p_sduserid,0) = 0 THEN
   raise exception 'p_sduserid must be specified'; end if;
  if coalesce(p_senseid,0) = 0 then
   raise exception 'p_senseid must be specified'; end if;

  if not 
    (p_sduserid = 1/*tsar*/ 
    or exists (select 1 from tsense where id = p_senseid 
    and (ownerid is null or coalesce(ownerid,0)=p_ownerid))) then
    raise exception 'You are not allowed to update this sense'; end if;
 
  update tsense set 
    oword = p_oword
    ,theme = p_theme
    ,phrase = p_phrase
    ,ownerid = p_ownerid
    where id = p_senseid;

  get diagnostics update_count = row_count;
  if update_count != 1 then
    raise exception 'expected to update just one record, which didn''t hapen'; end if;
  return query(select p_senseid); return; end;
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


create or replace function fnlanguageproposals(p_sduserid bigint, p_commonid bigint) 
  returns table (commonid bigint
  ,proposalid bigint
  ,senseid bigint
  ,proposalstatus enum_proposalstatus
  ,phrasecommon text
  ,word varchar(512)
  ,phantom bool 
  ,ownerid bigint
  ,sdusernickname varchar(128)
  ,languageslug text
  ,iscommon bool
  ,ismine bool
  ) language plpgsql as $$ 
begin
return query(
 	select s.commonid, s.proposalid, s.senseid
    ,cast('n/a' as enum_proposalstatus)
  	,s.phrase, s.word, s.phantom
    ,cast(0 as bigint) as ownerid, '<common>' as sdusernickname, s.languageslug
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
  ,sdusernickname varchar(256)
  ,languageslug text
  )
language plpgsql as $$
  declare someid bigint;
  begin 
    return query(
      -- ops is a proposal or a common sense. s is the same
      select s.commonid, s.proposalid, s.senseid
        ,s.proposalstatus
        ,s.phrase, s.word, s.phantom
        ,s.sdusernickname
        ,s.languageslug
	      from fnonepersonalsense(p_sduserid, p_commonid) ops
        -- actually it is an inner join to the same record
  		  left join vsense_wide as s on s.id = ops.r_senseid
        limit 1); end;
$$;

\echo *** language_and_sense.sql Done
