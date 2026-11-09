# Quickstart: validar H41

Desde la raíz del repositorio, con Go 1.27 (`go.mod`). Los escenarios 1 a 8 no usan modelo ni credencial, no escriben
en el índice ni en el historial de git y salen con 0 cuando el hito está hecho. Los escenarios 1 a 7 no usan red ni
escriben en el árbol de trabajo: solo en la caché de compilación de Go y, el 5, en el directorio temporal. El 8,
`make ci`, además: su objetivo `vuln` consulta la base de vulnerabilidades de Go, que requiere red (`Makefile:118`), y
`test` y `test-integration` dejan `coverage.out` y `coverage-integration.out` en la raíz (`Makefile:87` y `:91`),
ignorados por git (`.gitignore`: `/*.out`, `/coverage.*`). Lo que comprueba cada uno, en contracts/; los nombres de los
tests, en plan.md, «Gates». Antes del hito ya salen en verde el escenario 1, los tests de H39 y H40 de los escenarios
3, 4 y 5 y la parte de H39 y H40 del 6 (research V12); lo demás nombra tests o ficheros que entrega el hito.

## 1. Las skills mandan leer sus referencias y caben en su tope

```sh
make skills-check
```

Verde: ningún cuerpo de `skills/` pasa de 11 000 caracteres, ninguna deriva tras `make skills-sync`, y, con
`TestRemisionesDelRepositorio`, cada enlace a una referencia de `normativa` y `jurisprudencia` va en una frase con «lee»
antes, sin las tres expresiones, y cada frase es la de la lista del plan (contracts/remisiones.md §5 y §7);
`plazos-y-territorio` queda fuera, declarado en el test.

## 2. Lo que instala el binario

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h41-skills-instaladas$' ./internal/app/
```

Verde: `kitlegal skills install --dir` deja `normativa` y `jurisprudencia` con cada orden de leer en su sitio y lo
movido en su referencia, `plazos-y-territorio` como en `v0.9.0` y ninguna variante del laboratorio
(contracts/aceptacion.md §2). Construye los binarios del arnés e2e en el temporal del test.

## 3. Controles congelados, producto desde control y variantes declaradas

```sh
go test -count=1 -run '^(TestFrases|TestRemisionesDelRepositorio)$' ./internal/skills/
go test -count=1 -run '^(TestVariantesDeControl|TestSkillsDelProductoDesdeControl|TestVariantesDeNormativa|TestVariantesDePlazosYTerritorio|TestLaboratorioFueraDelProducto)$' ./internal/evals/
```

Verde: la frase de FR-001; cada `<skill>-control` es la de `v0.9.0`, fichero a fichero; `normativa` y `jurisprudencia`
son su control con las remisiones y los movimientos declarados y nada más, con el frontmatter idéntico;
`skills/plazos-y-territorio` es su control byte a byte; las variantes de H40 difieren de `normativa-control` solo en lo
suyo, y las tres de `plazos-y-territorio` de su control solo en las órdenes de `municipios.md` y `calculo_de_plazos.md`;
y ninguna variante se empotra, se instala ni la mide el job (contracts/variantes.md §4, contracts/remisiones.md §8).

## 4. El arnés, con el sustituto de `claude`

```sh
go test -count=1 -run '^(TestArnesDeLasFamilias|TestArnesDeNormativa|TestArnesDelLaboratorio|TestRecetasDelSondeo)$' ./internal/evals/
```

Verde: las cuatro invocaciones de la aceptación abren n/4 sesiones por eval en cada variante y modo, cada una con solo la
skill de su variante —la de `<skill>-producto`, byte a byte la de `skills/<skill>`—, y dejan la lista de sus
referencias en `sondeo.json`; una inválida no abre ninguna (contracts/arnes.md §5).

## 5. La tabla y la lectura con `cd`

```sh
go test -count=1 -run '^(TestTablaDeLasFamilias|TestLecturaConCd|TestTablaDeNormativa|TestTablaDelLaboratorio)$' ./internal/evals/
go run ./cmd/laboratorio tabla internal/evals/testdata/laboratorio/familias/plazos-y-territorio/sondeo > "${TMPDIR:-/tmp}/h41-tabla.md" \
  && cmp "${TMPDIR:-/tmp}/h41-tabla.md" internal/evals/testdata/laboratorio/familias/plazos-y-territorio/tabla.md \
  && echo igual
```

Verde e `igual`: los recuentos, turnos e intervalos de las sesiones fijas de las tres familias, también de las cinco
variantes de `plazos-y-territorio`, con las lecturas con `cd` que cuentan y las que no (contracts/tabla.md §5). Escribe
solo `h41-tabla.md` en el directorio temporal.

## 6. Las suites congeladas

```sh
go test -count=1 -run '^TestEntregaDelHito$/^h(39|40|41)-' ./internal/app/
```

Verde: las de H39 y H40 sin cambiar un fichero (FR-042) y las dos de H41 (contracts/aceptacion.md).

## 7. Los cuerpos de las variantes de `plazos-y-territorio` (FR-024)

```sh
for v in sin-condicion condicion-por-orden lista-con-verbo; do
  f=laboratorio/plazos-y-territorio/plazos-y-territorio-$v/skills/plazos-y-territorio/SKILL.md
  printf '%s ' "$v"; awk 'f; NR>1 && /^---$/ && !f {f=1}' "$f" | LC_ALL=en_US.UTF-8 wc -m
done
```

Da 10 991, 11 245 y 10 938 (contracts/variantes.md §3). Solo lee.

## 8. Todo

```sh
make ci
```

## 9. La aceptación de la persona (fuera del run; FR-052, FR-053)

Lo hace una persona, antes de fusionar, con la credencial del sondeo y Claude Code en el `PATH`; abre 760 sesiones con
modelo, que consumen su suscripción. Escribe las sesiones fuera del repositorio, bajo `../kitlegal-sondeos/h41/`; el
`DIRECTORIO` de cada invocación no puede existir antes. Las órdenes, las de contracts/arnes.md §1 (cuatro invocaciones y
tres tablas).

La persona deja las tres tablas en `evidencias/` con el modelo y la versión de Claude Code, y su lectura en la propuesta
de cambio. La propuesta vale si, en modo herramienta, el intervalo de `normativa-producto` en `con_herramientas.md` y en
`normas.md` no se toca con el de `normativa-control` y sí con el de `normativa-remisiones-en-orden`, y si en `normativa`
y `jurisprudencia`, en cada modo, `producto` no queda por debajo de `control` —con intervalos que no se tocan— en
«Activan antes» ni en las referencias a cuya condición llegan las evals (research D15). Si no, no se fusiona y lo medido
va a un ADR. La tabla de `plazos-y-territorio`, con la misma regla frente a `plazos-y-territorio-control` y el detalle
por sesión, va a un ADR que dice qué forma aplica H41.1.
