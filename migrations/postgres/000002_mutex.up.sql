create table themutex (id int);
comment on table themutex is
    'This table is locked in each operation on sdsusers_db that involves writing to avoid deadlocks';
insert into themutex (id) values (0);
