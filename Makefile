BIN := curator
CMD := ./cmd

.PHONY: build run tidy install-cron

build:
	go build -o $(BIN) $(CMD)

tidy:
	go mod tidy

run:
	@set -a && . ./.env && set +a && go run $(CMD)

# Instala entrada no crontab para rodar às 07h de seg, qua e sex.
# Ajuste o horário e os dias conforme sua preferência.
install-cron: build
	@BINARY_PATH=$$(pwd)/$(BIN); \
	ENV_PATH=$$(pwd)/.env; \
	CRON_LINE="0 7 * * 1,3,5 set -a && . $$ENV_PATH && set +a && $$BINARY_PATH >> /tmp/curator.log 2>&1"; \
	(crontab -l 2>/dev/null | grep -v "$$BINARY_PATH"; echo "$$CRON_LINE") | crontab -; \
	echo "Cron instalado: $$CRON_LINE"
