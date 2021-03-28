# "semdict" - E-mail based user registration in golang + postgresql

## Requirements
We're only currently installing via building from sources. That is suboptimal for servers, 
but we trying to save development effort just now :) Not all prerequisites are listed here, 
follow the manual and you'll find more.

### Golang
- golang 1.11.6 (other versions not tested), see golang home page for the instructions, we did the following
on the hosting PC controlled via ssh:
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
- postgresql 9.6.10 (other versions not tested). On Debian 9, it's just `apt-get install postgresql`
- tcl extension, `apt-get install postgresql-pltcl`


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

## Unit tests
```
cd pkg
go test ./...
cd ..
```

## Setup database

### Allow access for a root
We run service as root. Maybe it's a shame.
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

### Create a database

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

### Test run as an application

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
Now run application:
```
sudo ./semdict
```
Access http://your_server:8085 - it should welcome message. Kill app with ^C.


### Install
mc
I needed to install pkg-config package before this run successfully, your
mileage my vary.
```
sudo sh install.sh
```

### Write actual config file
```
sudo cp /etc/semdict/semdict.config.json.example /etc/semdict/semdict.config.json
sudo vi /etc/semdict/semdict.config.json
```
Fill your config with all the data you need.
