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

create or replace view vsense as select tsense.*,
  -- FIXME suboptimal!
  get_language_slug(tsense.languageid) as languageslug
  from tsense;


-- fnPersonalSenses returns all personal senses for the user. If the user is 0 or null,
-- then common senses are returned as well as unparallel personal
-- to copy-paste or complicate this one to have a good select plan for searches.
create or replace function fnpersonalsenses(p_sduserid bigint) 
  returns table(r_originid bigint, r_proposalid bigint, r_countofproposals bigint, r_addedbyme bool)
  language plpgsql as $$
  begin
  if coalesce(p_sduserid, 0) = 0 then
    return query(
      select cast(orig.id as bigint) as r_originid
      ,cast(null as bigint) as r_proposalid
      ,(select count(1) from tsense varic where varic.originid = orig.id) as r_countofproposals
      ,false as r_addedbyme
      from tsense orig where orig.originid is null and orig.ownerid is null); 
  else
    return query(
      select cast(orig.id as bigint) as r_originid
      ,cast(vari.id as bigint) as r_proposalid
      ,(select count(1) from tsense varic where varic.originid = orig.id) as r_countofproposals
      ,case when orig.ownerid = p_sduserid then true else false end as r_addedbyme
      from tsense orig 
      left join tsense vari on orig.id = vari.originid and vari.ownerid = p_sduserid 
      where orig.originid is null); end if; end;
$$;


-- fnOnePersonalSense returns a personal or common sense for the specific sense id
create or replace function fnonepersonalsense(p_sduserid bigint, p_originid bigint) 
  returns table(r_originid bigint, r_proposalid bigint)
  language plpgsql as $$
  begin
  return query(
    select cast(orig.id as bigint) as r_originid, cast(vari.id as bigint) as r_proposalid 
    from tsense orig 
    left join tsense vari on orig.id = vari.originid and vari.ownerid = p_sduserid 
    where orig.id = p_originid and orig.originid is null); end;
$$;

-- fnSavePersonalSense saves the sense. p_evenifidentical must be false for now
create or replace function fnsavepersonalsense(
    p_sduserid bigint, p_originid bigint, p_phrase text, p_word text, p_evenifidentical bool)
  returns table (success bool)
  language plpgsql as $$
  declare v_proposalid bigint;
  declare update_count int;
  begin
  if p_evenifidentical then
    raise exception 'invalid parameter p_evenifidentical'; end if;
  if exists (select 1 from tsense where id = p_originid and word = p_word and phrase = p_phrase) then
    -- nothing differs from the official version, delete our proposal
    delete from tsense where originid = p_originid and owner = p_sduserid;
    return query(select true); 
    return; end if; 
  select ensuresenseproposal(p_sduserid, p_originid) into v_proposalid;
  update tsense set 
  phrase = p_phrase,
  word = p_word
  where id = v_proposalid;
  get diagnostics update_count = row_count;
  if update_count != 1 then
    raise exception 'expected to update just one record, which didn''t hapen'; end if;
  end;
$$;

-- EnsureSenseProposal ensures that a user has his own proposal of a sense. One should not
-- make a proposal of user's unparallel sense.
create or replace function ensuresenseproposal(p_sduserid bigint, p_senseid bigint)
returns table (proposalsenseid bigint) 
language plpgsql as $$
  declare r_senseid bigint;
  declare v_ownerid bigint;
  begin
    lock table themutex;
    select ownerid from tsense where id = p_senseid into v_ownerid;
    if v_ownerid is not null then
      raise exception 
      'You can''t make a proposal of user''s new sense, until it is accepted to the language'; end if;
    select min(id) from tsense 
      where originid = p_senseid and ownerid = p_sduserid
      into r_senseid;
    if r_senseid is not null then 
      return query (select r_senseid); 
      return; end if;
    insert into tsense (languageid, phrase, word, originid, ownerid)
      select languageid, phrase, word, id, p_sduserid 
      from tsense where id = p_senseid returning id into r_senseid;
    if r_senseid is null then
      raise exception 
        'something went wrong, sense cloning failed'; 
    end if;
  return query (select r_senseid);
  end;
$$;

-- this is a mess...
select ensuresenseproposal(1,4);
update tsense set phrase = 'updated sense' where id=5;

-- end of mess

create or replace function explainSenseStatusVsProposals(
    p_id bigint, p_originid bigint, p_sduserid bigint, p_ownerid bigint, p_deleted bool) 
  returns
  table (commonorproposal varchar(128), whos varchar(512), kindofchange varchar(128))
  language plpgsql strict as $$
  declare r_commonorproposal varchar(128);
  declare r_whos varchar(512);
  declare r_kindofchange varchar(128);
begin
  r_commonorproposal = case
    when p_ownerid is null then 'common' 
    else 'proposal' end;
  r_whos = case 
    when p_ownerid is null then '' -- common - irrelevant
    when p_sduserid = p_ownerid then '<my>' 
    else 
      coalesce((select nickname from sduser where id = p_ownerid)
        ,'owner not found') end;
  r_kindofchange = case
    when p_ownerid is null then '' -- common - irrelevant
    when p_originid is null then 'addition'
    when p_delete then 'deletion'
    else 'change' end;
  return query(select r_commonorproposal, r_whos, r_kindofchange); end;
$$;


create or replace function fnsenseorproposalforview(p_sduserid bigint, p_id bigint, p_proposalifexists bool)
returns table (senseorproposalid bigint
  ,originid bigint
  ,phrase text
  ,word varchar(512)
  ,deleted bool
  ,languageslug text
  ,commonorproposal varchar(128)
  ,whos varchar(512)
  ,kindofchange varchar(128)
  )
language plpgsql as $$
  declare v_senseorproposalid bigint;
  declare v_originid bigint;
  declare v_ownerid bigint;
  declare v_deleted bool;
  begin
  if p_proposalifexists then
    select cast(s.id as bigint) as senseorproposalid
      ,cast(s.originid as bigint) as originid
      ,s.ownerid
  	  ,s.deleted 
	    from fnonepersonalsense(p_sduserid, p_id) ops
		  left join tsense as s on s.id = coalesce(ops.r_proposalid, ops.r_originid)
      limit 1
      into v_senseorproposalid, v_originid, v_ownerid, v_deleted;
  else
    select cast(s.id as bigint) senseorproposalid
      ,cast(s.originid as bigint) as originid
      ,s.ownerid
    	,s.deleted 
  	  from tsense as s where s.id = p_id
			limit 1  
      into v_senseorproposalid, v_originid, v_ownerid, v_deleted; end if;
  -- raise exception using message='keys: '||coalesce(v_originid,-1)||','||coalesce(v_senseorproposalid,-2);
  return query(
   select 
      v_senseorproposalid
			,coalesce(v_originid, cast(0 as bigint))
			,s.phrase
			,s.word
			,v_deleted 
			,s.languageslug
      ,essvp.commonorproposal
      ,essvp.whos
      ,essvp.kindofchange
      from vsense as s 
      -- inner join does not work here, I don't know why...
      left join explainSenseStatusVsProposals(
        v_senseorproposalid, v_originid, p_sduserid, v_ownerid, v_deleted) as essvp 
        on 1=1
      where s.id = v_senseorproposalid
      limit 1); end;
$$;

-- tests
create or replace function test_fnsensorproposalforview() returns void
language plpgsql strict as $$
begin
 if not exists (select originid, senseorproposalid from fnsenseorproposalforview(1,1,true) 
  where originid = 0 and senseorproposalid = 1) THEN
   raise exception 'test_fnsensorproposalforview failure 1'; end if; 
 if not exists (select originid, senseorproposalid from fnsenseorproposalforview(1,1,false) 
  where originid = 0 and senseorproposalid = 1) THEN
   raise exception 'test_fnsensorproposalforview failure 2'; end if; 
end;
$$;

select test_fnsensorproposalforview();



\echo *** language_and_sense.sql Done
