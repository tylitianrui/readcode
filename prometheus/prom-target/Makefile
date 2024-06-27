

.PHONY: mod
mod:
	@echo ">> fetch dependancy..."
	go  mod tidy
	go  mod vendor




.PHONY: run
run: mod
	@echo ">> api run... "
	go  run  main.go  run


