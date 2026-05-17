.PHONY: run build init clean

run:
	go run ./cmd/user-system/

build:
	go build -o bin/user-system ./cmd/user-system/

init:
	bash scripts/init.sh

clean:
	rm -rf bin/