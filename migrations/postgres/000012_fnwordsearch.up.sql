create or replace function fnwordsearch(
    p_sduserid bigint, p_wordpattern text, p_senseid integer, p_offset bigint, p_limit bigint)
    returns table (
                      senseid integer,
                      oword varchar(512),
                      theme varchar(512),
                      phrase text,
                      lwsjson jsonb,
                      hasfavoritelanguagetranslation bigint)
    language plpgsql as $$
begin
/* надо задать либо p_senseid, либо шаблон. Если всё пустое, то найдутся все слова */
    if coalesce(p_wordpattern,'') = '' then
     p_wordpattern = '%';
    end if;
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
        where (coalesce(p_senseid,0)<>0 and tsense.id = p_senseid
         or coalesce(p_senseid,0)=0 and tsense.oword like p_wordpattern)
        order by tsense.oword, tsense.theme, senseid
        offset p_offset limit p_limit); return; end;
$$;

