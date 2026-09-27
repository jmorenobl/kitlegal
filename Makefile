# Makefile de kitlegal.
#
# Es la única superficie de invocación de los controles: lo que ejecuta quien
# contribuye, lo que ejecuta el gancho de pre-commit y lo que ejecuta la
# integración continua son la misma orden. El contrato completo está en
# specs/001-h0-esqueleto-del-repo/contracts/make-targets.md.

# Parche de Go efectivo. La única fuente de verdad es la directiva `toolchain`
# de go.mod: el Makefile la lee y la exporta, de modo que todas las recetas
# —también las que operan sobre los módulos de herramienta— compilan y analizan
# con ese parche exacto, en local y en la integración continua (research.md D1).
# Se puede sobrescribir en la línea de órdenes (make ... GO_TOOLCHAIN=go1.26.99)
# para ejercer el camino de fallo de check-tools sin tocar go.mod.
GO_TOOLCHAIN := $(shell awk '$$1 == "toolchain" { print $$2; exit }' go.mod)

ifeq ($(GO_TOOLCHAIN),)
$(error go.mod no declara una directiva `toolchain`; sin ella el pin de la cadena de herramientas quedaría desactivado en silencio (research.md D1))
endif

export GOTOOLCHAIN := $(GO_TOOLCHAIN)

# Versión mínima de Go que hay que tener instalada: 1.21 es la primera capaz de
# cambiar de cadena de herramientas, que es lo que hace innecesario tener el
# parche concreto de la directiva `toolchain`.
GO_MIN_VERSION := 1.21

# Datos de construcción inyectados en el binario (FR-004, contracts/cli-version.md).
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0-dev)
COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo desconocido)
FECHA   ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# La cuarta inyección lleva esa misma VERSION a la identificación con la que
# kitlegal se presenta en la red, de modo que el User-Agent no pueda declarar
# una versión distinta de la que imprime el verbo `version`
# (FR-007 de H2, specs/003-h2-internal-httpx-cliente/research.md D5).
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.fecha=$(FECHA) \
	-X github.com/jmorenobl/kitlegal/internal/httpx.version=$(VERSION)

# Herramientas de los controles: un módulo Go por herramienta, sin instalación
# manual previa y con la versión fijada en su propio go.mod (research.md D2).
GOLANGCI    := go tool -modfile=tools/golangci-lint/go.mod golangci-lint
GOVULNCHECK := go tool -modfile=tools/govulncheck/go.mod govulncheck
GITLEAKS    := go tool -modfile=tools/gitleaks/go.mod gitleaks
LEFTHOOK    := go tool -modfile=tools/lefthook/go.mod lefthook

# Los modfiles de herramienta se descubren, no se enumeran: así mod-verify
# cubre los que existan en cada momento y también cualquiera que se añada
# después sin tocar este fichero (contracts/make-targets.md).
TOOL_MODULES := $(patsubst %/go.mod,%,$(wildcard tools/*/go.mod))

.DEFAULT_GOAL := help

.PHONY: build install test test-integration test-e2e lint lint-fast fmt fmt-check \
	vuln schema-check skills-check verify-sources evals skills-sync secrets mod-verify mod-tidy-check release \
	check-tools hooks ci help

## build: construye bin/kitlegal con los datos de versión inyectados
build: check-tools
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/kitlegal ./cmd/kitlegal

# El bucle de desarrollo, no la forma de instalar de quien usa: las skills se
# instalan con el binario recién instalado, por la ruta que da go list y no por
# el PATH, donde podría ir antes otro kitlegal. Sin comprobación previa: ante un
# conflicto, skills install sale con 1 sin cambiar nada y make también
# (specs/009-h19-instalar-sin-clonar/research.md D18).
## install: bucle de desarrollo; go install de kitlegal y, con ese binario, skills install -g --host claude
install: check-tools
	CGO_ENABLED=0 go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/kitlegal
	"$$(go list -f '{{.Target}}' ./cmd/kitlegal)" skills install -g --host claude

## test: tests unitarios con detector de carreras y perfil de cobertura
test: check-tools
	go test -race -shuffle=on -coverprofile=coverage.out ./...

## test-integration: tests con la etiqueta de compilación integration
test-integration: check-tools
	go test -race -tags=integration -coverprofile=coverage-integration.out ./...

## test-e2e: tests de extremo a extremo con testscript, contra el binario que construye el propio test
test-e2e: check-tools
	go test -race ./internal/app/

## lint: análisis estático completo, gosec incluido
lint: check-tools
	$(GOLANGCI) run ./...

## lint-fast: análisis estático rápido, el del gancho de pre-commit
lint-fast: check-tools
	$(GOLANGCI) run --fast-only ./...

## fmt: aplica el formato (gofumpt + goimports) corrigiendo los ficheros
fmt: check-tools
	$(GOLANGCI) fmt ./...

## fmt-check: comprueba el formato sin tocar ningún fichero
fmt-check: check-tools
	$(GOLANGCI) fmt --diff ./...

## vuln: análisis de vulnerabilidades conocidas (requiere red)
vuln: check-tools
	$(GOVULNCHECK) ./...

## schema-check: comprueba que schemas/ coincide con lo que emite --describe de cada verbo, sin escribir nada
schema-check: check-tools
	go test -count=1 -run '^TestEsquemasPublicados$$' ./internal/app/

## skills-check: comprueba skills, datos y evals sin red, sin modelo y sin escribir nada
skills-check: check-tools
	go test -count=1 -run '^(TestSkillsDelRepositorio|TestNormasDelRepositorio|TestTerritorioDelRepositorio|TestJerarquiaDelRepositorio|TestEvalsDelRepositorio|TestIdentificadoresDeLasNormas)$$' ./internal/app/ ./internal/skills/ ./internal/evals/

## verify-sources: comprueba contra la fuente real que sus respuestas se siguen interpretando (requiere red; fuera de ci)
verify-sources: check-tools
	scripts/verify-sources.sh

## evals: ejecuta las evals de una skill con Claude Code (Linux con strace, como root o con sudo; red solo del modelo; fuera de ci)
evals: check-tools
	scripts/evals.sh "$(SKILL)"

## skills-sync: regenera references/, la tabla de comandos de SKILL.md y los enlaces de scripts/ de cada skill
skills-sync: check-tools
	scripts/skills-sync.sh

## secrets: detección de secretos en todo el árbol
secrets: check-tools
	$(GITLEAKS) dir . --redact --no-banner

## mod-verify: verifica la integridad del módulo raíz y de cada módulo de herramienta
mod-verify: check-tools
	go mod verify
	@for d in $(TOOL_MODULES); do \
		echo "== $$d"; \
		(cd "$$d" && go mod verify) || exit 1; \
	done

## mod-tidy-check: comprueba que go.mod y go.sum están saneados
mod-tidy-check: check-tools
	go mod tidy -diff

## release: publicación de la release firmada (la aporta H19)
release:
	@echo "release: sin configurar hasta H19 (.goreleaser.yaml)" >&2
	@exit 1

## hooks: instala los ganchos de pre-commit
hooks: check-tools
	$(LEFTHOOK) install

## check-tools: comprueba go, git y que el toolchain fijado es obtenible
check-tools:
	@command -v go >/dev/null 2>&1 || { \
		echo "kitlegal: falta 'go'. Se necesita Go $(GO_MIN_VERSION) o superior; instálalo desde https://go.dev/dl/" >&2; \
		exit 1; \
	}
	@command -v git >/dev/null 2>&1 || { \
		echo "kitlegal: falta 'git'. Instálalo desde https://git-scm.com/downloads o con el gestor de paquetes del sistema" >&2; \
		exit 1; \
	}
	@if ! salida=$$(go version 2>&1) || ! echo "$$salida" | grep -qF "go version $(GO_TOOLCHAIN) "; then \
		echo "kitlegal: no se pudo usar el toolchain $(GO_TOOLCHAIN), que sale de la directiva 'toolchain' de go.mod." >&2; \
		echo "kitlegal: el go command responde: $$salida" >&2; \
		echo "kitlegal: comprueba que hay red para descargarlo, que el 'go' instalado es $(GO_MIN_VERSION) o superior y que esa directiva nombra un parche existente." >&2; \
		exit 1; \
	fi

## ci: el veredicto del repositorio; no modifica ningún fichero versionado
ci: fmt-check lint test test-integration vuln schema-check skills-check secrets mod-verify mod-tidy-check
	@echo "ci: todos los controles en verde"

## help: enumera las órdenes disponibles
help:
	@echo "kitlegal — órdenes disponibles:"
	@awk 'BEGIN { FS = ": " } /^## / { sub(/^## /, ""); printf "  %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
