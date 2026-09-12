## Objetivo

Construir **la única puerta por la que el módulo sale a Internet**. Hasta que existe `internal/httpx` no
se escribe ningún adaptador de fuente: es la condición previa de H4 (`boe`) y de todos los hitos de fuente
posteriores.

El proyecto consulta boletines oficiales, plataformas de contratación y bases de datos de subvenciones que
nadie nos ha invitado a consultar en masa. El principio I de la constitución convierte eso en una
obligación innegociable —identificarse, respetar `robots.txt`, no ir más deprisa de lo que el sitio tolera,
no insistir cuando la fuente pide parar, no leer nunca sin plazo—, y esa obligación no puede depender de
que quien escriba el adaptador número doce se acuerde de cumplirla. El criterio de aceptación del hito lo
dice en una frase: **un adaptador de prueba no puede hacer una petición sin contexto ni sin UA**.

Dos caras, por tanto:

1. **Un cliente que no se puede usar mal.** Un constructor (`httpx.New`) produce un cliente que ya trae
   identificación, `robots.txt`, ritmo por sitio, reintentos y plazo del contexto; la única superficie de
   red es una operación que exige `context.Context` como primer parámetro. No se exporta `*http.Client`,
   ni `http.RoundTripper`, ni `*http.Request`, ni `*http.Response`: las garantías son **por construcción**,
   no por disciplina.
2. **Una suite que nunca toca la red.** Grabación con `KITLEGAL_RECORD=1` y reproducción desde disco
   (`httpx.Replay(dir)`), para que desde H4 cada adaptador se pruebe contra lo que la fuente respondió un
   día concreto y `make ci` siga pasando en una máquina sin red.

Hito de **fundación** (principio VIII): no entrega ni cambia ninguna skill, y al cerrarlo la superficie
visible del binario es exactamente la misma que al abrirlo.

## Alcance

- **`internal/httpx`** (paquete nuevo, 17 ficheros de producción + 17 de test):
  - `cliente.go` (constructores, opciones y cadena), `peticion.go`, `sitio.go`, `agente.go`,
    `identificar.go`, `transporte.go`.
  - Cadena de decoradores sobre `http.RoundTripper`: identificación → `robots.txt` (`robots.go`) → ritmo
    por sitio (`ritmo.go`) → reintentos (`reintentos.go`).
  - Grabación y reproducción: `nombre.go` (nombre derivado de método + dirección), `grabar.go` (escritura
    atómica), `reproducir.go`.
  - Errores tipados: `errores.go`, con la tabla cerrada de **dieciséis** situaciones → códigos `{1, 2, 4, 5}`.
- **Kernel de H1, dos ampliaciones mínimas**: `schema.ConClase` (el error declara su clase sin que el
  adaptador importe `internal/cli`) y `schema.Resultado.Ensayo` (canal de la descripción de `--dry-run`).
  `internal/cli` las reconoce; ningún comportamiento visible cambia.
- **`Makefile`**: una cuarta inyección de `-ldflags` lleva la misma `VERSION` del verbo `version` a la
  identificación, para que el `User-Agent` no pueda declarar otra (FR-007).
- **Controles**: `internal/arch_test.go` endurecido, `internal/httpx/superficie_test.go` nuevo,
  `cmd/kitlegal/main_test.go` y `internal/app/main_test.go` ampliados.
- **`.golangci.yml`**: **solo comentarios** y cuatro palabras españolas bajo `misspell.ignore-rules`
  (`controles`, `inventario`, `legislacion`, `resolucion`). Ninguna regla, ningún linter, ninguna exclusión,
  ningún `//nolint`.
- **Documentación**: `docs/ADR/0011-ensayo-por-capas.md` (nuevo) y `docs/PENDIENTES.md`.
- **Artefactos del hito**: `specs/003-h2-internal-httpx-cliente/` completo (spec, plan, research con 22
  decisiones, data-model, cuatro contratos, quickstart, tasks y `gates/`).
- **Herramienta del workflow, de paso**: `.specify/workflows/hito/workflow.yml` 1.6.1, `scripts/paso.sh` y
  `docs/WORKFLOW.md` — `resume` reejecuta el paso de nivel superior, así que `siguiente_tarea` detecta una
  tarea ya marcada y sin commitear y emite una iteración de **cierre** en vez de arrastrar su trabajo a la
  tarea siguiente. Se arregló al tropezar con ello a mitad de H2 (`43c703d`); no toca el producto.

**Fuera de alcance, y por qué no aplican cuatro puntos de la Definition of Done**: H2 no añade applet ni
salida de applet (punto 4: `schemas/` no se toca), no cambia ningún comportamiento visible (punto 6: sin
guion e2e nuevo ni `CHANGELOG.md`), no consulta ninguna fuente real (punto 8: la primera fila de
`docs/SOURCES.md` llega en H4) y no entrega skill (punto 10). Tampoco se abre `internal/source/`.

## Dependencias (constitución §V)

Dos entradas nuevas en `go.mod`, **ambas de la lista fijada** de la constitución §V y de `docs/ROADMAP.md`
§3, así que ninguna requiere *Complexity Tracking*:

| Dependencia | Versión | Para qué | Alternativa descartada |
|---|---|---|---|
| `github.com/temoto/robotstxt` | v1.1.2 | Interpretar `robots.txt` con `FromStatusAndBytes`, que es exactamente el reparto por estado de RFC 9309 §2.3.1 | Escribir el intérprete a mano: más código y más manera de equivocarse en la obligación que el principio I declara innegociable |
| `golang.org/x/time` | v0.16.0 | `rate.NewLimiter(rate.Every(intervalo), 1)`, un limitador por sitio | Un `time.Ticker` propio: sin *burst* ni espera cancelable por contexto |

No se añade ninguna dependencia fuera de la lista. El supuesto S1 de `research.md` —que `go mod tidy` no
arrastra nada más— quedó verificado: `go mod tidy -diff` está en `make ci` desde H0 y
`TestDependenciasDelBinario` sigue en verde.

## Controles añadidos

Lo que hasta aquí era disciplina pasa a ser mecánico:

- **La regla R2 deja de pasar en vacío.** `internal/arch_test.go` ahora **exige** que el grafo transitivo
  real contenga `internal/httpx` además de comprobar que ningún otro paquete importa `net/http`. Antes de
  este hito la regla estaba activa y sin dueño: era verde sobre cero ficheros.
- **`TestElBinarioNoEnlazaHTTPX`**: `go list -deps` comprueba que el binario distribuido no enlaza el
  paquete. Con su retirada ya anotada para H4 en `docs/PENDIENTES.md`.
- **`TestSuperficieExportada`**: recorre con `go/parser` todas las declaraciones exportadas y falla si
  alguna nombra `*http.Client`, `http.RoundTripper`, `*http.Request` o `*http.Response`. La imposibilidad
  de esquivar las garantías es por construcción, y este test lo vigila.
- **`TestClasesDeError`**: dieciséis subpruebas, una por fila de la tabla cerrada, comprobando la clase y
  el código de salida sobre el error tal como sale del paquete **y** envuelto, que el cliente nunca produce
  «no encontrado» (3) ni «requiere identidad humana» (6), y que ninguna termina en `panic`.
- **`TestAdaptadorDePruebaConElKernel`**: el adaptador de prueba ejecutado con `app.Main` en proceso —
  ensayo con `--dry-run` (cero accesos, código 0), consulta y plazo vencido (código 4).
- **SC-012, demostrado, no afirmado**: las cuatro variantes del escenario 10 del quickstart (respuesta sin
  cerrar → `bodyclose`; petición sin contexto → `noctx`; `Pedir` sin contexto → no compila; importar la
  biblioteca HTTP en el kernel → `depguard` **y** R2) se ejecutaron sobre una copia desechable fuera del
  repositorio y cada una hace fallar `make ci` nombrando la regla. Salida literal en
  `specs/003-h2-internal-httpx-cliente/gates/evidencia-sc012.md`.

## Evidencia

- `specs/003-h2-internal-httpx-cliente/gates/evidencia-sc012.md` — los
  controles fallan ante cada intento de saltárselos (T017).
- `specs/003-h2-internal-httpx-cliente/gates/evidencia-cierre.md` —
  validación agregada (T018): escenarios 1, 7, 11 y 12 del quickstart, umbrales, sin supresiones nuevas,
  «solo direcciones locales» con once sondas que la orden delata, y el arreglo de un test inestable (3 de
  cada 1000) con el rojo reproducido antes y 0/1000 después.

| Umbral | Exigido | Medido |
|---|---|---|
| Global | ≥ 70 % | **93,3 %** |
| `internal/core/**` | ≥ 85 % | **90,1 %** |
| `internal/cli` | ≥ 90 % | **98,6 %** |
| `internal/httpx` | — | 94,8 % |

Ninguno se rebajó: `codecov.yml` no aparece en el diff frente a `main`.

**Aviso de método, por si se repite la medida**: el envoltorio de terminal usado en esta máquina reescribe
la salida de `go test`, de `git status --porcelain` (devuelve la cadena `ok`) y de `git diff` (sangra las
líneas añadidas). Una medida diferencial hecha sobre esas salidas **sale verde siempre**. Todo lo de arriba
está medido con el paso directo, y las evidencias incluyen sondas positivas que lo acreditan.

## Decisiones

Las 22 están razonadas en `specs/003-h2-internal-httpx-cliente/research.md`; las que cambian algo fuera del
paquete:

- **Superficie cerrada por construcción (D1, D2).** Un tipo, dos constructores (`New`, `Replay`), una
  operación con contexto obligatorio y cinco opciones. La respuesta es un tipo propio con el cuerpo ya
  leído: nunca un `*http.Response`, de modo que no queda cierre pendiente en manos del adaptador.
  Alternativa rechazada: exponer la cadena de `RoundTripper`, que es exactamente la salida por la que se
  esquivarían las garantías.
- **El error declara su clase, el kernel la reconoce (D4).** `httpx.Error` implementa `schema.ConClase` y
  `cli.Clasificar` lo lee con `errors.As`. Así el adaptador no importa `internal/cli` y la regla R1 (dominio
  puro) sigue intacta. Alternativa rechazada: un mapa de traducción en el kernel, que se desincroniza.
- **El ensayo lo describe cada capa y lo presenta el kernel** — **ADR 0011**, nuevo, porque cambia el
  contrato del applet que fijó ADR 0005. En H1 ninguna capa tenía efectos y la descripción de `--dry-run`
  era única; con red ya no lo es.
- **`robots.txt` según RFC 9309 §2.3.1 (D7).** 4xx y cuerpo vacío → «sin reglas» y se emite; 5xx que
  sobrevive a los reintentos → «todo desautorizado», código 5; 2xx ininterpretable → deniega igual; 429 →
  nunca es «sin reglas»; contexto vencido → código 4 sin llegar a evaluar regla. Caché por sitio.
- **Reintentos con *equal jitter* y `crypto/rand` (D9).** Tres intentos, espera interrumpible por el
  contexto, `Retry-After` en el error y **nunca** esperado por el cliente (D20).
- **El adaptador de prueba es material de test (D16).** Vive en `package httpx_test`, no se crea en el
  árbol de producción ni se registra como applet en ningún binario: el hito no abre `internal/source/`.
- **Fixtures del paquete junto al paquete (D17)**, escritos a mano contra hosts ficticios
  (`fuente.prueba`, `otra.prueba`). Lista cerrada de diez ficheros, cada uno con su fila en
  `TestNombreDeGrabacion`. **Ninguna grabación contra una fuente real en todo el hito** (FR-044).
- **La contingencia de `misspell` se resolvió sin atajo.** `legislacion` y `resolucion` no son
  identificadores sino nombres de fichero que el contrato de grabación fija; como el término no podía
  reescribirse, se anotó, se detuvo el trabajo y se redelimitó la tarea para declarar `.golangci.yml`, en
  vez de un `//nolint` o de retocar el contrato para contentar a un diccionario inglés.

## Pendientes

- **Para H4, ya anotado en `docs/PENDIENTES.md`:** retirar `TestElBinarioNoEnlazaHTTPX` al enlazar el
  paquete, con la ampliación **justificada** de `modulosDelBinario` (`golang.org/x/time`,
  `github.com/temoto/robotstxt`); y el ritmo por fuente (`ConIntervalo`), que saldrá de `docs/SOURCES.md`,
  la tabla que crea H4 y que H2 no toca.
- **Dónde viven los fixtures grabados de fuentes reales** (`testdata/<fuente>/` en la raíz o
  `internal/source/<fuente>/testdata/`) sigue abierto: el pendiente pasa de «Antes de H2» a **«Antes de la
  primera tarea `[datos]` de H4»**, con la recomendación intacta (junto al paquete). El cliente es
  agnóstico (FR-064), así que la elección no cambia una línea de este hito.
- **Codecov (S3)**: H2 no declara componente propio; rigen los umbrales generales. Si un estado no
  apareciera o saliera verde sobre cero ficheros, la causa sería la configuración, no el umbral.
- **Revisión pendiente del ritual** (§6 del roadmap, punto 4): `/code-review` y `/security-review` sobre
  esta propuesta. No toca `docs/SOURCES.md` porque no hay fuente.
- **La fusión es humana.** Squash-merge; el gancho `pre-push` rechaza `main`, los push forzados, los
  borrados y las etiquetas (ADR 0007).

🤖 Generated with [Claude Code](https://claude.com/claude-code)
