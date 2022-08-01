dep:
	go mod download

gen:
	go install github.com/google/wire/cmd/wire@v0.5.0
	wire ./cmd/gowarden

build: dep gen
	CGO_ENABLED=0 \
	go build \
	--trimpath \
	--ldflags "-s -w" \
	-o ./gowarden \
	./cmd/gowarden

dev:
	docker compose up -d --build warden
	docker compose logs warden -f