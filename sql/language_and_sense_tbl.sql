--/*
\connect sduser_db
\set ON_ERROR_STOP on
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

comment on table tlanguage is 'tlanguage is a language, or a dialect, or a source of translation';
comment on column tlanguage.slug is 'slug is a human-readable abbreviated identifier';
comment on column tlanguage.commentary is 'commentary is a full descriptive name of the dialect';
comment on column tlanguage.ownerid is 'ownerid specifies an owner of the language. If NULL, language is "common", so that everyone can add tlws records referencing the language';

insert into tlanguage (id, slug, commentary, ownerid) 
  values (2,'ру','русский',1), (3,'中','中文',null)
    ,(4, 'ру-1С', '1С предприятие',2)
    ,(5, 'ру-excel', 'Microsoft Excel',null);

alter table sduser_profile add constraint fk_sduser_profile_favorite_tlanguageid
  foreign key (favorite_tlanguageid) references tlanguage (id);

create table tsense (
  id serial primary KEY,
  oword varchar(512) not null,
  theme varchar(512) not null,
  phrase text not null,
  ownerid bigint not null references sduser
);

comment on table tsense is 'tsense stored a record for a specific sense of an English word. (Sense X Language X Word) is a (many X many X many) relation. ';
comment on column tsense.id is 'id serves as a slug of a sense';
comment on column tsense.oword is 'oword = original word. English word or a collocation';
comment on column tsense.phrase is 'Phrase in «Common Russian» that expesses one specific sense of the word';
comment on column tsense.theme is 'Theme is useful as a search criteria in a combination with the word. Theme is set in Russian';
comment on column tsense.ownerid is 'Owner of the sense. Normally, tzar owns senses, except for new ones.';

insert into tsense (oword, theme, phrase, ownerid)
  VALUES
  ('golang','ЯП','Язык программирования, созданный гуглом в 2000s', 1);

insert into tsense (oword, theme, phrase, ownerid)
  VALUES
  ('canvas','ГПИ','пространство на экранной форме, на котором можно рисовать', 1);

insert into tsense (oword, theme, phrase, ownerid)
  VALUES
  ('operator','ЯП','+, -, *, /, >>, «и», «или» и тому подобное', 1);

create table tlws (
  id serial primary KEY,
  languageid bigint not null references tlanguage,
  senseid bigint not null references tsense,
  word varchar(512) not null,
  commentary text null,
  ownerid bigint null references sduser
);

comment on table tlws is 'tlws is a language-word-sense relation, that is, translation variant';
comment on column tlws.id is 'id is a surrogate key and serves as slug';
comment on column tlws.languageid is 'dialect we are translating the sense to';
comment on column tlws.word is 'translation, that is, word or a phrase in the language referenced which can be used to express a sense';
comment on column tlws.commentary is 'comment establishing a chosen translation';
comment on column tlws.ownerid is 'Owner of the relation. If none, language''s owner is implied.';

insert into tlws (languageid, senseid, word) values
  (2,2,'холст'), (4,2,'канва'),(5,2,'Гоу'),(2,3,'операция'),(4,3,'оператор');


\echo *** language_and_sense_tbl.sql Done