test: build
	go test tests/db_test.go
build: main.c
	mkdir -p bin
	gcc main.c -o bin/db
clean:
	rm -rf bin/*
