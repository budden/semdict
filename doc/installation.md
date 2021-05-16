# "semdict" - Регистрация пользователя по электронной почте в golang + postgresql

## Требования
В настоящее время установка идёт только через сборку из источников. Это неоптимально для серверов,
но сейчас мы пытаемся сэкономить усилия по разработке :) Не все необходимые предпосылки перечислены здесь,
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

# способ настройки параметров env зависит от вашей оболочки и от того, является ли она локальной или удаленной машиной. 
# это для дистанционного управления
vi ~/.profile
# добавьте эти строки в конце ~/.profile
PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
# конец строк для добавления ~/.profile

# выход из системы и повторный вход в систему

# должно сработать следующее
go version 
# следующее должно быть непустым
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

# FIXME вместо этого используйте вендоринг!
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
# мы ушли из psql
exit
# мы покинули sudo su - postgres и 
# теперь снова находимся в учетной записи пользователя нашего сервера 

sudo vi /etc/postgresql/9.6/main/pg_hba.conf

# Добавьте следующую строку в начале значимых строк в файле, 
# отформатированном по аналогии с другими (пробелами)
host    all             root            127.0.0.1/32            trust

# Перезагрузите postgres и проверьте, все ли у нас в порядке
sudo service postgresql restart
sudo psql postgres://localhost/postgres

# должно появиться приветственное сообщение psql и приглашение. 
# Теперь выйдите из psql
\quit
```

### Создание базы данных

```
cd $GOPATH/src/github.com/budden/semdict
# просто загрузка скрипта не работает, потому что (я думаю) Я забыл описать, как включить вызов оболочки из сценариев Postgres.
# Так 
vi sql/recreate_sduser_db.sql
# в определении :thisdir замените `echo $GOPATH...` на его фактическое значение, то есть /root/go... (без кавычек)
# ESC :wq!

sudo psql -f sql/recreate_sduser_db.sql postgres://localhost/postgres

# Должен пройти без ошибок и закончиться "CREATE VIEW"
```

### Тестовый запуск в качестве приложения

Создайте semdict.config.json следующим образом:
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
Доступ http://your_server:8085 - там должно быть приветственное сообщение. Убить приложение можно с помощью ^C.

### Установка
mc
Мне нужно было установить пакет pkg-config до успешного запуска, ваш
пробег может варьироваться.
```
sudo sh install.sh
```

### Написать фактический конфигурационный файл
```
sudo cp /etc/semdict/semdict.config.json.example /etc/semdict/semdict.config.json
sudo vi /etc/semdict/semdict.config.json
```
Заполните свою конфигурацию всеми необходимыми данными.
