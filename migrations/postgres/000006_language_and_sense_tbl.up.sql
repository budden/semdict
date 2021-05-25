drop table if exists tlws cascade;
drop table if exists tsense cascade;
drop table if exists tlanguage cascade;
--*/

create table tlanguage (
                           id serial primary KEY,
                           slug varchar(128) not null unique,
                           commentary text,
                           ownerid bigint references sduser
);

comment on table tlanguage is 'tlanguage-это язык, или диалект, или источник перевода';
comment on column tlanguage.slug is 'slug-это сокращенный идентификатор, читаемый человеком';
comment on column tlanguage.commentary is 'комментарий - это полное описательное название диалекта';
comment on column tlanguage.ownerid is 'ownerid указывает владельца языка. Если NULL, язык является "общим", так что каждый может добавлять записи tlws, ссылающиеся на язык';

alter table sduser_profile add constraint fk_sduser_profile_favorite_tlanguageid
    foreign key (favorite_tlanguageid) references tlanguage (id);

create table tsense (
                        id serial primary KEY,
                        oword varchar(512) not null,
                        theme varchar(512) not null,
                        phrase text not null,
                        ownerid bigint not null references sduser
);

comment on table tsense is 'tsense хранит запись для определённого смысла английского слова. (Смысл X Язык X Слово) - это отношение (много X много X много). ';
comment on column tsense.id is 'id служит слизняком(slug) смысла';
comment on column tsense.oword is 'oword = оригинальное слово. Английское слово или словосочетание';
comment on column tsense.phrase is 'Фраза на «общерусском языке», выражающая один специфический смысл этого слова';
comment on column tsense.theme is 'Тема полезна в качестве критерия поиска в сочетании со словом. Тема задана на русском языке';
comment on column tsense.ownerid is 'Обладатель чувства. Обычно царь владеет чувствами, за исключением новых.';

create table tlws (
                      id serial primary KEY,
                      languageid bigint not null references tlanguage,
                      word varchar(512) not null,
                      senseid bigint not null references tsense,
                      commentary text not null default '',
                      ownerid bigint null references sduser
);

-- one can (in a future) have several possible translations for a sense
create unique index tlws_key on tlws (languageid, senseid, word);

comment on table tlws is 'tlws - это отношение язык-слово-смысл, то есть вариант перевода';
comment on column tlws.id is 'id является суррогатным ключом и служит в качестве слизняка(slug)';
comment on column tlws.languageid is 'диалект, на который мы переводим смысл';
comment on column tlws.word is 'перевод, то есть слово или фраза на указанном языке, которые могут быть использованы для выражения смысла';
comment on column tlws.commentary is 'комментарий к выбранному переводу';
comment on column tlws.ownerid is 'Владелец отношения. Если нет, подразумевается владелец языка.';
