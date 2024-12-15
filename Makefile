# go makefile

program != basename $$(pwd)

version != cat VERSION

gitclean = if git status --porcelain | grep '^.*$$'; then echo git status is dirty; false; else echo git status is clean; true; fi

build: fmt
	fix go build

fmt: go.sum
	fix go fmt .

go.mod:
	go mod init

go.sum: go.mod
	go mod tidy

test:
	fix -- go test -v .

release:
	@$(gitclean)
	gh release create v$(version) --notes "v$(version)"

clean:
	rm -f $(program)
	go clean
	rm -f /tmp/*.recording

sterile: clean
	go clean -r
	go clean -cache
	go clean -modcache
	rm -f go.mod go.sum
