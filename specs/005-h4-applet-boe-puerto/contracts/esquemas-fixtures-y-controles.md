# Contrato: esquemas publicados, fixtures, referencias, golden y controles del hito

Qué ficheros protegidos crea H4 (`schemas/`, `internal/source/boe/testdata/`, `internal/app/testdata/script/`),
con qué forma, quién los crea y qué los vigila. Decisiones en [research.md](../research.md) D11-D15.

## 1. `schemas/norma.json` y `schemas/bloque.json`

| Fichero | Verbos | `$id` |
|---|---|---|
| `schemas/norma.json` | `analisis`, `buscar`, `indice`, `metadatos` | `https://ventanillalegal.es/schemas/norma.json` |
| `schemas/bloque.json` | `articulo`, `articulos` | `https://ventanillalegal.es/schemas/bloque.json` |

Forma (claves ordenadas; dos espacios; sin escape HTML; salto final):

```json
{
  "$defs": {
    "<verbo>": {
      "$defs": { "…": "…" },
      "$id": "https://ventanillalegal.es/schemas/<fichero>/<verbo>",
      "$schema": "https://json-schema.org/draft/2020-12/schema",
      "description": "<ayuda del verbo>",
      "properties": { "entrada": { "…": "…" }, "salida": { "…": "…" } },
      "title": "boe <verbo>"
    }
  },
  "$id": "https://ventanillalegal.es/schemas/<fichero>",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "description": "Salidas de los verbos <lista> del applet boe. Generado desde --describe con make schema-check; no editar.",
  "title": "boe · <norma|bloque>"
}
```

- Cada `$defs.<verbo>` es la forma canónica de `kitlegal boe <verbo> --describe` **más** la clave `$id`.
- **Regenerar** (solo en una tarea `[datos]`): `go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/ -args -actualizar-esquemas`.
  La línea de esa tarea nombra `TestEsquemasPublicados` y `-actualizar-esquemas` y remite a esta sección sin copiar la
  orden: el extractor de rutas del workflow toma toda ruta del texto de la línea, y `./internal/app/` declararía
  `internal/app/` entero, código y guiones incluidos (plan, obligación 3).
- **Vigilar**: `make schema-check` (= `go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/`). Mensajes:
  «schemas/<fichero>: la parte de «<verbo>» no coincide con lo que emite `kitlegal boe <verbo> --describe`» y
  «schemas/<fichero>: el fichero no es la serialización canónica de sus partes». `TestEsquemasCubrenTodosLosVerbos`:
  cada verbo registrado en producción está en exactamente un fichero.
- **Validar** (FR-111): `santhosh-tekuri/jsonschema/v6` con `AssertFormat`, recurso = contenido del fichero,
  `Compile("https://ventanillalegal.es/schemas/<fichero>#/$defs/<verbo>/properties/salida")`.

## 2. Golden de `data`

- Ruta: `internal/source/boe/testdata/golden/<caso>.json`.
- Contenido: `{"data": <Resultado.Datos>, "url": <Resultado.Procedencia.URL>}` en forma canónica (claves ordenadas, dos
  espacios, sin escape HTML, salto final). Sin fecha: no forma parte de lo que se fija.
- Cada caso se ejecuta con la fuente sobre la reproducción de `testdata/boe.legislacion-consolidada` y una caché vacía.
- `TestGolden` compara byte a byte **cada** fichero existente con su caso y falla ante un fichero sin caso; con
  `-args -actualizar-golden` escribe los de la lista. `TestGoldenCubreTodosLosCasos` exige los trece y los seis verbos.
- **Regenerar** (solo en la tarea `[datos]` de golden): `go test -count=1 -run '^TestGolden$' ./internal/source/boe/ -args -actualizar-golden`.
  La línea de esa tarea lleva `internal/source/boe/testdata/golden/`, nombra `TestGolden` y `-actualizar-golden` y remite
  a esta sección sin copiar la orden: `./internal/source/boe/` declararía el paquete entero, con `terminos.go`, las
  grabaciones y las referencias que fija la persona (plan, obligación 3.2).

| Caso | Consulta |
|---|---|
| `buscar-procedimiento-administrativo-comun` | `buscar procedimiento administrativo común` |
| `buscar-sin-resultados` | `buscar zzqxkwvjh` |
| `indice-BOE-A-2015-10565` | `indice BOE-A-2015-10565` |
| `articulo-BOE-A-2015-10565-a21` | `articulo BOE-A-2015-10565 a21` |
| `articulo-BOE-A-2015-10565-a1` | `articulo BOE-A-2015-10565 a1` |
| `articulo-BOE-A-1985-5392-a22` | `articulo BOE-A-1985-5392 a22` |
| `articulo-BOE-A-2017-12902-a118` | `articulo BOE-A-2017-12902 a118` |
| `articulo-BOE-A-2017-12902-da3` | `articulo BOE-A-2017-12902 da3` |
| `articulo-BOE-A-1992-26318-a42` | `articulo BOE-A-1992-26318 a42` (norma derogada: avisos reales) |
| `articulos-BOE-A-2015-10565-a21-a22-a23` | `articulos BOE-A-2015-10565 a21 a22 a23` |
| `metadatos-BOE-A-2015-10565` | `metadatos BOE-A-2015-10565` |
| `metadatos-BOE-A-1992-26318` | `metadatos BOE-A-1992-26318` |
| `analisis-BOE-A-2015-10565` | `analisis BOE-A-2015-10565` |

## 3. Grabaciones reales

### 3.1 Manifiesto `internal/source/boe/testdata/grabaciones.json`

```json
{
  "fuente": "boe.legislacion-consolidada",
  "recursos": [
    {"recurso": "busqueda", "texto": "procedimiento administrativo común", "para": "…"},
    {"recurso": "bloque", "norma": "BOE-A-2015-10565", "bloque": "a21", "para": "…"}
  ]
}
```

`recurso` ∈ `busqueda` (con `texto`), `indice`, `metadatos`, `analisis` (con `norma`), `bloque` (con `norma` y `bloque`).
La dirección y el `Accept` los construye el código (`direcciones.go`, `busqueda.go`), nunca el manifiesto.

| # | Recurso | Argumentos | `Accept` | Para qué |
|---|---|---|---|---|
| 1 | busqueda | `procedimiento administrativo común` | JSON | `buscar` con resultados; golden; e2e |
| 2 | busqueda | `zzqxkwvjh` | JSON | `buscar` sin resultados, guardado (FR-032) |
| 3 | indice | `BOE-A-2015-10565` | JSON | `indice`; golden; e2e; gramática |
| 4 | indice | `BOE-A-1985-5392` | JSON | gramática contra índices grabados |
| 5 | indice | `BOE-A-2017-12902` | JSON | gramática contra índices grabados |
| 6 | bloque | `BOE-A-2015-10565` `a21` | XML | aceptación (FR-116); US1; e2e; < 200 ms |
| 7 | bloque | `BOE-A-2015-10565` `a1` | XML | aceptación (candidato «sin modificaciones») |
| 8 | bloque | `BOE-A-2015-10565` `a22` | XML | `articulos` (US4-2) |
| 9 | bloque | `BOE-A-2015-10565` `a23` | XML | `articulos` (US4-2, US4-3) |
| 10 | bloque | `BOE-A-2015-10565` `a9999` | XML | bloque inexistente → 3 (S2) |
| 11 | bloque | `BOE-A-1985-5392` `a22` | XML | aceptación (candidato «varias versiones») |
| 12 | bloque | `BOE-A-2017-12902` `a118` | XML | aceptación (candidato «varias versiones») |
| 13 | bloque | `BOE-A-2017-12902` `da3` | XML | aceptación (disposición) |
| 14 | bloque | `BOE-A-1992-26318` `a42` | XML | avisos reales de norma derogada (US1-3) |
| 15 | metadatos | `BOE-A-2015-10565` | JSON | avisos de 6-10; `metadatos`; golden |
| 16 | metadatos | `BOE-A-1985-5392` | JSON | avisos de 11 |
| 17 | metadatos | `BOE-A-2017-12902` | JSON | avisos de 12-13 |
| 18 | metadatos | `BOE-A-1992-26318` | JSON | norma derogada (US5-1); avisos de 14 |
| 19 | metadatos | `BOE-A-2099-99999` | JSON | norma inexistente → 3 (S2) |
| 20 | indice | `BOE-A-2099-99999` | JSON | índice inexistente → 3 |
| 21 | analisis | `BOE-A-2015-10565` | JSON | `analisis`; golden |
| 22 | analisis | `BOE-A-2099-99999` | JSON | análisis inexistente → 3 |

Suplentes, solo si al grabar no se cumple S7: bloque `BOE-A-2015-10565` `a5` y bloque `BOE-A-1985-5392` `a1` (con sus
metadatos ya en la lista), sustituyendo a 7 o a 11 en el manifiesto, en las referencias y en los golden.

### 3.2 Grabación

- Directorio: `internal/source/boe/testdata/boe.legislacion-consolidada/`; un fichero por petición con el nombre y el
  formato del contrato de grabación de H2 (§1, §2), más `GET_https_www.boe.es_robots.txt.json`, que la grabación
  guarda y la reproducción no usa.
- Orden (la ejecuta una persona, con red, en la pausa de la tarea `[datos]` del manifiesto):

```bash
scripts/grabar-fixtures.sh
```

  que es `KITLEGAL_RECORD=1 go test -tags=grabacion -count=1 -run '^TestGrabarFixtures$' ./internal/source/boe/`.
- Procedimiento de la pausa, en este orden (research.md D12 y D15):
  1. revisa los términos de uso y el `robots.txt` del BOE; si prohíben el acceso automatizado, rechaza la pausa y el
     hito se detiene (FR-123);
  2. deja la fila **definitiva** de `docs/SOURCES.md` (§8) con la fecha real de la revisión en «Revisado»;
  3. alinea con ella `IntervaloEntrePeticiones` y `terminosDeUso` en `internal/source/boe/terminos.go`;
  4. comprueba `go test -count=1 -run '^TestFuenteCoincideConSources$' ./internal/source/boe/` en verde;
  5. graba con `scripts/grabar-fixtures.sh`, que ya pide con el intervalo revisado;
  6. comprueba S1, S2, S4 y S7 sobre lo grabado (suplentes en el manifiesto si hace falta);
  7. escribe a mano y revisa las cinco referencias (§5);
  8. confirma en la rama la fila, `terminos.go`, las grabaciones, las referencias y, si cambió, el manifiesto, y aprueba.

  Que ninguna tarea posterior cambie lo confirmado aquí depende de una regla sobre el texto entero de las líneas de tarea
  (plan, obligación 3.2; research.md D15), no de que ninguna tarea pretenda declarar esos ficheros: el extractor de rutas
  toma como declarada toda ruta completa que aparezca en cualquier parte de la línea, también la de un fichero que la
  tarea solo lee, compara o dice no tocar y la de un paquete dentro de una orden (`workflow.yml` 583 y 606); el guardián
  admite esa ruta y todo lo que cuelga de ella (`docs/WORKFLOW.md` 71; `workflow.yml` 695-698), y `[datos]` basta para
  tocar `testdata/` (704). La línea de la tarea del manifiesto lleva, de todo ello, solo la ruta del manifiesto y cita
  este procedimiento por su sección, sin copiar sus rutas ni sus órdenes. Desde la tarea siguiente, ninguna línea
  contiene, en ninguna parte de su texto ni dentro de una orden (`./internal/source/boe/`), la ruta completa de
  `docs/SOURCES.md`, `internal/source/boe/terminos.go`, `internal/source/boe/testdata/grabaciones.json`,
  `internal/source/boe/testdata/boe.legislacion-consolidada/` o `internal/source/boe/testdata/referencias/`, de ninguna
  ruta bajo ellos ni de un directorio que los contenga (`docs/`, `internal/`, `internal/source/`, `internal/source/boe/`,
  `internal/source/boe/testdata/`): se nombran sin directorio (`SOURCES.md`, `terminos.go`, `grabaciones.json`, «las
  grabaciones», «las referencias») o por el test que los lee, que el guardián no casa con la ruta completa (698). Los
  ficheros del paquete que una tarea cambia van por su ruta completa y los subdirectorios de datos también
  (`internal/source/boe/testdata/sintetico/`, `internal/source/boe/testdata/golden/`), sin llaves ni comodines.
- `TestGrabacionesCompletas` (tarea de código siguiente): para cada entrada del manifiesto, `httpx.Replay` sobre el
  directorio sirve su petición. En la misma tarea, `TestReferenciasCompletas` (§5) y `TestFuenteCoincideConSources` sin
  la admisión de `pendiente` (§8).

## 4. Sintéticos

Copias de una grabación de §3 con el mínimo cambio, en `internal/source/boe/testdata/sintetico/<escenario>/boe.legislacion-consolidada/`.
La petición grabada (método y dirección) no cambia, para que la reproducción la empareje.

| Escenario | Ficheros (grabación de origen) | Cambio | Lo usan |
|---|---|---|---|
| `bloque-ilegible` | bloque a21 (6) | estado 200, `Content-Type: text/html`, cuerpo `<html><body>Mantenimiento<br></body></html>` | `TestArticulo/bloque-ilegible`, `TestFallosNoSeGuardan`, `TestVerificacionDeFuentesDetectaCambios` |
| `bloque-sin-elemento` | bloque a21 (6) | estado 200, cuerpo `<?xml version="1.0" encoding="utf-8"?><response><status><code>200</code><text>ok</text></status><data></data></response>` | `TestArticulo/bloque-sin-elemento` |
| `metadatos-caidos` | bloque a21 (6, sin cambios) y metadatos LPAC (15) | metadatos con estado 503 y cuerpo vacío | `TestArticulo/metadatos-caidos`, `TestCodigosDeSalidaDeBoe`, SC-012 |
| `metadatos-ilegibles` | bloque a21 (6, sin cambios) y metadatos LPAC (15) | metadatos con estado 200 y cuerpo `no es JSON` | `TestArticulo/metadatos-ilegibles` |
| `fuente-caida` | índice LPAC (3) | estado 503, cuerpo vacío | `TestIndice/fuente-caida`, `TestCodigosDeSalidaDeBoe` |
| `limite` | metadatos LPAC (15) | estado 429, `Retry-After: 60`, cuerpo vacío | `TestMetadatos/limite`, `TestCodigosDeSalidaDeBoe` |
| `avisos` | bloque a21 (6, sin cambios) y metadatos LPAC (15) | en metadatos: `estado_consolidacion.codigo` `"4"`, `estatus_derogacion` `"S"`, `vigencia_agotada` `"S"` | `TestArticulo/tres-avisos`, `TestMetadatos/tres-avisos` |

Los casos de lectura que no necesitan el kernel (sin versiones, última sin fecha, `tail`, CRLF, CDATA, atributos
con saltos, objeto suelto, índice anidado o plano, envoltorios, tipos inesperados) son tablas en `bloque_test.go` y
`lectura_test.go`, sin fichero.

## 5. Referencias del diff de aceptación (FR-116)

- Ruta: `internal/source/boe/testdata/referencias/<norma>-<bloque>.json`, cinco ficheros: `BOE-A-2015-10565-a21`,
  `BOE-A-2015-10565-a1`, `BOE-A-1985-5392-a22`, `BOE-A-2017-12902-a118`, `BOE-A-2017-12902-da3`.
- Forma:

```json
{
  "norma": "BOE-A-2015-10565",
  "bloque": "a21",
  "grabacion_bloque": "<nombre del fichero de la grabación del bloque>",
  "grabacion_metadatos": "<nombre del fichero de la grabación de los metadatos>",
  "campos": {
    "titulo":             {"valor": "…", "boe_py": "106"},
    "tipo":               {"valor": "…", "boe_py": "107"},
    "fecha_version":      {"valor": "…", "boe_py": "116 (112 sin versiones)"},
    "fecha_vigencia":     {"valor": "…", "boe_py": "117"},
    "norma_modificadora": {"valor": "…", "boe_py": "118"},
    "texto":              {"valor": "…", "boe_py": "120, 132-136"},
    "avisos":             {"valor": ["…"], "boe_py": "197-211 (lista avisos antes de la línea 213)"},
    "url":                {"valor": "https://www.boe.es/buscar/act.php?id=BOE-A-2015-10565#a21", "boe_py": "430"}
  }
}
```

- Las deriva **a mano una persona** de las grabaciones con la lógica de `xml_bloque_to_text`, `_elem_all_text`,
  `_check_vigencia` y `cmd_articulo`, y las revisa, en la pausa de la tarea `[datos]` del manifiesto —la que graba los
  fixtures—, tras grabar y antes de confirmar (FR-116, Q2; §3.2, punto 7). Quedan confirmadas antes de todo código de
  lectura y de `articulo`. El ejecutor nunca las crea, las completa ni las ajusta al código.
- `TestReferenciasCompletas` (`casos_test.go`, tarea de código que sigue a la pausa): exige exactamente cinco ficheros en
  `testdata/referencias/`, de tres normas distintas, uno de ellos `BOE-A-2015-10565-a21`, cada uno de la lista anterior o
  con los suplentes de §3.1 aplicados; que cada uno se lea con la forma de arriba sin claves desconocidas, con los ocho
  campos y un `boe_py` no vacío en cada uno; y que `grabacion_bloque` y `grabacion_metadatos` nombren ficheros de
  `testdata/boe.legislacion-consolidada/`. No compara contenido: eso lo hace `TestArticuloCoincideConBoePy`.
- `TestArticuloCoincideConBoePy` (subtests `BOE-A-2015-10565-a21` … `BOE-A-2017-12902-da3`): proyecta `data` sobre esos
  ocho campos (los avisos como la lista ordenada de sus `texto`) y exige igualdad campo a campo, nombrando el campo
  distinto.

## 6. Guiones e2e (`internal/app/testdata/script/`)

`TestEntregaDelHito` (`internal/app/e2e_test.go`): `Setup` copia `internal/source/boe/testdata/boe.legislacion-consolidada`
a `$WORK/reproduccion/boe.legislacion-consolidada` con `os.CopyFS` y fija `KITLEGAL_CACHE_DIR=$WORK/cache`; `Cmds`
añade `cronometra <máximo> <programa> <argumentos…>`, que ejecuta con `TestScript.Exec` y falla si el código no es 0
o si la duración alcanza el máximo.

| Guion | Comprueba |
|---|---|
| `boe-verbos.txtar` | los seis verbos con `--json` sobre la reproducción: código 0, `"ok":true`, `"fuente":"boe.legislacion-consolidada"`, `url` `https://www.boe.es/datosabiertos/api/legislacion-consolidada…`, `hash` y un fragmento de `data` de cada uno; `--describe` de los seis (título `boe <verbo>`); `kitlegal --help` lista `boe` y `kitlegal boe --help` sus seis verbos |
| `boe-multicall.txtar` | `boe articulo BOE-A-2015-10565 a21 --json` por el enlace y `kitlegal boe articulo …` dan la misma salida (`cmp`; la segunda sale de la caché con la misma `fecha_consulta`) |
| `boe-offline.txtar` | tras sembrar `metadatos BOE-A-2015-10565`, con `--offline` la misma salida y código 0, también tras borrar la reproducción; `indice BOE-A-1985-5392 --offline` → código 4 y `"clase":"fuente-no-disponible"` |
| `boe-codigos.txtar` | 2: `boe`, `boe articulo`, `boe articulo BOE-A-2015 a21`, `boe articulo BOE-A-2015-10565 ../a21`, `boe buscar ""`; 3: `boe articulo BOE-A-2015-10565 a9999`, `boe metadatos BOE-A-2099-99999`; 4: `--offline` sin entrada; cada uno con su clase en el sobre de fallo y la `url` que fija FR-101; el código se comprueba con el mismo mecanismo que `argumentos.txtar` |
| `boe-cache-rapida.txtar` | siembra `articulo BOE-A-2015-10565 a21`, deja la reproducción vacía (`rm` + `mkdir`) y ejecuta diez veces `cronometra 200ms $KITLEGAL_BIN boe articulo BOE-A-2015-10565 a21 --json` (SC-002) |
| `argumentos.txtar` (H1, ←) | líneas 24 y 30: `stderr 'applets disponibles: (boe, )?contar, echo'` en la tarea `[datos]` anterior al registro de `boe` en el binario de e2e, y `stderr 'applets disponibles: boe, contar, echo'` en la tarea `[datos]` de los guiones nuevos, que sigue inmediatamente a la de registro y precede a golden, esquemas y contratos (plan, pasos 11-12). Los códigos se comprueban con `exec sh -c '<orden>; test $? -eq <código>'`, que es el mecanismo de este guion |

## 7. Verificación contra la fuente real

- `scripts/verify-sources.sh`: `go test -tags=fuentes -count=1 -run '^TestVerificarFuentes$' ./internal/app/`.
- `Makefile`: `verify-sources: check-tools` → `scripts/verify-sources.sh` (no está en `ci`).
- `internal/app/fuentes_red_test.go` (`//go:build fuentes`): `TestVerificarFuentes` llama a
  `verificarArticulo(t.Context(), registro, esquemaDeArticulo)` con `AppletBoe(DependenciasDeRed())` y la caché en
  `t.TempDir()`.
- `internal/app/fuentes_test.go`: `verificarArticulo` (invoca `boe articulo BOE-A-2015-10565 a21 --json` por `app.Main`,
  exige código 0, sobre válido contra `schemas/bloque.json#/$defs/articulo/properties/salida` y `data.texto` no vacío;
  devuelve un error que empieza por «caso «boe articulo»:») y `TestVerificacionDeFuentesDetectaCambios` (con la
  reproducción grabada → sin error; con `sintetico/bloque-ilegible` → error que nombra el caso).
- `.github/workflows/nightly.yml`, trabajo nuevo `fuentes` (junto al de `make ci`):

```yaml
  fuentes:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      issues: write
    steps:
      - name: Obtener el código
        uses: actions/checkout@v7
      - name: Instalar Go y restaurar la caché
        uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: |
            go.sum
            tools/*/go.sum
      - name: Verificar las fuentes
        run: make verify-sources
      - name: Abrir o actualizar la incidencia
        if: failure()
        env:
          GH_TOKEN: ${{ github.token }}
          TITULO: "verify-sources: boe articulo"
          CUERPO: "La verificación nocturna contra la fuente real ha fallado: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
        run: |
          numero=$(gh issue list --state open --search "\"$TITULO\" in:title" --json number --jq '.[0].number // empty')
          if [ -n "$numero" ]; then gh issue comment "$numero" --body "$CUERPO"; else gh issue create --title "$TITULO" --body "$CUERPO"; fi
```

## 8. `docs/SOURCES.md`

```markdown
| Fuente | Applet | Base | Licencia | Términos de uso | robots.txt | Ritmo | Formato | Revisado |
|---|---|---|---|---|---|---|---|---|
| `boe.legislacion-consolidada` | `boe` | <https://www.boe.es/datosabiertos/api/legislacion-consolidada> | <licencia revisada> | <URL de los términos> | <resultado de la revisión> | `1s` | XML (bloque) y JSON; sin autenticación | AAAA-MM-DD |
```

`TestFuenteCoincideConSources` (`internal/source/boe/terminos_test.go`) localiza la fila por su primera celda y exige:
`time.ParseDuration` de la celda «Ritmo» = `IntervaloEntrePeticiones`; la URL entre `<…>` de «Términos de uso» =
`terminosDeUso.URL`; «Revisado» = una fecha `AAAA-MM-DD` igual a `terminosDeUso.Revisados`, que no es cero.
`TestFuenteNombreVigenciasYTerminos` exige que `Terms()` devuelva `terminosDeUso`.

| Momento | Quién | Fila | `terminos.go` | `TestFuenteCoincideConSources` |
|---|---|---|---|---|
| Tarea del arnés de grabación (código) | ejecutor | **propuesta**: ritmo `1s`, URL de términos propuesta, licencia y `robots.txt` por revisar, «Revisado» `pendiente` | iguales a la propuesta; `Revisados` cero | admite exactamente la pareja `pendiente`/cero |
| Pausa del manifiesto, antes de grabar (§3.2) | persona | **definitiva**, con la fecha real de la revisión | alineados con la fila | la persona lo ejecuta en verde |
| Tarea de código siguiente | ejecutor | no la toca | no los toca | deja de admitir `pendiente` |
| Todas las posteriores, de código o `[datos]` | ejecutor | no la toca | no los toca | sin la admisión de `pendiente` |

En las dos últimas filas, «no la toca» no se sigue de que la tarea no pretenda declarar el fichero: el extractor de rutas
toma como declarada toda ruta completa que aparezca en cualquier parte del texto de la línea, también la de un fichero
que la tarea solo lee o dice no tocar y la de un paquete dentro de una orden (`workflow.yml` 583 y 606), y el guardián
admite esa ruta y todo lo que cuelga de ella (`docs/WORKFLOW.md` 71; `workflow.yml` 695-698). Depende de la regla de la
obligación 3.2 del plan: desde la tarea que sigue a la pausa, ninguna línea contiene en ninguna parte la ruta completa de
`docs/SOURCES.md` o `internal/source/boe/terminos.go`, ni de un directorio que los contenga (`docs/`, `internal/`,
`internal/source/`, `internal/source/boe/`), tampoco en una orden (`./internal/source/boe/`). La tarea de código
siguiente, cuyos tests leen la fila, la nombra como `SOURCES.md` o por `TestFuenteCoincideConSources`, y `terminos.go`
sin directorio; los ficheros del paquete que una tarea cambia van por su ruta completa, sin llaves ni comodines.

## 9. `internal/source/boe/doc.go` (FR-120)

El comentario del paquete lleva una sección «Comportamientos no obvios de refs/boe.py» con las 32 entradas de FR-120,
una por línea de lista:

```
//   - [N] líneas <rango> · <comportamiento> · <se porta | se adapta: <motivo> | no se porta: <motivo>> · <FR-xxx | Fuera de alcance>
```

`TestDocAnotaElPorte` analiza `doc.go` con `go/parser` y exige las entradas 1 a 32, una vez cada una, con líneas, destino
y requisito.

## 10. Controles de arquitectura

- `internal/arch_test.go`: se retiran `TestElBinarioNoEnlazaHTTPX` y `TestElBinarioNoEnlazaCache`; `modulosDelBinario`
  gana los módulos medidos al enlazar, cada uno con su justificación en el comentario; nuevo
  `TestLasFuentesNoFirmanComoKitlegal` (literales `kitlegal.`/`kitlegal:` en `internal/source/**`, sin pasar en vacío).
- `.golangci.yml`: `run.build-tags: [integration, fuentes, grabacion]`; `misspell.ignore-rules` + `administrativo`,
  `capitulo`, `disposicion`, `materias`, `regulares` (tarea del paso 3) y `dependencias` (tarea de `fuente.go`), cada una
  con su comentario (research.md D18). Ninguna exclusión nueva.
