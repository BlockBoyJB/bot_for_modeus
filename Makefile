compose-up:
	docker-compose up -d --build

compose-down:
	docker-compose down

mocks:
	mockgen -source=pkg/crypter/crypter.go -destination=internal/mocks/cryptermocks/crypter.go -package=cryptermocks
	mockgen -source=internal/repo/repo.go -destination=internal/mocks/repomocks/repo.go -package=repomocks

#mockgen -source=pkg/modeus/modeus.go -destination=internal/mocks/modeusmocks/modeus.go -package=modeusmocks
#mockgen -source=internal/parser/parser.go -destination=internal/mocks/parsermocks/parser.go -package=parsermocks
#mockgen -source=internal/service/service.go -destination=internal/mocks/servicemocks/service.go -package=servicemocks
#mockgen -source=pkg/bot/context.go -destination=internal/mocks/botmocks/context.go -package=botmocks
#mockgen -source=internal/service/service.go -destination=internal/mocks/servicemocks/service.go -package=servicemocks

mongo-tests:
	docker run --name mongo --rm -d -p 27017:27017 mongo:5.0-rc-focal

redis-tests:
	docker run --name redis --rm -d -p 6379:6379 redis:latest

init-test-containers: mongo-tests redis-tests

stop-test-containers:
	docker stop mongo
	docker stop redis

init-tests:
	go test -v ./...

tests: init-test-containers init-tests stop-test-containers
