-- no sense to drop tables here, but we must reset sequences (FIXME)


insert into sduser (nickname, registrationemail, salt, hash, registrationtimestamp)
values ('tsar','tsar@example.com','Fr5ISNGBVjsNUX1C5Q--Vw',
        'qZwRJrl9O_VwBuQKJrMTYW1bh4zqNUAhMcmPyh5kBpo',current_timestamp);
-- password is aA$9bbbb
insert into sduser_profile (id) values (1);
select grantuserprivilege(1,1);

insert into sduser (nickname, registrationemail, salt, hash, registrationtimestamp)
values ('user2','user2@example.com','Fr5ISNGBVjsNUX1C5Q--Vw',
        'qZwRJrl9O_VwBuQKJrMTYW1bh4zqNUAhMcmPyh5kBpo',current_timestamp);
-- password is aA$9bbbb
select grantuserprivilege(2,1);

insert into tlanguage (id, slug, commentary, ownerid)
values (2,'ру','русский',1), (3,'中','中文',null)
     ,(4, 'ру-1С', '1С предприятие',2)
     ,(5, 'ру-excel', 'Microsoft Excel',null);

insert into sduser_profile(id,favorite_tlanguageid) values (2,4);

insert into tsense (oword, theme, phrase, ownerid)
VALUES
('golang','ЯП','Язык программирования, созданный гуглом в 2000s', 1);

insert into tsense (oword, theme, phrase, ownerid)
VALUES
('canvas','ГПИ','пространство на экранной форме, на котором можно рисовать', 1);

insert into tsense (oword, theme, phrase, ownerid)
VALUES
('operator','ЯП','+, -, *, /, >>, «и», «или» и тому подобное', 1);

insert into tlws (languageid, senseid, word) values
(2,2,'холст'), (4,2,'канва'),(5,2,'Гоу'),(2,3,'операция'),(4,3,'оператор');
