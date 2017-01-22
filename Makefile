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

doc: pdf
	godoc -http=:6060 -index

pdf:
	@pandoc README.md --latex-engine=xelatex -o README.pdf

analyse: vet lint
