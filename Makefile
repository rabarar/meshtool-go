
.PHONY: all
all: tidy audit check-git-clean


.PHONY: tidy
tidy:
	@echo "tidy and fmt..."
	go mod tidy -v
	go fmt ./...


.PHONY: audit
audit:
	@echo "running audit checks..."
	go mod verify
	go vet ./...
	go list -m all
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

.PHONY: test
test:
	@echo "running tests..."
	go test -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html


.PHONY: check-git-clean
check-git-clean:
	@echo "Checking git status..."
	@git diff --quiet || (echo "Git working directory is not clean" && exit 1)
	@git diff --cached --quiet || (echo "Git index is not clean" && exit 1)