# Quickstart: validar H7.1

Escenarios para comprobar la entrega una vez implementada. Se ejecutan desde la raíz del repositorio, sin red y en
este orden (el primero prepara el binario y la carpeta de trabajo de los demás). Ninguno toca el árbol de trabajo, el
índice ni el historial de git: el binario y la caché van a un directorio temporal, `go test` trabaja en los suyos, y el
único que escribe dentro del repositorio es el §9 (`make ci`), en `coverage.out` y `coverage-integration.out`, que git
ignora (`/*.out`). Formatos y reglas: [contracts/applet-graph.md](./contracts/applet-graph.md); guiones y tests:
[contracts/arnes-e2e.md](./contracts/arnes-e2e.md).

## 0. Binario y caché temporales

```bash
d=$(mktemp -d)
CGO_ENABLED=0 go build -trimpath -o "$d/kitlegal" ./cmd/kitlegal
export KITLEGAL_CACHE_DIR="$d/cache"
```

## 1. La ayuda de `check` dice su ámbito y su cota (FR-001, FR-007)

```bash
"$d/kitlegal" graph check --help
```

Esperado: `Usage: graph check [<norma> [<bloques> ...]] [flags]` y la descripción de contracts/applet-graph.md §1,
con «como mucho 50 hallazgos».

## 2. Una norma que el grafo no conoce, sin grafo (FR-003, SC-009)

```bash
"$d/kitlegal" graph check BOE-A-2099-99999 --json; echo "código $?"
test ! -e "$d/cache" && echo "no se ha creado nada"
```

Esperado: `código 0`, `"ok":true` y
`"data":{"norma":"BOE-A-2099-99999","bloques":[],"version-obsoleta":0,"fuente-caducada":0,"omitidos":0,"hallazgos":[]}`;
después, `no se ha creado nada`.

## 3. Argumentos sin forma (FR-004, SC-009)

```bash
"$d/kitlegal" graph check a21; echo "código $?"
"$d/kitlegal" graph check ''; echo "código $?"
"$d/kitlegal" graph check BOE-A-2015-10565 ' '; echo "código $?"
```

Esperado: las tres con `código 2`, y en la salida de error
`argumentos inválidos: la norma "a21" no tiene la forma BOE-A-<año>-<número>, …`,
`argumentos inválidos: la norma "" no tiene la forma …` y
`argumentos inválidos: el bloque " " está vacío o solo tiene espacio en blanco`.

## 4. Salida legible sobre el grafo vacío (FR-061, FR-063)

```bash
"$d/kitlegal" graph stats
"$d/kitlegal" graph check
```

Esperado: `El grafo del mundo tiene 0 nodos, 0 aristas y 0 textos.`; y
`No hay nada que volver a comprobar en todo lo consultado.`, una línea en blanco y
`Para acotar la comprobación a una norma y a sus bloques: kitlegal graph check <norma> [<bloque>...]`. Ninguna línea
empieza por `fuente`, `url`, `fecha_consulta` ni `hash`.

## 5. La regla genérica (FR-070, SC-011)

```bash
mkdir -p "$d/cache" && printf 'Este fichero no es una base de datos SQLite.\n' > "$d/cache/world.db"
"$d/kitlegal" graph stats; echo "código $?"
"$d/kitlegal" graph check BOE-A-2015-10565 --json; echo "código $?"
rm -r "$d/cache"
```

Esperado: `código 1` las dos veces, con `grafo: "<ruta>/world.db" no es una base de datos utilizable: file is not a
database (26)` en la salida de error y, con `--json`, `"data":{"clase":"inesperado",…}`.

## 6. La entrega, con las grabaciones: suite congelada, guiones de H7 y tests de integración

```bash
go test -count=1 -run '^TestEntregaDelHito$/^h7-1-' ./internal/app/
go test -count=1 -run '^TestEntregaDelHito$/^h7-grafo-' ./internal/app/
go test -count=1 -tags=integration -run '^TestMedidaDelGrafo$' -v ./internal/app/
go test -count=1 -tags=integration -run '^TestIntegracionGrafoDeH7$' ./internal/graph/
go test -count=1 -run '^(TestJuzgar|TestComprobarFormasDeHallazgo|TestEtiquetasDeHallazgo)$' ./internal/evals/
```

Esperado: todos `ok`. Los guiones `h7-1-grafo-lecturas`, `h7-1-grafo-check-acotado`, `h7-1-grafo-legible` y
`h7-1-grafo-regla-generica` son la suite congelada, ya activada (SC-003, SC-004, SC-009, SC-010, SC-011); los
`h7-grafo-*`, los de H7 con los cambios de contracts/arnes-e2e.md §5; `TestMedidaDelGrafo` publica con `-v` los bytes
de la salida sin argumentos, ≤ 40 000 (SC-001, SC-002); `TestIntegracionGrafoDeH7`, SC-012; los de `internal/evals`,
SC-008.

## 7. Lo retirado no está (FR-095, SC-013)

```bash
for f in internal/graph/publicar.go internal/graph/publicar_test.go internal/graph/integracion_enlace_test.go; do
  test ! -e "$f" && echo "retirado: $f"
done
grep -rniE 'EvalSymlinks|os\.Symlink|immutable=1|world\.db-nuevo|READONLY_ROLLBACK|journal_mode.(delete|persist|truncate)|sufijoDiario|comprobarEscritura|modoInmutable|sinPermisoDe|permisosDeSoloLectura|baseDeFuera|tablaAjena|ShmSuelto|DiarioCaliente|LecturaConDiario|enRollback' internal/graph internal/core/grafo
echo "coincidencias: $?"
```

Esperado: los tres `retirado: …` y `coincidencias: 1` (ninguna).

## 8. La skill, la eval y los esquemas (FR-045, FR-046, FR-082, FR-093)

```bash
make skills-check
make schema-check
wc -l < skills/boe-legislacion/SKILL.md
grep -c '⚠ REDACCIÓN MODIFICADA:' skills/boe-legislacion/SKILL.md
grep -n 'hallazgos' evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml
```

Esperado: los dos objetivos en verde (incluidos los subtests `hallazgos-de-la-skill` y `hallazgos-del-esquema`); menos
de 300 líneas; al menos una forma fija; y la eval 19 con `hallazgos:` y `- version-obsoleta`.

## 9. El veredicto del repositorio (FR-090, FR-096)

```bash
make ci
```

Esperado: `ci: todos los controles en verde`. La cobertura, contra los umbrales de `codecov.yml`, la mide CI en la
propuesta de cambio.
