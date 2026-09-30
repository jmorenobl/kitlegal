# Contrato: el sondeo local (`make evals-sondeo`)

FR-060 a FR-068, FR-096; SC-002, SC-009. Decisiones en [research.md](../research.md) D9, D10, D16, D17 y D18; tipos en
[data-model.md](../data-model.md) §8. No es un veredicto: el veredicto sigue siendo del job (ADR 0016). Ninguna tarea ni
ningún paso del workflow lo ejecuta (FR-068).

## 1. La orden

```bash
make evals-sondeo SKILL=<skill> EVALS=<nn>[,<nn>…] MODELO=<id> REPETICIONES=<n> [CONCURRENCIA=<n>]
```

En el `Makefile`, con su línea de `make help`:

```make
## evals-sondeo: sondeo local de unas evals de una skill con Claude Code, sin strace ni veredicto (macOS o Linux; consume la suscripción; CLAUDE_CODE_OAUTH_TOKEN)
evals-sondeo: check-tools
	@scripts/evals-sondeo.sh "$(SKILL)" "$(EVALS)" "$(MODELO)" "$(REPETICIONES)" "$(CONCURRENCIA)"
```

`SKILL`, `EVALS`, `MODELO` y `REPETICIONES` no tienen valor por defecto; `CONCURRENCIA`, vacía, es la del job para esa
skill (4 en `boe-legislacion`, 1 en `legal-core`), leída de `.github/workflows/evals.yml` (research D16).

## 2. `scripts/evals-sondeo.sh`

Guion de bash sin nada posterior a bash 3.2. Con cinco argumentos (si no, uso y código 1):

1. crea el directorio temporal con `mktemp -d "${TMPDIR:-/tmp}/kitlegal-sondeo.XXXXXX"` y lo borra con `trap … EXIT`,
   también cuando termina con otro código (FR-064), y dentro `tmp/`;
2. ejecuta, con `TMPDIR=<directorio>/tmp`, `go test -tags evals -count=1 -timeout 0 -run '^TestSondeo$' ./internal/evals/
   -args -skill … -evals … -modelo … -repeticiones … -concurrencia … -temporal <directorio>`, con su salida estándar y
   de error en `<directorio>/go-test.log`: lo que escriben en su `TMPDIR` el directorio de trabajo de `go test` y el de
   `go install`, y la preparación de cada sesión, que `TestSondeo` hace en proceso (copias de las grabaciones con
   `os.MkdirTemp`), queda dentro del directorio y lo borra el `trap`, también si el proceso muere sin retirarlo; fuera
   quedan solo las cachés de Go;
3. si termina con 0, imprime `<directorio>/salida.txt` y sale con 0; si no, imprime el registro por la salida de error
   y sale con 1.

## 3. `TestSondeo` (etiqueta `evals`)

En este orden, y sin abrir ninguna sesión hasta el paso 4:

1. **Argumentos** (FR-067), cada error nombrando el argumento de `make`:

   | Argumento | Error |
   |---|---|
   | `SKILL` | `SKILL: «<valor>» no es ninguna skill con evals` (forma de `name` y carpeta `evals/<skill>/` legible sin ficheros mal formados); `SKILL: el job de evals no ejecuta «<valor>»` (no está en la matriz) |
   | `EVALS` | `EVALS: «<valor>» no es una lista de números de eval de dos cifras separados por comas` (también con un número repetido); `EVALS: <nn> no es ninguna eval de <skill>`, una línea por número |
   | `MODELO` | `MODELO: está vacío`; `MODELO: «<valor>» no tiene la forma de un id de modelo` |
   | `REPETICIONES` | `REPETICIONES: «<valor>» no es un entero mayor o igual que 1` |
   | `CONCURRENCIA` | `CONCURRENCIA: «<valor>» no es un entero mayor o igual que 1` (vacía no es error) |

   Un entero mayor o igual que 1 tiene la forma `^[1-9][0-9]*$`, como `CONCURRENCIA_DE_EVALS` en `scripts/evals.sh`.
   Los errores de varios argumentos salen juntos, uno por línea y en el orden de la tabla; con un `SKILL` que no vale,
   de `EVALS` solo se comprueba la forma.

2. **Credencial** (FR-063): si `CLAUDE_CODE_OAUTH_TOKEN` no está en el entorno o está vacía,
   `falta la credencial: CLAUDE_CODE_OAUTH_TOKEN, el token de la suscripción que da claude setup-token, no está en el
   entorno o está vacía`. Sin llamar al servicio y sin ninguna otra comprobación de la credencial.
3. **El árbol de trabajo** (FR-062): `go install -trimpath ./cmd/kitlegal` en la raíz del repositorio, con el entorno
   de quien lo lanza, `GOBIN=<temporal>/bin` y `CGO_ENABLED=0`; y `<temporal>/bin/kitlegal skills install -g --host
   claude` con el entorno de §5, que lleva `HOME=<temporal>/home`. Las skills quedan en
   `<temporal>/home/.claude/skills/`, como las deja `make install`.
4. **Las sesiones**: el plan de `PlanDeEvals` con las evals pedidas en su orden, `MODELO` como modelo que decide, sin
   modelos informativos, `REPETICIONES` y sin prueba de red; el repartidor ([ejecucion-del-job.md](./ejecucion-del-job.md)
   §3) con esa concurrencia, `Traza: false`, `Skills` `<temporal>/home/.claude/skills`, sesiones en
   `<temporal>/sesiones` y el entorno de §5. La interrupción es el `Done()` del contexto que cancelan `SIGINT` y `SIGTERM`.
5. **El juicio** (FR-061; research D17): cada sesión con `LeerSesion`, `Juzgar` sobre la eval sin `comandos` ni
   `prohibidos` y la sesión sin invocaciones, `exigirElModeloPedido` y la clasificación de límites; series y recuento
   con `repartirEnSeries` y `recontarExpresiones`, como el informe.
6. **La salida** (§4) en `<temporal>/salida.txt`. Ningún `informe.json` ni veredicto (FR-066).

## 4. La salida

Texto, sin nada por sesión de las expresiones (FR-065). Ejemplo con las evals del quickstart (≈ 1 KB):

```text
Esto es un sondeo, no un veredicto: el veredicto de la skill lo da el job de evals.
No comprueba lo que el job lee de la traza de strace —qué órdenes se ejecutaron (los comandos esperados y los prohibidos de cada eval) y las llegadas a la red— ni la ausencia de Python: por eso sus tasas no son las del job.

Tasa de cada serie (sesiones que pasan de las medidas):
- 03-lrbrl-atribuciones-del-pleno.yaml con claude-sonnet-5: 3 de 3
- 06-irpf-rendimientos-del-trabajo.yaml con claude-sonnet-5: 3 de 3
- 13-lrbrl-atribuciones-por-materia.yaml con claude-sonnet-5 (informativa): 3 de 3
- 14-trlrhl-impuestos-por-materia.yaml con claude-sonnet-5 (informativa): 2 de 3
- 15-irpf-rendimientos-por-materia.yaml con claude-sonnet-5 (informativa): sin medir

Respuestas con alguna expresión prohibida en las evals que activan la skill: 1 de 13 (7,7 %); referencia: como mucho el 5 %.

Sesiones sin medir por límite de uso:
- 15-irpf-rendimientos-por-materia-claude-sonnet-5-02: mensaje del límite de uso: You've hit your session limit · resets 5pm
- 15-irpf-rendimientos-por-materia-claude-sonnet-5-03: sin abrir tras el límite de uso

Sesiones que no terminaron por otra causa:
- 14-trlrhl-impuestos-por-materia-claude-sonnet-5-01: la sesión no terminó: código 1: result con is_error: Failed to authenticate. API Error: 401 OAuth access token is invalid.
```

Reglas: la primera línea es siempre la del ejemplo; la segunda, siempre esa. Una serie con alguna sesión sin medir dice
`sin medir`; las demás, `<pasan> de <sesiones>`, con las que no terminaron por otra causa como no pasadas, igual que el
job. El motivo de una sesión que no terminó es el del job (`la sesión no terminó: ` y `MotivoSinTerminar`, data-model
§2): con un `result` con `is_error` al final, lleva su texto también cuando `claude -p` sale con un código distinto de 0,
como con la credencial del ejemplo (research V18). El recuento es el de `recontarExpresiones` (sin las sesiones sin medir), con el porcentaje con un decimal y coma;
una skill sin lista dice `Respuestas con alguna expresión prohibida: la skill no tiene lista de expresiones
prohibidas.`. Sin sesiones en un apartado, `ninguna.` en su línea (`Sesiones sin medir por límite de uso: ninguna.`).
Tamaño: 2 líneas fijas, una por serie (≤ 19 evals; un modelo por sondeo), el recuento y, a lo sumo, una por sesión
pedida: con 5 evals, 3 repeticiones y todas terminadas, ≈ 0,9 KB; con las 15 listadas, < 3 KB.

## 5. El entorno de las sesiones del sondeo (FR-062, FR-063)

Base: de quien lo lanza, solo `PATH`, `LANG`, `LC_ALL`, `LC_CTYPE`, `LC_MESSAGES`, `TERM`, `USER`, `LOGNAME`, `SHELL`, `TZ`
y `CLAUDE_CODE_OAUTH_TOKEN`; `PATH` con `<temporal>/bin` delante y `HOME` en `<temporal>/home`. Encima, lo de cada sesión
([ejecucion-del-job.md](./ejecucion-del-job.md) §4). Ni `ANTHROPIC_API_KEY`, ni `ANTHROPIC_AUTH_TOKEN`, ni ninguna otra
variable de quien lo lanza. Con su `CLAUDE_CONFIG_DIR`, cada sesión no carga el `CLAUDE.md`, los hooks, los plugins ni
las skills de la persona, y el servicio del llavero que consulta Claude Code no es el suyo (research V5).

## 6. Códigos de salida

`make evals-sondeo` sale con 0 si ha podido abrir y juzgar las sesiones, sean cuales sean las tasas, el recuento o las
sesiones sin medir (FR-066); con 1 si un argumento no vale, falta la credencial, no se pudo construir el binario o
instalar las skills, el directorio de las sesiones no se pudo crear o listar, o el repartidor devolvió un error; y con el código de `make` (2) detrás.

## 7. Tests (en `make ci`; FR-096)

| Test | Qué comprueba | Requisito |
|---|---|---|
| `TestJuicioDelSondeo` | sobre sesiones sintéticas con transcript y traza, el juicio del sondeo es, eval a eval, el del job salvo `comandos_ejecutados`, `comandos_ausentes`, `comandos_prohibidos_ejecutados`, `invocaciones`, `fuera_de_lo_grabado`, `otras_fallidas`, `llegadas_a_la_red` y los motivos y el `pasa` que dependen de ellos | FR-061; US4-1 |
| `TestSalidaDelSondeo` | la primera línea; la segunda nombra los comandos esperados y los prohibidos, las llegadas a la red y Python; ninguna expresión de ninguna sesión aparece; series, recuento con el 5 %, sin medir y no terminadas como §4, con la sesión del ejemplo leída con `LeerSesion` de un transcript que acaba en el `result` con `is_error` de research V18 y código 1: su línea lleva `código 1: result con is_error: Failed to authenticate. …`; la skill sin lista | FR-063, FR-065; US4-2 |
| `TestComprobarElSondeo` (tabla) | cada error de §3.1, nombrando su argumento; la concurrencia por omisión, 4 y 1, leída de la definición real; sin `CLAUDE_CODE_OAUTH_TOKEN` o vacía, el error de §3.2 | FR-060, FR-063, FR-067; US4-5 |
| `TestSondear` (tabla, con sustitutos y sin construir el binario) | todas las sesiones fallan: devuelve la salida sin error (el guion sale con 0); la sesión k da (a): no abre más, las que faltan salen sin medir y devuelve sin error; un argumento o la credencial que no valen: ningún sustituto se invoca; las sesiones no ven `ANTHROPIC_API_KEY` ni `ANTHROPIC_AUTH_TOKEN` aunque estén en la base, sí `CLAUDE_CODE_OAUTH_TOKEN`, `HOME` es el del temporal y el primer `kitlegal` del `PATH` es el de `<temporal>/bin`; el `HOME` y el `TMPDIR` de la base no cambian: las sesiones no escriben en el entorno de quien lo lanza. El `TMPDIR` del proceso, donde escribe la preparación de cada sesión, no lo mira: lo fija `TestGuionDelSondeo` | FR-061 a FR-066; SC-009; US4-3, US4-4, US4-6 |
| `TestGuionDelSondeo` | `scripts/evals-sondeo.sh` con un `go` sustituto en el `PATH` y `TMPDIR` temporal: el sustituto se ejecuta con `TMPDIR` en `<directorio>/tmp` y deja en él un directorio sin borrar, como un `go test` que muere sin retirar su preparación; si escribe `salida.txt` y sale con 0, el guion sale con 0, su salida estándar es esa y `TMPDIR` queda vacío; si sale con 1, el guion sale con 1, su salida de error lleva el registro y `TMPDIR` queda vacío | FR-064, FR-066; SC-009; US4-4 |
| `TestPrepararElArbolDelSondeo` (etiqueta `integration`) | el `go install` y el `skills install` reales: `<temporal>/bin/kitlegal` existe y `<temporal>/home/.claude/skills/boe-legislacion/SKILL.md` es la del árbol; el `HOME` de la base no gana nada fuera de la telemetría del go command (`go env GOTELEMETRYDIR`), que no se apaga desde el entorno y que el `go test` del guion escribe igual | FR-062, FR-064 |

## 8. Uso, de fuera adentro

- **Quién y cuántas veces**: quien ajusta `SKILL.md`, antes de pedir el job, una vez por cambio que quiere medir.
- **Consumo**: las sesiones del plan pedido, `EVALS × REPETICIONES`, de la suscripción de quien lo lanza; nada más: ni
  llamadas previas al servicio ni reintentos de sesiones.
- **Salida**: §4; no deja informe, ni historial, ni ficheros: cada sondeo empieza y termina en su temporal.
