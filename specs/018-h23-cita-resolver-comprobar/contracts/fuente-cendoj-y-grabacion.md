# Contrato · La fuente `cendoj.jurisprudencia`, su grabación y su comprobación

El adaptador `internal/source/cendoj`: lo que declara, lo que pide, lo que reconoce y cómo se graba y se comprueba. La
fila de la fuente en `docs/SOURCES.md` no cambia (FR-023, FR-117).

## 1. Lo que la fuente declara, atado a su fila

| Constante | Valor | Celda de la fila |
|---|---|---|
| `NombreDeLaFuente` | `cendoj.jurisprudencia` | Fuente |
| base | `https://www.poderjudicial.es/search` | Base |
| página del buscador | base + `/indexAN.jsp` | Formato («sesión con cookie de `/search/indexAN.jsp`») |
| formulario | base + `/search.action` | Formato («formulario `POST /search/search.action`») |
| campos de consulta | `ECLI`, `ROJ`, `NUMERORESOLUCION` | Formato («con el campo `ECLI`, `ROJ` o `NUMERORESOLUCION` y fechas») |
| `IntervaloEntrePeticiones` | `5 * time.Second` | Ritmo (`5s`) |
| `Terms().URL` | `https://www.poderjudicial.es/search/indexAN.jsp` | Términos de uso |
| `Terms().Revisados` | 2026-10-03, 00:00 UTC | Revisado |

`TestFuenteCoincideConSources`, en `internal/source/cendoj/terminos_test.go`, lee la fila y falla si cualquiera de
ellos diverge: la dirección de los términos, el día, el ritmo, la base, la dirección del formulario y cada campo de
consulta, que tiene que estar nombrado en la celda «Formato» (FR-023).

`cendoj.OpcionesDelCliente()` da las opciones con las que la fuente tiene que pedirse: `httpx.ConFuente` con su nombre,
`httpx.ConFormulario` con su formulario y `httpx.ConIntentos(1)`. Quien compone añade el ritmo y el registrador
(`internal/app/cita.go`).

## 2. Una consulta

Dentro de una `httpx.Consulta` ([httpx-formulario.md](./httpx-formulario.md)):

1. `GET` de la página. Si su estado no es 200, la consulta termina ahí (§3) y el formulario no se envía.
2. `POST` del formulario, con los cinco campos fijos y los de la referencia:

| Referencia | Campos |
|---|---|
| ECLI | `ECLI` |
| ROJ, con fecha o sin ella | `ROJ` |
| número de resolución con su fecha | `NUMERORESOLUCION`, `FECHARESOLUCIONDESDE` y `FECHARESOLUCIONHASTA`, las dos con la fecha como `dd/mm/aaaa` |

   Fijos: `action=query`, `databasematch=AN`, `recordsPerPage=10`, `sort=IN_FECHARESOLUCION:decreasing`, `start=1`.

Nunca una segunda página, el documento de una resolución ni otra dirección; ninguna petición se repite (FR-006,
FR-020, FR-021).

## 3. Lo que se reconoce

**Clase de la respuesta del formulario** (`clasificar`):

| Respuesta | Clase |
|---|---|
| estado 200 y el texto «No se ha encontrado ningún resultado» | sin resultados |
| estado 200, sin ese texto y con algún enlace `/search/AN/openDocument/` | lista |
| cualquier otra, con el estado que sea | no reconocida |

**Lectura de la lista** (`leerResultados`): de cada resultado, su ECLI, su ROJ, su órgano y sala, su fecha, su número
de resolución, su número de recurso, su ponente y la dirección de su documento. La fecha se entrega como
`AAAA-MM-DD`. Si de algún resultado no se leen su ECLI, su ROJ, su fecha o su dirección —o no tienen su forma—, toda la
respuesta es «no reconocida». El resumen del CENDOJ no se lee. La escribe la tarea que tiene las grabaciones delante,
con expresiones de la biblioteca estándar y sin ninguna dependencia nueva (research D10, S1).

**Qué da cada cosa** ([cita-resolver.md §4](./cita-resolver.md)):

| Lo que ocurre | Clase del fallo |
|---|---|
| sin resultados; o lista en la que ninguna resolución tiene la fecha pedida | `no-encontrado` |
| página con estado distinto de 200; formulario «no reconocido» | `limite-o-tos` |
| fallo de `internal/httpx` | la que traiga: `limite-o-tos` (429, `robots.txt`) o `fuente-no-disponible` (5xx, conexión, plazo, turno) |

## 4. El manifiesto y sus grabaciones

`internal/source/cendoj/testdata/grabaciones.json`:

```json
{
  "fuente": "cendoj.jurisprudencia",
  "consultas": [
    {"ecli": "ECLI:ES:TS:2023:3144", "para": "golden por ECLI; e2e; eval (d)"},
    {"roj": "STS 3144/2023", "para": "golden por ROJ; e2e con su fecha y con otra; eval (f)"},
    {"resolucion": "1088/2023", "fecha": "2023-07-04", "para": "golden por número con fecha; e2e; eval (a)"},
    {"ecli": "ECLI:ES:TS:2023:999999", "para": "cero resultados: golden; e2e"},
    {"resolucion": "9999/2023", "fecha": "2023-01-01", "para": "eval (b), primer intento"},
    {"roj": "STS 9999/2023", "para": "eval (b), segundo intento"},
    {"resolucion": "3144/2023", "fecha": "2023-01-01", "para": "eval (f), primer intento"}
  ]
}
```

Cada entrada lleva una referencia como la de `cita resolver` —una entrada con `roj` no lleva fecha, porque esa
consulta no la envía— y `para`. Son las siete respuestas que reproducen los tests y las evals, y ninguna más (FR-092).
Lo que queda en `internal/source/cendoj/testdata/cendoj.jurisprudencia/`, con los nombres del prototipo (research M1):

| Petición | Fichero |
|---|---|
| página | `GET_https_www.poderjudicial.es_search_indexAN.jsp.json` |
| `robots.txt` (la pide el cliente; la reproducción no la usa) | `GET_https_www.poderjudicial.es_robots.txt.json` |
| 1 · `ECLI=ECLI:ES:TS:2023:3144` | `POST_https_www.poderjudicial.es_search_search.action_c_ECLI_ECLI_3AES_3ATS_3A2023_3A3144_action_quer-2740d948.json` |
| 2 · `ROJ=STS 3144/2023` | `POST_https_www.poderjudicial.es_search_search.action_c_ROJ_STS_3144_2F2023_action_query_databasematc-f743cbdf.json` |
| 3 · `NUMERORESOLUCION=1088/2023`, 04/07/2023 | `POST_https_www.poderjudicial.es_search_search.action_c_FECHARESOLUCIONDESDE_04_2F07_2F2023_FECHARESO-13d3b725.json` |
| 4 · `ECLI=ECLI:ES:TS:2023:999999` | `POST_https_www.poderjudicial.es_search_search.action_c_ECLI_ECLI_3AES_3ATS_3A2023_3A999999_action_qu-27a990b2.json` |
| 5 · `NUMERORESOLUCION=9999/2023`, 01/01/2023 | `POST_https_www.poderjudicial.es_search_search.action_c_FECHARESOLUCIONDESDE_01_2F01_2F2023_FECHARESO-5f2d7f9d.json` |
| 6 · `ROJ=STS 9999/2023` | `POST_https_www.poderjudicial.es_search_search.action_c_ROJ_STS_9999_2F2023_action_query_databasematc-d01ce834.json` |
| 7 · `NUMERORESOLUCION=3144/2023`, 01/01/2023 | `POST_https_www.poderjudicial.es_search_search.action_c_FECHARESOLUCIONDESDE_01_2F01_2F2023_FECHARESO-16c2f425.json` |

Las grabaciones llevan la página de resultados con el resumen del CENDOJ: lo asume el ADR 0036, y son las mínimas.

## 5. El test de grabación

`TestGrabarConsultas`, en `internal/source/cendoj/grabacion_test.go` (`//go:build grabacion`). Lo ejecuta solo el paso
`grabar_datos`; sin `KITLEGAL_RECORD=1` falla antes de pedir nada.

1. Construye el cliente con `OpcionesDelCliente()`, el ritmo de la fila y la raíz de grabación en un directorio
   temporal del test.
2. Abre **una** `Consulta`, pide la página y, con ella, envía las siete consultas del manifiesto, en su orden: nueve
   peticiones con el `robots.txt`, separadas 5 s.
3. Clasifica cada respuesta (§3). Con la página en un estado que no es 200, o con un envío «no reconocido», falla
   nombrando la consulta, y lo grabado se queda en el temporal, que el test borra.
4. Solo si todas se reconocen copia las grabaciones a `testdata/cendoj.jurisprudencia/`.

Los pasos 2 a 4 son `grabarConsultas`, de `casos_test.go`, que se compila siempre. `TestGrabacionRechazaLoNoReconocido`
lo ejecuta en `make ci`, sin red, con un cliente en reproducción sobre grabaciones sintéticas escritas en el propio
test: con un envío cuya respuesta es una página que no se reconoce y con una página en 403, devuelve un error y el
destino queda vacío (FR-093, FR-094). Ningún paso provoca un CAPTCHA ni un bloqueo.

Si la grabación falla, `grabar_datos` la repite una vez al minuto y, si vuelve a fallar, detiene el run con su
dossier: es lo que hace con cualquier fuente, y no cambia.

## 6. Tests del adaptador (en `make ci`, sin red)

| Test | Qué fija |
|---|---|
| `TestLecturaDeLasGrabaciones` | golden de las grabaciones 1 a 4: una resolución por ECLI, por ROJ y por número con fecha, y sin resultados (FR-110) |
| `TestLaSentenciaConocida` | las grabaciones 1, 2 y 3 dan la sentencia de `evidencias/adr-0036/`: `ECLI:ES:TS:2023:3144`, `STS 3144/2023`, 2023-07-04, `1088/2023`, `4703/2019`, y una dirección bajo `/search/AN/openDocument/` |
| `TestClasificar` | §3, con respuestas escritas en el test: las dos reconocidas, una página sin ninguna marca, un 403, un 404, una redirección |
| `TestPeticionesDeUnaConsulta` | contra un sitio de prueba (`httpxtest`): una petición de `robots.txt`, una de la página y un envío, y ninguna otra dirección; con el sitio en 503, una petición por dirección y ningún envío; con la página en 403, ningún envío (FR-021; SC-004) |
| `TestPaginaCompleta` | una lista sintética de 10 resultados declara `pagina_completa`, y `data` nunca lleva más de 10 (FR-014; SC-004) |
| `TestFiltroPorFecha` | con `--fecha`, lo de otra fecha no entra en `data`, ni en el mensaje, ni en la caché (FR-011, FR-012) |
| `TestVigenciaDeLaCache` | para una entrega y para un «no encontrado»: repetida antes de 30 días, no pide nada; después, vuelve a la fuente; con `--offline` y vencida, 4 (FR-040, FR-041) |
| `TestNiResumenNiPagina` | tras una consulta con la grabación 1, que entrega al grafo: ninguna secuencia de 30 o más letras y espacios de la respuesta grabada que no esté en un valor de `data` —el resumen es prosa, y lo es casi toda la página— aparece en ningún fichero de la carpeta de la caché, que es también la del grafo; tras un «no encontrado» por la fecha con la grabación 2, tampoco el ECLI, la fecha ni la dirección de la resolución encontrada (FR-044; SC-005) |
| `TestOperacionesDeGrafo` | un nodo `Resolucion` por resolución, con su ECLI como id; ninguna operación sin procedencia tras la entrega; aplicarla dos veces no duplica nada; un fallo no emite nada (FR-042) |
| `TestFuenteCoincideConSources` | §1 |
| `FuzzECLI`, `FuzzROJ` (`internal/core/ids`) | ninguna entrada hace fallar al programa; lo aceptado tiene la forma del requisito (FR-113) |

## 7. La comprobación de la fuente, a petición

```text
make verify-sources                 # lo de hoy: boe articulo; nada del CENDOJ
make verify-sources FUENTE=cendoj   # resuelve ECLI:ES:TS:2023:3144 contra la fuente real
```

`scripts/verify-sources.sh` recibe la fuente como primer argumento. Sin argumento ejecuta exactamente la orden de hoy.
Con `cendoj`, `go test -tags=fuentes -count=1 -run '^TestVerificarCendoj$' ./internal/app/`: el applet `cita` del
binario distribuido, con la caché en un directorio temporal y sin grafo, exige código 0, un sobre válido contra
`schemas/resolucion.json` y una resolución con ese ECLI. No graba nada. Con cualquier otro valor sale con 2. La
cabecera del guion dice cuándo se lanza: antes de una release y cuando alguien avisa de que `cita resolver` no reconoce
la respuesta (FR-090).

`TestElCendojSoloSeCompruebaAPeticion` (`internal/app/fuentes_test.go`, en `make ci`) ejecuta el guion con un `go`
sustituto que apunta sus argumentos: sin fuente, son los de `TestVerificarFuentes` y ninguno nombra el CENDOJ; con
`cendoj`, los de `TestVerificarCendoj`. Y lee `Makefile` y cada flujo de `.github/workflows/`: `FUENTE` no tiene valor
por defecto, y ningún flujo nombra `FUENTE=cendoj`, `verify-sources.sh cendoj` ni `TestVerificarCendoj` (FR-091;
SC-003). `TestVerificacionDelCendojDetectaCambios` ejerce la misma comprobación sin red: pasa sobre las grabaciones y
falla sobre una respuesta sintética que no se reconoce.
