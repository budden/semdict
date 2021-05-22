up:
	@echo "➡️ Configure postgres & migrates"
	docker-compose up configure
	@echo "➡️ Infrastructure is started and configured"
	@echo "➡️ Done"

setup:
	docker-compose up --detach --force-recreate --renew-anon-volumes --remove-orphans postgres
	@echo "➡️ Wait 30 seconds to start the infrastructure (pg)"
	sleep 30
	@echo "➡️ Rebuild configure service"
	docker-compose build --force-rm configure

initial-setup-ssl:
	@echo "➡️ Primary setup SSL (once for domain)"
	docker-compose up -d certbot

rerun-ssl:
	@echo "➡️ TODO"

run-proxy:
	@echo "➡️ Launch reverse-proxy"
	docker-compose up -d webserver443

run-docker: up
	docker-compose up -d --build semdict-server

down:
	docker-compose down --volumes --remove-orphans

run-local: up
	go run ./main.go
