up:
	docker-compose up --detach --force-recreate --renew-anon-volumes --remove-orphans postgres
	@echo "➡️ Wait 30 seconds to start the infrastructure (pg)"
	sleep 30
	@echo "➡️ Configure postgres & migrates"
	docker-compose build --force-rm configure
	docker-compose up configure
	@echo "➡️ Infrastructure is started and configured"
	@echo "➡️ Done"

down:
	docker-compose down --volumes --remove-orphans

run:
	go run ./main.go
