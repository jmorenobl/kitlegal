# Quickstart · H23 · validar `cita resolver` y la skill `jurisprudencia`

Guía para comprobar el hito una vez implementado. Las órdenes se ejecutan desde la raíz del repositorio. Las de §1 a §9
no usan la red ni abren sesiones con modelo, y no dejan nada en el árbol de trabajo, en el índice ni en el historial:
lo que escriben va a un directorio temporal que §10 borra, y a la caché de compilación de Go. §11 y §12 no son para un
paso del run: los lanza una persona o el workflow.

Detalle de cada salida en [contracts/](./contracts/) y de los tipos en [data-model.md](./data-model.md).

## 0. Preparación

```bash
Q="$(mktemp -d)"
go build -o "$Q/kitlegal" ./cmd/kitlegal
export KITLEGAL_CACHE_DIR="$Q/cache"
```

Todo lo que sigue usa `"$Q/kitlegal"`: no toca el `kitlegal` instalado ni la caché de la cuenta.

## 1. El verbo existe y se describe (FR-001)

```bash
"$Q/kitlegal" --help | grep -E '^\s+cita\s'
"$Q/kitlegal" cita resolver --describe | grep -c '"title": "cita resolver"'
"$Q/kitlegal" cita resolver --describe | grep -c '"title": "--'
"$Q/kitlegal" skills install --describe | grep -c '"title": "--'
make schema-check
git diff --numstat main -- schemas/instalacion.json
```

Esperado: la línea de `cita` en la ayuda; `1`; `3`, las banderas propias de `cita resolver` (`--roj`, `--resolucion`,
`--fecha`); `3`, las de `skills install` (`--global`, `--host`, `--dir`), que hoy da `0`; `schema-check` en verde, con
`schemas/resolucion.json` y `schemas/instalacion.json` iguales a lo que emite `--describe`; y
`7	0	schemas/instalacion.json`: siete líneas `title` añadidas y ninguna quitada (research D16, M5).

## 2. Una referencia mal dada termina con 2, sin pedir nada (FR-002 a FR-005; US2)

```bash
for referencia in "ECLI:ES:TS:2023" "ECLI:FR:CC:2023:1" "ecli:es:ts:2023:3144"; do
  "$Q/kitlegal" cita resolver "$referencia" --json; echo "código $?"
done
"$Q/kitlegal" cita resolver --resolucion 1088/2023 --json; echo "código $?"
"$Q/kitlegal" cita resolver ECLI:ES:TS:2023:3144 --fecha 2023-07-04 --json; echo "código $?"
"$Q/kitlegal" cita resolver ECLI:ES:TS:2023:3144 --roj "STS 3144/2023" --json; echo "código $?"
"$Q/kitlegal" cita resolver --json; echo "código $?"
```

Esperado: las siete, `código 2` y un sobre con `"clase":"argumentos"`. Ninguna abre una conexión: son rechazos
anteriores a la caché y a la red.

## 3. El Tribunal Constitucional, fuera de cobertura y sin petición (FR-015)

```bash
"$Q/kitlegal" cita resolver ECLI:ES:TC:2024:79 --json --no-graph; echo "código $?"
```

Esperado: `código 0`, `"fuente":"kitlegal.cita"`, `"resoluciones":[]` y
`"cobertura":"tribunal-constitucional-no-cubierto"`. No se pide nada a ninguna fuente.

## 4. `--dry-run` y `--offline` no piden nada (FR-017, FR-019)

```bash
"$Q/kitlegal" cita resolver ECLI:ES:TS:2023:3144 --dry-run; echo "código $?"
"$Q/kitlegal" cita resolver ECLI:ES:TS:2023:3144 --offline --json; echo "código $?"
```

Esperado: la primera, `código 0`, nada en la salida estándar y, en la de error, las dos peticiones que no ha emitido
(`GET https://www.poderjudicial.es/search/indexAN.jsp` y `POST https://www.poderjudicial.es/search/search.action`); la
segunda, `código 4` y `"clase":"fuente-no-disponible"`.

## 5. Los guiones de aceptación, con las grabaciones (FR-111; SC-002)

```bash
go test -count=1 -run '^TestEntregaDelHito$/h23-' ./internal/app/
```

Esperado: `ok`. Son los cinco guiones de `specs/018-h23-cita-resolver-comprobar/aceptacion/`, activados con el prefijo
`h23-`: las tres formas y `--roj` con su fecha (0); el ECLI que no existe y `--roj` con otra fecha (3); el ECLI mal
formado y `--resolucion` sin `--fecha` (2); el del Tribunal Constitucional (0); la página que no se reconoce y el 403,
sintéticos (5); la repetición con `--offline` (0); `graph show`; y la herramienta `cita_resolver`.

## 6. El formulario en `internal/httpx` (FR-030 a FR-034; SC-006)

```bash
go test -count=1 -race -run '^(TestFormularioAdmitido|TestEnvioDelFormulario|TestCookiesDeLaConsulta|TestConsultaSinRedirecciones|TestGrabacionDeFormularios|TestNombreDeGrabacion|TestPedirRechazaMetodo)$' ./internal/httpx/
go test -count=1 -run '^TestArquitectura$' ./internal/
```

Esperado: `ok` en los dos. Solo `internal/httpx` importa `net/http`, y el único método que no es GET ni HEAD es el
`POST` de una consulta a la dirección declarada.

## 7. El adaptador (FR-010 a FR-024, FR-040 a FR-044; SC-004, SC-005)

```bash
go test -count=1 -race ./internal/source/cendoj/ ./internal/core/ids/
go test -count=1 -run '^(TestNiResumenNiPagina|TestSalidasDeCitaContraSuEsquema|TestVerificacionDelCendojDetectaCambios|TestElCendojSoloSeCompruebaAPeticion)$' ./internal/app/
go test -count=1 -run '^FuzzECLI$' -fuzz '^FuzzECLI$' -fuzztime 10s ./internal/core/ids/
```

Esperado: `ok`. El primero ejecuta, entre otros, `TestLecturaDeLasGrabaciones`, `TestLaSentenciaConocida`,
`TestPeticionesDeUnaConsulta`, `TestPaginaCompleta`, `TestVigenciaDeLaCache`, `TestFuenteCoincideConSources` y
`TestGrabacionRechazaLoNoReconocido`. El fuzz no deja ningún fichero si no encuentra un fallo.

## 8. La skill y sus evals, sin modelo (FR-060, FR-070 a FR-075; SC-009)

```bash
make skills-check
wc -l skills/jurisprudencia/SKILL.md
git diff --stat main -- skills/boe-legislacion skills/legal-core
```

Esperado: `skills-check` en verde; menos de 300 líneas; y el `git diff` sin ninguna línea.

## 9. El umbral y el job, sin modelo (FR-080 a FR-083; SC-007)

```bash
go test -count=1 -run '^(TestUmbralDeCitaSinResolver|TestUmbralesDeJurisprudencia|TestUmbralesDelInforme|TestDefinicionDelJob|TestPrepararYComprobar|TestEsquemaDeEval|TestFormaDelComando)$' ./internal/evals/
```

Esperado: `ok`. Una sola respuesta con un ECLI que no viene de una entrega ni de la pregunta da veredicto `fallo`; con
ninguna, los cuatro umbrales de `jurisprudencia` se cumplen.

## 10. Limpieza

```bash
unset KITLEGAL_CACHE_DIR
rm -rf "$Q"
git status --short
```

Esperado: `git status --short` no enseña nada que no estuviera antes de §0.

## 11. Solo una persona: la fuente real (FR-090)

```bash
make verify-sources FUENTE=cendoj
```

Hace tres peticiones al CENDOJ, separadas 5 s, con el agente del proyecto. Se lanza antes de una release y cuando
alguien avisa de que `cita resolver` no reconoce la respuesta. `make verify-sources`, sin fuente, no lo consulta.

## 12. El cierre: lo mide el workflow y lo lee una persona (SC-001, SC-008)

Tras la revisión final, el workflow abre la propuesta de cambio y espera a `evals (jurisprudencia)`,
`evals (boe-legislacion)` y `evals (legal-core)`. En el informe de `jurisprudencia`, `umbrales` lleva
`sin_activar:<modelo>:<modo>` y `cita_sin_resolver:<modelo>:<modo>` en los dos modos, con `decide: true` y cumplidos, y
`red` está vacío. Después, una persona lee las respuestas del modelo que decide a las evals 02, 04 y 06 en los dos
modos y anota en `docs/USO.md` si alguna resume o caracteriza una sentencia cuyo texto no tenía delante.
