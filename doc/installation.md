# "semdict" - Регистрация пользователей на основе электронной почты в golang + postgresql

## Требования
В настоящее время мы устанавливаемся только через сборку из исходников. Это неоптимально для серверов, 
но сейчас мы пытаемся сэкономить усилия на разработке :) Здесь перечислены не все необходимые условия, 
следуйте руководству, и вы найдёте больше.

### Golang
- golang 1.16.2 (другие версии не тестировались), инструкции см. на домашней странице golang,
на хостинговом компьютере, управляемом через ssh мы сделали следующее:
```
cd ~
mkdir install_golang
cd install_golang
wget https://dl.google.com/go/go1.16.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.16.2.linux-amd64.tar.gz
mkdir ~/go

# способ настройки параметров env зависит от вашей оболочки и от того, 
# локальная это машина или удалённая.
# это для дистанционного управления
vi ~/.profile
# добавьте эти строки в конец ~/.profile
PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
# конец строк для добавления в ~/.profile

# выйти из системы и войти снова

# должно сработать следующее
go version 
# следующее должно быть непустым
echo $GOPATH
```

### Postgresql
- postgresql 9.6.10 (другие версии не тестировались). В Debian 9 это просто `apt-get install postgresql`
- расширение tcl, `apt-get install postgresql-pltcl`


## Построение

```
go get -d github.com/stretchr/testify/assert
cd $GOPATH/src/github.com/stretchr/testify/assert
git checkout v1.3.0
go get ./... 


go get github.com/budden/semdict
cd $GOPATH/src/github.com/budden/semdict

# FIXME используйте вендоринг вместо этого!
go get ./...

go generate
go build
```

### CKEditor
```
cd static
curl -L -O https://download.cksource.com/CKEditor/CKEditor/CKEditor%204.11.3/ckeditor_4.11.3_basic.zip
unzip ckeditor_4.11.3_basic.zip
mv ckeditor ckeditor_4.11.3_basic
cd .. 
```

## Модульные тесты
```
cd pkg
go test ./...
cd ..
```

## Настройка базы данных

### Разрешить доступ для корня
Мы запускаем службу от имени root. Может быть, это позор.
```
sudo su - postgres

psql
# в psql:
create role root;
alter role root login;
alter role root createdb;
\quit
# мы покинули psql
exit
# мы вышли из sudo su - postgres и теперь снова находимся 
# в учётной записи пользователя нашего сервера 

sudo vi /etc/postgresql/9.6/main/pg_hba.conf

# Добавьте в начало значимых строк файла следующую строку, 
# отформатированную по аналогии с другими (через пробелы)

host    all             root            127.0.0.1/32            trust

# Перезапустите postgres и проверьте, всё ли у нас в порядке
sudo service postgresql restart
sudo psql postgres://localhost/postgres

# должно появиться приветственное сообщение и приглашение psql. 
# Теперь выйдите из psql
\quit
```

### Создание базы данных

```
cd $GOPATH/src/github.com/budden/semdict
# просто загрузка скрипта не работает, потому что (я думаю) я забыл описать, как включить вызов shell из скриптов Postgres.
# Так что 
vi sql/recreate_sduser_db.sql
# в определении :thisdir, заменить `echo $GOPATH...` с его реальным значением, то есть /root/go... (без кавычек)
# ESC :wq!

sudo psql -f sql/recreate_sduser_db.sql postgres://localhost/postgres

# Должен пройти без ошибок и завершиться "CREATE VIEW"
```

### Тестовый запуск в качестве приложения

Создайте файл semdict.config.json следующим образом:
```
{"Comment":["My"
 ,"config"]
,"SiteRoot": "localhost"
,"ServerPort": "8085"
,"SMTPServer":""
,"SMTPUser":""
,"SMTPPassword":""
,"SenderEMail":""
,"PostgresqlServerURL": "postgresql://localhost:5432"
,"TLSCertFile": ""
,"TLSKeyFile": ""
}

```
Теперь запустите приложение:
```
sudo ./semdict
```
Зайдите на сайт http://your_server:8085 - он должен приветствовать сообщение. Убейте приложение с помощью ^C.


### Установка
mc  
Мне нужно было установить пакет pkg-config для успешного запуска, ваш
способ может быть иным.
```
sudo sh install.sh
```

### Запись фактического файла конфигурации
```
sudo cp /etc/semdict/semdict.config.json.example /etc/semdict/semdict.config.json
sudo vi /etc/semdict/semdict.config.json
```
Заполните свой конфиг всеми необходимыми данными.

