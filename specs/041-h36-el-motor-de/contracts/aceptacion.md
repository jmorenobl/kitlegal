# Contrato: la suite de aceptación de H36

Dos guiones testscript que escribe la tarea `[aceptacion]` en `specs/041-h36-el-motor-de/aceptacion/` y que quedan
congelados (ADR 0018). Al activarse, `aceptacion.sh activar H36` los copia como
`internal/app/testdata/script/h36-<nombre>.txtar` y `TestEntregaDelHito` (`internal/app/e2e_test.go`) los ejecuta en
`make ci`. Cada uno tiene que fallar hoy por una aserción (rojo primero, `scripts/workflow/aceptacion.sh`). Lo demás de
la entrega no se ve desde el binario —el civil no lo pide ningún tipo (US2.5), el ensayo es un fichero de prueba, y
`LeerTipos` y `TestPreceptosDeLosPlazos` son internos—, y lo comprueban los tests Go de plan.md.

## 1. La captura con el binario de `main` (FR-021)

La tarea `[aceptacion]` corre antes de cualquier cambio de producto, cuando el código del árbol es el de `cd3ad148`
(`git diff --name-only cd3ad148 HEAD` solo nombra `docs/` y `specs/`; research V2). La tarea:

1. comprueba esa condición con la misma orden, y no captura si nombra algo más;
2. construye el binario con `make build`, que lo deja en `./bin/kitlegal` (`bin/` está en `.gitignore`);
3. escribe el guion de §2 en `aceptacion/plazos-de-hoy.txtar`, sin sus ficheros, y detrás añade el sobre de cada
   orden de §2 con una sola orden por golden, que lo lleva del binario al guion sin pasar por el contexto de la sesión:

   ```sh
   ./bin/kitlegal plazos calcular <orden> --json \
     | sed -E -e 's/"fecha_consulta":"[^"]*"/"fecha_consulta":""/' -e '1i -- <golden> --' \
     >> specs/041-h36-el-motor-de/aceptacion/plazos-de-hoy.txtar
   ```

   El `1i` de `sed` pone la cabecera del fichero de txtar; la salida de error del binario, con líneas que no son del
   sobre, no entra en el guion.

El código de cada orden es el de su tabla de §2 y §3, el que da el binario de `main` (research V24). Nada de eso
escribe fuera de `bin/`, que git ignora, y de `aceptacion/`. Por qué `make build` y no un binario en `$TMPDIR` ni
`go run`, research S2.

## 2. `aceptacion/plazos-de-hoy.txtar` (US1; FR-005, FR-006, FR-008, FR-011, FR-012, FR-021; SC-001)

Cabecera: con los datos de hoy, lo único que ve quien usa el binario es que la herramienta ya no nombra los tipos; cada
sobre es el de `main`, salvo `fecha_consulta`.

1. **La aserción que falla hoy**: `exec kitlegal plazos calcular x x x --describe` y `! stdout 'alzada'` y
   `! stdout 'reposicion'` (hoy los nombran la ayuda del tipo y el enumerado de la salida; ni la descripción del verbo
   ni nada de §C.7 y §C.8 los nombra).
2. **Los tres casos** (US1.2): cada orden, `exec sh -c` con la salida estándar por la misma sustitución de
   `fecha_consulta` de §1 (`sed -E 's/"fecha_consulta":"[^"]*"/"fecha_consulta":""/'`, sin el `1i`) y
   `test $? -eq 0` sobre el código de `kitlegal` (no el de la tubería), y `cmp stdout <golden>`:

   | Golden | Orden (con `kitlegal plazos calcular` delante y `--json` detrás) |
   |---|---|
   | `caso-11.json` | `silencio 2026-07-09 Leganés Leganés --duracion 3 --unidad meses` |
   | `caso-15.json` | `contencioso 2026-10-27 Madrid` |
   | `caso-16.json` | `contencioso 2026-10-31 Madrid` |

3. **Los errores de §C.6** (US1.3), igual, con su código:

   | Golden | Orden | Código |
   |---|---|---|
   | `error-tipo.json` | `inventado 2026-09-09 Leganés Leganés` | 2 |
   | `error-inicio.json` | `reposicion 2026-02-30 Leganés Leganés` | 2 |
   | `error-duracion.json` | `generico 2026-09-09 Leganés Leganés --unidad dias-habiles` | 2 |
   | `error-unidad.json` | `generico 2026-09-09 Leganés Leganés --duracion 3 --unidad semanas` | 2 |
   | `error-municipios-lpac.json` | `reposicion 2026-09-09 Leganés` | 2 |
   | `error-municipios-contencioso.json` | `contencioso 2026-06-15 Madrid Leganés` | 2 |
   | `error-ambiguo.json` | `reposicion 2026-09-09 Arroyomolinos Leganés` | 2 |
   | `error-inexistente.json` | `reposicion 2026-09-09 Inexistente Leganés` | 3 |
   | `error-fuera-del-calendario.json` | `generico 9999-12-31 Leganés Leganés --duracion 1 --unidad dias-naturales` | 2 |

   La fila «no cargan» de §C.6 no se captura: el binario de `main` carga sus datos empotrados (spec, «Assumptions»); la
   cubre el caso «sin-los-tipos» de `TestAppletPlazos`. La nueva («el cómputo no la admite») no la puede dar `main` ni
   la rama con los datos de hoy; la cubren C7 y E12.

En rojo, además de hoy (mutantes de US1.4 a US1.6, en `mutantes.md`): una palabra del texto del art. 183 cambiada en
`data/plazos.yaml` rompe `caso-15.json`; el silencio con `da: vencimiento`, `caso-11.json`; el texto del municipio de
`contencioso`, `error-municipios-contencioso.json`.

## 3. `aceptacion/describe.txtar` (US4.1 con los datos de hoy, US4.2, US4.3; FR-012; SC-005)

1. `exec kitlegal plazos calcular x x x --describe`, código 0, y para cada argumento la ayuda de §C.7, byte a byte, con
   la forma de las líneas 39 a 43 de `h39-describe.txtar` (`"<argumento>": \{\n … "description": "<ayuda>"`). Falla hoy
   en la primera.
2. `stdout -count=2` del `"pattern"` de §C.8 (el de `tipo` y el de `avisos[].codigo`); `! stdout '"dias-del-art-183-lopj"'`
   y `! stdout '"alzada"'`; y los enumerados que siguen: `"nacional-y-autonomico"`, `"sin-festivos-locales"` y
   `"autonomico"`.
3. La descripción del verbo sigue: `stdout 'Devuelve la fecha de un plazo —un recurso, el acceso a la información
   pública, el silencio o uno genérico—'`.
4. `exec kitlegal plazos calcular --help`, código 0 (de su texto no se mira nada: research S5).

El mismo `--describe` con un `data/plazos.yaml` con un tipo más o que no carga no se puede dar al binario de e2e, que
lleva sus datos empotrados: lo comprueba `TestDescribeDePlazosSinDatos` (contracts/applet-calcular.md §3).

## 4. Lo que la suite no cubre y quién lo cubre

| Historia | Lo cubre |
|---|---|
| US2 (el civil) | `TestComputoCivil` (C1 a C10, y que ningún tipo lo lleva) |
| US3 (el ensayo) | `TestComputoDeEnsayo` (E1 a E13) y `TestNingunCodigoDePlazosEnGo` |
| US5 (`LeerTipos`) | `TestLeerTipos` |
| US6 (los preceptos del civil) | `TestPreceptosDeLosPlazos` |
