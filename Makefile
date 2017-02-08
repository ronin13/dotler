.PHONY: all dotler vet test clean lint race bench dep

# Credit: https://github.com/rightscale/go-boilerplate/blob/master/Makefile
DEPEND=golang.org/x/tools/cmd/cover github.com/Masterminds/glide github.com/golang/lint/golint

dep:
	go get -u -v $(DEPEND)
	glide install

all: dotler

dotler: dep vet
	go build -v -o dotler

vet: lint
	go vet *.go

test: dotler
	go test -v .

bench: dotler
	go test -v -run=XXX -bench=. -benchtime=60s

race: dep
	go build -race -v -o dotler

clean:
	@rm -f dotler

lint:
	golint *.go

doc: pdf
	godoc -http=:6060 -index

pdf:
	@pandoc README.md --latex-engine=xelatex -o README.pdf

analyse: vet lint
