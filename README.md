# Simple gomon
Gomon monitors your project directory and runs commands if it detects changes.

## Usage
- Download the binary and place it in your project folder or build it yourself using `go build`
- If your project is started by `go run main.go` you can use `./gomon` directly
- Gomon will compile your project and run the binary. If it detects changes it will recompile and rerun
- If you need custom build commands use the `-build`, `-run`, and `-teardown` parameters

## Parameters of ./gomon:
- build string
    - Custom build command (default "go build -o=__temp__")
- run string
    - Custom run command (default "./__temp__")
- tearDown string
    - Custom teadown command (default "rm __temp__")
- types string
    - File types to monitor (default "go,html,css,js,tmpl")

Only the files defined in `types` (default "go,htmp,css,js,tmpl") are monitored. At startup and if a change is detected the build command and the run command are executed. When the program is stopped the teardown command is executed to clean up.