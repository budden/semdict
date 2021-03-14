--/*
\connect sduser_db
\set ON_ERROR_STOP on
--*/ 

/*create or replace function get_language_slug(p_languageid int) returns text
 language plpgsql strict as $$
 declare v_result text;
 declare v_len_limit int;
  begin
  
  select slug from tlanguage 
  where id = p_languageid is null 
  into v_result;

  return v_result;
  end;
 $$;*/

create or replace view vsense_wide as select s.*
  ,u.nickname as sdusernickname
  from tsense s left join sduser u on s.ownerid=u.id;

\echo *** language_and_sense_view_1.sql Done
