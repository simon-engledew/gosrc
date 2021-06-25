## gosrc

Prints all the source files used to build a Go target:

`gosrc .`

Can be useful in Makefiles for go build targets:

```makefile
bin/command: $(shell gosrc cmd/command)
    go build -o $@ ./cmd/command
```
