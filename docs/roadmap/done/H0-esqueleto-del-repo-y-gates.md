---
id: H0
fase: "0"
---

# H0 · Esqueleto del repo y gates de CI

- **Objetivo**: un repo vacío pero blindado. Cualquier línea de Go que entre después pasa por los controles.
- **Entrega**: `make build test lint vuln` en verde; `kitlegal version` imprime versión, commit y fecha.
- **Alcance**: `go.mod` (módulo `github.com/jmorenobl/kitlegal`, Go estable actual), `cmd/kitlegal/main.go` mínimo, `Makefile` (`build test test-integration test-e2e lint fmt vuln schema-check skills-sync release ci install`), `.golangci.yml`, `lefthook.yml`, `.github/workflows/{ci,codeql,nightly}.yml`, `.github/dependabot.yml`, `codecov.yml`, `LICENSE` (decisión open-core: Apache-2.0 en el núcleo), `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `docs/ADR/0001-multicall.md`, `0002-sqlite-sin-cgo.md`, `0003-no-cendoj-masivo.md`, `0004-frontera-humana.md`.
- **Controles**: todos los de §3 marcados H0.
- **Aceptación**: PR de prueba con un `fmt.Println` en `internal/` falla en lint; PR limpia pasa en < 3 min.
