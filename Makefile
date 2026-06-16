BINARY := go-local-git-stats-cli
COVERAGE_OUT := coverage.out
COVERAGE_HTML := coverage.html

.PHONY: build test coverage coverage-html clean

build:
	go build -o $(BINARY) .

test:
	go test ./...

# Run tests with coverage and print a per-function summary.
coverage:
	go test -coverprofile=$(COVERAGE_OUT) ./...
	go tool cover -func=$(COVERAGE_OUT)

# Generate a browsable HTML coverage report.
coverage-html: coverage
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Wrote $(COVERAGE_HTML)"

clean:
	rm -f $(BINARY) $(COVERAGE_OUT) $(COVERAGE_HTML)
