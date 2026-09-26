# Research: H19 · Instalar sin clonar

**Feature**: `009-h19-instalar-sin-clonar` · **Fecha**: 2026-09-26 · **Plan**: [plan.md](./plan.md)

Modo desatendido. Cada decisión técnica se tomó con el «Criterio de decisión autónoma» de la constitución: la mejor
solución sin atajos, lo no especificado fuera, y la lectura conservadora cuando la duda toca alcance, frontera
humana, privacidad o una decisión cerrada. Este documento tiene tres partes:

- **Decisiones** (D1-D34): qué se elige, por qué y qué se rechaza.
- **Verificaciones** (V1-V48): toda afirmación sobre una herramienta o dependencia externa, con dónde se comprobó en
  local (`fichero:línea` del módulo en la caché, `go doc`, o la orden ejecutada). Nada de lo que figura aquí se
  comprobó con red.
- **Supuestos no verificados** (S1-S15): lo que no se puede comprobar sin red, sin la plataforma o sin una persona.
  El plan **no los afirma como hechos**; dice qué los comprueba y cuándo.

Las rutas `goreleaser:` son relativas a `/Users/jorge/go/pkg/mod/github.com/goreleaser/goreleaser/v2@v2.18.1/`; las
`testscript:`, a `/Users/jorge/go/pkg/mod/github.com/rogpeppe/go-internal@v1.16.0/testscript/`; las `kong:`, a
`/Users/jorge/go/pkg/mod/github.com/alecthomas/kong@v1.16.1/`; las `jsonschema:`, a
`/Users/jorge/go/pkg/mod/github.com/invopop/jsonschema@v0.14.0/`.

---

## Decisiones

### D1 · Dominio puro en `internal/core/instalacion`, con el disco detrás de un puerto

**Decisión.** La lógica del applet —qué es de quién en el disco (conflictos de FR-041), qué se escribe y en qué orden,
el manifiesto, los hallazgos de `doctor` con su orden, la igualdad de versiones y el aviso— vive en un paquete de
dominio nuevo, `internal/core/instalacion`, que **no hace entrada ni salida**: ve el disco a través de un puerto
(`Disco`, solo lectura y sin seguir enlaces) que define él mismo y recibe las skills empotradas ya leídas como bytes.
La escritura la ejecuta el adaptador siguiendo el plan que devuelve el dominio.

**Por qué.** FR-041, FR-046, FR-065 y FR-066 son tablas de decisión con decenas de casos, y FR-044 y FR-066 (i)-(ii)
son propiedades («tras cualquier fallo, una nueva ejecución completa sin conflictos»; «ejecutar las órdenes de
`doctor` en su orden deja `doctor` limpio») que solo se pueden probar de forma exhaustiva sobre un disco en memoria.
Es la arquitectura hexagonal de la constitución (§IV): el dominio define el puerto y el adaptador lo implementa, y
`internal/core/**` gana además el umbral de cobertura del 85 % sobre la parte más delicada del hito. Los tests del
dominio son sintéticos por construcción porque R1 se amplía en este hito (D32): hoy ni la lista `core` de `depguard` ni
`paquetesInternos` de `TestArquitectura` nombran `internal/disco` ni el paquete raíz, y los `_test.go` solo los vigila
`depguard` (el grafo de `TestArquitectura` sale de `go list -deps ./...`, sin tests: V47), así que sin esa ampliación un
`package instalacion_test` que importara el adaptador o lo empotrado pasaría todos los controles.

**Alternativas rechazadas.** (a) Todo en `internal/app/skills.go` con `os` directamente: cada caso de FR-041 exigiría
un árbol real en `t.TempDir()`, las propiedades de FR-044 y FR-066 no se podrían recorrer de forma sistemática y el
applet mezclaría composición con lógica. (b) Un solo paquete adaptador `internal/instalacion` que haga I/O y lógica:
misma pega, y dejaría la lógica fuera del umbral de `internal/core/**`. (c) Reutilizar `internal/skills`: es la
herramienta de desarrollo que el binario **no puede enlazar** (arrastra `yaml.v3` y `jsonschema` de validación;
`TestDependenciasDelBinario`, V35).

### D2 · Adaptador de disco nuevo, `internal/disco`

**Decisión.** Un paquete adaptador `internal/disco` implementa el puerto de lectura (`Disco`), el ejecutor del plan
(`Escritor`) y el creador de enlaces del sistema (`Enlazador`) sobre el sistema de ficheros real.

**Por qué.** Ningún adaptador de `docs/ROADMAP.md` §2 cubre el disco local (`source`, `httpx`, `cache`, `store`,
`graph`, `render`), y la escritura atómica, la lectura sin seguir enlaces y el recurso de copia necesitan tests
propios contra un árbol real (enlaces, tuberías con nombre, permisos). Va a *Complexity Tracking* por ser un paquete
fuera de la lista de §2.

**Alternativas rechazadas.** (a) Meterlo en `internal/app`: la raíz de composición pasaría a hacer I/O de ficheros y
sus tests se mezclarían con los del kernel. (b) `internal/store`: es el repositorio del grafo del asunto sobre SQLite
(H10), no un sistema de ficheros.

### D3 · El empotrado: `skills.go` en la raíz del módulo, paquete `kitlegal`

**Decisión.** Un único fichero Go en la raíz, `skills.go` (paquete `kitlegal`, ruta de importación
`github.com/jmorenobl/kitlegal`), con `//go:embed skills/*/SKILL.md skills/*/references/*` y una función `Skills()
fs.FS` que devuelve lo empotrado tal cual. La lectura a `[]instalacion.SkillEmpotrada` (nombre, ficheros y huellas)
la hace `internal/app`, que es quien puede importar `io/fs` (el dominio no: V26). Una skill es un directorio de
`skills/` con `SKILL.md` en lo empotrado (FR-004); un directorio con `references/` y sin `SKILL.md` no es skill.

**Por qué.** Es lo que fija el hito literalmente (FR-002) y lo único posible: un patrón de `//go:embed` no puede subir
de directorio (V20). Función y no variable exportada, para que nadie pueda sustituir lo empotrado desde fuera, que es
el estilo del kernel (`cli.ProcedenciaKernel`). El patrón del hito embebe también ficheros con punto en `references/`
(V20); no hace falta `all:` ni un filtro, porque `skills-check` ya rechaza en `references/` toda entrada que no sea una
referencia declarada (`DerivaFicheroSobrante`, V38), y `TestSkillsEmpotradas` compara lo empotrado con el árbol.

**Alternativas rechazadas.** (a) `all:skills`: embebería `scripts/` y cualquier fichero de la skill (FR-001 dice «y
nada más»). (b) `skills/*/references/*.md`: más estrecho que el literal del hito («todos los ficheros de su
`references/`»).

### D4 · La versión del binario llega al registro al construirlo

**Decisión.** `app.RegistroDeProduccion(version string)` y `app.Arrancar(argv, construir func(version string)
(*Registro, error), …)`: `Arrancar` pasa a `construir` la versión que ya recibe de `-ldflags`, y el registro compone
el applet `skills` y el aviso con ella. Lo mismo el registro del binario de e2e. Los llamadores que no tienen versión
(`internal/evals/trazas.go`, tests) pasan la cadena vacía, que no tiene forma SemVer y no compara nunca (FR-073).

**Por qué.** El manifiesto registra la versión del binario (FR-031) y el aviso la compara (FR-071), y la versión solo
existe como `main.version` inyectada; FR-091 prohíbe una quinta inyección. Pasarla al construir es explícito, no usa
estado global y no toca `internal/cli` (spec, *Fuera de alcance*).

**Alternativas rechazadas.** (a) Una quinta `-X` para el paquete del applet: la prohíbe FR-091. (b) Llevarla en
`schema.Contexto`: cambiaría el kernel de `internal/cli` para todos los applets. (c) Leer `runtime/debug.BuildInfo`:
no lleva la versión inyectada con `-X` y daría otra versión que `kitlegal version`.

### D5 · El aviso lo emite la composición, tras el análisis del verbo y nunca en la ayuda

**Decisión.** El registro guarda un avisador opcional (`Registro.Avisar(func() (string, bool))`), que componen
`RegistroDeProduccion` y el registro de e2e con el disco real, `HOME` y las skills empotradas. `internal/app/main.go`
lo llama **una vez** en `resolverApplet`, justo después de `cli.Analizar`, cuando el applet no es `skills` y la
decisión no es la ayuda —también si el análisis falla, que es el exit 2 de «`kitlegal boe articulo` sin norma»—, y
escribe la línea con `Presentador.Aviso`. `version`, la ayuda del binario y los fallos anteriores al despacho no llegan
nunca ahí (FR-070).

**Por qué.** Es exactamente el conjunto de FR-070: el despacho ya separa `version` (`DestinoReservado`), la ayuda sin
applet y el applet desconocido; y `Analizar` distingue la ayuda de un error aunque Kong devuelva error tras escribirla
(V23). Registrar el avisador en la composición deja `Main` comprobable con un avisador sintético y sin disco.

**El error de escritura del aviso no se propaga**, y es una decisión, no un silencio: FR-072 prohíbe que el aviso
cambie el código de salida o la salida estándar. Si la salida de error está rota, el fallo que sí cuenta —el mensaje de
un error, si lo hay— lo encuentra el kernel en su propia escritura. Un test lo fija (presentador cuyo `Aviso` falla:
mismo código, misma salida estándar).

**Alternativas rechazadas.** (a) Emitirlo desde cada applet: repetido y olvidable. (b) Emitirlo en `Main` tras la
invocación: no sabría si fue ayuda sin ampliar el desenlace del kernel. (c) Propagar su error: viola FR-072.

### D6 · Lo que se instala no sigue enlaces: `Lstat`, y abrir solo ficheros regulares comprobando que son el mismo

**Decisión.** El adaptador examina toda ruta por debajo de la raíz del ámbito con `os.Lstat` y `os.Readlink`; solo
comprueba si el destino de un enlace resuelve con `os.Stat` sobre el enlace, sin leer nada (la única excepción de
FR-028). Para leer un fichero (huella o manifiesto): `Lstat` → exige fichero regular → `os.Open` → `f.Stat` → exige
fichero regular y `os.SameFile` con lo examinado; si no, es «no es un fichero regular». Nada que no sea regular según
`Lstat` se abre. Lo que hay por encima de la raíz del ámbito o de la ruta de `--dir` se usa tal cual (FR-027).

**Por qué.** Cumple FR-028 en todas las plataformas con la biblioteca estándar y cierra la carrera de sustituir el
fichero por un enlace entre el examen y la apertura (`SameFile` compara lo abierto con lo examinado). La carrera
residual —un fichero regular cambiado por una tubería con nombre entre `Lstat` y `Open`— exige escritura en el
directorio del proyecto por otro proceso a la vez y queda anotada como límite conocido.

**Alternativas rechazadas.** (a) `os.Root` (Go 1.24+): confina a una raíz pero **sigue** los enlaces que quedan
dentro de ella, justo lo que FR-028 prohíbe. (b) `O_NOFOLLOW|O_NONBLOCK`: solo existe en Unix y obligaría a una
variante de Windows sin CI que la pruebe (el hito no tiene CI en Windows).

### D7 · Orden de aplicación que hace cierto FR-044

**Decisión.** El dominio devuelve el plan ya ordenado en cuatro fases y lo aplica en ese orden a través del puerto
`Escritor` (el adaptador solo implementa las operaciones sueltas), parándose en el primer fallo, que sale con exit 1
nombrándolo:

1. **Retirar** lo declarado que va a cambiar o desaparecer: los ficheros con huella intacta que se van a reescribir o
   que ya no se empotran, y las copias de host que pasan a enlace (con sus directorios, ya vacíos).
2. **Enlazar** las entradas de host que tocan (nuevas, que faltan o que sustituyen a una copia), creando antes
   `<Raiz>/.claude` y `<Raiz>/.claude/skills` si faltan (FR-023); si el creador de enlaces falla, esa entrada pasa a
   copia (FR-024).
3. **Escribir el manifiesto final**, de una vez y de forma atómica, con los modos que resultaron de la fase 2.
4. **Escribir** los ficheros (del directorio neutro y de las copias), creando los directorios que falten.

**Por qué.** Tras un fallo en cualquier punto, el estado del disco es uno que `install` completa sin conflictos: en
1 y 4, un fichero declarado que falta se repone (FR-046); en 2, un enlace creado antes de declararlo tiene el destino
literal de FR-021 y se adopta (FR-041), y un `.claude/skills` recién creado y vacío es un directorio real que no es de
ninguna skill ni conflicto; en 3, o no cambió nada del manifiesto o ya declara el estado final, y lo que
falta se repone. Lo que nunca ocurre es un fichero **con contenido nuevo y huella vieja** en el manifiesto, o **no
declarado en una ruta que `install` escribiría**, que serían «fichero editado» o «fichero ajeno». El test
`TestFalloAMitadSeCompleta` del dominio lo recorre: para cada operación del plan, falla en ella, vuelve a planificar
sobre el estado resultante y exige cero conflictos y el estado final.

**Alternativas rechazadas.** (a) Escribir ficheros y al final el manifiesto: un fallo del manifiesto deja ficheros
nuevos no declarados → conflicto. (b) Manifiesto primero y reescribir encima: un fallo deja huella nueva declarada y
contenido viejo → conflicto. (c) Construir el directorio nuevo aparte y renombrarlo: se llevaría por delante los
ficheros no declarados que FR-047 manda conservar. (d) Tolerar como no conflicto un fichero cuyo contenido ya es el
empotrado: contradice el literal de FR-041 (e) y (f).

### D8 · Escritura atómica de cada fichero y permisos

**Decisión.** Cada fichero (también el manifiesto) se escribe en un temporal del mismo directorio (`os.CreateTemp`) y
se renombra encima; el temporal se retira si algo falla. Permisos: `0o600` para ficheros, `0o750` para directorios,
los mismos que ya usa `internal/skills` (V38) y los que acepta `gosec` (G306, G301). Ningún bit de ejecución
(FR-015).

**Por qué.** Un corte nunca deja un fichero a medias con el nombre final. Los permisos son los del repositorio; los que
queden en un clon del proyecto los fija git, que solo versiona el bit de ejecución.

### D9 · El creador de enlaces: `Disponible` sondea en el sistema de ficheros del host, y recurso en la aplicación

**Decisión.** El puerto `Enlazador` tiene dos operaciones: `Disponible(directorio string) (bool, error)` y
`Enlazar(destino, ruta string) error`. La pregunta es siempre **por un directorio del propio ámbito**, el del sistema de
ficheros donde vivirá la entrada de host: el dominio pasa el directorio real existente más próximo a
`<Raiz>/.claude/skills`, en este orden y sin subir nunca por encima de la raíz: `<Raiz>/.claude/skills`, `<Raiz>/.claude`
o `<Raiz>`. Los dos primeros ya están examinados sin seguir enlaces (si uno existe y no es un directorio real, es el
conflicto de FR-023 o FR-026 y no se pregunta nada); la raíz se usa tal cual (FR-027), también alcanzada por un enlace, y
en local es el directorio de trabajo (data-model §3). Una entrada nueva vive en el sistema de ficheros de su directorio padre,
y los directorios que falten se crean dentro del más próximo que existe: la respuesta es la de la ruta donde `install`
va a enlazar.

El `Enlazador` del sistema responde con una **sonda en ese directorio**: crea con `os.Symlink` un enlace de nombre
`.kitlegal-sonda-<16 hexadecimales aleatorios>` y destino literal `kitlegal-sonda` (no resuelve y no se sigue), y lo
retira con `os.Remove`. Si crearlo falla, `false` (con `fs.ErrExist` prueba otro nombre, hasta 8 veces, antes de darse
por vencido con error); si retirarlo falla, devuelve el error y la orden sale con 1 nombrando la ruta de la sonda, porque
el disco ya no quedaría como estaba. Recuerda la respuesta por directorio durante la invocación, y el dominio solo
pregunta cuando tiene que decidir entre enlace y copia: una entrada de host que hay que crear, o una declarada `copia`
que sigue siendo un directorio real (`install` y la clase 4 de `doctor`). Nunca se pregunta en el caso idempotente (un
enlace declarado que sigue siendo el de FR-021), ni en `list`, ni en el aviso, ni con `--dir`. Al terminar la sonda el
directorio tiene exactamente las mismas entradas y los mismos bytes: `--dry-run` y `doctor` siguen sin dejar ningún
cambio en disco (FR-048, FR-068), y la sonda no crea directorios.

La predicción del plan sale de `Disponible` —es lo que da `--dry-run` (FR-048) y lo que usa `doctor` para «copia»
(FR-069)—, y en la fase 2 de D7, si `Enlazar` falla igualmente, la entrada pasa a copia (FR-024 literal). El applet
no compone el `Enlazador` dentro: lo recibe en `app.DependenciasDeSkills` (con la versión y lo empotrado), como `boe`
recibe su cliente en `app.DependenciasDeBoe`, y `app.DependenciasDeSkillsDelSistema(version)` pone el del sistema de
`internal/disco`; así se sustituye en `TestAppletSkills` y en e2e, un binario compuesto con un enlazador que siempre
falla (D24).

**Raíz que no existe.** Solo ocurre con `-g` y un `HOME` que no existe. Dentro del ámbito no hay ningún directorio donde
sondear, y lo que hay por encima de la raíz no se toca (FR-027): el dominio no pregunta y predice `enlace`; si ese
sistema de ficheros no admite enlaces, la fase 2 cae a copia (FR-024) y solo el `--dry-run` de esa primera instalación
habría dicho `enlace`. No afecta a FR-045 ni a `doctor`: la segunda ejecución encuentra la raíz y sondea en ella, y sin
manifiesto `doctor` no tiene hallazgos. Es la lectura conservadora, anotada en `gates/supuestos.md`.

**Por qué.** Sin predicción no hay `--dry-run` ni hallazgo «copia»; sin el recurso en la aplicación, un sistema de
ficheros sin enlaces (exFAT, FAT, un recurso de red sin extensiones Unix) dejaría la orden a medias. Y la predicción solo
es cierta si se hace en el mismo sistema de ficheros que la aplicación: con el ámbito en uno sin enlaces y la sonda en
otro que sí los admite, cada `install` retiraría la copia, fallaría al enlazar y la volvería a crear (`actualizada`
siempre, contra FR-045), `--dry-run` diría `enlace` y la orden dejaría `copia` (FR-048), y `doctor` daría un hallazgo
«copia» cuya orden no lo arregla (FR-066 (i), FR-069). La misma sonda en las tres plataformas se prueba en Unix.

**Alternativas rechazadas.** (a) **Sondear en un directorio temporal propio** (`os.MkdirTemp`, en `TMPDIR`), que fue la
primera versión de esta decisión: responde por el sistema de ficheros de `TMPDIR`, no por el del ámbito, y produce
exactamente los tres fallos anteriores cuando están en sistemas distintos (el proyecto en un disco exFAT y `TMPDIR` en
APFS). (b) **No retirar la copia hasta haber creado el enlace** (apartarla con un `rename`, enlazar y, si falla,
devolverla): arregla la oscilación de FR-045, pero `--dry-run` seguiría prediciendo `enlace` y `doctor` seguiría dando
el hallazgo «copia» con una orden que no lo arregla, porque la predicción seguiría saliendo de otro sistema de ficheros;
y añade a la fase 1 una operación más que `TestFalloAMitadSeCompleta` tendría que cubrir sin ganar nada sobre la
sonda en el propio ámbito, que ya evita que la copia se retire cuando no se puede enlazar. (c) Deducir la capacidad del tipo de sistema de ficheros (`statfs`): distinto en cada plataforma, sin
equivalente en la biblioteca estándar para Windows y no concluyente en recursos de red, donde decide el servidor. (d)
Sondear por encima de la raíz cuando no existe: escribiría fuera del ámbito (FR-027). (e) Devolver siempre `true` en Unix
y sondear solo en Windows: código de Windows sin ningún test que lo ejerza, y falso en Unix sobre exFAT.

### D10 · El manifiesto: JSON canónico, estricto al leer

**Decisión.** `kitlegal.json` se serializa con `encoding/json/v2` en forma determinista (`json.Deterministic(true)`,
claves de mapa ordenadas), sangrado de dos espacios y salto final, y se lee con `json.RejectUnknownMembers(true)` y
nombres duplicados rechazados (lo que `jsontext` hace por omisión), más las reglas de forma de
[contracts/manifiesto.md](./contracts/manifiesto.md) (nombres de skill, rutas relativas limpias, huellas `sha256:` de 64
hexadecimales, un único host `claude`, modos `enlace`/`copia`, versiones sin caracteres de control). Todo incumplimiento
es «manifiesto ilegible» (FR-035). Forma: `version`, y `skills` como objeto por nombre, cada una con `version`,
`ficheros` (ruta relativa al directorio neutro → huella) y, si tiene, `hosts` (por host: `ruta` relativa a la raíz del
ámbito, `modo` y, en copia, `ficheros`).

**Por qué.** Mapas por nombre y claves ordenadas dan manifiestos byte a byte iguales en cualquier máquina (FR-032) sin
código de ordenación propio (V21). La lectura estricta es lo que hace que un manifiesto manipulado —una clave repetida
que esconda otra, una ruta con `..`— sea ilegible y no una instrucción para tocar lo que no es suyo. Las versiones sin
caracteres de control garantizan que el aviso sea una sola línea (FR-071).

**Alternativas rechazadas.** (a) `encoding/json` v1: admite claves duplicadas (gana la última) y habría que
reimplementar el rechazo. (b) YAML: dependencia en el binario y un formato menos estricto. (c) Un esquema publicado en
`schemas/`: fuera de alcance (spec).

### D11 · Rutas que se presentan: como se alcanzan desde el directorio de trabajo

**Decisión.** Toda ruta de la salida (`data`, conflictos, hallazgos y órdenes) se escribe con `/` y como se alcanza
desde el directorio de trabajo de la invocación: relativa en el ámbito local (`.agents/skills/boe-legislacion`),
absoluta con `-g` (`HOME` ya expandido) y, con `--dir`, colgando de la ruta tal como se pasó (relativa o absoluta),
limpia y sin barra final. El ámbito local no llama a `os.Getwd`: la raíz es el propio directorio de trabajo.

**Por qué.** Es la regla de FR-066 para las órdenes, y usarla en todo hace que la ruta de un hallazgo sea la que su
orden retira. Rutas relativas en local hacen la salida independiente de dónde esté el proyecto.

### D12 · Conflictos y hallazgos viajan como texto en el sobre de fallo, con un formato fijo

**Decisión.** Con conflictos o hallazgos, el dominio devuelve un error que declara la clase `inesperado`
(`schema.ConClase`) y cuyo mensaje es una cabecera y una línea por entrada, en orden determinista:
`<clase>: <ruta>` en `install`, `list` y `doctor` (fallos de ámbito), y `<clase>: <ruta>: <orden>` en los hallazgos de
`doctor`. El kernel lo emite sin cambios: `ok:false`, `data` `{clase, mensaje}` y el mismo texto en la salida de error
(V25). El formato exacto está en [contracts/applet-skills.md](./contracts/applet-skills.md) §5.

**Por qué.** Es la clarificación del spec (FR-052): el ADR 0006 no se enmienda. Declarar la clase explícitamente, en
lugar de dejarla caer en «lo no previsto», la hace comprobable en test.

### D13 · `--dry-run` por el canal de ensayo del kernel

**Decisión.** Con `--dry-run`, `install` hace el mismo plan y, si no hay conflictos, lo describe en
`Resultado.Ensayo`, una línea por skill pedida con su ruta, su estado y sus enlaces con su modo; el kernel las
presenta en la salida de error con su prefijo (V24) y no hay sobre en la salida estándar. Con conflictos, devuelve el
mismo error que sin la bandera: mismo mensaje, mismo exit 1. Nada toca el disco, ni directorios vacíos (FR-048).

**Por qué.** `--dry-run` es una bandera global cuyo contrato fija el kernel desde H1 (ADR 0011): «se habría hecho»
por la salida de error y sin sobre. Emitir un sobre solo para `skills` cambiaría el kernel para un applet. «La misma
salida» de FR-048 se lee así: la misma información —estado por skill, enlaces con su modo y conflictos— y el mismo
código.

**Alternativa rechazada.** Emitir el sobre en `--dry-run`: cambio del kernel (`internal/cli`/`internal/app`) que
afecta a todos los applets y que el spec deja fuera.

### D14 · Validación de la invocación, antes de tocar el disco

**Decisión.** En este orden, sin leer nada del disco: `-g` con `--dir` → exit 2 («`-g` y `--dir` se excluyen»);
`--host` con `--dir` → exit 2; `--host` distinto de `claude` → exit 2; un nombre que no es de una skill empotrada →
exit 2 con la lista de las disponibles (un repetido cuenta una vez); `-g` sin `HOME` o con `HOME` vacío → exit 1. La
validación la hace el verbo con errores del dominio que declaran su clase (`argumentos` → 2), no Kong: el mensaje de
Kong para `xor` o `enum` sale en inglés y no dice lo que pide FR-013.

**Por qué.** FR-010, FR-012, FR-013 y FR-020 exigen «nada escrito ni leído del disco». El `enum` del esquema de
`--describe` para `--host` se declara con la etiqueta `jsonschema` (V22).

### D15 · Salida de los tres verbos

**Decisión.** `install`: `data` es la **lista** de skills pedidas, cada una con `nombre`, `ruta`, `estado` y `enlaces`
(`host`, `ruta`, `modo`), que es exactamente FR-051. `list`: objeto con `directorio`, `manifiesto` (booleano), `version`
(cadena o nula) y `skills` (`nombre`, `ruta`, `version`, `empotrada`, `enlaces`). `doctor` sin hallazgos: objeto con
`directorio`, `manifiesto`, `version`, `version_del_binario` y `hallazgos` (lista, siempre vacía en éxito). La versión
nula se declara con `jsonschema:"nullable"` (V22). Procedencia de applet calculado: `fuente` `kitlegal.skills`, `url`
`kitlegal:applet/skills`, fecha del reloj del kernel (FR-050; ADR 0006).

**Por qué.** Claves en español y en `snake_case`, como `fecha_consulta`; la forma de cada una es la de su FR, sin
campos no pedidos.

### D16 · El contrato publicado: `schemas/instalacion.json`

**Decisión.** Un fichero publicado nuevo, `schemas/instalacion.json` (entidad `instalacion`, applet `skills`, verbos
`doctor`, `install` y `list`), generado con la receta existente (`TestEsquemasPublicados -actualizar-esquemas`) y con
su fila en la tabla de `internal/app/esquemas_test.go`. Registrar el applet y publicar su esquema van juntos: en
cuanto el verbo está en el registro, `TestEsquemasCubrenTodosLosVerbos` exige su parte (V27), el mismo patrón de H6.

### D17 · `skills-sync` y `skills-check` sin enlaces; `scripts/` es un defecto

**Decisión.** En `internal/skills`: la tabla de comandos se titula y se escribe con `kitlegal <applet>` (constante que
sustituye a `carpetaDeLosScripts`); desaparecen `enlaces.go`, `EnlacesEsperados`, las tres clases de deriva de enlaces
y los arreglos que creaban enlaces; y **tener `scripts/`** (cualquier entrada con ese nombre en el directorio de una
skill, vista con `Lstat`) pasa a ser un `DefectoDeSkill` que `skills-sync` y `skills-check` presentan y que hace fallar
los dos (FR-082). `skills-sync` no lo borra: retirar lo que alguien puso a mano no está especificado.

**Por qué.** Un defecto, y no una deriva, porque no hay nada regenerado con lo que arreglarlo.

### D18 · `make install`, el bucle de desarrollo

**Decisión.** `go install` con las mismas inyecciones y, a continuación, el binario recién instalado —por la ruta que
da `go list -f '{{.Target}}' ./cmd/kitlegal`, no por el `PATH`— con `skills install -g --host claude`. Sin comprobación
previa: si hay un conflicto, la segunda orden sale con 1 y `make` también. `TestInstalacion` (integración) pasa a
comprobar lo nuevo sobre la copia mínima del árbol con `HOME`, `GOBIN` y `GOPATH` temporales (FR-126, SC-018).

**Por qué.** Es FR-125 literal. La ruta de `go list` es la del binario que se acaba de construir y no la de otro que
esté antes en el `PATH`.

### D19 · El job de evals: el binario en el `PATH` de forma explícita

**Decisión.** En `.github/workflows/evals.yml`, el paso de instalación hace `make install` y añade a `$GITHUB_PATH` el
directorio de `go list -f '{{.Target}}'`; el paso que retira Python localiza el binario con `command -v kitlegal` (y
falla si no está) en lugar de `bin/instalado/`, y añade `$HOME/.agents` a lo que no puede retirar. `scripts/evals.sh`
comprueba además, antes de la primera sesión, que `kitlegal` está en el `PATH`. El texto de la sesión de prueba de red
(`internal/evals/preparar.go`) invoca `kitlegal boe articulo …`.

**Por qué.** FR-127. No se confía en que `setup-go` ponga `GOPATH/bin` en el `PATH` (S7): se hace explícito.

### D20 · goreleaser como módulo de herramienta: `tools/goreleaser`, v2.18.1

**Decisión.** `tools/goreleaser/go.mod` con `tool github.com/goreleaser/goreleaser/v2` en v2.18.1, invocado como las
demás (`go tool -modfile=tools/goreleaser/go.mod goreleaser`), cubierto por `make mod-verify` (que descubre los
`tools/*/go.mod`) y con su entrada en `.github/dependabot.yml`, como las otras cuatro herramientas.

**Por qué.** FR-096. v2.18.1 es la versión presente en la caché local (publicada el 2026-09-05) y declara `go 1.27.1`,
el parche que fija el repositorio (V1). Su código es el que se ha leído para todas las afirmaciones V2-V17.

**Consecuencia.** El grafo de módulos de goreleaser **no** está entero en la caché local: `go mod tidy` sin proxy falla
en `github.com/Azure/go-autorest@v14.2.0+incompatible` (V33). Crear `tools/goreleaser/go.sum` necesita el proxy de
módulos de Go, igual que lo necesitaron las cuatro herramientas de H0 y lo necesita `make vuln` en cada `make ci`. No
es una fuente de datos y no pasa por `grabar_datos` (S10).

### D21 · `.goreleaser.yaml`: sin ninguna propiedad obsoleta, con `homebrew_casks`

**Decisión.** La configuración se escribe contra v2.18.1 sin ninguna propiedad obsoleta (FR-093, FR-094):

- `version: 2`, `project_name: kitlegal` y `release.github` con `owner: jmorenobl` y `name: kitlegal` fijados, de
  modo que `goreleaser check` no dependa de git ni de un remote (V2).
- `builds`: `main: ./cmd/kitlegal`, `binary: kitlegal`, `env: [CGO_ENABLED=0]`, `flags: [-trimpath]`, `goos: [darwin,
  linux, windows]`, `goarch: [amd64, arm64]` (el valor por omisión incluiría 386, V16) y **las cuatro** `-X` del
  `Makefile`, con la versión `{{ if .IsSnapshot }}{{ .Version }}{{ else }}{{ .Tag }}{{ end }}` en `main.version` y en
  `internal/httpx.version`, `{{ .FullCommit }}` en `main.commit` y `{{ .Date }}` en `main.fecha` (V6).
- `archives`: `name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"` (sin versión, D22), `formats: [tar.gz]` y
  `format_overrides` de `windows` a `zip` (V7); ficheros del archivo, los de omisión (licencia, README y CHANGELOG).
- `checksum`: `name_template: checksums.txt`, SHA-256 (por omisión), y `extra_files` con `scripts/install.sh`, que
  así también lleva su huella (D23).
- `sboms`: los valores por omisión, un SBOM de syft por archivo (V9).
- `signs`: `cmd: cosign`, `artifacts: checksum`, firma keyless en un *bundle* (`sign-blob --bundle=${signature}
  ${artifact} --yes`), el patrón del propio `.goreleaser.yaml` de goreleaser (V10; flags de cosign: S3).
- `homebrew_casks` (no `brews`, obsoleto: V3) en `jmorenobl/homebrew-tap`, con `token: "{{ .Env.PUBLISHER_TOKEN }}"`,
  `homepage`, `description` y un `hooks.post.install` que retira la cuarentena de macOS del binario instalado (D25).
- `scoops` en `jmorenobl/scoop-bucket` con el mismo token, `homepage`, `description` y `license: Apache-2.0` (V12).
- `nfpms`: `formats: [deb, rpm]`, `maintainer` (D26), `description`, `homepage`, `license: Apache-2.0` (V13).
- `release.extra_files`: `scripts/install.sh` (FR-109).
- `changelog`: `use: git` con grupos por expresión regular de Conventional Commits (V14). Es lo que el hito llama
  «notas de release desde los Conventional Commits»; `CHANGELOG.md` sigue a mano (spec, *Fuera de alcance*).

La plantilla del token solo se evalúa al publicar (V11, V12): ni `check` ni el snapshot necesitan `PUBLISHER_TOKEN`
(FR-097).

**Alternativas rechazadas.** (a) `brews` (fórmula): `goreleaser check` saldría con 2 (V2, V3) y FR-094 no se
cumpliría. (b) `goreleaser-action` en `release.yml`: una segunda forma de fijar la versión de la herramienta; con el
módulo de `tools/`, la versión la fija `go.sum`, como las demás (FR-096).

### D22 · Archivos sin versión en el nombre, y `install.sh` por las URL de descarga de GitHub

**Decisión.** Los seis archivos se llaman `kitlegal_<os>_<arch>.tar.gz` (`.zip` en Windows) y el fichero de
checksums, `checksums.txt`. `install.sh` descarga de `<base>/latest/download/<archivo>` sin versión y de
`<base>/download/v<versión>/<archivo>` con ella, donde `<base>` es `https://github.com/jmorenobl/kitlegal/releases`.

**Por qué.** Con el nombre sin versión, «la última release» es una URL fija de GitHub y `install.sh` no tiene que
consultar la API ni interpretar JSON (sin `jq`, sin más requisitos: FR-100). La forma de esas URL no se puede
comprobar sin red (S1): la comprueba el trabajo de humo de `release.yml` tras la primera etiqueta (humano).

**Alternativa rechazada.** Consultar `api.github.com/repos/…/releases/latest`: exige interpretar JSON en `sh`, añade
el límite de peticiones anónimas de la API y una segunda URL que puede fallar.

### D23 · La huella de `install.sh` también va en `checksums.txt`

**Decisión.** `checksum.extra_files: [{glob: scripts/install.sh}]`, además de `release.extra_files`.

**Por qué.** FR-093 pide la huella de «todos los artefactos subidos», y `install.sh` se sube; `release.extra_files`
solo lo adjunta y no entra en los checksums (V8, V14). Así la firma de los checksums cubre también el instalador.

### D24 · El arnés e2e: binarios con versión, enlazador que falla, `arbol` y un origen de release local

**Decisión.** `internal/app/e2e_test.go` y el binario de e2e ganan lo que los guiones de aceptación necesitan
([contracts/arnes-e2e.md](./contracts/arnes-e2e.md)):

- además del `kitlegal` de siempre (versión `dev`, un binario de desarrollo), tres construcciones del mismo binario
  con `-ldflags` constantes: versión `v0.1.0`, versión `v0.2.0`, y versión `v0.1.0` con el enlazador que siempre
  falla, cuyas rutas absolutas exporta en `KITLEGAL_V1_BIN`, `KITLEGAL_V2_BIN` y `KITLEGAL_SIN_ENLACES_BIN`. Invocados
  por un enlace llamado `kitlegal`, el despacho toma el applet del primer argumento (el nombre no está registrado).
  El enlazador que falla es un tipo del `package main` de e2e (nunca de `internal/disco`); cuando la variable `-X` lo
  elige, `registroDeE2E` lo pone en el campo `Enlazador` de `app.DependenciasDeSkillsDelSistema(version)` antes de
  pasarlas a `app.AppletSkills` (D9), sin tocar el applet;
- `KITLEGAL_SKILLS` (el `skills/` del repositorio, para comparar lo instalado byte a byte con lo empotrado) y
  `KITLEGAL_INSTALADOR` (`scripts/install.sh`);
- un **origen de release local**, construido una vez por el arnés con el binario de versión `v0.1.0` empaquetado como
  `kitlegal_<GOOS>_<GOARCH>.tar.gz` y su `checksums.txt` (con una línea señuelo cuyo nombre contiene el del archivo),
  en la disposición de D22, en `KITLEGAL_ORIGEN`, con `KITLEGAL_ORIGEN_VERSION` y `KITLEGAL_ORIGEN_ARCHIVO`. Con
  `KITLEGAL_DIST` definido, el origen se construye con el archivo y los checksums **del snapshot** de `dist/` y la
  versión de `dist/metadata.json` (FR-108, FR-120);
- la orden `arbol <directorio>`: una línea por entrada, sin seguir enlaces y en orden de bytes, con tipo, ruta y, según
  el tipo, huella y permisos o destino literal; `cp stdout` y `cmp` sobre ella comparan el disco «byte a byte»;
- `http_proxy`, `https_proxy`, `HTTP_PROXY`, `HTTPS_PROXY` y `ALL_PROXY` a `http://127.0.0.1:9` y `NO_PROXY` vacío en
  todos los guiones: toda petición HTTP(S) que se colara fallaría (V32).

**Por qué.** Sin binarios de versiones distintas no hay aviso ni «actualizada» que probar; sin enlazador inyectable no
hay SC-012 en e2e (FR-143); sin una instantánea sin seguir enlaces no hay «0 cambios en disco» comprobable. Las
órdenes de ejecución siguen escritas con constantes, como exige el análisis de seguridad del arnés actual.

**Alternativas rechazadas.** (a) Instantáneas con `find`, `ls -l` o `stat`: difieren entre BSD y GNU y `ls -l` no
compara contenidos. (b) Variables de entorno en el binario de e2e para elegir el enlazador: el binario de e2e no lee
del entorno nada que decida su comportamiento (comentario de su `main.go`); `-X` sobre una variable de cadena del
`package main` de e2e, fijada al construirlo, lo mantiene así (`-X` «add string value definition of the form
importpath.name=value»: `go tool link -help`; por eso es una variable de cadena y no una constante).

### D25 · El cask retira la cuarentena de macOS

**Decisión.** `homebrew_casks[0].hooks.post.install` retira `com.apple.quarantine` del binario instalado, solo en
macOS.

**Por qué.** El roadmap da la notarización por innecesaria «porque Homebrew no la necesita», premisa cierta para una
fórmula; `brews` está obsoleto y el cask aplica cuarentena a lo que descarga (S5), así que sin el gancho el binario sin
notarizar no arrancaría en un Mac y FR-093 («lo que el tap necesita para que `brew install` instale kitlegal») no se
cumpliría. El campo existe en v2.18.1 (V11); lo comprueba la persona en un Mac limpio tras la etiqueta (FR-150).

**Alternativa rechazada.** Notarizar: fuera de alcance (spec).

### D26 · `nfpms.maintainer`

**Decisión.** `maintainer: "Jorge <jmorenobl@gmail.com>"`, la identidad de autor que ya llevan todos los commits del
repositorio.

**Por qué.** Sin `maintainer`, goreleaser marca la configuración como obsoleta y `check` sale con 2 (V3, V13). Es la
identidad de autoría que el historial ya publica; no se introduce ningún dato personal nuevo. Queda como supuesto
visible en el informe (`gates/supuestos.md`) por si la persona prefiere otra dirección.

**Alternativa rechazada.** Una dirección `noreply` de GitHub: su forma exacta para esta cuenta no se puede comprobar
sin red.

### D27 · `install.sh`

**Decisión.** POSIX `sh` con `set -eu`, sin `bash`, `jq`, Go ni git. Orden: argumentos (0 o 1; versión SemVer 2.0.0
con `v` opcional, comprobada con `grep -E` y `LC_ALL=C`) → sistema y arquitectura (`uname -s`/`uname -m`, FR-101) →
directorio de instalación (`$KITLEGAL_INSTALL_DIR` no vacío, si no `$HOME/.local/bin`; sin ninguno de los dos, fallo
**antes de descargar**, FR-103) → temporal con `mktemp -d` y `trap` que lo retira → `curl -fsSL` del archivo y de
`checksums.txt` → huella de la línea cuyo **segundo campo es exactamente** el nombre del archivo (`awk '$2 == f'`), con
`sha256sum` o, si no está, `shasum -a 256` → extracción solo del miembro `kitlegal` → copia a un temporal del
directorio de instalación y `mv -f` encima (un `kitlegal` anterior sigue intacto hasta el final) → si el directorio no
está en el `PATH`, la línea exacta `export PATH='<dir>':"$PATH"` (con la comilla simple escapada) → última línea,
`kitlegal skills install`. Todo error va a la salida de error con código distinto de 0 y un mensaje que nombra la
versión (o «la última») y lo que falló.

**Mecanismo para el origen local (FR-107).** `KITLEGAL_INSTALL_URL`, la base de las URL de D22 (por omisión la de
GitHub). Los guiones de aceptación la apuntan a `file://…` de su copia del origen, y los proxies del arnés (D24)
hacen fallar cualquier petición HTTP(S).

**Por qué.** Cada paso es un FR de 100 a 106; `curl` es el requisito que la propia orden de entrega ya tiene
(`curl … | sh`). No hay alternativa `wget`: no está especificada.

### D28 · Dónde se prueban la release y el snapshot

**Decisión.**

- `TestConfiguracionDeLaRelease` (raíz del módulo, en `make test`): lee `.goreleaser.yaml`, `Makefile`,
  `.github/workflows/release.yml` y `ci.yml` con `yaml.v3` (solo en test) y comprueba lo mecánico: las seis
  plataformas, `CGO_ENABLED=0`, `-trimpath`, que las `-X` de goreleaser son **las mismas cuatro** que las del
  `Makefile` (FR-091), las plantillas de versión (FR-092), checksums, SBOM, firma, cask, bucket, nfpm y
  `extra_files` declarados (FR-093), `PUBLISHER_TOKEN` nombrado solo en los tokens del tap y del bucket y usado solo en
  el paso de publicación (FR-097, FR-114), `release.yml` solo en etiquetas `v*` con `id-token: write` y
  `attestations: write`, la atestación y el trabajo de humo con sus seis comprobaciones (FR-110 a FR-113), y el trabajo
  de snapshot de `ci.yml` sin `id-token` ni secretos y con `make release` y `make snapshot-check` (FR-120).
- `TestSinInstalacionPorEnlaces` (raíz, en `make test`): ni `SKILL.md`, ni el `Makefile`, ni `.github/` contienen
  `scripts/boe`, `scripts/territorio` ni `bin/instalado`, y no existen `scripts/instalar-skills.sh` ni
  `internal/skills/enlaces.go` (SC-014, FR-083).
- `TestSnapshot` (raíz, etiqueta de compilación `snapshot`, en `make snapshot-check`): sobre `dist/` —con
  `metadata.json` y `artifacts.json` (V15)— los seis archivos, cada línea de `checksums.txt` que los nombra con la
  huella calculada, ni SBOM ni firmas, y el binario de la plataforma que ejecuta el test imprime en `version` la versión
  y el commit de `metadata.json` (en el trabajo de CI, `ubuntu`: linux/amd64, SC-016).
- Las pruebas de `install.sh` son **guiones de aceptación** (`instalador-*.txtar`) contra el origen local del arnés en
  `make ci`, y los mismos contra el snapshot en `make snapshot-check` con `KITLEGAL_DIST` (FR-108, FR-120). El arnés
  exige, con `KITLEGAL_DIST` definido, que haya al menos un guion `*instalador-*` que ejecutar, para que el objetivo no
  pase en vacío.
- **Cómo se verifica dentro del run.** Los guiones `instalador-*` solo están en `internal/app/testdata/script/` después
  de `scripts/workflow/aceptacion.sh activar`, que el workflow ejecuta tras el bucle de tareas (V19), y `install.sh` no
  existe hasta su tarea. Por eso `make snapshot-check` completo **no** puede pasar solo con lo que hay en
  `testdata/script/` durante el bucle, y el plan lo resuelve con el mismo mecanismo que `aceptacion.sh rojo-primero`:
  las tareas que lo necesitan copian **un momento** cada `aceptacion/instalador-*.txtar` congelado del directorio del
  feature a `internal/app/testdata/script/zz-<nombre>.txtar` (el nombre sigue conteniendo `instalador-` y no choca con
  el `h19-` de la activación), ejecutan, y las retiran con `rm -f` **antes** de `make ci` y del commit; `git status
  --porcelain internal/app/testdata/script` vacío lo confirma. Las copias no se editan ni se declaran: son las
  congeladas, byte a byte. Por pasos del plan:
  - `install.sh` (paso 13), **antes** que la release: `.goreleaser.yaml` lo nombra en `checksum.extra_files` y el
    snapshot falla con una ruta literal que no existe (V48). Con las copias, `go test -count=1 -run
    '^TestEntregaDelHito$/instalador-' ./internal/app/` contra el origen local;
  - la release (paso 14): `make release` (S9); `make snapshot-check` **sin** las copias sale con error y su salida
    contiene el mensaje del arnés que dice que no hay ningún guion `instalador-` (la garantía de no pasar en vacío,
    ejercida); y **con** ellas, `make snapshot-check` completo en verde;
  - CI (paso 15): `make release` y `make snapshot-check` completo con las copias; y `TestConfiguracionDeLaRelease`
    gana en esta tarea lo que fija de `release.yml` y del trabajo `snapshot` de `ci.yml`, que antes no existen (en el
    paso 14 comprueba `.goreleaser.yaml` y el `Makefile`).

  Tras la activación, los guiones son `h19-instalador-*` y `make snapshot-check` los encuentra sin copias: es lo que
  ejecuta el trabajo de snapshot de CI en la propuesta de cambio.

**Por qué.** Lo que se puede comprobar con un programa no se deja a un juez (constitución, «Gates»). La raíz del
módulo es donde viven `.goreleaser.yaml`, el `Makefile` y `skills.go`; los tests de la raíz no añaden ningún paquete
nuevo al árbol. `yaml.v3` es dependencia de §V y en un `_test.go` no entra en el binario (`TestDependenciasDelBinario`
mide `./cmd/kitlegal`).

**Alternativa rechazada.** Un paquete de solo tests (`internal/release/`): un directorio más sin código de
producción, fuera de la estructura de §2.

### D29 · Objetivos del `Makefile`

**Decisión.** `GORELEASER := go tool -modfile=tools/goreleaser/go.mod goreleaser`;

- `goreleaser-check`: `$(GORELEASER) check`; entra en `ci` (FR-094);
- `release`: `$(GORELEASER) release --snapshot --clean --skip=publish,sign,sbom` (FR-095, la única definición del
  snapshot); `--skip=publish` es redundante con `--snapshot` pero es el literal de la clarificación y es válido (V4);
- `snapshot-check`: `go test -count=1 -tags=snapshot -run '^TestSnapshot$$' .` y los guiones del instalador contra
  `dist/` (`KITLEGAL_DIST=$(CURDIR)/dist go test -count=1 -run '^TestEntregaDelHito$$/instalador-' ./internal/app/`);
  lo ejecuta el trabajo de snapshot de CI tras `make release`; dentro del run, antes de activar la suite, se verifica
  como fija D28 («Cómo se verifica dentro del run»);
- `install`: D18; `skills-sync` y `help` con el texto nuevo (FR-121).

**Por qué.** El `Makefile` es la única superficie de invocación (CLAUDE.md; `ci.yml` no admite un `go test` suelto).
`snapshot-check` es el nombre que sigue a `schema-check` y `skills-check`, y lo separa de `goreleaser-check`, que
valida la configuración y no el resultado.

### D30 · CI: un trabajo de snapshot y `release.yml`

**Decisión.**

- `ci.yml` gana el trabajo `snapshot` (en `pull_request` y en `push` a `main`): `contents: read`, sin secretos y sin
  `id-token`; `checkout` con `fetch-depth: 0` (para que la versión de snapshot salga de la historia; V5), `setup-go`
  como el trabajo `ci`, `make release` y `make snapshot-check`.
- `release.yml`, solo `on: push: tags: ['v*']`: trabajo `publicar` (`contents: write`, `id-token: write`,
  `attestations: write`) que instala syft y cosign, ejecuta `go tool -modfile=tools/goreleaser/go.mod goreleaser
  release --clean` con `GITHUB_TOKEN` y `PUBLISHER_TOKEN` en el entorno **solo de ese paso**, y atesta con
  `actions/attest-build-provenance` los seis archivos y `checksums.txt`; trabajo `humo` (`needs: publicar`,
  `contents: read`, `attestations: read`) que descarga el archivo linux/amd64 publicado y `checksums.txt` con `gh
  release download`, verifica su huella, `gh attestation verify … --repo jmorenobl/kitlegal`, `kitlegal version`
  imprime la etiqueta, `kitlegal boe articulo BOE-A-2015-10565 a21 --offline` con `KITLEGAL_CACHE_DIR` vacío sale con
  4 y `kitlegal skills install` en un directorio vacío deja `.agents/skills/boe-legislacion/SKILL.md` (FR-113).

**Por qué.** FR-110 a FR-115 y FR-120. Las acciones de terceros se fijan por su etiqueta mayor, como el resto de flujos
del repositorio (V36); sus versiones no se pueden comprobar sin red (S2) y las mantiene Dependabot
(`github-actions`). Nada de `release.yml` se ejecuta en el run (FR-115).

### D31 · Igualdad de versiones y forma SemVer

**Decisión.** `MismaVersion(a, b)`: quitar **una** `v` inicial a cada una y comparar byte a byte (FR-077).
`FormaSemVer(v)`: la gramática de SemVer 2.0.0 (la expresión regular de semver.org, con `v` opcional), sin ordenar
nada. La misma gramática, en ERE, la usa `install.sh` (D27).

**Por qué.** Es FR-073 y FR-077 literales; no se ordena (spec, *Fuera de alcance*).

### D32 · Paquetes y comprobaciones nuevas de arquitectura

**Decisión.** Dos cambios en las reglas ejecutables:

1. **R1 alcanza al adaptador nuevo y al paquete raíz**, en sus dos capas y en la misma tarea que crea
   `internal/disco` (paso 3 del plan), antes de que exista ningún test que pudiera saltársela:
   - `.golangci.yml`, lista `core` de `depguard` (la que vigila también los `_test.go`, V26): dos entradas nuevas en
     `deny`, `github.com/jmorenobl/kitlegal/internal/disco` («R1: el dominio no importa adaptadores; el disco llega por
     los puertos que define internal/core/instalacion») y **`github.com/jmorenobl/kitlegal$`** («R1: el dominio no
     importa lo empotrado; recibe las skills como bytes desde internal/app»). El `$` final es la coincidencia exacta de
     `depguard` (V45): sin él, el prefijo del paquete raíz capturaría el módulo entero, `internal/core` incluido. El
     comentario de la regla deja de decir «los ocho paquetes de internal/».
   - `internal/arch_test.go`: `disco` entra en `paquetesInternos` (nueve paquetes; el comentario se actualiza), y
     `compruebaDominioPuro` gana una lista de **denegaciones exactas** con la ruta del módulo (`g.modulo`), comparadas
     con `==` y no con `cuelgaDe`, que es de prefijo (V47): así una arista del dominio al paquete raíz se nombra como
     tal, y no solo de forma transitiva por su `io/fs`.

   Los `_test.go` los ve solo `depguard`, porque el grafo de `TestArquitectura` sale de `go list -deps ./...` sin tests
   (V47): la entrada de `depguard` es la que hace fallar `make lint` ante un `package instalacion_test` que importe
   `internal/disco` o el paquete raíz. Se demuestra con una sonda temporal: un `_test.go` externo en
   `internal/core/instalacion` con una importación en blanco sola en su línea (uniq-by-line agrupa por línea), `make
   lint` en rojo nombrando R1, y la sonda retirada antes de `make ci`: la de `internal/disco` en el paso 3 y la del
   paquete raíz en el paso 4, que es cuando ese paquete existe (antes, la importación no compilaría y el lint no
   llegaría a `depguard`).
2. `TestArquitectura` gana una subprueba: `internal/core/instalacion`, `internal/disco` y el paquete raíz no alcanzan
   `net`, `net/http` ni `internal/httpx` (directa ni transitivamente dentro del módulo). Es la garantía mecánica de
   SC-022 («el applet `skills` y el aviso abren 0 conexiones»): R2 solo prohíbe `net/http` fuera de `httpx`, no `net`.

El dominio nuevo queda dentro de R1 sin tocar la selección de orígenes (recorre todo `internal/core/**`, V26).

**Alternativas rechazadas.** (a) Confiar en la detección transitiva (`internal/disco` → `os`, raíz → `io/fs`): solo
existe en el grafo de producción; un `_test.go` externo del dominio no entra en `go list -deps ./...` y `depguard` solo
mira importaciones directas. (b) Denegar el paquete raíz por prefijo, sin `$` en `depguard` o con `cuelgaDe` en el test:
capturaría todo el módulo y R1 fallaría sobre el propio dominio. (c) Dejarlo para la tarea de la subprueba D32 (paso 9):
entre los pasos 3 y 9 el dominio podría ganar un test que importe el adaptador sin que nada lo detecte.

### D33 · `.golangci.yml`: la etiqueta `snapshot` y R1 ampliada

**Decisión.** `run.build-tags` gana `snapshot`, como ya tiene `integration`, `fuentes`, `grabacion` y `evals`: sin
ella, `TestSnapshot` quedaría fuera de todo lint. La lista `core` de `depguard` gana las dos entradas de D32. Ninguna
regla se relaja ni se exceptúa: las dos únicas ediciones de reglas **añaden** denegaciones. Si `misspell` marca alguna
palabra española nueva, se resuelve sin tocar `.golangci.yml`, con la técnica ya documentada en el repositorio (S14).

### D34 · Lo que no entra

Siguiendo el criterio conservador, y además de la lista de *Fuera de alcance* del spec: ni `wget` en `install.sh`, ni
verificación de firma o atestación en `install.sh`, ni `USERPROFILE` como sustituto de `HOME` en Windows (FR-012 dice
`$HOME`), ni un verbo por omisión para `skills` (nombrar el verbo es obligatorio, como en `territorio`), ni borrar
`scripts/` desde `skills-sync`, ni cambios en `CLAUDE.md`, la constitución, `docs/ROADMAP.md` o los ADR (las menciones
desfasadas van a *Pendientes* de la propuesta de cambio). Ninguna fuente externa se graba: el hito no tiene paso
`grabar_datos` ni fila nueva en `docs/SOURCES.md`.

---

## Verificaciones

| # | Afirmación | Dónde se comprobó |
|---|---|---|
| V1 | goreleaser v2.18.1 está en la caché local, publicado el 2026-09-05, y declara `go 1.27.1` | `/Users/jorge/go/pkg/mod/cache/download/github.com/goreleaser/goreleaser/v2/@v/v2.18.1.info`; `goreleaser:go.mod:3` |
| V2 | `goreleaser check` solo ejecuta el pipe de valores por omisión; sale con 1 si la configuración es inválida y con 2 si usa propiedades obsoletas; la carga del YAML es estricta (campo desconocido = error); con `project_name` y `release.github.owner/name` fijados no necesita git, remote, red ni tokens | `goreleaser:cmd/check.go:47-81`; `goreleaser:pkg/config/load.go:59`; `goreleaser:internal/yaml/yaml.go:13-16`; `goreleaser:internal/pipe/release/scm.go:11-24`; `goreleaser:internal/pipe/project/project.go:23-39` |
| V3 | Obsoletas en v2.18.1 relevantes: `snapshot.name_template`, `archives.format`, `archives.format_overrides.format`, `archives.builds`, `brews`, `homebrew_casks.{binary,manpage,url.verified,conflicts.formula}`, `nfpms.maintainer` vacío, `nfpms.builds`, `builds.gobinary`; sin avisos en scoops, signs, sboms, checksum, changelog y release | `goreleaser:internal/pipe/snapshot/snapshot.go:26`; `archive/archive.go:63,72,80`; `brew/brew.go:61`; `cask/cask.go:73,83,87,91`; `nfpm/nfpm.go:76-80`; `build/build.go:95`; `internal/deprecate/deprecate.go:47` |
| V4 | `publish`, `sign` y `sbom` son valores válidos de `--skip` en `release`; `--snapshot` ya salta publish pero no sign ni sbom; `--clean` vacía `dist` | `goreleaser:internal/skips/skips.go:83-85,114-119`; `cmd/release.go:161-163`; `internal/pipe/sign/sign.go:38-40`; `internal/pipe/sbom/sbom.go:36-38`; `internal/pipe/dist/dist.go:16-54` |
| V5 | Snapshot: plantilla `{{ .Version }}-SNAPSHOT-{{ .ShortCommit }}`; sin etiquetas, `.Tag` = `v0.0.0` y la versión `0.0.0-SNAPSHOT-<corto>`; sin remote, `0.0.0-SNAPSHOT-none`; no exige árbol limpio y funciona con clon superficial | `goreleaser:internal/pipe/snapshot/snapshot.go:22-40`; `internal/pipe/git/git.go:58,63-86,132-145,176-185,275-280` |
| V6 | `.Version` quita la `v` (fuera de snapshot) o es la del snapshot; `.Tag` la conserva; `.IsSnapshot`; `.FullCommit` es el SHA completo; `.Date` es RFC 3339 en UTC; `.Env` con clave ausente es error | `goreleaser:internal/tmpl/tmpl.go:112-147,275`; `internal/pipe/git/git.go:58,162-163` |
| V7 | Archivos: plantilla `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}` válida (solo exige nombres únicos); `formats` y `format_overrides.formats`; binario en la raíz del archivo; ficheros por omisión licencia, readme y changelog | `goreleaser:internal/pipe/archive/archive.go:30-31,83-92,187-190,234-247,346-361`; `pkg/config/config.go:595-601,646-664` |
| V8 | Checksums: `name_template` literal válido; SHA-256 por omisión; línea `<hash>  <nombre>` ordenada por nombre; incluye archivos, paquetes nfpm y SBOM; `release.extra_files` no entra; `checksum.extra_files` sí | `goreleaser:internal/pipe/checksums/checksums.go:41-50,100,170-206,226`; `pkg/config/config.go:1194-1200` |
| V9 | SBOM por omisión: `syft`, un documento `{{ .ArtifactName }}.sbom.json` por archivo; con `--skip=sbom` no se invoca | `goreleaser:internal/pipe/sbom/sbom.go:36-90,300` |
| V10 | Firma: `signs` (no `binary_signs`) con `artifacts: checksum`, `cmd`, `args`, `signature`, `certificate`; con `--skip=sign` no se invoca; el propio goreleaser firma con `cosign sign-blob --bundle=${signature} ${artifact} --yes` | `goreleaser:internal/pipe/sign/sign.go:38-124,179-211`; `pkg/config/config.go:1024-1044`; `goreleaser:.goreleaser.yaml:369-378` |
| V11 | Cask: `repository.{owner,name,token}`; la plantilla del token solo se evalúa al publicar; cubre darwin y linux; `hooks.post.install`; directorio `Casks` | `goreleaser:internal/pipe/cask/cask.go:54-98,190,240-265,468-480`; `internal/client/config.go:50`; `internal/pipe/publish/publish.go:85`; `pkg/config/config.go:218-285` |
| V12 | Scoop: mismo `RepoRef`, token solo al publicar, un archivo de Windows por arquitectura (amd64, arm64) | `goreleaser:internal/pipe/scoop/scoop.go:99-145,249`; `pkg/config/config.go:455-473` |
| V13 | nfpm: `formats` obligatorio en la práctica, `maintainer` vacío es obsoleto; deb y rpm se construyen en snapshot con binarios linux | `goreleaser:internal/pipe/nfpm/nfpm.go:38,46-48,70-80,94-96,238-242` |
| V14 | `release.extra_files` con `glob`; se adjunta solo al publicar; `changelog` con `use: git` y `groups` (`title`, `regexp`, `order`); el changelog se salta en snapshot | `goreleaser:pkg/config/config.go:676-724,1280-1295`; `internal/pipe/release/release.go:162-176`; `internal/pipe/changelog/changelog.go:45-57,167-220` |
| V15 | Tras un snapshot, `dist/metadata.json` (`version`, `tag`, `commit`…) y `dist/artifacts.json` (`name`, `path`, `goos`, `goarch`, `type`…); binarios en `dist/kitlegal_<os>_<arch>_<variante>/kitlegal` | `goreleaser:internal/pipeline/pipeline.go:87-88,175`; `internal/pipe/metadata/metadata.go:59-96`; `internal/artifact/artifact.go:170-294`; `internal/pipe/build/build.go:236-251`; `internal/builders/golang/targets.go:34-40` |
| V16 | Builds: `goarch` por omisión incluye 386; `ldflags` por omisión se sustituyen enteros al declararlos; `builds.gobinary` es lo único obsoleto | `goreleaser:internal/builders/golang/build.go:31-33,120-143`; `internal/pipe/build/build.go:93-96` |
| V17 | Ningún pipe exige `GITHUB_TOKEN` en snapshot; dos o más tokens de forja definidos a la vez son un error incluso en snapshot | `goreleaser:internal/pipe/env/env.go:87-100,127-137` |
| V18 | testscript: `exec` que falla da «unexpected command failure»; `exists` usa `os.Stat` (sigue enlaces); `symlink` no convierte el destino a absoluto; `stdout`/`stderr`/`grep` compilan con `(?m)`; una orden propia escribe su salida con `ts.Stdout()` y `cmp stdout` la ve; `HOME=/no-home` por omisión; no hay prefijo `?` | `testscript:cmd.go:257-276,292,461-472,690`; `testscript:testscript.go:968-1015`; `testscript:doc.go:59-65,98-100` |
| V19 | El rojo-primero rechaza como inválido un guion cuyo fallo contiene `unknown command`, `cannot parse`, `unexpected command` o `usage: `; la activación copia cada guion a `internal/app/testdata/script/` con el prefijo `h19-` | `scripts/workflow/aceptacion.sh` (subcomandos `rojo-primero` y `activar`) |
| V20 | `//go:embed`: patrones relativos al directorio del paquete, sin `..`; `dir/*` embebe también ficheros con punto; no casan enlaces simbólicos | `go doc embed` (toolchain go1.27.1) |
| V21 | `encoding/json/v2`: `RejectUnknownMembers`, `Deterministic` (mapas ordenados); `jsontext` rechaza nombres duplicados salvo `AllowDuplicateNames`; `WithIndent` | `go doc encoding/json/v2 RejectUnknownMembers`, `… Deterministic`; `go doc encoding/json/jsontext AllowDuplicateNames`, `… WithIndent`; ya lo importa `internal/skills/comandos.go` |
| V22 | `jsonschema:"nullable"` produce `oneOf` con `null` | `jsonschema:reflect.go:536-545,971-976` |
| V23 | Kong admite `short:"g"`; `cli.Analizar` convierte todo fallo de análisis en `ErrArgumentos` (exit 2) y detecta la ayuda aunque Kong devuelva error después | `kong:tag.go:283-285`; `internal/cli/parse.go:111-124` |
| V24 | `--dry-run`: el applet se ejecuta con la bandera en el contexto, el kernel escribe en la salida de error «--dry-run: …» y una línea «se habría pedido …» por entrada de `Resultado.Ensayo`, sin sobre; con error, sale su código | `internal/app/main.go:360-424` |
| V25 | El sobre de fallo lleva `{clase, mensaje}` con la clase de `cli.Clasificar` (sentinelas, después `schema.ConClase`) y el mismo mensaje va a la salida de error | `internal/cli/errors.go:90-112`; `internal/cli/sobre.go:100-168`; `internal/core/schema/error.go:61-80` |
| V26 | R1: `depguard` deniega a `internal/core/**` (también en tests) `os`, `io` —y por prefijo `io/fs`— y los paquetes internos; `TestArquitectura` sigue las importaciones del módulo y deniega `log`, `log/slog`, `os`, `io`, `net/http`, `database/sql` | `.golangci.yml:88-122`; `internal/arch_test.go:469-507,588-590`; H6 `research.md` V5-V7 |
| V27 | `TestEsquemasCubrenTodosLosVerbos` exige una parte publicada por cada verbo del registro de producción | `internal/app/esquemas_test.go:333-420` |
| V28 | `TestTablaDeComandosCoincideConLaGramatica` recorre todos los verbos del registro de producción y escribe hoy `scripts/boe …` | `internal/app/skills_test.go:165-210` |
| V29 | `argumentos.txtar` fija la lista literal `applets disponibles: boe, contar, echo, territorio` | `internal/app/testdata/script/argumentos.txtar:24,30` |
| V30 | El job de evals localiza el binario por `bin/instalado/kitlegal`; la sesión de prueba de red invoca `scripts/boe`; `evals.sh` comprueba `~/.claude/skills/<skill>/SKILL.md` con `-f`, que sigue el enlace | `.github/workflows/evals.yml:147`; `internal/evals/preparar.go:38-40`; `scripts/evals.sh:103` |
| V31 | El `Makefile` inyecta exactamente `main.version`, `main.commit`, `main.fecha` e `internal/httpx.version` | `Makefile:28-37` |
| V32 | En este macOS (darwin/arm64): `curl -fsSL file://…` lee ficheros locales y sale con 37 si falta; con `https_proxy=http://127.0.0.1:9`, `curl` sale con 7; `env -u HOME` deja `HOME` sin definir; `readlink` sin `-f` da el destino literal; `mkfifo`, `sha256sum`, `shasum`, `tar`, `mktemp` existen | `/tmp/h19-gorel/herramientas.sh`, ejecutado con `rtk proxy sh` |
| V33 | `go mod tidy` del módulo de herramienta de goreleaser, sin proxy, falla: falta `github.com/Azure/go-autorest@v14.2.0+incompatible` en la caché | `/tmp/h19-gorel/probar.sh` (`GOPROXY=off`, toolchain go1.27.1 de la caché) |
| V34 | `TestInstalacion` instala sobre una copia mínima del árbol que sale de `go list -deps` (con `EmbedFiles`), con `HOME`, `GOBIN` y `GOPATH` temporales y `GOPROXY=off` | `internal/skills/instalacion_test.go:39-140` |
| V35 | `internal/evals/trazas.go` llama a `app.RegistroDeProduccion()`; `TestDependenciasDelBinario` mide `./cmd/kitlegal` en las seis plataformas | `internal/evals/trazas.go:342`; `internal/arch_test.go:200-252` |
| V36 | Los flujos fijan las acciones por su etiqueta mayor (`actions/checkout@v7`, `actions/setup-go@v7`, `github/codeql-action@v4`) | `.github/workflows/ci.yml:27,36`; `.github/workflows/codeql.yml:44,63` |
| V37 | `/bin/` y `/dist/` están en `.gitignore` | `.gitignore:2-3` |
| V38 | `internal/skills` crea directorios con `0o750` y ficheros con `0o600`; rechaza en `references/` lo que no es una referencia declarada (`DerivaFicheroSobrante`) | `internal/skills/sincronia.go:33-36,100-104` |
| V39 | Dependabot tiene una entrada `gomod` por cada módulo de `tools/` | `.github/dependabot.yml` |
| V40 | La ayuda del binario escribe `uso: …` (no `usage: `) y una línea `  <applet>  <descripción>` por applet, alineada al más largo | `internal/app/ayuda.go:31-50,146-164` |
| V41 | `TestInstalacion` y sus guiones de `internal/skills/testdata/script/` comprueban la instalación por enlaces (`bin/instalado`, `scripts/boe`) | `internal/skills/testdata/script/instalar*.txtar` |
| V42 | Un applet calculado firma en el espacio reservado desde `internal/app`, que `TestLasFuentesNoFirmanComoKitlegal` no vigila | `internal/app/territorio.go:19-22`; `internal/arch_test.go:254-280` |
| V43 | El binario de e2e se construye con `go install` y una orden escrita entera con constantes | `internal/app/e2e_test.go:139-170` |
| V44 | El nombre de una skill: `a-z`, `0-9` y `-`, sin guion al principio, al final ni dos seguidos | `internal/skills/frontmatter.go:386-422` |
| V45 | `depguard` v2.2.1 (la del módulo de `golangci-lint` v2.13.2 fijado en `tools/golangci-lint`) compara `deny` por prefijo, salvo una entrada terminada en `$`, que casa solo con el paquete exacto; ordena la lista y la busca con el `$` recortado | `~/go/pkg/mod/github.com/!open!pee!dee!p/depguard/v2@v2.2.1/settings.go:113,224-247`; `…/README.md:61-66`; `tools/golangci-lint/go.mod:91` |
| V46 | `golangci-lint` v2.13.2 pasa cada `deny[].pkg` a `depguard` tal cual, sin normalizarlo | `~/go/pkg/mod/github.com/golangci/golangci-lint/v2@v2.13.2/pkg/golinters/depguard/depguard.go:33-36` |
| V47 | `TestArquitectura` construye el grafo con `go list -deps … ./...` (sin `-test`: ningún `_test.go` entra), R1 compara con `primerPrefijo`/`cuelgaDe` (igual o prefijo seguido de `/`), `paquetesInternos` son `app, cache, cli, graph, httpx, render, source, store`; los `_test.go` los cubre `depguard` con `run.tests: true` | `internal/arch_test.go:1-24,469-507,578-582,594-600,702-716`; `.golangci.yml:10,83-122` |
| V48 | En snapshot, el pipe de checksums resuelve `checksum.extra_files` con `extrafiles.Find`, que devuelve el error de `fileglob.Glob`; para un patrón sin comodines cuyo fichero no existe, `fileglob` v1.4.0 devuelve `matching "<ruta>": file does not exist`: `make release` falla si `scripts/install.sh` no existe | `goreleaser:internal/pipe/checksums/checksums.go:54-60,187-198`; `goreleaser:internal/extrafiles/extra_files.go:17-32`; `~/go/pkg/mod/github.com/goreleaser/fileglob@v1.4.0/glob.go:135-142`; `goreleaser:go.mod:33` |

## Supuestos no verificados

| # | Supuesto | Por qué no se puede comprobar ahora | Qué lo comprueba y cuándo |
|---|---|---|---|
| S1 | GitHub sirve `…/releases/latest/download/<asset>` (última release no preliminar) y `…/releases/download/<etiqueta>/<asset>` | Sin red; el repositorio es privado y sin releases | Trabajo de humo de `release.yml` y la instalación en un Mac limpio, tras la etiqueta (humano, FR-150) |
| S2 | Las versiones mayores vigentes de `actions/attest-build-provenance`, `sigstore/cosign-installer` y `anchore/sbom-action/download-syft` | Sin red | `release.yml` en la etiqueta; Dependabot las mantiene |
| S3 | `cosign sign-blob --bundle=<f> <artifact> --yes` firma sin clave con el `id-token` de Actions | Flags de una CLI que no está en local | `release.yml` en la etiqueta |
| S4 | `gh attestation verify <archivo> --repo jmorenobl/kitlegal` sale con 0 para un archivo atestado por `attest-build-provenance` | Sin red ni release | Trabajo de humo (FR-113) |
| S5 | Homebrew pone en cuarentena lo que instala un cask y admite casks en Linux | Sin Homebrew de pruebas ni red | La persona, en un Mac limpio y en Linux, tras la etiqueta (FR-150) |
| S6 | Scoop exige dar de alta un bucket que no es de los predeterminados antes de `scoop install kitlegal` | Sin Windows | README lo documenta; lo comprueba una persona en Windows |
| S7 | `actions/setup-go` pone `GOPATH/bin` en el `PATH` | Sin red | No se depende de ello: D19 lo hace explícito |
| S8 | En `ubuntu-24.04`, `curl` lee `file://`, y existen `sha256sum`, `mkfifo`, `env -u` y `readlink` con el mismo comportamiento que en macOS | Sin runner | El trabajo `ci` de la propuesta de cambio (los guiones e2e corren en `make ci`) |
| S9 | nfpm acepta la versión de snapshot (`0.0.0-SNAPSHOT-<corto>`) para `.deb` y `.rpm` | Sin trazar en el fuente de nfpm | `make release` en la tarea que lo introduce y el trabajo de snapshot |
| S10 | El proxy de módulos de Go sirve el grafo completo de goreleaser v2.18.1 | V33 | La tarea que crea `tools/goreleaser/`, como H0 con las otras herramientas |
| S11 | Codex y Antigravity leen `.agents/skills/` | Pendiente en `CLAUDE.md` | Ningún test depende de ello; se sigue sin `--host` para ellos (spec) |
| S12 | El secreto `PUBLISHER_TOKEN` existe en el repositorio | Secreto de la plataforma | Ninguna tarea lo necesita (FR-097); lo usa la primera publicación |
| S13 | En Windows sin permiso de enlaces, `os.Symlink` falla y el recurso de copia actúa | Sin CI en Windows (fuera de alcance) | El recurso se prueba con el enlazador que falla (D9, D24) |
| S14 | `misspell` no marca las palabras nuevas del dominio (`instalacion`, `manifiesto`, `hallazgos`…) o, si lo hace, se resuelve sin tocar `.golangci.yml` | Diccionario de la herramienta, sin ejecutar sobre código que aún no existe | `make lint` en cada tarea |
| S15 | El runner `ubuntu-latest` de GitHub es linux/amd64, de modo que el binario que `TestSnapshot` ejecuta en el trabajo de snapshot es el linux/amd64 que pide SC-016 | Sin red ni plataforma | El registro del trabajo de snapshot: `TestSnapshot` escribe con `t.Log` la plataforma cuyo binario ejecutó |
