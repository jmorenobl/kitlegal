# Quickstart: validar H23.1

Guía para comprobar la entrega sobre la cabeza del hito. Todas las órdenes se dan desde la raíz del repositorio, con
`bash` o `zsh`. Contratos: [umbral-documento-sin-cotejar](./contracts/umbral-documento-sin-cotejar.md),
[skill-jurisprudencia](./contracts/skill-jurisprudencia.md) y [documentos](./contracts/documentos.md).

Ningún escenario abre una sesión con modelo ni pide nada a ninguna fuente. La única red es la de las herramientas de Go
en `make ci` (`vuln`), en el 7. El 8 lo ejecuta el workflow, y el 9, una persona.

**Lo que dejan en el árbol.** Los escenarios 0 a 3 y el 6 no dejan nada: solo escriben en la caché de compilación de
Go. El 4 crea un worktree en `$TMPDIR`, con su registro en `.git/worktrees/`, y lo retira al acabar: no toca el árbol
de trabajo, el índice ni el historial. El 5 escribe en `$TMPDIR/h231-skill` y lo borra al acabar. El 7 deja
`coverage.out` y `coverage-integration.out`, que git ignora (`.gitignore`: `/coverage.*`).

**Lo que vale ya antes de implementar nada**, en `ef35b40`: el escenario 0; `make skills-check` y `make schema-check`
del 2; el 3; el 5 —sin diferencias, porque la skill es la de `main`—; y las búsquedas del 6, que dan las líneas que el
hito corrige (cuatro en la primera, 0 en las demás). Entonces `TestUmbralesDeJurisprudencia` pasa con trece umbrales y
sin los casos nuevos, `TestEvalsDelRepositorio` pasa sin el subtest `piden-el-cotejo` y las sustituciones del 4 no
casan con nada. Sobre la cabeza del hito,
cada escenario da lo que su «Esperado» dice.

**Con `rtk` delante**, la salida de `go test`, `git diff` y `diff` llega resumida («`ok`» es `Go test: N passed`):
`rtk proxy <orden>` da la original.

## 0. Antes de empezar

```bash
go version
git status --short
```

Esperado: Go 1.27; y la lista de lo que haya sin commitear, que el escenario 4 tiene que dejar igual.

## 1. El umbral, en `jurisprudencia` (US1, US3)

```bash
go test -count=1 -v -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/
```

Esperado: `ok` y quince subtests en `PASS`: los diez casos de antes, `el-elemento-del-contrato` y los cuatro nuevos de
contracts §5 —`una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta`,
`una-sin-activar-y-sin-cotejar` y `una-sin-medir-y-sin-cotejar`—. Cada caso exige los quince umbrales de contracts §1
con sus valores escritos a mano, once que deciden; los dos `documento_sin_cotejar` sobre 12 respuestas por modo (11 con
la sesión sin medir); y, con una sesión sin cotejo, `fallo` con el motivo de contracts §3.

## 2. Las evals que piden el cotejo y la forma de la skill (US1, US2, US3)

```bash
go test -count=1 -v -run '^TestEvalsDelRepositorio$' ./internal/evals/
make skills-check
make schema-check
```

Esperado:

1. `ok`, con el subtest `piden-el-cotejo` en `PASS`: la 04, la 06, la 08 y la 09 de `jurisprudencia`, y ninguna de
   `boe-legislacion` ni de `legal-core`.
2. `ok`: `SKILL.md` de `jurisprudencia` por debajo de 300 líneas, su tabla de comandos la regenerada y ninguna
   herramienta nueva.
3. `ok`: `schemas/` no cambia.

## 3. Lo que no cambia (US3)

```bash
go test -count=1 -run '^(TestUmbralesDelInforme|TestUmbralesConJuezYConSentencias|TestUmbralConjunto|TestInforme|TestInformeConElJuez|TestInformeMarkdownDeLosUmbrales|TestInformeEnDosModos|TestInformeConSesionesSinMedir|TestInformeFinalDeLosUmbrales|TestJuzgar|TestJuzgarLasLlamadas|TestJuzgarSentencias|TestCitaSinDocumento|TestCopiasDelJuez|TestMedidaVersionada|TestDefinicionDelJob|TestJuicioDelSondeo|TestSalidaDelSondeo|TestSondear)$' ./internal/evals/
git diff --name-only main -- internal/evals/sondeo.go scripts skills/boe-legislacion skills/legal-core evals schemas evidencias testdata internal/app/testdata .github .specify cmd docs/ROADMAP.md docs/ADR docs/USO.md docs/HISTORIAL.md docs/WORKFLOW.md README.md web
git diff --name-only main -- internal skills
```

Esperado:

1. `ok`. Los trece de `boe-legislacion`, nueve que deciden, y `legal-core` con `[]` (`TestUmbralesDelInforme`,
   `TestInformeEnDosModos`); el orden y los motivos de los de cada modo con juez y sentencias; el juicio de cada sesión
   y la tasa de cada serie; la medida, las copias del juez, la definición del job y las líneas del sondeo: todo como
   hoy, sin que ninguno de esos tests haya cambiado lo que espera. `TestInformeFinalDeLosUmbrales` sigue esperando
   «jurisprudencia: 13 umbrales», los de H26 (research D11).
2. Sin salida.
3. Solo `internal/evals/conjunto_test.go`, `internal/evals/doc.go`, `internal/evals/informe.go`,
   `internal/evals/informe_test.go`, `internal/evals/umbrales.go` y `skills/jurisprudencia/SKILL.md`.

## 4. Los dos mutantes, en una copia fuera del árbol (FR-064)

```bash
copia="$TMPDIR/h231-mutantes"
git worktree add --detach "$copia" HEAD

f="$copia/internal/evals/informe.go"
sed -i.orig '/^func (g grupoJuzgado) sinCotejar()/,/^}/ s/if !respuesta\.cotejo {/if !slices.ContainsFunc(g.queLePidenElCotejo(), func(otra respuestaDelJuez) bool { return otra.cotejo }) {/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE 1 NO SE HA APLICADO"
go -C "$copia" test -count=1 -v -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/ | grep -e '--- FAIL' -e '^ok' -e '^FAIL'
git -C "$copia" checkout -- internal/evals/informe.go

f="$copia/internal/evals/umbrales.go"
sed -i.orig '/^func umbralDeDocumentoSinCotejar(/,/^}/ s/Decide:\( *\)true,/Decide:\1false,/' "$f"
git diff --no-index --quiet "$f.orig" "$f" && echo "EL MUTANTE 2 NO SE HA APLICADO"
go -C "$copia" test -count=1 -v -run '^TestUmbralesDeJurisprudencia$' ./internal/evals/ | grep -e '--- FAIL' -e '^ok' -e '^FAIL'
git -C "$copia" checkout -- internal/evals/umbrales.go

git worktree remove --force "$copia"
git status --short
```

Cada sustitución cambia una sola línea y deja el original en `<fichero>.orig`, dentro de la copia: Go no lo compila, y
se va con ella.

Esperado:

- Ninguna de las dos líneas «NO SE HA APLICADO». Si sale alguna, el código no se llama como fija el plan (`sinCotejar`
  con su `if !respuesta.cotejo {`; `umbralDeDocumentoSinCotejar` con su `Decide: true,`), y la sustitución se ajusta a
  él: el mutante es el de contracts §6.
- Las dos órdenes de `go test` terminan en `FAIL`, que aquí es lo que se busca (research M8):
  - con el **mutante 1**, fallan `una-sin-cotejar-en-el-modo-orden`, `una-sin-cotejar-en-el-modo-herramienta` y
    `una-sin-activar-y-sin-cotejar`, que esperan 1 de 12 y reciben 0;
  - con el **mutante 2**, fallan los catorce casos de la tabla.
- `git status --short`, lo mismo que en el escenario 0: ningún mutante ha estado en el árbol.

Si `$copia` quedó de una ejecución interrumpida, `git worktree remove --force "$copia"` la retira. La tarea del run que
ejecuta este escenario deja lo que ha visto en `specs/022-h23-1-jurisprudencia-coteja-el/mutantes.md`.

## 5. La skill: qué cambia y qué no (US2)

```bash
wc -l skills/jurisprudencia/SKILL.md
d="$TMPDIR/h231-skill"
mkdir -p "$d"
git show main:skills/jurisprudencia/SKILL.md > "$d/main.md"
sed -n '1,38p' "$d/main.md" > "$d/main-cabecera.md"
sed -n '1,38p' skills/jurisprudencia/SKILL.md > "$d/hito-cabecera.md"
awk '/^### 2\. /{p=1} /^### 3\. /{p=0} p' "$d/main.md" > "$d/main-paso2.md"
awk '/^### 2\. /{p=1} /^### 3\. /{p=0} p' skills/jurisprudencia/SKILL.md > "$d/hito-paso2.md"
sed -n '/^### 4\. /,$p' "$d/main.md" > "$d/main-resto.md"
sed -n '/^### 4\. /,$p' skills/jurisprudencia/SKILL.md > "$d/hito-resto.md"
cmp "$d/main-cabecera.md" "$d/hito-cabecera.md"
cmp "$d/main-paso2.md" "$d/hito-paso2.md"
cmp "$d/main-resto.md" "$d/hito-resto.md"
rm -r "$d"
git diff main -- skills/jurisprudencia/SKILL.md
```

Esperado:

1. Menos de 300 líneas (hoy, 197).
2. Las tres órdenes `cmp`, sin salida: no cambian el frontmatter, el título, la entrada y la introducción del
   «Protocolo» (líneas 1 a 38), el paso 2, ni los pasos 4 a 7, «Comandos» y «Reglas» (FR-023, FR-024).
3. El diff, solo en el paso 1 y en el paso 3, con lo de contracts/skill-jurisprudencia.md §1: el cotejo es lo primero,
   nada del documento antes de él, y si es la pedida lo dice `es_la_pedida`, nunca a ojo, con el caso del número de
   resolución que es el ROJ pedido. Lo lee la revisión final (SC-007).

## 6. Los documentos (US4)

```bash
grep -n -e 'otros trece' -e 'son también trece' -e 'cuatro de los trece' -e 'los cuatro del modo orden' CONTRIBUTING.md
grep -c 'documento_sin_cotejar' CONTRIBUTING.md docs/JURISPRUDENCIA.md CLAUDE.md
sed -n '/^## \[Unreleased\]/,/^## \[0\.6\.0\]/p' CHANGELOG.md | grep -c -e 'documento_sin_cotejar' -e 'jurisprudencia` v0\.1\.1'
git diff --name-only main -- docs/WORKFLOW.md
git diff -U0 main -- CHANGELOG.md | grep '^@@'
```

Esperado:

1. Sin salida: ninguna de las frases que el hito deja falsas (en `main`, `ef35b40`, son las líneas 1145, 1172, 1187 y
   1190).
2. Un número mayor que 0 en cada uno de los tres (en `main`, 0 en los tres).
3. Un número mayor que 1: *Unreleased* nombra el umbral y la v0.1.1 (en `main`, 0).
4. Sin salida: `docs/WORKFLOW.md` no cambia.
5. Todas las cabeceras con su primer número —la línea de `main`— por debajo de 65, la de `## [0.6.0]`: las versiones
   publicadas no cambian.

Es una ayuda para la lectura, no su sustituto: que cada cifra, fichero, test y umbral exista en el árbol lo comprueba la
revisión final contra [documentos](./contracts/documentos.md) (SC-008).

## 7. Todo junto

```bash
make ci
```

Esperado: en verde, con `fmt-check`, `lint`, `test`, `schema-check`, `skills-check` y las reglas del conjunto (FR-060).

## 8. El job de cierre (lo lanza el workflow)

El workflow empuja la rama, abre la propuesta de cambio, pone la etiqueta `evals` y guarda el informe de cada skill en
`gates/evals/`. Sobre ellos:

```bash
jq -r '.veredicto, (.umbrales | length), ([.umbrales[] | select(.decide)] | length), ([.umbrales[] | select(.decide and (.cumple | not))] | length), (.red | length), (.juez.sin_juzgar | length)' specs/022-h23-1-jurisprudencia-coteja-el/gates/evals/jurisprudencia.json
jq -r '.umbrales[] | select(.nombre | test("^documento_sin_cotejar:")) | [.nombre, .medida, .total, .umbral, .cumple, .decide] | @tsv' specs/022-h23-1-jurisprudencia-coteja-el/gates/evals/jurisprudencia.json
jq -r '.veredicto, (.umbrales | length), (.red | length)' specs/022-h23-1-jurisprudencia-coteja-el/gates/evals/legal-core.json
jq -r '.veredicto, (.umbrales | length), ([.umbrales[] | select(.decide)] | length)' specs/022-h23-1-jurisprudencia-coteja-el/gates/evals/boe-legislacion.json
```

Esperado:

1. `aprobado`, `15`, `11`, `0`, `0` y `0`: quince umbrales, once que deciden, ninguno de ellos sin cumplir, `red`
   vacío y ninguna respuesta sin juzgar (SC-001).
2. Dos líneas, `documento_sin_cotejar:claude-sonnet-5-5:orden` y `…:herramienta`, cada una con `0`, `12`, `0`, `true` y
   `true`.
3. `aprobado`, `0` y `0`.
4. El veredicto que sea, `13` y `9`: si es `fallo`, no es motivo de este hito y queda en el informe final (FR-050).

El informe final da los dos `evals:jurisprudencia:documento_sin_cotejar:claude-sonnet-5-5:<modo>` como «comprobado por
su control» (SC-009).

Si `documento_sin_cotejar` sale en rojo, su motivo nombra cada sesión sin cotejo. Una reparación del cierre arregla lo
que entrega el hito —el código del umbral, el paso del cotejo de `SKILL.md` o un documento—, y nunca cambia el valor
del umbral, su `decide`, las sesiones de su `total`, las evals, la rúbrica, los casos ni la medida (FR-050).

## 9. Después del run (una persona)

1. Antes de fusionar, el sondeo de la eval 06 con la skill del hito (SC-010): 0 sesiones sin cotejar de 30, frente a 6
   de 48. Cómo lo mide lo decide la persona: el sondeo no cambia en este hito (FR-033) y su salida no dice qué sesiones
   ejecutaron `cita cotejar` (spec, «Fuera de alcance»).
2. Corregir, fuera del hito, el comentario de `.github/workflows/evals.yml:157`, que sigue diciendo que cada una de las
   dos skills con juez lleva «trece umbrales, nueve que deciden» (research V23).
3. Escribir en `CLAUDE.md` la viñeta de H23.1 y la frase del siguiente hito, al fusionar.
