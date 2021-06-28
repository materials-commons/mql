.PHONY: bin test all fmt deploy docs server cli setup

all: fmt bin

fmt:
	-go fmt ./...

bin: cli server

cli:
	(cd ./cmd/mql; go build)

server:
	(cd ./cmd/mqlservd; go build)

deploy: deploy-cli deploy-server

deploy-cli: cli
	sudo cp cmd/mql/mql /usr/local/bin
	sudo chmod a+rx /usr/local/bin/mql

deploy-server: server
	@sudo supervisorctl stop mqlservd:mqlservd_00
	sudo cp cmd/mqlservd/mqlservd /usr/local/bin
	sudo chmod a+rx /usr/local/bin/mqlservd
	sudo cp operations/supervisord.d/mqlservd.ini /etc/supervisord.d
	@sudo supervisorctl start mqlservd:mqlservd_00
