version = 0.0.0
filename = gomon

test:
		go test ./...

vet:
		go vet ./...

build_w: test vet
		env GOOS=windows GOARCH=amd64 go build -o=bin/$(filename)-$(version)-windwos-amd64.exe

build_a: test vet
		env GOOS=darwin GOARCH=amd64 go build -o=bin/$(filename)-$(version)-darwin-amd64

build_l: test vet
		env GOOS=linux GOARCH=amd64 go build -o=bin/$(filename)-$(version)-linux-amd64

build_r: test vet
		env GOOS=linux GOARCH=arm GOARM=7 go build -o=bin/$(filename)-$(version)-linux-raspberry

build_all: clean build_w build_a build_l build_r

clean:
		rm -rf bin/*