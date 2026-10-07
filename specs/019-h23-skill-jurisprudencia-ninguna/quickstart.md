# Quickstart: validar H23

Guía para comprobar la entrega sobre la cabeza del hito. Todas las órdenes se dan desde la raíz del repositorio.
Los escenarios 1 a 9 no tocan la red ni abren ninguna sesión con modelo; el 10 lo ejecuta el workflow y el 11, una
persona. Contratos: [applet-cita](./contracts/applet-cita.md), [evals-jurisprudencia](./contracts/evals-jurisprudencia.md)
y [skill-jurisprudencia](./contracts/skill-jurisprudencia.md).

**Lo que dejan en el árbol.** `make build` deja `bin/kitlegal`, y `make ci`, `coverage.out` y
`coverage-integration.out`: los tres los ignora git (`.gitignore`: `/bin/`, `/coverage.*`). Ningún escenario toca el
índice, el historial ni la caché o el grafo de la cuenta: cada orden del binario lleva `KITLEGAL_CACHE_DIR` en un
directorio temporal, y los tests usan los suyos.

## 0. Antes de empezar

```bash
go version
shasum -a 256 evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt
make build
export KITLEGAL_CACHE_DIR="$(mktemp -d)"
```

Esperado: Go 1.27; la huella `4886e0c8ca9836527ec08d8732b50315640a76803033374af5287b2a8321ee27`, la de
`evidencias/adr-0036/manifiesto.json`; y `bin/kitlegal`. Las dos primeras órdenes valen también hoy, antes de
implementar nada.

## 1. La consulta preparada (US1)

```bash
bin/kitlegal cita preparar ECLI:ES:TS:2023:3144 --json
bin/kitlegal cita preparar --roj "STS 3144/2023" --json
bin/kitlegal cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json
bin/kitlegal cita preparar --texto "cláusula suelo" --json
bin/kitlegal cita preparar ECLI:ES:TC:2024:79 --json
```

Esperado, las cinco con código 0 y `"fuente":"kitlegal.cita"`:

1. `"direccion":"https://www.poderjudicial.es/search/indexAN.jsp"`, una casilla `{"nombre":"ECLI","valor":"ECLI:ES:TS:2023:3144"}` y `"equivalente":{"forma":"roj","valor":"STS 3144/2023"}`.
2. Una casilla «Nº ROJ» y `"equivalente":{"forma":"ecli","valor":"ECLI:ES:TS:2023:3144"}`.
3. Tres casillas: «Nº Resolución» con `1088/2023` y «Fecha resolución», «Desde» y «Hasta», con `04/07/2023`; sin `equivalente`.
4. `"direccion":"https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo/1/AN"` y `"casillas":[]`.
5. `"cobertura":{"cendoj":"no-cubierto",…}`, sin `direccion` y con `"casillas":[]`.

Solo la quinta lleva `cobertura`: las cuatro primeras no tienen esa clave.

## 2. Los errores de argumentos (US1.6)

```bash
bin/kitlegal cita preparar ECLI:ES:TS:2023 --json; echo "código $?"
bin/kitlegal cita preparar ECLI:FR:CC:2023:1 --json; echo "código $?"
bin/kitlegal cita preparar --resolucion 1088/2023 --json; echo "código $?"
bin/kitlegal cita preparar ECLI:ES:TS:2023:3144 --texto "cláusula suelo" --json; echo "código $?"
bin/kitlegal cita preparar ECLI:ES:TS:2023:3144 --texto "" --json; echo "código $?"
bin/kitlegal cita preparar ECLI:ES:TS:2023:3144 --roj "" --json; echo "código $?"
bin/kitlegal cita --json; echo "código $?"
```

Esperado: las siete con `código 2`, `"ok":false` y `"clase":"argumentos"`; el mensaje de la segunda dice que el ECLI
no es español. La quinta y la sexta llevan un argumento escrito vacío, que no es un argumento no dado: una
referencia junto a `--texto`, y más de una forma de referencia.

## 3. El cotejo (US2)

```bash
F=evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt
bin/kitlegal cita cotejar --json < "$F"
bin/kitlegal cita cotejar ECLI:ES:TS:2023:3144 --json < "$F"
bin/kitlegal cita cotejar --roj "STS 1088/2023" --json < "$F"; echo "código $?"
bin/kitlegal cita cotejar --resolucion 3144/2023 --fecha 2023-07-04 --json < "$F"
tail -n 12 "$F" | bin/kitlegal cita cotejar --json; echo "código $?"
bin/kitlegal cita cotejar --documento "" --json < "$F"; echo "código $?"
head -n 11 "$F" | bin/kitlegal cita cotejar --roj "STS 1088/2023" --json
```

Esperado:

1. Los ocho datos en `ficha`, `"correspondencia":"se-corresponden"`, `"hallazgos":[]`, y
   `"url":"kitlegal:documento/sha256:4886e0c8…ee27"`, la huella del escenario 0.
2. `"es_la_pedida":true`.
3. `código 0`, `"ok":true`, `"es_la_pedida":false` y un hallazgo `documento-distinto` con
   `"cruce":"numero-de-resolucion"`.
4. `"cruce":"numero-del-roj"`.
5. `código 2`, `argumentos`: el final del fragmento no lleva ficha.
6. `código 2`, `argumentos`: con `--documento`, también vacío, la entrada no se lee.
7. La ficha sola, que es lo que la skill pasa: el mismo `data` que la tercera, con otra `url`.

## 4. Las herramientas (US4.3, US4.4)

```bash
go test -count=1 -run '^TestHerramientasDelServidor$' ./internal/app/
go test -count=1 -run '^TestEntregaDelHito$/^h23-cita-herramientas$' ./internal/app/
```

Esperado: `ok` las dos. La primera exige doce herramientas, las de los verbos del registro menos los de `skills` y
`mcp`, con los esquemas de `--describe`; la segunda, que `cita_cotejar` con la ficha y el ROJ `STS 1088/2023`
devuelve un resultado con su hallazgo, que sin `documento`, o con él vacío, devuelve el error `argumentos`, y que una
llamada con resultado a cada una de las dos herramientas da el sobre de su orden salvo `fecha_consulta`:
`cita_cotejar` con la ficha sola y con la ficha y ese ROJ, y `cita_preparar` con `resolucion` `1088/2023` y `fecha`
`2023-07-04` —el sobre de `cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json`, con sus tres casillas—,
con `texto` `cláusula suelo`, con `roj` `STS 1088/2023` y con `ecli` `ECLI:ES:TS:2023:3144`.

## 5. Sin red, sin caché y sin grafo (US4.1, US4.2)

```bash
go test -count=1 -run '^TestArquitectura$' ./internal/
go test -count=1 -run '^TestEntregaDelHito$/^h23-cita-sin-efectos$' ./internal/app/
ls -A "$KITLEGAL_CACHE_DIR"
```

Esperado: `ok` las dos; y el directorio de los escenarios 1 a 3, vacío.

## 6. Golden, aceptación y fuzz (SC-003, SC-004)

```bash
go test -count=1 -run '^(TestCitaPreparar|TestCitaCotejar)$' ./internal/app/
go test -count=1 -run '^TestEntregaDelHito$/^h23-cita-(preparar|cotejar)$' ./internal/app/
go test -count=1 -run '^(FuzzECLI|FuzzROJ|TestEquivalencia)$' ./internal/core/ids/
go test -count=1 -run '^(FuzzLeerFicha|TestLeerFicha|TestCotejar|TestPreparar)$' ./internal/core/cita/
make schema-check
```

Esperado: `ok` en todas. Sin `-fuzz`, cada `Fuzz…` ejecuta su corpus versionado.

## 7. La skill (US3, en lo que no pide modelo)

```bash
make skills-check
wc -l skills/jurisprudencia/SKILL.md
git diff --stat main -- skills/boe-legislacion skills/legal-core
```

Esperado: `ok`; menos de 300 líneas; y ninguna línea de diff en las otras dos skills.

## 8. El juicio y los umbrales, sin modelo (US5)

```bash
go test -count=1 -run '^(TestFormatoDeSentencias|TestJuzgarSentencias|TestCitaSinDocumento|TestPreguntasConElFragmento)$' ./internal/evals/
go test -count=1 -run '^(TestUmbralesDeJurisprudencia|TestUmbralesDelInforme|TestDefinicionDelJob|TestEvalsDelRepositorio)$' ./internal/evals/
```

Esperado: `ok`. Con sesiones sintéticas: una cita sin documento cotejado o un ECLI sin origen dan `fallo`; los
umbrales de `jurisprudencia` son cuatro; los de las otras dos skills, los de antes.

## 9. Todo

```bash
make ci
```

Esperado: verde, con `schema-check` y `skills-check` sin diferencias (SC-009).

## 10. El cierre (lo ejecuta el workflow, no una tarea)

Tras la revisión final, el workflow publica la rama, abre la propuesta de cambio y espera a los trabajos
`evals (jurisprudencia)`, `evals (boe-legislacion)` y `evals (legal-core)`. Esperado (SC-001): en el informe de
`jurisprudencia`, `cita_sin_documento:claude-sonnet-5-5:orden`, `…:herramienta`, `sin_activar:claude-sonnet-5-5:orden`
y `…:herramienta` con `cumple` y `decide` verdaderos y veredicto `aprobado`; los otros dos, aprobados; `red`, vacío
en los tres.

## 11. Después del run (una persona; SC-002)

Leer en los informes las respuestas del modelo que decide a las evals 02, 04 y 06, en los dos modos, y anotar en
`docs/USO.md` si alguna resume o caracteriza una sentencia cuyo texto no tenía delante; comprobar la dirección del
buscador del Tribunal Constitucional (research S1). Ninguna release lleva `jurisprudencia` hasta que H25 esté en
`main`.
