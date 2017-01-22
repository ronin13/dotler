.PHONY: all dotler vet test clean lint race bench

all: dotler

dotler: vet
	go build -v -o dotler

vet: lint
	go vet *.go


test: dotler
	go test -v .

bench: dotler
	go test -v -run=XXX -bench=. -benchtime=60s

race:
	go run -race dotler.go

clean:
	@rm -f dotler


lint:
	golint *.go

doc:
	godoc -http=:6060 -index
