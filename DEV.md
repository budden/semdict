# Develop and Quck Start Guide


First to setting variables `.env` from `.env.example`

```bash
# run once at the start of work
make setup

# run every time migrations changes?
make up

# start local server (outside docker) for develop
# make run-local

# start server in docker infra
# use command for restart server after changes
make run-docker

# run once at setup SSL for domain
make initial-setup-ssl
# Should be an `Exit 0` for the certbot container - means that the certificate has been installed successfully
#
# MacBook-Pro-George:semdict gebv$ docker-compose ps | grep certbot
# certbot               certbot certonly --webroot ...   Exit 0
# MacBook-Pro-George:semdict gebv$

# start reverse proxy with configured ssl
make run-proxy
```


Просмотр логов:

docker logs -f semdict-server

Подключение к postgresq нужно делать через докер, но я делаю его прямо в локальной машине, 
```
psql -h localhost -p 5432 -U semdict sduser_db
```
