install:
	@go install -ldflags="-s -w" .

build:
	docker build -t btcs .