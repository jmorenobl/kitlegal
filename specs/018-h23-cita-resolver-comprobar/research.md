# Research · H23 · `cita resolver`: comprobar que una sentencia existe, por el formulario del CENDOJ + skill `jurisprudencia`

Fase 0 del plan. Modo desatendido: cada decisión se tomó con el «Criterio de decisión autónoma» de la constitución y
lleva su alternativa rechazada. Toda afirmación sobre una herramienta, una dependencia o el propio repositorio remite a
la tabla V (comprobada en local, en esta sesión), las medidas a la tabla M, y lo que no se pudo comprobar sin red o sin
las grabaciones, a la tabla S, como supuesto.

**Lo que no se ha hecho en esta sesión**: no se ha ejecutado `make ci` ni ningún test del repositorio; no se ha
consultado ninguna fuente —la sesión no tiene red—; no se ha abierto ninguna sesión con modelo ni se ha lanzado ningún
subagente; y no se ha leído ninguna credencial. Lo comprobado es lectura del repositorio, de la biblioteca estándar de
go1.27.1 y de la caché de módulos, y un prototipo de un solo uso.

**Material de un solo uso, fuera de lo versionado**, en `h23-proto/` del directorio temporal de la sesión, donde queda:
`main.go`, con la regla de `internal/httpx/nombre.go` copiada y extendida con los campos de un formulario (D6) y los
ejemplos de los contratos serializados con `encoding/json` —solo usa la biblioteca estándar, y de él salen M1 a M4—; y
`precheck.awk`, una réplica de las reglas de la fase plan de `scripts/workflow/precheck.sh` sobre «Controles de
umbral», que enseña cada fila y cada control (22 y 33). El guion de verdad se ejecutó al terminar
(`bash scripts/workflow/precheck.sh plan`): `precheck_plan ok`.

**Corrección del plan, ronda 1 (2026-10-06)**. El juez rechazó V17 y D16: daban `cita resolver` por el primer verbo con
banderas propias y concluían que marcarlas en `--describe` no cambiaba ningún esquema publicado. Las dos están
rehechas. En la sesión del corrector se ejecutó, en el repositorio, `go run ./cmd/kitlegal <applet> <verbo> --describe`
de los 14 verbos del registro (V17); y, sobre una copia del árbol de trabajo en `h23-corr/wt/` del directorio temporal
—sin `.git`, `.github`, `specs/` ni `web/`, con un bucle añadido a `internal/cli/describe.go` que pone `title` a cada
bandera propia y, después, con la tabla de `internal/skills/comandos.go` leyéndolo—, los tests y la regeneración que
nombra M5. Esa copia es otro material de un solo uso, que queda en el directorio temporal: nada de ella entra en el
repositorio. Tampoco en esa sesión se ha ejecutado `make ci` ni ningún test sobre el repositorio.

## V · Verificado en local, en esta sesión

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | `internal/httpx` rechaza con la clase `argumentos`, antes de abrir nada, todo método que no sea GET o HEAD, y `Peticion` no lleva cuerpo ni cabeceras: solo `Metodo`, `URL` y `Acepta` | `internal/httpx/cliente.go:559-562`; `internal/httpx/peticion.go:15-25` |
| V2 | El cliente HTTP del paquete se construye con `Jar: nil` y con `CheckRedirect` que devuelve la última respuesta; las redirecciones las sigue `Pedir`, hasta 10 saltos y con el mismo método | `internal/httpx/transporte.go:40-62`; `internal/httpx/cliente.go:20-25`, `:625-674` |
| V3 | Con `Jar` no nulo, `http.Client` inserta en cada petición las cookies del almacén para su dirección y lo actualiza con las de cada respuesta; con `Jar` nulo solo viajan las que lleve la petición. `cookiejar.New(nil)` es válido: un `Options` nulo equivale al cero | `go doc net/http.Client`; `src/net/http/client.go:81-90`, `:182-193` de go1.27.1; `go doc net/http/cookiejar.New`; `go doc net/http/cookiejar.Options` |
| V4 | `http.NewRequestWithContext` con un cuerpo `*strings.Reader` o `*bytes.Reader` rellena `ContentLength` y `GetBody` | `go doc net/http.NewRequestWithContext` |
| V5 | `url.Values.Encode` codifica en la forma `application/x-www-form-urlencoded`, ordenada por clave | `go doc net/url.Values.Encode` |
| V6 | Lo que `Pedir` entrega según el estado final: 429, fallo de clase `limite-o-tos`; 5xx, `fuente-no-disponible`; un 3xx que no se puede seguir, `fuente-no-disponible`; cualquier otro —403, 404, 410, 400— se entrega al adaptador con su `Estado` | `internal/httpx/cliente.go:794-821`; `internal/httpx/errores.go:183-212` |
| V7 | `ConIntentos(1)` es la forma de no reintentar; el decorador de reintentos baja cada intento con `peticion.Clone(ctx)`, que no repone un cuerpo ya leído. Solo se repiten un 5xx y un fallo de transporte | `internal/httpx/cliente.go:211-232`; `internal/httpx/reintentos.go:82-125` |
| V8 | La cadena de `New` es transporte → marca de emisión → grabación (si está activa) → ritmo → reintentos → `robots.txt` → identificación; la de `Replay`, transporte de reproducción → marca → identificación, sin ritmo, reintentos ni `robots.txt`. El `robots.txt` lo pide el decorador con su propia petición, que no pasa por `http.Client` | `internal/httpx/cliente.go:306-336`, `:360-383`; `internal/httpx/robots.go:184-196` |
| V9 | No obtener `robots.txt` sin que venza el contexto es `limite-o-tos` (5); el turno que no cabe en el plazo y el contexto vencido, `fuente-no-disponible` (4) | `internal/httpx/errores.go:183-197`; `internal/httpx/cliente.go:640-643`, `:745-757` |
| V10 | Una grabación se nombra con el método y la dirección —`<MÉTODO>_<esquema>_<host><ruta>[_q_<consulta>]`, saneado; si pasa de 120 caracteres, los 100 primeros, un guion y 8 hexadecimales de `sha256("<MÉTODO> <dirección>")`— y se empareja por `peticion.metodo` y `peticion.url` del fichero, exactos. Guarda las cabeceras de la petición y de la respuesta. `formato` es 1 | `internal/httpx/nombre.go:14-60`, `:147-156`; `internal/httpx/grabar.go:27-30`, `:53-81`, `:166-190`, `:226-256`; `internal/httpx/reproducir.go:51-115` |
| V11 | El patrón de una fuente: nombre, ritmo y términos en `terminos.go`, atados a su fila por `TestFuenteCoincideConSources`; manifiesto `testdata/grabaciones.json` con `fuente`; `TestGrabarFixtures` con `//go:build grabacion`, que falla sin `KITLEGAL_RECORD=1`; un `Ritmo` por dependencias, compartido por los clientes de un proceso | `internal/source/boe/terminos.go`; `internal/source/boe/testdata/grabaciones.json`; `internal/source/boe/grabacion_test.go:1-66`; `internal/app/boe.go:41-71` |
| V12 | La fuente de `boe` valida antes de abrir nada, abre la caché en solo lectura con `--offline` o `--dry-run`, sirve la entrada vigente sin construir el cliente, con `--offline` y sin entrada da `fuente-no-disponible`, y bajo `--dry-run` devuelve en `Resultado.Ensayo` la línea de cada petición que no emitió. La clave es `<fuente>\|<versión>\|<verbo>\|<dirección>` | `internal/source/boe/fuente.go:340-378`, `:438-440`, `:471-521`; `internal/source/boe/entradas.go:26-45`, `:108-111` |
| V13 | El kernel traduce la clase del error al código (2, 3, 4, 5, 6, 7 y 1) en un solo `switch`; un adaptador declara su clase con `schema.ConClase` | `internal/cli/errors.go:94-161`; `internal/core/schema/error.go:9-76` |
| V14 | Un applet que no consulta ninguna fuente firma con `kitlegal.<nombre>` y `kitlegal:applet/<nombre>`; el ADR 0006 prevé `cita` entre ellos | `docs/ADR/0006-*.md:88-93`; `internal/app/territorio.go:19-22` |
| V15 | Las operaciones de grafo viajan en `Resultado.Grafo`, con la procedencia del sobre; los tipos de nodo son constantes de `internal/core/grafo/vocabulario.go` y `Resolucion` no está; `graph check` solo da `fuente-caducada` de una `Norma`, un `Bloque` o una `BloqueVersion`; `ValidarID` admite un id con dos puntos; `schemas/grafo.json` no enumera tipos de nodo | `internal/core/schema/sobre.go:90-123`; `internal/core/grafo/vocabulario.go:5-22`; `internal/core/grafo/comprobar.go:285-323`; `internal/core/grafo/id.go:25-41`; `grep -n Municipio schemas/grafo.json` (sin resultados) |
| V16 | Las herramientas MCP salen del registro, una por verbo, salvo las de `mcp` y `skills`; `LineaDeLlamada` escribe cada campo que no es de posición como `--<nombre>=<valor>`, el terminador `--` y los de posición | `internal/app/herramientas.go:35-73`; `internal/cli/herramienta.go:118-178` |
| V17 | El documento de `--describe` da la `entrada` como un objeto plano, sin decir qué propiedad va por su posición y cuál es una bandera propia; la tabla de comandos toma por argumento de posición toda propiedad que no se llama como una bandera global, y lo dice: «todos los argumentos propios de un verbo de una tabla son de posición». De los 14 verbos del registro, tres tienen banderas propias: `skills install` (`--global`, `--host`, `--dir`) y `skills list` y `skills doctor` (`--global`, `--dir`), siete propiedades entre los tres; los otros once solo tienen argumentos de posición o ninguno. Ninguna skill declara el applet `skills`, así que ningún verbo de una tabla de comandos de hoy tiene banderas propias. `TestTablaDeComandosCoincideConLaGramatica` sí genera una fila por cada verbo del registro, también de esos tres: hoy las escribe como argumentos de posición opcionales, y su invocación mínima omite todo lo que empieza por `[` | `internal/cli/describe.go:180-197`; `internal/skills/comandos.go:273-292`, `:585-610`; `internal/app/instalacion.go:163-167`, `:184-186`; `grep -n kitlegal-applets skills/*/SKILL.md` (`boe graph` y `territorio`); `internal/app/skills_test.go:158-197`, `:1466-1486`; `go run ./cmd/kitlegal <applet> <verbo> --describe` de los 14 verbos (sesión del corrector) |
| V18 | `jsonschema.Schema` de `invopop/jsonschema` v0.14.0 tiene `Title`, que se escribe como `title` | `~/go/pkg/mod/github.com/invopop/jsonschema@v0.14.0/schema.go:70` |
| V19 | Una skill declara sus applets en `metadata.kitlegal-applets` y sus referencias en `kitlegal-referencias`, que puede faltar; lo empotrado es `skills/*/SKILL.md` y `skills/*/references/*`; `skillsExigidas` son las que `skills/` tiene que tener | `internal/skills/frontmatter.go:42-43`, `:98-103`; `skills.go:24`; `internal/app/skills_test.go:56-60`, `:93-100` |
| V20 | El formato de eval: `comandos` con cinco formas que decide el verbo (`resolver` es hoy la de territorio); una eval que activa exige `comandos` y `citas` o `territorio`; la cita se extrae con `formaDeCita`; las formas fijas, con `patronDeEtiqueta` | `internal/evals/formato.go:103-214`; `schemas/eval.yaml.json:18-98`; `internal/evals/citas.go:8-45`; `internal/evals/avisos.go:16-93` |
| V21 | Un comando esperado solo lo satisface una invocación que consultó y terminó con 0; la caché de cada sesión se prepara ejecutando en proceso las consultas necesarias de todas las evals con el applet `boe` sobre `httpx.Replay`, y una que no termina en 0 es una falta | `internal/evals/juzgar.go:943-971`; `internal/evals/preparar.go:147-197`, `:401-431`; `internal/evals/consultas.go:87-123` |
| V22 | Los textos que devolvieron las herramientas de una sesión —la orden de Bash que nombra `kitlegal` y la llamada a una herramienta del registro, también si fallan— están en `Sesion.Textos` | `internal/evals/sesion.go:153-185`, `:640-700` |
| V23 | Los umbrales de las respuestas (`sin_activar` y los de las clases) solo existen si la skill tiene juez: `legal-core` no tiene ninguno. Los grupos de respuestas por modo los da `respuestasQueSeJuzgan`, que no depende del juez | `internal/evals/umbrales.go:159-196`, `:234-246`; `internal/evals/informe.go:418-437`, `:1399-1437`; `specs/017-h24-las-evals-juzgan/research.md:135-137` |
| V24 | Las sesiones de eval corren con los proxies apuntando a `127.0.0.1:9`: una consulta que no está en la caché termina con 4 o con 5, queda en `fuera_de_lo_grabado` y no cambia si la eval pasa; `red` solo lleva conexiones de clase `red` | `internal/evals/sesiones.go:69-73`, `:545-560`; `internal/evals/juzgar.go:188-201`, `:245-256` |
| V25 | La matriz del job es `skill: [boe-legislacion, legal-core]`, con `concurrencia` y `objetivo_de_duracion` por `include`; `timeout-minutes: 352`; el peor caso de un trabajo es `485 + Σ⌈N/C⌉ × (22 + 240 + 10)` s y `TestDefinicionDelJob` lo recalcula | `.github/workflows/evals.yml:162-186`; `internal/evals/definicion.go:57-75`, `:576-612` |
| V26 | `make verify-sources` ejecuta `scripts/verify-sources.sh`, que lanza `go test -tags=fuentes -run '^TestVerificarFuentes$' ./internal/app/`; el flujo nocturno ejecuta `make verify-sources` | `Makefile` (`verify-sources`); `scripts/verify-sources.sh`; `internal/app/fuentes_red_test.go`; `.github/workflows/nightly.yml:91-113` |
| V27 | `grabar_datos` exige la fila de la fuente en `docs/SOURCES.md` de `main` con «Revisado» fechado, ejecuta `go test -tags=grabacion -run '^TestGrabar' ./<paquete>/` con `KITLEGAL_RECORD=1`, lo repite una vez al minuto y, si falla dos veces, se detiene con su dossier; solo admite ficheros bajo `testdata/` del paquete y `evidencias/<hito>/` | `scripts/workflow/grabar-datos.sh:57-65`, `:115-134` |
| V28 | La fila de `cendoj.jurisprudencia` está en `docs/SOURCES.md` de `main`, con «Revisado» 2026-10-03, ritmo `5s`, base `https://www.poderjudicial.es/search`, términos `https://www.poderjudicial.es/search/indexAN.jsp` y el formulario `POST /search/search.action` con los campos `ECLI`, `ROJ` o `NUMERORESOLUCION` y fechas | `git show main:docs/SOURCES.md` (una fila con `cendoj.jurisprudencia` y `2026-10-03`) |
| V29 | Lo probado a mano del formulario: `GET /search/indexAN.jsp` (200, cookie de sesión) y `POST /search/search.action` con la cookie, `X-Requested-With: XMLHttpRequest` y los campos `action=query`, `sort=IN_FECHARESOLUCION:decreasing`, `recordsPerPage=10`, `databasematch=AN`, `start=1` y el de la referencia; las fechas, en `FECHARESOLUCIONDESDE` y `FECHARESOLUCIONHASTA` como `04/07/2023`; el documento, en `/search/AN/openDocument/{hash}/{fecha}`; cero resultados, «No se ha encontrado ningún resultado» | `docs/JURISPRUDENCIA.md` §3 |
| V30 | «No entra ninguna dependencia nueva en el binario» es una consecuencia aceptada del ADR 0036; `golang.org/x/net` no está en `go.mod` | `docs/ADR/0036-*.md` («Consecuencias»); `go.mod` |
| V31 | Un guion de aceptación tiene que fallar hoy por una aserción y no por una orden desconocida; al activarse se copia a `internal/app/testdata/script/` con el prefijo del hito, `h23-` | `scripts/workflow/aceptacion.sh:36-63`, `:74-92` |
| V32 | `testscript` desempaqueta el archivo del guion antes de llamar a `Setup`; el arnés copia en `Setup` las grabaciones a `$WORK/reproduccion/<fuente>` con `os.CopyFS`, que no sobrescribe; un guion pone en juego otra respuesta copiándola con `cp` sobre la de su reproducción | `~/go/pkg/mod/github.com/rogpeppe/go-internal@v1.16.0/testscript/testscript.go:528-545`; `go doc os.CopyFS`; `internal/app/e2e_test.go:556-573`, `:660-690` |
| V33 | El binario de e2e registra los applets de producción con `boe` sobre la reproducción y las skills empotradas de verdad; su reloj se fija con `-ldflags` | `internal/app/ejemplo/kitlegal-e2e/main.go:128-164`; `internal/app/e2e_test.go:120-143` |
| V34 | Registrar un applet cambia cuatro listas literales de applets y un ejemplo: `internal/app/registro_test.go:249`, `cmd/kitlegal/main_test.go:20`, `internal/app/ejemplo/kitlegal-e2e/main_test.go` y `internal/app/testdata/script/argumentos.txtar:24` y `:30`; `CONTRIBUTING.md:85`. Y como el servidor MCP anuncia una herramienta por verbo, cambia también lo que cuenta las herramientas: los dos recuentos de `TestHerramientasDelServidor` (`internal/app/herramientas_test.go:134-135`, 10 y 13) y tres guiones de H21 cuya salida esperada lleva `1 herramientas 13` y se compara con `cmp stdout`: `h21-mcp-herramientas.txtar:40`, con su lista de trece líneas y los comentarios que las cuentan; `h21-mcp-proceso.txtar:69`; y `h21-mcp-protocolo.txtar:63`, `:69` y `:79`. Los tres corren con el binario de e2e. Ninguno de los dos últimos contiene `applets disponibles`, `graph_stats` ni una lista de applets: solo el recuento | `grep -rn "applets disponibles"`; `grep -rn '"boe", "graph"'`; `git grep -n 'herramientas 1[0-9]'`; M6 |
| V35 | Una skill nueva en `skills/` cambia lo que instala el binario de e2e y lo que instala `make install`. En `internal/app/testdata/script/`, 17 guiones: los 15 que enumeran las dos skills de hoy en la misma línea, 72 líneas en total (salidas de `skills install`, `list` y `doctor`, árboles y el mensaje «skills disponibles»), y dos que no las nombran juntas y cuentan una versión por skill en el manifiesto, `h19-skills-aviso.txtar:144`, `:145` y `:158` y `h19-skills-aviso-sin-aviso.txtar:94` (`grep -count=`). En `internal/skills/testdata/script/`, tres de los cuatro guiones que ejecuta `TestInstalacion` (`//go:build integration`; en `make ci` por `test-integration`, `Makefile:91` y `:264`) sobre una copia de `skills/` entera (`internal/skills/instalacion_test.go:73`), con `make install`: `instalar.txtar:28` y `:33` e `instalar-sin-gobin.txtar:19` comparan con `\A…\z` el listado completo de lo instalado, e `instalar-de-nuevo.txtar:19-20` lo enumera línea a línea, sin fallar; `instalar-con-conflicto.txtar` no instala nada y no cambia. La primera cuenta de esta fila, 15, solo miraba las líneas con las dos skills juntas en un directorio | `grep -c "boe-legislacion.*legal-core" internal/app/testdata/script/*.txtar`; `git grep -n 'legal-core' -- '*.txtar'`; M6 |
| V36 | `ficherosDeEsquemas` da a cada applet su fichero de `schemas/`, con el nombre de su entidad, y `TestEsquemasPublicados -actualizar-esquemas` lo escribe desde `--describe`. Cada parte publicada es el documento entero de su verbo, con su `entrada`: las de `doctor`, `install` y `list` en `schemas/instalacion.json` llevan `dir` y `global`, y la de `install`, `host`. La web sirve `schemas/` tal cual en `https://kitlegal.es/schemas/`, y su flujo se lanza con un cambio en `schemas/**` | `internal/app/esquemas_test.go:23-70`; `schemas/instalacion.json:115`, `:121`, `:306`, `:312`, `:315`, `:548`, `:554`; `web/src/pages/schemas/[nombre].json.ts`; `.github/workflows/web.yml:19`, `:27` |
| V37 | El texto de la eval (d) existe y ocupa 2 353 bytes | `wc -c evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` |
| V38 | `depguard` reserva `net/http` a `internal/httpx/**`, por prefijo de ruta de fichero | `.golangci.yml:156-171`; `internal/arch_test.go:81-90` |
| V39 | El README dice hoy, en «¿Y las sentencias?», que la jurisprudencia del CENDOJ «no llegará por ahora» | `README.md:259-277` |

## M · Medido en esta sesión (prototipo `h23-proto`; M5, en la del corrector, con `h23-corr`; M6, en la del corrector de tareas, con `h23-tareas`)

| # | Medida | Resultado |
|---|---|---|
| M1 | Nombre de la grabación de cada envío con la regla de D6 | los siete de [contracts/fuente-cendoj-y-grabacion.md §4](./contracts/fuente-cendoj-y-grabacion.md); el de una petición sin cuerpo no cambia (`GET_https_www.boe.es_…_bloque_a21.json`) |
| M2 | Cuerpo de un envío, codificado | 126 B (ECLI), 112 B (ROJ), 193 B (número con sus dos fechas) |
| M3 | JSON compacto: una resolución, 318 B; `data` con una, 384 B; el sobre con una, 625 B; con diez, 3 495 B; el del Tribunal Constitucional, 297 B; el de «no encontrado» por la fecha, 422 B | `encoding/json` sobre los ejemplos de [contracts/cita-resolver.md](./contracts/cita-resolver.md), con una URL de documento supuesta de 93 caracteres |
| M4 | La entrada de caché de una entrega con una resolución, 502 B; la de un «no encontrado», 270 B; los datos del nodo `Resolucion`, 318 B; un elemento de `umbrales`, 355 B; la cita, 71 caracteres; la línea, 147 | ídem |
| M5 | Lo que cambia al marcar con `title` cada bandera propia de todo verbo en `--describe` y al escribirla la tabla como `[--<nombre>=<nombre>]` detrás de los argumentos de posición (D16) | Con la marca y `schemas/` sin regenerar fallan dos tests, y ningún otro de `internal/cli`, `internal/skills`, `internal/mcp/...`, `internal/app/...`, `internal/core/instalacion`, `internal/source/...` y `cmd/...`: `TestEsquemasPublicados/schemas`, por las partes `doctor`, `install` y `list` de `schemas/instalacion.json`, y `TestEsquemasDeHerramienta/los_dos_esquemas_son_las_partes_del_documento_de_--describe`. Con `-actualizar-esquemas`, de `schemas/` solo cambia `instalacion.json` (`diff -rq`): siete líneas añadidas y ninguna quitada —`"title": "--dir"` y `"title": "--global"` en `doctor`, `install` y `list`, y `"title": "--host"` en `install`—, de 15 977 a 16 211 bytes; y con él regenerado, `internal/app/...` vuelve a pasar. Con la tabla leyendo la marca, los tests de `internal/skills` pasan sin tocarlos, y también `TestSkillsDelRepositorio` —las tablas de `boe-legislacion` y de `legal-core` no cambian—, `TestOrdenesDeLasSkillsEmpotradas` y `TestTablaDeComandosCoincideConLaGramatica`, con sus tres subtests de `skills`. `internal/evals` no se midió: la copia no lleva `.github/`, y ningún fichero Go de ese paquete nombra `cli.Describir`, `EsquemasDeHerramienta`, `LeerDescripcionDeVerbo` ni `instalacion.json` (`grep`) |
| M6 | Qué tests y guiones de hoy fallan al registrar un applet `cita` y al añadir una skill sin referencias. Un árbol de trabajo desechable de la cabeza de la rama, bajo el directorio temporal y ya retirado, con un applet `cita` de relleno —un verbo `resolver` con un argumento de posición y tres banderas, sin fuente— registrado en producción y en el binario de e2e, y después con el prototipo de la skill copiado a `skills/jurisprudencia/SKILL.md`; `go test -count=1 -tags=integration ./...`, sin `-race` y sin las dos medidas de tiempo | **Con el applet** fallan `TestPuntoDeEntrada`, `TestRegistroDeProduccion`, `TestRegistroDeE2E`, `TestEsquemasCubrenTodosLosVerbos`, las cuatro subpruebas de `TestHerramientasDelServidor` y cuatro guiones de `TestEntregaDelHito` —`argumentos`, `h21-mcp-herramientas`, `h21-mcp-proceso` y `h21-mcp-protocolo`, los tres últimos con `1 herramientas 14` donde esperan 13—, y ningún otro test de ningún paquete: `TestHerramientasAnunciadas`, `TestEjecutar` y los del paquete raíz pasan, porque `herramientasDeHoy` solo pide «al menos» (`Subset`). **Con la skill, además**: `TestEvalsDelRepositorio/linea-sin-consulta`; `TestInstalacion/instalar` e `instalar-sin-gobin`; diecisiete guiones de `TestEntregaDelHito`, los quince de V35 más `h19-skills-aviso` y `h19-skills-aviso-sin-aviso`; y, de `TestSkillsDelRepositorio`, `regenerar-dos-veces/jurisprudencia`, que escribe `references/antigua.md` sin crear la carpeta, y las trece subpruebas de `frontmatter/jurisprudencia`, cuyo ayudante `metadataDeKitlegal` escribe `kitlegal-referencias:` sin valor («no es una cadena»). Con esos dos ayudantes corregidos en la copia —la línea, solo si la skill declara referencias; la carpeta, creada antes de escribir—, `TestSkillsDelRepositorio`, `TestOrdenesDeLasSkillsEmpotradas` y `TestTablaDeComandosCoincideConLaGramatica` pasan con la skill. Una tercera carpeta en `evals/`, copia de la de `legal-core`, no hace fallar ningún otro test de `internal/evals`, `internal/skills`, `internal/empaquetado` ni del paquete raíz; `skillsDeHoy` pide «al menos», como `herramientasDeHoy`. **No medido**: los tests con las etiquetas `snapshot`, `evals`, `fuentes` y `grabacion`; el lint y el resto de `make ci`; las seis evals del hito, que piden el formato que aún no existe; ni el applet real, del que el relleno solo tiene el nombre, el verbo y sus argumentos |

## S · Supuestos no verificados

| # | Supuesto | Por qué no se pudo comprobar | Qué lo mide o lo acota |
|---|---|---|---|
| S1 | La estructura HTML de las dos respuestas del formulario | No hay ninguna respuesta del buscador en el repositorio, y la sesión no tiene red | Las grabaciones de `grabar_datos`; la lectura fina se escribe con ellas delante (D10) |
| S2 | El buscador responde lo mismo con los campos en orden alfabético y con `:`, `/` y el espacio codificados (`%3A`, `%2F`, `+`) | La prueba a mano no dice cómo se codificaron | El test de grabación: una respuesta que no se reconoce no se graba y el run se detiene (D27) |
| S3 | Sin `X-Requested-With: XMLHttpRequest` el formulario respondería lo mismo | Solo se probó con ella | No se depende de ello: se envía (D5) |
| S4 | Las respuestas a las tres consultas que solo usan las evals (b) y (f) —`NUMERORESOLUCION=9999/2023` y `3144/2023` con el 1 de enero de 2023, y `ROJ=STS 9999/2023`— | No se han probado | Cualquiera de las dos reconocidas vale: con resultados de otra fecha o sin ellos, la consulta termina con 3 (D24) |
| S5 | Qué cookie da la página y con qué atributos | No hay grabación | No se depende de su nombre: la guarda el almacén de la consulta (D3) |
| S6 | La dirección del buscador del Tribunal Constitucional es `https://hj.tribunalconstitucional.es/` | No está en el repositorio (el spec lo declara) y no hay red | La persona, al leer el informe final (`gates/supuestos.md`); está en un solo sitio (D29) |
| S7 | Las cotas de un ECLI: órgano de 1 a 7 caracteres y número de orden de hasta 25 | `DOUE-Z-2019-70039` no está en el repositorio | El spec las fija (FR-003) y el plan las toma de él |
| S8 | `gitleaks` no marca la cookie de sesión que guardan las grabaciones | Las grabaciones no existen todavía | `make ci` (`secrets`) en la tarea que parte de ellas |
| S9 | El peor caso del trabajo `evals (jurisprudencia)`, 20 613 s (D25), y lo que tardan sus sesiones | No se ha ejecutado `TestDefinicionDelJob` con la skill nueva ni ningún job | `TestDefinicionDelJob` en `make ci`; el job de cierre |
| S10 | El modelo que decide ejecuta las órdenes de la tabla con `--json` y sin filtrar su salida | Exige abrir una sesión con modelo | `cita_sin_resolver`, que falla hacia el rojo (D21) |
| S11 | El cliente de la app de escritorio acepta el esquema de entrada de `cita_resolver` | Como S6 de H21 | La aceptación humana (SC-008) |

## D · Decisiones

### D1 · Dónde vive cada pieza

`internal/core/ids` gana el reconocimiento de ECLI y de ROJ (`ecli.go`, `roj.go`): son identificadores naturales
(constitución VII), el paquete ya tiene INE y DIR3 con su error de clase `argumentos`, y es donde `docs/ROADMAP.md` §3
pone el fuzzing de identificadores. El adaptador es `internal/source/cendoj`. El applet, `internal/app/cita.go`. El
formulario, en `internal/httpx`. El formato de eval y el umbral, en `internal/evals`.
**Rechazado**: `internal/core/cita`, que es de H8 —las citas de normas y `validar`— y hoy tendría dos funciones.

### D2 · El formulario de consulta en `internal/httpx`: una opción y un campo

`ConFormulario(direccion string)` declara, al construir el cliente, la única dirección a la que ese cliente puede
enviar un formulario. `Peticion` gana `Campos map[string]string`. Un `POST` solo se admite desde una `Consulta` (D3) de
un cliente que declaró el formulario, a esa dirección exacta y con al menos un campo; cualquier otro `POST`, cualquier
otro método y unos `Campos` con GET o HEAD siguen siendo `argumentos` antes de abrir nada (FR-030, FR-031). El cuerpo es
`url.Values.Encode` de los campos (V5), con `Content-Type: application/x-www-form-urlencoded`.
**Rechazado**: un juego de cabeceras o un cuerpo libres en `Peticion`, que abrirían el paquete a cualquier `POST`; y un
método `EnviarFormulario` aparte de `Pedir`, que duplicaría el ensayo, las comprobaciones y el instante.

### D3 · La cookie vive en una `Consulta`

`(*Cliente).Consulta()` devuelve una `Consulta` con un `http.Client` propio sobre **la misma cadena** del cliente y un
`cookiejar` nuevo (V3). Sus peticiones —la página y el formulario— comparten las cookies que el sitio les dé; al
soltarla no queda nada, y la siguiente `Consulta` empieza sin ninguna (FR-032). Un cliente sin `ConFormulario` no da
consultas (`argumentos`) y sigue con `Jar: nil`. La identificación, `robots.txt`, el ritmo, la grabación y el contexto
son los de la cadena, que no cambia (FR-033); el `robots.txt` no recibe la cookie, porque no pasa por `http.Client` (V8).
**Rechazado**: un almacén en el cliente, que llevaría la cookie de una consulta a la siguiente en el servidor MCP,
donde el ritmo y las dependencias viven todo el proceso; y copiar a mano `Set-Cookie` en `Cookie`, que es reescribir lo
que la biblioteca ya hace con dominio, ruta y caducidad.

### D4 · Dentro de una `Consulta` no se sigue ninguna redirección

La respuesta 3xx se entrega con su `Estado`, sin pedir nada más (FR-021, FR-033). El adaptador la trata como cualquier
estado que no es 200: respuesta que no se reconoce (5). Fuera de una `Consulta`, `Pedir` sigue como hoy, y `robots.txt`
se obtiene como hoy.
**Rechazado**: una opción `SinRedirecciones` aparte, que habría que acordarse de declarar junto al formulario.

### D5 · El envío lleva `X-Requested-With: XMLHttpRequest`

Es la petición que se probó a mano (V29) y la que hace la propia página del buscador. No cambia el agente, que sigue
siendo el del proyecto. Es una constante de `internal/httpx` para todo envío de formulario; la fila de
`docs/SOURCES.md` no la nombra, y queda como supuesto en `gates/supuestos.md` para que la persona lo vea.
**Rechazado**: omitirla. No se sabe qué responde el buscador sin ella (S3), y descubrirlo en `grabar_datos` detendría
el run.

### D6 · Dos envíos a la misma dirección se graban por separado

El nombre de la grabación de una petición con campos es el de hoy seguido de `_c_` y del cuerpo codificado, saneado con
la misma regla; si pasa de 120 caracteres, el resumen se calcula sobre `"<MÉTODO> <dirección>\n<cuerpo>"`. El fichero
guarda el cuerpo en `peticion.cuerpo`, y la grabación y la reproducción lo comparan además del método y la dirección.
Una petición sin campos no lleva la clave, su nombre y su resumen son los de hoy y `formato` sigue en 1, así que las
grabaciones que hay se reproducen sin tocarlas (FR-034; M1). El decorador de grabación y el transporte de reproducción
leen el cuerpo con `GetBody` (V4), sin consumir el que se envía.
**Rechazado**: subir `formato` a 2, que obligaría a regrabar o a leer dos formatos; y nombrar solo con el resumen, que
deja en `testdata/` ficheros que no se sabe qué consultan.

### D7 · Reintentos: el cliente de la fuente declara uno

`ConIntentos(1)` (V7; FR-021). El decorador de reintentos no cambia. Un cliente con formulario y más de un intento no
existe ni lo crea este hito; no se le escribe ninguna guarda (regla genérica).

### D8 · El ritmo: 5 s, uno por dependencias

`cendoj.IntervaloEntrePeticiones = 5 * time.Second`, en `terminos.go`, y un `httpx.NuevoRitmo` por
`DependenciasDeCita`, como `boe` (V11): en la orden, de un proceso; en el servidor MCP, de todas las llamadas (FR-024).
Nada se guarda entre procesos.

### D9 · Qué campo consulta cada forma

| Forma | Campos, además de los cinco fijos |
|---|---|
| ECLI | `ECLI=<ecli>` |
| `--roj`, con `--fecha` o sin ella | `ROJ=<roj>` |
| `--resolucion` con `--fecha` | `NUMERORESOLUCION=<número/año>`, y `FECHARESOLUCIONDESDE` y `FECHARESOLUCIONHASTA` con la fecha como `dd/mm/aaaa` |

Los fijos: `action=query`, `databasematch=AN`, `recordsPerPage=10`, `sort=IN_FECHARESOLUCION:decreasing`, `start=1`
(V29; FR-006, FR-020).

### D10 · Reconocer en dos pasos, con la biblioteca estándar

1. **Clase de la respuesta**, con las dos marcas que el repositorio documenta (V29): con estado 200, «No se ha
   encontrado ningún resultado» es «sin resultados»; si no, un enlace `/search/AN/openDocument/` es «lista»; cualquier
   otra cosa, y cualquier estado que no es 200, «no reconocida». Se puede escribir antes de grabar, y es lo que usa el
   test de grabación (FR-092).
2. **Lectura de la lista**: de cada resultado, los ocho campos. Se escribe con las grabaciones delante, sobre sus
   goldens (FR-110), con expresiones de la biblioteca estándar ancladas en lo que traigan. Un resultado sin ECLI, ROJ,
   fecha o URL hace «no reconocida» toda la respuesta (FR-022). El resumen no se lee.

Sin dependencia de HTML: el ADR 0036 lo fija (V30).
**Rechazado**: fijar ahora selectores sobre una estructura que no se ha visto (S1); y `golang.org/x/net/html`, que el
ADR descarta.

### D11 · La forma de `data`

Un objeto con `resoluciones`, `pagina_completa` y `cobertura`
([contracts/cita-resolver.md §3](./contracts/cita-resolver.md)). `fecha` va como `AAAA-MM-DD`, la misma forma de
`--fecha`, y es la que se compara (FR-011). `organo` lleva órgano y sala en un texto, como los da la fuente.
`pagina_completa` es verdadero cuando la fuente devolvió 10 resultados, antes de filtrar por fecha (FR-014).
`cobertura` es `cubierto` o `tribunal-constitucional-no-cubierto` (FR-015), siempre presente.
**Rechazado**: una lista desnuda en `data`, que no deja sitio a las dos declaraciones; y `organo` y `sala` separados,
que obligaría a partir un texto cuya forma no se ha visto.

### D12 · El mensaje de «no encontrado»

Nombra la forma, el valor y la fecha pedidos, y nada de lo encontrado con otra fecha (FR-012): «no se ha encontrado
ninguna resolución con el ROJ STS 3144/2023 y la fecha 2023-01-01: ninguna de las encontradas con ese ROJ tiene esa
fecha». Los fallos 4 y 5 nombran la petición, como en `boe`.

### D13 · La caché

Clave `cendoj.jurisprudencia|1|resolver|<forma>|<valor>|<fecha>`, con `forma` `ecli`, `roj` o `resolucion` y la fecha
vacía si no se pidió. Contenido: `fecha_consulta`, `url` y, o bien `datos` (la entrega), o bien `no_encontrado` (el
mensaje). Vigencia, 30 días (FR-040, FR-041). No se guardan los fallos 4 y 5 ni la declaración del Tribunal
Constitucional.
**Rechazado**: guardar la lista de la fuente sin filtrar para servir dos consultas por ROJ con una petición: dejaría en
la caché la resolución de otra fecha, contra FR-041.

### D14 · El grafo

Un `schema.Nodo` por resolución de la entrega: id, su ECLI; tipo, `Resolucion` (constante nueva); datos, los ocho
campos; `Observado.Vigencia`, la de la caché. Sin aristas, sin texto y sin `Persona` (FR-042). `graph show` lo enseña
sin cambios y `graph check` no da hallazgos de él (V15; FR-043).

### D15 · El Tribunal Constitucional lo reconoce el applet

Tras validar la referencia y antes de componer la fuente: procedencia `kitlegal.cita` / `kitlegal:applet/cita` (V14),
`data` sin resoluciones y con su `cobertura`; ninguna petición, ni caché, ni grafo (FR-015).
**Rechazado**: que lo devuelva la fuente `cendoj.jurisprudencia`, cuyo nombre no puede firmar lo que no ha consultado.

### D16 · `--describe` dice qué argumento propio es una bandera, en todo verbo, y la tabla de comandos lo lee

`cita resolver` es el primer verbo de una tabla de comandos con banderas propias: los tres de `skills` ya las tienen,
pero ninguna skill declara ese applet y ninguna tabla los escribe (V17). Hoy la tabla escribiría `--roj`,
`--resolucion` y `--fecha` como argumentos de posición, una orden que no funciona, porque la `entrada` de `--describe`
no distingue una cosa de la otra.

El documento de `--describe` marca cada bandera propia con `title`, que lleva cómo se escribe (`--roj`). La regla es
una y no mira el applet: la lleva todo campo de los argumentos de un verbo que no va por su posición. Las banderas
globales y los argumentos de posición no cambian. La tabla escribe los argumentos de posición como hoy y, detrás, cada
bandera propia: `[--roj=<roj>]`. Los esquemas de una herramienta MCP no llevan la marca: una herramienta recibe sus
argumentos por su nombre.

**Lo que cambia de lo publicado** (M5): el documento de `skills install`, `skills list` y `skills doctor` gana `title`
en `global`, `dir` y `host`, y con él `schemas/instalacion.json`, que es ese documento publicado y que la web sirve
(V36): siete líneas más, 234 bytes, sin cambiar nada de lo que el esquema valida. Los demás ficheros de `schemas/` no
cambian: sus verbos no tienen banderas propias. El fichero lo regenera la misma tarea que cambia `--describe`, que por
eso es `[datos]`: con el documento cambiado y el fichero sin regenerar, `TestEsquemasPublicados` deja `make ci` en rojo.

Una bandera propia sin valor o que se repite solo la tienen esos tres verbos de `skills`. Su fila solo existe dentro de
`TestTablaDeComandosCoincideConLaGramatica`, que la convierte en una invocación mínima sin nada de lo opcional: no la
lee ninguna persona ni ninguna skill, y no se le diseña ni se le prueba una forma propia (umbral de materialidad).

**Rechazado**: no marcar los verbos de `skills` y `mcp`, o marcar solo los de un applet con tabla, para dejar
`schemas/instalacion.json` como está: es un caso por applet en el kernel, y dos verbos con banderas propias se
describirían de dos maneras; marcar también las globales, que cambiaría los seis ficheros publicados para decir lo que
la tabla ya sabe por el documento de un verbo sin argumentos; una palabra propia (`x-bandera`), que cada validador
trata a su manera; marcar los de posición, que cambiaría cinco de los seis ficheros en lugar de uno; y dejar la fila
como salga y explicar las banderas en prosa, que deja en la tabla una orden que no funciona.

### D17 · El esquema publicado

`schemas/resolucion.json`: applet `cita`, entidad `resolucion`, verbo `resolver` (V36). Lo escribe
`TestEsquemasPublicados -actualizar-esquemas`.

### D18 · `--dry-run`

Como `boe` (V12): caché en solo lectura; si la consulta no está en ella, `Resultado.Ensayo` lleva las líneas
`GET …/indexAN.jsp` y `POST …/search.action`, y no sale nada (FR-019).

### D19 · La salida legible

`Resultado.Legible`, compuesta de los mismos datos: una línea por campo de cada resolución, la declaración de página
completa cuando la hay y la de cobertura en el caso del Tribunal Constitucional (FR-018; ADR 0026).

### D20 · Lo que gana el formato de eval

Una forma más de `comandos`, la de resolución (`applet: cita`, `verbo: resolver`, una referencia, `fecha` y
`no_encontrado`), y cuatro claves: `sentencias`, `sentencia_no_comprobada`, `sin_sentencias` y `direcciones`
([contracts/evals-jurisprudencia.md §1](./contracts/evals-jurisprudencia.md)). El comando de resolución lo satisface la
invocación con esa referencia que termina con 0, o con 3 si declara `no_encontrado`. `formaDelComando` mira el applet
además del verbo, porque `resolver` es también el de `territorio`. Las evals de hoy validan igual (FR-072).
**Rechazado**: reutilizar `citas`, cuya pareja es norma y bloque; y un comando sin referencia («consultó algo»), que no
dice qué preparar en la caché de la sesión.

### D21 · `cita_sin_resolver`: de dónde sale lo entregado

De `Sesion.Textos` (V22): cada sobre con `ok` verdadero y `fuente` `cendoj.jurisprudencia` que haya en la salida de una
herramienta entrega los `ecli` de sus `resoluciones`. Un sobre de fallo no entrega nada, y el del Tribunal
Constitucional tampoco: no trae resoluciones (FR-081). De la respuesta se cuentan los ECLI de sus citas que no están
entregados y los ECLI sueltos —fuera de una cita y de una línea que empieza por `⚠`— que no están entregados ni en la
pregunta, sin distinguir mayúsculas. Un ECLI, en un texto, es `ECLI:` y sus cuatro partes, con los puntos del número
solo entre letras o cifras, para que el punto final de una frase no entre en él.
Si una sesión filtra la salida de la orden y el sobre no llega entero, sus ECLI no cuentan como entregados y el umbral
falla: falla hacia el rojo (S10).
**Rechazado**: volver a ejecutar las consultas de la sesión contra su caché para saber qué entregaron: es otro
mecanismo, y el requisito habla de «la salida» de la consulta.

### D22 · Los umbrales de las respuestas existen con juez o con sentencias

Hoy solo existen con juez (V23). Pasan a existir también si el conjunto de la skill declara sentencias —alguna eval
con un comando de resolución—: `sin_activar:<modelo>:<modo>` y, detrás, `cita_sin_resolver:<modelo>:<modo>`, sobre los
mismos grupos de respuestas. `boe-legislacion` y `legal-core` no declaran ninguna y sus umbrales no cambian (12 y 0).
`jurisprudencia` tiene exactamente cuatro (FR-083).
**Rechazado**: un caso por nombre de skill; y un fichero de declaración por skill, que es un mecanismo más para decir
lo que las evals ya dicen.

### D23 · La caché de las sesiones de `jurisprudencia`

`Preparar` y `ComprobarSinRed` registran, junto a `boe`, el applet `cita` sobre `httpx.Replay` de las grabaciones del
CENDOJ (`internal/source/cendoj/testdata/cendoj.jurisprudencia`). Las consultas necesarias de una eval son sus
comandos de resolución; la que declara `no_encontrado` tiene que terminar con 3, y cualquier otro código es una falta
(FR-074). El binario de producción no gana ningún modo de reproducción.

### D24 · Las seis evals

Preguntas, comandos y esperado en [contracts/evals-jurisprudencia.md §2](./contracts/evals-jurisprudencia.md). Cada
comando es un paso que FR-061 manda dar con esa pregunta, y es además lo que se prepara en la caché. En (d) el texto
trae el ECLI y la skill resuelve por él (D29). En (e) la pregunta da el ECLI del Tribunal Constitucional
(`ECLI:ES:TC:2024:79`, el de `docs/JURISPRUDENCIA.md` §1), de modo que nombrarlo no cuenta en el umbral, y no exige
ninguna consulta (FR-070).

### D25 · El trabajo `evals (jurisprudencia)`

Una entrada más en la matriz, con `concurrencia: 1` y `objetivo_de_duracion: 0`, como `legal-core`. Con una
concurrencia de 4 habría nueve sesiones a la vez contra la suscripción en lugar de las cinco con las que se midió
H7.3. Peor caso, calculado y no ejecutado (S9): 72 sesiones y la prueba de red, como mucho 74 tandas de 272 s más
485 s, 20 613 s, 343,6 minutos; cabe en los 352 de hoy, que no cambian.

### D26 · La comprobación de la fuente, a petición

`scripts/verify-sources.sh` recibe la fuente como primer argumento, que `make verify-sources` le pasa desde `FUENTE`,
sin valor por defecto. Sin argumento hace lo de hoy. Con `cendoj`, ejecuta `TestVerificarCendoj`, con la misma etiqueta
`fuentes`. Cualquier otro valor, salida 2. Un test de `make ci` ejecuta el guion con un `go` sustituto y lee los
flujos de `.github/workflows/` (FR-091).
**Rechazado**: una etiqueta de compilación propia: la orden sin fuente ya ancla el test por su nombre, y el control la
comprueba.

### D27 · La grabación

`internal/source/cendoj/testdata/grabaciones.json`, con `fuente` y las siete consultas, y `TestGrabarConsultas`
(`//go:build grabacion`). Graba en un directorio temporal: una página y, con su misma `Consulta`, los siete envíos.
Cada respuesta se clasifica (D10); con una «no reconocida» el test falla y no se copia nada; solo si todas se
reconocen se copian a `testdata/cendoj.jurisprudencia/` (FR-092). Nueve peticiones con el `robots.txt`, a 5 s: unos
40 s. Ningún material va a `evidencias/`.

### D28 · El arnés e2e

El binario de e2e registra `cita` con su cliente en reproducción sobre `reproduccion/cendoj.jurisprudencia`, que
`Setup` copia como la de `boe`. El 403 es una grabación sintética, escrita en el guion, que se copia sobre la de la
página, y la página que no se reconoce, sobre la del envío por ECLI; los nombres son los de M1 (V32; FR-094, FR-111).

### D29 · La skill

Texto completo en [contracts/skill-jurisprudencia.md](./contracts/skill-jurisprudencia.md). Dos precisiones sobre el
spec, las dos dentro de FR-061: si una sentencia viene con su ECLI, se resuelve por el ECLI aunque venga también su
ROJ; y la dirección del buscador del Tribunal Constitucional (S6) está solo en `SKILL.md`, de donde la toma la eval (e)
con un test que las ata. No nombra operadores del buscador que no consten en el repositorio.

### D30 · Las listas literales

`cita` entra en las cuatro listas de applets y en el ejemplo de `CONTRIBUTING.md`, y `cita_resolver`, en lo que cuenta
las herramientas: los dos recuentos de `TestHerramientasDelServidor` y los tres guiones de H21 que comparan lo que
anuncia el servidor (V34). `jurisprudencia` entra en `skillsExigidas`, en los diecisiete guiones de e2e que enumeran o
cuentan las skills instaladas y en tres de los de `make install` (V35). Son datos de prueba que dicen qué lleva el
binario; cambian con él, en tareas `[datos]`, y lo que cambia de cada guion es eso: la enumeración, el recuento y el
comentario que los cuenta.

Las listas que solo piden «al menos» no fallan (M6). `herramientasDeHoy` gana `cita_resolver`, porque es donde FR-051
se comprueba; `skillsDeHoy` no se toca: FR-051 solo admite en `internal/empaquetado` datos de prueba que enumeren
herramientas.

Dos comprobaciones de hoy no fallan por una lista, sino por dar por hecho algo que la skill nueva no tiene: D33.

### D31 · Lo que no se especifica

Una caché, un grafo o una grabación tocados a mano; una respuesta del buscador a medio llegar; un cliente con
formulario y varios intentos: regla genérica, defecto `inesperado` y código 1, sin caso propio.

### D32 · Un sitio de prueba para contar peticiones: `internal/httpx/httpxtest`

El control de FR-021 cuenta las peticiones de una consulta «contra un sitio de prueba», con el `robots.txt` y sin
reintentos: eso no lo da la reproducción, cuya cadena no lleva ni `robots.txt` ni reintentos (V8), y `httptest` solo
se puede importar bajo `internal/httpx` (V38). `httpxtest.NuevoSitio` levanta un servidor local y apunta lo que recibe;
`TestPeticionesDeUnaConsulta`, en el paquete del adaptador, lo usa con las opciones de red de la fuente —las de
producción, con la base del sitio de prueba y un intervalo corto—. Es el patrón de `internal/mcp/mcptest`: código de
apoyo que el binario no enlaza, y un test lo comprueba.
**Rechazado**: un cliente de mentira que cuente lo que el adaptador pide, que no vería ni el `robots.txt` ni un
reintento; y poner el test del adaptador dentro de `internal/httpx`, que obligaría a exportar del adaptador una opción
solo para tests.

### D33 · Dos comprobaciones de hoy que daban por hecho algo que la skill nueva no tiene

M6 las encontró al añadir la skill; las dos se corrigen en el test, no en la skill.

**`linea-sin-consulta`** (`TestEvalsDelRepositorio`, H21) exige hoy a toda skill de `skills/` la línea
`⚠ SIN CONSULTA AL BOE: <causa>. Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/`. Esa
línea es la respuesta que espera la eval sin binario ni servidor de un conjunto: la reconoce `ExtraerSinConsulta`, el
mismo juicio de esa eval, y la subprueba existe para atar las dos cosas. `jurisprudencia` no tiene esa eval —queda
fuera de alcance en el spec— y sin herramienta ni binario responde con `⚠ SENTENCIA NO COMPROBADA:` y el motivo «no se
ha podido consultar» (FR-063, casos límite). La subprueba pasa a exigir la línea a las skills cuyo conjunto de evals
lleva una eval con `sin_binario_ni_servidor: true`, leído de `evals/<skill>/`, y `skillsConLineaSinConsulta` queda como
premisa: `boe-legislacion` y `legal-core` tienen que estar entre ellas, para que no pase en vacío. No hay ningún caso
por nombre de skill, y una skill que gane esa eval queda obligada a la línea sin tocar el test.
**Rechazado**: escribir la línea del BOE en `skills/jurisprudencia/SKILL.md`, que contradice FR-063 y manda a una skill
del CENDOJ decir que no ha consultado el BOE; eximir a `jurisprudencia` por su nombre, que es un caso por skill; y dar a
la línea la fuente como parámetro, que cambia la forma fija que el juicio reconoce y obliga a tocar las otras dos
skills (FR-069).

**Una skill sin referencias.** El producto la admite (V19), pero dos ayudantes de `TestSkillsDelRepositorio` no:
`metadataDeKitlegal` escribe siempre la línea `kitlegal-referencias:`, que sin nombres queda sin valor y no se lee, y
`probarRegenerarDosVeces` escribe `references/antigua.md` sin crear la carpeta. Los dos dejan de darlo por hecho: la
línea se escribe solo si la skill declara referencias, y la carpeta se crea antes de escribir. Ningún caso se retira:
los de una skill con referencias siguen igual y los de `jurisprudencia` pasan a ejecutarse.
**Rechazado**: dar a `jurisprudencia` una carpeta `references/` para no tocar el test: el hito no le da ningún dato de
`data/`.
