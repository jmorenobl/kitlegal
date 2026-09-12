# 0010 · Los applets de ejemplo viven en un paquete normal, no bajo `testdata/`

- **Estado**: aceptada
- **Fecha**: 2026-09-12
- **Hito**: transversal (revisión de estructura tras H1, antes de H2)

## Contexto y problema

H1 dejó los dos applets de ejemplo (`echo`, `contar`) y la raíz de composición del binario de e2e bajo
`internal/app/testdata/` (research.md D19 de `specs/002-h1-kernel-cli-multicall/`). La garantía que se
buscaba era que el binario distribuido no pudiera enlazarlos ni por descuido, y se apoyaba en que los
comodines de Go (`./...`) nunca descienden a un directorio `testdata`.

Esa misma propiedad es el problema. Go reserva `testdata/` para fixtures, no para código compilable, y
tener paquetes ahí obligaba a enumerarlos a mano en todo lo que usa comodines: la variable
`TESTDATA_PKGS` del `Makefile` en `lint`, `fmt` y `fmt-check`, una función `paquetesDeEjemplo` en el test
de arquitectura, y aun así `go vet ./...` y el perfil de cobertura no los veían. La implementación de
referencia que copiará cada applet era, en la práctica, el código con más maquinaria especial del árbol.

## Decisión

- Los applets de ejemplo viven en `internal/app/ejemplo/`, un paquete normal del módulo que todos los
  controles alcanzan con `./...` sin enumerarlo: vet, lint, formato, tests, cobertura y test de
  arquitectura.
- La raíz de composición del binario de e2e vive en `internal/app/ejemplo/kitlegal-e2e/`. Es un
  `package main` bajo `internal/`, no bajo `cmd/`: `go install ./cmd/...` y goreleaser no lo alcanzan, y
  la excepción de lint para `os.Exit` y los descriptores sigue acotada a ese directorio exacto.
- La garantía de «el binario distribuido no los enlaza» deja de apoyarse en una propiedad del sistema de
  construcción y pasa a dos controles explícitos, en las dos capas que ya usa el proyecto para las reglas
  de arquitectura:
  - `depguard`, lista `ejemplo`: nadie puede importar `internal/app/ejemplo` salvo su propio árbol y los
    ficheros `_test.go`.
  - `internal/arch_test.go`, `TestElBinarioNoEnlazaLosEjemplos`: el cierre transitivo real de
    `cmd/kitlegal` no contiene ningún paquete bajo `internal/app/ejemplo`.
- `internal/app/testdata/` queda solo para lo que Go espera ahí: los guiones `.txtar` del e2e.

## Consecuencias

- Desaparecen `TESTDATA_PKGS` del `Makefile` y `paquetesDeEjemplo` del test de arquitectura.
- `internal/app/ejemplo` entra en la cobertura global (umbral del 70 % en `codecov.yml`). El `main` del
  binario de e2e no tiene cobertura unitaria, como tampoco la tiene `cmd/kitlegal/main.go`.
- Los artefactos de H1 en `specs/002-h1-kernel-cli-multicall/` siguen citando la ubicación antigua
  (FR-009, D19). Son el registro de aquel hito y no se reescriben; este ADR es la decisión que los
  sustituye en ese punto.
- Un applet de ejemplo nuevo se añade en `internal/app/ejemplo` y no requiere tocar ningún fichero de
  configuración.
