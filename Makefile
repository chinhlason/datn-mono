build:

run:
	docker-compose up -d
	go run .

stop:
	docker-compose down
	pkill -f go