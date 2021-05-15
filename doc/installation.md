# "semdict" - Регистрация пользователя по электронной почте в golang + postgresql

## Требования
В настоящее время мы устанавливаем только через сборку из источников. Это неоптимально для серверов,
но сейчас мы пытаемся сэкономить усилия по разработке :) Не все необходимые условия перечислены здесь,
следуйте руководству, и вы найдёте больше.

### Golang
- golang 1.16.2 (другие версии не тестировались), инструкции см. на домашней странице golang, мы сделали следующее
на хостинговом компьютере, управляемом через ssh:
```
cd ~
mkdir install_golang
cd install_golang
wget https://dl.google.com/go/go1.16.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.16.2.linux-amd64.tar.gz
mkdir ~/go

# the way env vars are set up depends on your shell and if it is local
# or remote machine. this one is for remote
vi ~/.profile
# add these lines at the end of ~/.profile
PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
# end of lines to add to ~/.profile

# logout and login again

# the following must work
go version 
# the following must be non-empty
echo $GOPATH
```

### Postgresql
- postgresql 9.6.10 (другие версии не тестировались). В Debian 9 это просто `apt-get install postgresql`
- расширение tcl, `apt-get install postgresql-pltcl`

## Building

```
go get -d github.com/stretchr/testify/assert
cd $GOPATH/src/github.com/stretchr/testify/assert
git checkout v1.3.0
go get ./... 


go get github.com/budden/semdict
cd $GOPATH/src/github.com/budden/semdict

# FIXME use vendoring instead!
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

### Разрешить доступ для root
Мы запускаем службу от имени root. Может быть, это позор.
```
sudo su - postgres

psql
# in psql:
create role root;
alter role root login;
alter role root createdb;
\quit
# we left psql
exit
# we left sudo su - postgres and 
# now we're again in our server user's account 

sudo vi /etc/postgresql/9.6/main/pg_hba.conf

# Add the following line at the beginning of meaningful 
# lines in the file, formatted by an 
# analogy with others (by spaces)
host    all             root            127.0.0.1/32            trust

# Restart postgres and check if we're ok
sudo service postgresql restart
sudo psql postgres://localhost/postgres

# psql welcome message and prompt must appear. 
# Now quit psql
\quit
```

### Создание базы данных

```
cd $GOPATH/src/github.com/budden/semdict
# just loading the script does not work, because (I guess) I forgot to describe how to enable calling shell from Postgres scripts.
# So 
vi sql/recreate_sduser_db.sql
# in the definition of the :thisdir, replace `echo $GOPATH...` with its actual value, that is, /root/go... (no quotes)
# ESC :wq!

sudo psql -f sql/recreate_sduser_db.sql postgres://localhost/postgres

# Must pass w/o errors and end with "CREATE VIEW"
```

### Тестовый запуск в качестве приложения

Create semdict.config.json like this:
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
Доступ http://your_server:8085 - там должно быть приветственное сообщение. Убить приложение с помощью ^C.


### Установка
mc
Мне нужно было установить пакет pkg-config до успешного запуска, ваш
пробег может варьироваться.
```
sudo sh install.sh
```

### Write actual config file
```
sudo cp /etc/semdict/semdict.config.json.example /etc/semdict/semdict.config.json
sudo vi /etc/semdict/semdict.config.json
```
Заполните свою конфигурацию всеми необходимыми данными.
