.PHONY: all dotler vet test clean lint race bench dep

# Credit: https://github.com/rightscale/go-boilerplate/blob/master/Makefile
DEPEND=golang.org/x/tools/cmd/cover github.com/Masterminds/glide github.com/golang/lint/golint

dep:
	go get -u $(DEPEND)
	glide install

all: dotler

dotler: dep vet
	go build -v -o build/dotler

vet: lint
	go vet ./dotler ./tests

test: dep
	cd tests && go test -v .

bench: dep
	cd tests && go test -v -run=XXX -bench=. -benchtime=60s

race: dep
	go build -race -v -o build/dotler.race

clean:
	@rm -f build/dotler build/dotler.race

lint:
	golint dotler tests

doc: pdf
	godoc -http=:6060 -index

pdf:
	@bash -c 'tail +3 README.md | pandoc - --latex-engine=xelatex -o README.pdf'

analyse: vet lint
