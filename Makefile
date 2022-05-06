build_w:
	env GOOS=windows GOARCH=amd64 go build -o=bin/gomon-windwos-amd64.exe
build_a:
	env GOOS=darwin GOARCH=amd64 go build -o=bin/gomon-darwin-amd64
build_l:
	env GOOS=linux GOARCH=amd64 go build -o=bin/gomon-linux-amd64
build_all: build_w build_a build_l