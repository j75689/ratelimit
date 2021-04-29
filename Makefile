.PHONY: tools run build-image deploy-local

tools:
	@go get github.com/google/wire/cmd/wire

run:
	@go run main.go http

build-image:
	@read -p "Enter Image Name: " IMAGE_NAME; \
	docker build . -f ./build/Dockerfile -t "$$IMAGE_NAME"

deploy-local:
	@docker-compose -f ./deployment/docker-compose/docker-compose.yml build
	@docker-compose -f ./deployment/docker-compose/docker-compose.yml up