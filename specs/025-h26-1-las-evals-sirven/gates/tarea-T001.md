# T001, intento 1: sin verde por `make vuln` (GO-2026-6809, biblioteca estándar de go1.27.2)

## Lo que queda hecho

`schemas/eval.yaml.json` lleva ya `lecturas` y `sin_servir` en `properties` del objeto raíz y `lectura` y
`lectura-sin-servir` en `$defs`, tal como las escribe contracts/formato-de-eval.md §1. `norma` y `bloque` ya se
llamaban así en el fichero (`#/$defs/norma`, `#/$defs/bloque`, los del comando de bloque), así que no hay supuesto que
anotar. Ninguna otra regla del esquema cambia y ningún código cambia.

Rojo → verde con un arnés temporal (`internal/evals/zz_t001_temporal_test.go`, borrado antes de `make ci`), que valida
contra `esquemaDeEval()` con `skills.ValidarDocumentoYAML[map[string]any]`:

- antes del cambio, las 7 evals válidas con las claves nuevas fallaban (`additionalProperties` de la raíz);
- después, las 7 pasan (lectura, dos lecturas, `sin_servir` con bloque, con `verbo: indice`, con `verbo: metadatos`,
  con un bloque que es comando de la eval, y las dos claves juntas), y las 17 inválidas se rechazan con su ruta:
  `lecturas` vacía o que no es lista, sin `motivo`, motivo `""`, solo blancos (`"  \t "`) o número, con `fecha`, sin
  `bloque`, norma o bloque sin su patrón; `sin_servir` vacía, solo con `norma`, con `verbo: analisis`, con `bloque` y
  `verbo`, con `motivo`, sin `norma`; y otra clave de raíz. Las rutas salen sin barra inicial
  (`lecturas/0/motivo, línea 8: '' does not match pattern '\\S'`; `sin_servir/0, línea 8: …`).

`go test ./internal/evals/ -run 'TestEsquemaDeEval$|TestEvalsDelRepositorio$' -count=1 -v`: los dos `--- PASS`.

`make ci`, en primer plano: `fmt-check`, `lint` (`0 issues.`), `test`, `test-integration` y `test-tiempos` en verde;
falla en `vuln`, y por eso `make` no llega a los pasos siguientes, que se ejecutaron después uno a uno, todos en verde:
`make schema-check skills-check goreleaser-check secrets mod-verify mod-tidy-check` (`no leaks found`,
`all modules verified`, `go mod tidy -diff` sin diferencias).

## Por qué no está en verde

```
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
=== Symbol Results ===
Vulnerability #1: GO-2026-6809
    Distrust compromised intermediate certificates in crypto/x509
  More info: https://pkg.go.dev/vuln/GO-2026-6809
  Standard library
    Found in: crypto/x509@go1.27.2
    Fixed in: crypto/x509@go1.27.3
    Example traces found:
      #1: internal/empaquetado/ejecutar.go:71:21: empaquetado.Ejecutar calls fmt.Fprintf, which eventually calls x509.Certificate.Verify
Your code is affected by 1 vulnerability from the Go standard library.
make: *** [vuln] Error 3
```

No depende de la tarea: el hallazgo está en la biblioteca estándar del parche fijado en `go.mod` y su traza pasa por
`internal/empaquetado`, que la tarea no toca (no se corrió `make vuln` sin el cambio del esquema). Mientras el parche
no suba, `make ci` saldrá en rojo en cualquier tarea del hito.

## El arreglo y por qué no se pudo hacer en esta sesión

El arreglo de raíz es el de la #132 (`c88c1d0`) con go1.27.2: subir la directiva `toolchain` de `go.mod` a
`go1.27.3`, la única fuente del parche (el Makefile la lee y la exporta en `GOTOOLCHAIN`), y volver a correr `make ci`
(la #132 tuvo que subir además `golang.org/x/tools` en `tools/golangci-lint` porque cambió la versión de los datos de
exportación; un parche no debería, pero lo dice `make lint`). `go.mod` se admite en cualquier tarea.

Se intentó: con `toolchain go1.27.3`, `go version` sale con

```
go: downloading go1.27.3 (darwin/arm64)
go: download go1.27.3: golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64: Get "https://storage.googleapis.com/proxy-golang-org-prod/…": Forbidden
deny network-outbound storage.googleapis.com:443
```

El proxy de módulos redirige la descarga a `storage.googleapis.com`, que el sandbox de la sesión no deja salir, y
go1.27.3 no está en la caché de módulos (`~/go/pkg/mod/golang.org/` tiene go1.26.1, go1.27.0, go1.27.1 y go1.27.2).
`go.mod` se devolvió a `toolchain go1.27.2`: con go1.27.3 declarado y sin descargar, falla cualquier orden de `go`.
Es una dependencia externa inaccesible desde el sandbox, no un defecto de la tarea ni una ruta que le falte.

## Para el intento siguiente

Si go1.27.3 ya está en la caché de módulos (lo baja una persona fuera del sandbox, por ejemplo con `go version` en el
repositorio tras subir la directiva, o la #132 equivalente fusionada en `main` y traída a la rama), basta con subir
`toolchain` a `go1.27.3` en `go.mod`, comprobar `go version`, correr `make ci` en primer plano y, en verde, marcar T001:
el esquema ya está hecho y verificado.

## Hecho fuera del run (2026-10-21)

La subida está ya en la rama: `801c435`, cherry-pick de la #180 (`chore(deps): go1.27.3`), commit anterior a la base
de este intento; go1.27.3 está en la caché de módulos y `go version` lo resuelve sin red. No hay que tocar `go.mod`:
basta con correr `make ci` en primer plano y, en verde, marcar T001.

## Cierre (2026-10-21)

Con go1.27.3 en la rama (`go version go1.27.3 darwin/arm64`), sin tocar el esquema del intento anterior:

- rojo → verde repetido con el mismo arnés temporal (borrado antes de `make ci`): sin el cambio del esquema, las 5
  evals válidas con `lecturas` o `sin_servir` se rechazan (`additional properties 'lecturas' not allowed`); con él,
  pasan, y las 12 inválidas se rechazan cada una con su ruta;
- `TestEsquemaDeEval` y `TestEvalsDelRepositorio`: `--- PASS`;
- `make ci` en primer plano: `ci: todos los controles en verde` (`vuln` incluido).

T001 marcada `[X]`.
