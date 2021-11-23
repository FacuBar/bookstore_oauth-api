mysql:
	docker run --name oauth-mysql -p 9001:3306 -e MYSQL_ROOT_PASSWORD=secret -d mysql

createdb: 
	docker exec -it oauth-mysql mysql --user='root' --password='secret' --execute='CREATE DATABASE oauth_db'

dropdb:
	docker exec -it oauth-mysql mysql --user='root' --password='secret' --execute='DROP DATABASE oauth_db'

migrateup:
	migrate -path migration/ -database "mysql://root:secret@tcp(localhost:9001)/oauth_db" -verbose up

migratedown: 
	migrate -path migration/ -database "mysql://root:secret@tcp(localhost:9001)/oauth_db" -verbose down

server:
	go run cmd/main.go

.PHONY: mysql createdb dropdb	migrateup migratedown server