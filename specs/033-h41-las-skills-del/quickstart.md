# Quickstart: validar H41

Desde la raíz del repositorio, con Go 1.27 (`go.mod`). Los escenarios 1 a 7 no usan modelo ni credencial, no escriben en
el índice ni en el historial de git y salen con 0 cuando el hito está hecho. Los escenarios 1 a 6 no usan red ni
escriben en el árbol de trabajo —solo en la caché de compilación de Go y, el 5, en el directorio temporal—. El 7,
`make ci`, además: su objetivo `vuln` consulta la base de vulnerabilidades de Go, que requiere red (Makefile:117), y
`test` y `test-integration` dejan `coverage.out` y `coverage-integration.out` en la raíz (Makefile:87 y 91), ignorados
por git (`.gitignore`: `/*.out`, `/coverage.*`). Lo que comprueban, en contracts/; los nombres de los tests, en plan.md,
«Gates».

## 1. Las skills mandan leer sus referencias y caben en su tope

```sh
make skills-check
```

Verde: ningún cuerpo de `skills/` pasa de 11 000 caracteres, ninguna deriva tras `make skills-sync`, y, con
`TestRemisionesDelRepositorio`, cada enlace a una referencia va en una frase con «lee» antes y sin «si la tienes», «antes
de escribir la respuesta» ni «antes de dar la consulta», en mayúscula o minúscula (contracts/remisiones.md §6). Antes
del hito también salía en verde, porque ese test no existía; con el SKILL.md de cada `<skill>-control`, su subtest
«control» da los fallos de contracts/remisiones.md §6.4.

## 2. Lo que instala el binario

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h41-remisiones$' ./internal/app/
```

Verde: `kitlegal skills install --dir` deja las tres skills con cada orden de leer en su sitio y lo movido en su
referencia (contracts/aceptacion.md §2). Construye los binarios del arnés e2e en el temporal del test.

## 3. Control congelado y producto calculado desde control

```sh
go test -count=1 -run '^(TestVariantesDeControl|TestSkillsDelProductoDesdeControl|TestVariantesDeNormativa|TestLaboratorioFueraDelProducto)$' ./internal/evals/
```

Verde: cada `<skill>-control` es la de `v0.9.0`, fichero a fichero; `skills/<skill>` es su control con las remisiones y
los movimientos declarados y nada más, con el frontmatter idéntico; las variantes de H40 difieren de `normativa-control`
solo en lo suyo; y ninguna variante se empotra, se instala ni la mide el job (contracts/variantes.md §2,
contracts/remisiones.md §7).

## 4. El arnés, con el sustituto de `claude`

```sh
go test -count=1 -run '^(TestArnesDeLasFamilias|TestArnesDeNormativa|TestArnesDelLaboratorio|TestRecetasDelSondeo)$' ./internal/evals/
```

Verde: una petición válida de cada familia abre n/4 sesiones por eval en cada variante y modo, cada una con solo la
skill de su variante —la de `<skill>-producto`, byte a byte la de `skills/<skill>`—, y deja la lista de sus referencias
en `sondeo.json`; una inválida no abre ninguna (contracts/arnes.md §5).

## 5. La tabla y la lectura con `cd`

```sh
go test -count=1 -run '^(TestTablaDeLasFamilias|TestLecturaConCd|TestTablaDeNormativa|TestTablaDelLaboratorio)$' ./internal/evals/
go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo > "${TMPDIR:-/tmp}/h41-tabla.md" \
  && cmp "${TMPDIR:-/tmp}/h41-tabla.md" internal/evals/testdata/laboratorio/familias/plazos-y-territorio/tabla.md \
  && echo igual
```

Verde y `igual`: los recuentos, turnos e intervalos de las sesiones fijas de las tres familias, con las lecturas con `cd`
que cuentan y las que no (contracts/tabla.md §5). Escribe solo `h41-tabla.md` en el directorio temporal.

## 6. Las suites congeladas

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h(39|40|41)-' ./internal/app/
```

Verde: las de H39 y H40 sin cambiar un fichero (FR-042) y las dos de H41 (contracts/aceptacion.md).

## 7. Todo

```sh
make ci
```

## 8. La aceptación de la persona (fuera del run; FR-052)

Lo hace una persona, antes de fusionar, con la credencial del sondeo y Claude Code en el `PATH`; abre 520 sesiones con
modelo, que consumen su suscripción. Escribe las sesiones fuera del repositorio, bajo `../kitlegal-sondeos/h41/`; el
`DIRECTORIO` de cada lanzamiento no puede existir antes.

```sh
V=$(claude --version | awk '{print $1}')
make evals-laboratorio VARIANTES=normativa-control,normativa-producto MODOS=orden,herramienta SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/normativa/a
make evals-laboratorio VARIANTES=normativa-remisiones-en-orden MODOS=herramienta SESIONES=40 \
  MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/normativa/b
make evals-laboratorio VARIANTES=plazos-y-territorio-control,plazos-y-territorio-producto MODOS=orden,herramienta \
  SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/plazos-y-territorio
make evals-laboratorio VARIANTES=jurisprudencia-control,jurisprudencia-producto MODOS=orden,herramienta \
  SESIONES=40 MODELO=claude-sonnet-5-5 CLAUDE_CODE=$V CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41/jurisprudencia
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/normativa
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/plazos-y-territorio
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41/jurisprudencia
```

La persona deja las tres tablas en `evidencias/` con el modelo y la versión de Claude Code, y su lectura en la propuesta
de cambio. Vale si, en modo herramienta, el intervalo de `normativa-producto` en `con_herramientas.md` y en `normas.md`
no se toca con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden`, y si en ninguna familia y modo
`producto` queda por debajo de `control` —con intervalos que no se tocan— en «Activan antes» ni en las referencias a
cuya condición llegan las evals (research D12). Si no, la propuesta no se fusiona y lo medido va a un ADR.
