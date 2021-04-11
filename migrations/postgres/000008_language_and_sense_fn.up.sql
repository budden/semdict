-- fncanuserchangetlws says if the user can change this tlws
create or replace function fncanuserchangetlws(
    p_sduserid bigint, p_tlws_ownerid bigint, p_tlanguage_ownerid bigint)
    returns int
    language plpgsql immutable as $$
declare tlws_ownerid bigint;
begin
    if coalesce(p_sduserid,0) = 0 then return 0; end if;
    if coalesce(p_tlws_ownerid,0) <> 0 then
        tlws_ownerid = p_tlws_ownerid;
    elsif coalesce(p_tlanguage_ownerid,0) <> 0 then
        tlws_ownerid = p_tlanguage_ownerid;
    else
        tlws_ownerid = 0;
    end if;
    if p_sduserid = tlws_ownerid then
        return 1;
    elsif tlws_ownerid = 0 then
        return 1;
    else
        return 0; end if;
end;
$$;


create or replace function fnsavelws(
    p_sduserid bigint,
    p_lwsid bigint,
    p_languageid bigint,
    p_word varchar(512),
    p_senseid bigint,
    p_commentary text,
    action varchar(64) -- create, save, delete
)
    returns table (r_tlwsid bigint)
    language plpgsql as $$
declare v_dialect_ownerid bigint;
    declare v_ownerid bigint;
    declare v_affected_count int;

begin

    select ownerid from tlanguage
    where id = p_languageid
    into v_dialect_ownerid;

    -- check rights (later)
    -- perform an action

    if coalesce(v_dialect_ownerid,0) = p_sduserid then
        v_ownerid = null;
    else
        v_ownerid = p_sduserid; end if;

    if action = 'create' then
        insert into tlws
        (languageid, word, senseid, commentary, ownerid)
        values
        (p_languageid, p_word, p_senseid, p_commentary, v_ownerid)
        returning id into p_lwsid;

        get diagnostics v_affected_count = row_count;
        if v_affected_count != 1 then
            raise exception 'expected to insert exactly one record, which didn''t happen'; end if;
        return query (select p_lwsid);
    elsif action='save' then
        if p_lwsid is null then
            raise exception 'p_lwsid is null'; end if;
        update tlws
        set word = p_word, commentary = p_commentary, ownerid = v_ownerid
        where id = p_lwsid and
          --- next conditions are an additional checks
                languageid = p_languageid and senseid = p_senseid;
        if v_affected_count != 1 then
            raise exception 'expected to update exactly one record, which didn''t hapen'; end if;
        return query (select p_lwsid);
    elsif action='delete' then
        raise exception 'sorry, not implemented yet';
    else
        raise exception 'bad action'; end if;
    return; end;
$$;

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


create or replace function fnwordsearchmasterrecord(
    p_sduserid bigint)
    returns table (
                      favoritelanguageid int,
                      favoritelanguageslug varchar(128))
    language plpgsql as $$
begin
    if coalesce(p_sduserid,0) = 0 then
        return query(select cast(0 as int), cast('' as varchar(128)));
        return;
    else
        return query(select
                         coalesce(sduser_profile.favorite_tlanguageid,0),
                         coalesce(tlanguage.slug,'')
                     from sduser_profile
                              left join tlanguage on sduser_profile.favorite_tlanguageid = tlanguage.id
                     where sduser_profile.id = p_sduserid);
        return;
    end if; end
$$;


create or replace function fnwordsearch(
    p_sduserid bigint, p_wordpattern text, p_offset bigint, p_limit bigint)
    returns table (
                      senseid integer,
                      oword varchar(512),
                      theme varchar(512),
                      phrase text,
                      lwsjson jsonb,
                      hasfavoritelanguagetranslation bigint)
    language plpgsql as $$
begin
    return query(
        select tsense.id as senseid,
               tsense.oword,
               tsense.theme,
               tsense.phrase,
               (select jsonb_agg(row_to_json(detail))
                from
                    (select tlws.*, tlanguage.slug languageslug,
                            fncanuserchangetlws(p_sduserid,tlws.ownerid,tlanguage.ownerid) as canedit,
                            case when tlws.languageid = coalesce(sduser_profile.favorite_tlanguageid,0)
                                     then 0 else 1 end as prefer_favorite_language
                     from tlws
                              left join tlanguage on tlws.languageid = tlanguage.id
                     where tlws.senseid=tsense.id order by prefer_favorite_language, languageslug
                    ) as detail
               ) as lwsjson,
               (select count(1)
                from tlws
                where tlws.senseid=tsense.id and tlws.languageid = sduser_profile.favorite_tlanguageid) as hasfavoritelanguagetranslation
        from tsense
                 left join sduser_profile on sduser_profile.id = p_sduserid
        where tsense.oword like p_wordpattern
        order by tsense.oword, tsense.theme, senseid
        offset p_offset limit p_limit); return; end;
$$;

