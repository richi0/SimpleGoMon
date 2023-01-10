# Simple gomon

Gomon monitors your project directory and runs commands if it detects changes.

## Usage

- Download the binary and place it in your project folder or build it yourself using `go build`.
- If your project is started by `go run main.go` you can use `./gomon` directly.
- Gomon will compile your project and run the binary. If it detects changes it will recompile and rerun.
- If you need custom build commands use the `-setup`, `-update`, or `-teardown` parameters.

```
  -debug
        Activates debug logs
  -root string
        Root folder to monitor (default ".")
  -setup string
        Custom build command (default "go build -o=__temp__")
  -teardown string
        Custom teardown command (default "rm __temp__")
  -types string
        File types to monitor (default "go,html,css,js,tmpl")
  -update string
        Custom run command (default "./__temp__")
```

Only the file types defined in `types` (default `go, html, css, js, tmpl`) are monitored. At startup the `setup` command and the `update` command run once. After every detected file change the update command is stopped if still running and restarted. When the programm stops the `teardown` command is run once.
