gosrc: $(shell ./gosrc .)
	go build -o $@ .
