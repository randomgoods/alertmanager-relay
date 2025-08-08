PREFIX?=/usr
BINDIR?=$(PREFIX)/bin
SHAREDIR?=$(PREFIX)/share

NAME=alermanager-relay

default: run

build:
	go build -o bin/alertmanager-relay

clean:
	rm -Rf ./bin/

docs:
	go doc

lint:
	go vet

install: build man
	@install -Dm 755 bin/$(NAME) $(DESTDIR)$(BINDIR)/$(NAME)

man:
	scdoc < $(NAME).1.scd | tail --lines=+8 > $(NAME).1
	@install -Dm 644 $(NAME).1 $(DESTDIR)$(SHAREDIR)/man/man1/$(NAME).1
	@rm $(NAME).1
	mandb --quiet

test:
	go test

run:
	go run main.go

uninstall:
	@rm -f $(DESTDIR)$(BINDIR)/$(NAME)
	@rm -f $(DESTDIR)$(SHAREDIR)/man/man1/$(NAME).1
	mandb --quiet
