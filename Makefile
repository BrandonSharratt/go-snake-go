get:
	@echo ">> Getting any missing dependencies.."
	go get -t ./...
.PHONY: get

install:
	glide install
	go build -o starter-snake-go .
	chmod ugo+x starter-snake-go
.PHONY: install

run: install
	./starter-snake-go server
.PHONY: run

test:
	go test ./...
.PHONY: test

fmt:
	@echo ">> Running Gofmt.."
	gofmt -l -s -w .
