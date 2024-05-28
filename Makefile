
all: build

build:
	go build cmd/trace/logo-trace.go
	go build cmd/visual/logo-visual.go

run:
	cat example.logo | go run cmd/trace/logo-trace.go

run-visual:
	cat example.logo | go run cmd/visual/logo-visual.go

clean:
	go clean

