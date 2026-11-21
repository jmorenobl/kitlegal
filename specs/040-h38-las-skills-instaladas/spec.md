# Feature Specification: H38 · Las skills instaladas se actualizan con el binario (ADR 0060)

**Feature Branch**: `040-h38-las-skills-instaladas`

**Created**: 2026-11-21

**Status**: Draft

**Input**: el fichero del hito, `docs/roadmap/in-progress/H38-las-skills-instaladas-se-actualizan.md` (entrada de la
persona, ADR 0018), y el ADR 0060, que aplica.

## Contexto

Actualizar el binario con el gestor de paquetes (`brew upgrade`, `scoop update`, el instalador) deja hoy las skills
globales (`~/.agents/skills`) en la versión anterior. El aviso de versión de H19 lo dice en cada invocación, pero lo lee
el agente, no la persona, y no dice que la conversación ya cargó las skills anteriores. Evidencia de partida, tal como
la da el hito (no medida en esta sesión): el 2026-10-21, binario v0.8.0 y skills v0.7.0 (`docs/USO.md`); el 2026-11-21,
en otro equipo, binario v0.9.0 y skills v0.8.0, y el aviso salió en cada consulta.

Lo que este hito cambia del contrato de H19 (`specs/009-h19-instalar-sin-clonar/`): `contracts/aviso.md` §3 y §4,
FR 071, FR 077 (la regla de igualdad sigue valiendo para la línea de hoy y para `doctor`; el orden SemVer solo decide si
se reinstala) y SC 013. Lo demás de ese contrato sigue igual (ADR 0060, punto 4). Los requisitos de H19 se citan aquí
sin guion («FR 044 de H19») para no confundirlos con los de este spec.

Comprobado en esta sesión, leyendo el repositorio (base de la rama: `40413591`):

- la línea de hoy la compone `Aviso` (`internal/core/instalacion/aviso.go`): `aviso: las skills instaladas son de
  kitlegal <instalada> y este binario es kitlegal <binario>; ejecuta: kitlegal skills install`, con ` -g` si el
  manifiesto es el global; `<instalada>` sale de `versionQueDifiere`;
- el kernel la busca una sola vez por invocación, tras el análisis y antes de atender la decisión, termine como termine
  el análisis, salvo la ayuda y el applet `skills` (`buscaElAviso` y `resolverApplet`, `internal/app/main.go`);
- `Planificar` comprueba todos los conflictos antes de devolver el plan, y `Aplicar` se para en la primera escritura
  que falla dejando un estado que `install` completa (`internal/core/instalacion/plan.go` y `aplicar.go`; FR 044 de
  H19);
- los nombres retirados son hoy `boe-legislacion` → `normativa` y `legal-core` → `plazos-y-territorio`
  (`internal/app/retiradas.go`);
- `cache.Directorio()` resuelve `KITLEGAL_CACHE_DIR` o `~/.cache/kitlegal` sin crear nada (`internal/cache/ruta.go`);
- el applet `mcp` tiene un solo verbo, `serve` (`internal/app/mcp.go`);
- los únicos guiones de `internal/app/testdata/script/` que instalan con `$KITLEGAL_V1_BIN skills install -g` y esperan
  la línea de hoy con ` -g` son `h19-skills-aviso.txtar`, `h19-skills-aviso-sin-aviso.txtar` y
  `h19-skills-conflictos-dentro.txtar` (`grep -l`);
- ninguna `SKILL.md` de `skills/` lee, cita ni trata la línea del aviso (`grep -n 'skills install\|aviso'`).

## Clarifications

Ninguna en esta sesión: el fichero del hito y el ADR 0060 fijan las condiciones, las líneas, el candado, los controles y
la aceptación.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - El binario nuevo pone al día las skills globales en la primera consulta (Priority: P1)

Quien actualiza el binario con su gestor de paquetes y tiene las skills globales de la versión anterior, sin tocar,
hace su primera consulta (una orden de cualquier applet que no sea `skills`). Esa orden deja `~/.agents/skills` y sus
hosts como los dejaría `kitlegal skills install -g` del binario nuevo, y dice en la salida de error una sola línea que
lo cuenta y que avisa de que se aplica en la próxima conversación. La segunda consulta no dice nada.

**Why this priority**: es el objetivo del hito; sin ella, las skills se quedan viejas hasta que alguien ejecuta la orden.

**Independent Test**: e2e con los binarios v1 y v2 del arnés de H19: `V1 skills install -g`, una orden de v2, y se
comparan el manifiesto global, el árbol, la salida de error y la salida estándar.

**Acceptance Scenarios**:

1. **Dado** `~/.agents/skills` instalado con `kitlegal skills install -g` de la v1 y el binario v2 en el `PATH`,
   **Cuando** se hace la primera consulta con v2, **Entonces** `~/.agents/skills/kitlegal.json` queda en la v2, la
   salida de error es exactamente la línea `aviso: kitlegal v0.2.0 ha actualizado las skills globales, que eran de
   v0.1.0; se aplican en la próxima conversación` y su salto, y la salida estándar y el código de salida son, byte a
   byte, los de la misma orden sin skills instaladas.
2. **Dado** el estado que deja el escenario 1, **Cuando** se hace una segunda consulta con v2, **Entonces** la salida
   de error está vacía y nada escrito (FR-013).
3. **Dado** el estado que deja el escenario 1, **Cuando** se compara con el que deja `V2 skills install -g` directo
   sobre una instalación v1 igual, **Entonces** los dos árboles del global son iguales y `V2 skills doctor -g` no da
   ningún hallazgo.
4. **Dado** una instalación global v1 a la que le falta un fichero de una skill, **Cuando** se hace una consulta con
   v2, **Entonces** reinstala, como en el escenario 1.
5. **Dado** una instalación global en la que el manifiesto está en la v2 y una skill declarada sigue en la v1 (un
   `V2 skills install -g <skill>` de un subconjunto sobre una v1), **Cuando** se hace una consulta con v2,
   **Entonces** reinstala y la línea nombra `v0.1.0`, la versión que nombraría la línea de hoy.
6. **Dado** un manifiesto global de una versión anterior que declara una skill con nombre retirado e intacta,
   **Cuando** se hace una consulta con el binario nuevo, **Entonces** queda instalada su heredera y la retirada se va
   del disco y del manifiesto.
7. **Dado** una instalación global v1, **Cuando** se hace con v2 una orden que termina con exit 4, **Entonces**
   reinstala y el código sigue siendo 4.
8. **Dado** una instalación global v1, **Cuando** se invoca v2 por el nombre de un applet (multicall), **Entonces**
   reinstala igual que por el primer argumento.

---

### User Story 2 - Lo que la persona ha tocado no se pisa (Priority: P1)

Si algo de lo instalado no es lo que instaló kitlegal —un fichero editado, uno ajeno, una entrada ajena en un host, una
skill con nombre retirado que se ha editado y no se puede retirar—, la consulta no escribe nada (FR-013) y da la
línea de hoy, que sigue pidiendo `kitlegal skills install -g`.

**Why this priority**: es la condición que hace segura la reinstalación (ADR 0060, opción 3).

**Independent Test**: e2e con cada estado tocado: la salida de error es la línea de hoy y el árbol del global no
cambia.

**Acceptance Scenarios**:

1. **Dado** una instalación global v1 con un fichero editado, **Cuando** se hace una consulta con v2, **Entonces** la
   salida de error es exactamente la línea de hoy (`aviso: las skills instaladas son de kitlegal v0.1.0 y este binario
   es kitlegal v0.2.0; ejecuta: kitlegal skills install -g`), `kitlegal.json` sigue en la v1 y el árbol queda como
   estaba.
2. **Dado** una instalación global v1 con un fichero ajeno dentro de una skill, o una entrada ajena en un host, o una
   skill de nombre retirado editada («sin retirar»), **Cuando** se hace una consulta con v2, **Entonces** la línea de
   hoy y el árbol sin cambios.
3. **Dado** un binario más viejo que lo instalado (v1 sobre una instalación v2), **Cuando** se hace una consulta,
   **Entonces** la línea de hoy y nada escrito.
4. **Dado** una instalación y un binario cuyas versiones solo difieren en los metadatos de construcción (`+…`),
   **Cuando** se hace una consulta, **Entonces** la línea de hoy y nada escrito.

---

### User Story 3 - Solo la CLI reinstala, y solo en el ámbito global (Priority: P1)

El servidor MCP, las órdenes que solo describen o simulan, las que no se analizan bien, la ayuda, el applet `skills` y
el ámbito local se quedan como hoy; `--dry-run` cuenta lo que se haría.

**Why this priority**: el binario del plugin (`kitlegal mcp serve`) podría volver atrás lo que instaló otro gestor
(ADR 0060, opción 2), y una orden que no ejecuta no debe escribir.

**Independent Test**: e2e de cada invocación excluida sobre una instalación global v1 con v2.

**Acceptance Scenarios**:

1. **Dado** un manifiesto local viejo, **Cuando** se hace una consulta, **Entonces** la línea de hoy, sin ` -g`, y
   nada escrito.
2. **Dado** un manifiesto local al día y uno global viejo, **Cuando** se hace una consulta, **Entonces** ninguna línea
   y nada escrito.
3. **Dado** una instalación global v1, **Cuando** v2 arranca `kitlegal mcp serve`, **Entonces** solo da la línea de
   hoy, una vez por arranque, y nada escrito (FR-013).
4. **Dado** una instalación global v1, **Cuando** se invoca v2 con `--describe` o con un análisis que falla (exit 2),
   **Entonces** la línea de hoy y nada escrito; **cuando** se pide la ayuda, ninguna línea, como hoy.
5. **Dado** una instalación global v1 sin nada tocado, **Cuando** se invoca una orden de v2 con `--dry-run`,
   **Entonces** la salida de error es la de esa orden sin skills instaladas con una línea de aviso más, `aviso:
   kitlegal v0.2.0 se actualizarían las skills globales, que eran de v0.1.0; se aplican en la próxima conversación`, y
   nada escrito (FR-013).

---

### User Story 4 - El candado y los fallos no dejan nada roto (Priority: P2)

Dos invocaciones a la vez no reinstalan las dos; un candado que dejó una invocación interrumpida no bloquea para
siempre; y una escritura que falla deja lo que repara `kitlegal skills install -g`.

**Why this priority**: el agente puede lanzar varias órdenes a la vez, y una orden puede interrumpirse a mitad.

**Independent Test**: e2e con `$KITLEGAL_CACHE_DIR/skills-global.lock` creado a mano y con `touch -t`, sin directorio
de caché, con `~/.agents/skills` sin permiso de escritura y con dos consultas a la vez; unitario del fallo a mitad.

**Acceptance Scenarios**:

1. **Dado** `$KITLEGAL_CACHE_DIR/skills-global.lock` con una fecha de modificación de hace menos de 10 minutos,
   **Cuando** se hace una consulta con v2 sobre una instalación v1, **Entonces** la línea de hoy, sin esperar, y el
   árbol y el candado sin cambios.
2. **Dado** un candado huérfano de hace más de 10 minutos (`touch -t`), **Cuando** se hace la consulta, **Entonces**
   reinstala y el candado no queda.
3. **Dado** que `$KITLEGAL_CACHE_DIR` no existe todavía, **Cuando** se hace la consulta, **Entonces** se crea el
   directorio, reinstala y el candado no queda.
4. **Dado** `~/.agents/skills` y lo que contiene sin permiso de escritura y la caché escribible, **Cuando** se hace la
   consulta, **Entonces** la línea de hoy, el árbol sin cambios y el candado borrado; con los permisos devueltos, la
   consulta siguiente reinstala.
5. **Dado** una reinstalación que falla a mitad de sus escrituras, **Cuando** se ejecuta después `kitlegal skills
   install -g`, **Entonces** completa la instalación sin conflictos.
6. **Dado** una instalación global v1, **Cuando** se lanzan a la vez dos consultas de v2, **Entonces** como mucho una
   da la línea de FR-006; la otra da la línea de hoy o ninguna, y el árbol queda como en el escenario 1 de US1.

---

### User Story 5 - La documentación dice cómo se actualiza ahora (Priority: P3)

Quien lee `README.md` sabe que las skills globales se ponen al día solas en la primera consulta salvo que se hayan
tocado, y que las skills nuevas y las locales siguen pidiendo `kitlegal skills install`.

**Why this priority**: sin ella la persona sigue ejecutando a mano lo que ya no hace falta, o espera lo que no pasa.

**Independent Test**: lectura del párrafo «Actualizar» de `README.md` y de «Cambiado» en `CHANGELOG.md`.

**Acceptance Scenarios**:

1. **Dado** el `README.md` de la rama, **Cuando** se lee su párrafo «Actualizar», **Entonces** dice lo de FR-017.
2. **Dado** el `CHANGELOG.md` de la rama, **Cuando** se lee «Unreleased» > «Cambiado», **Entonces** hay una entrada
   con lo de FR-017, que declara Windows sin probar.

### Edge Cases

- **Binario de desarrollo** (versión sin forma SemVer): ni avisa ni reinstala, como hoy (FR 073 de H19).
- **`HOME` sin definir, vacío o `/`**: no hay global; nada cambia respecto de hoy.
- **Manifiesto local ilegible, o `./.agents` que no es un directorio real**: hoy no hay ningún efecto y el global no se
  mira; sigue igual, sin reinstalar (FR-003 a).
- **Una skill empotrada nueva que no hereda de ninguna declarada**: no se instala; tras la reinstalación, nada la
  nombra y sigue pidiendo `kitlegal skills install -g` según `README.md` (FR-005, FR-017).
- **Una skill declarada que el binario no empotra y no es un nombre retirado**: no se compara (FR 036 de H19) ni se
  toca.
- **`KITLEGAL_CACHE_DIR` presente y vacío, o que apunta a algo que no es un directorio**: `cache.Directorio()` falla;
  la línea de hoy y nada escrito (FR-010).
- **Lo que llega de fuera manipulado a mano**, fuera de los estados que enumera este spec, lo cubre la regla genérica,
  con una precisión: un fallo de la reinstalación nunca es un error de la invocación, así que sale la línea de hoy
  (FR-011, FR-012).

## Requirements *(mandatory)*

### Functional Requirements

Definiciones que usan los requisitos:

- **La línea de hoy**: la línea que, en el mismo estado del disco y con la misma invocación, da el aviso de H19 con
  `contracts/aviso.md` §1 a §4 tal como están en `main`, o ninguna si ese aviso no da ninguna.
- **Las versiones comparadas**: la `version` del manifiesto global y la de cada skill que declara y que el binario
  empotra, las mismas que compara hoy el aviso (`versionQueDifiere`).
- **La versión que nombra**: la que nombra la línea de hoy en ese estado (la del manifiesto si difiere; si no, la de la
  primera skill declarada y empotrada, por orden de nombre, que difiere).
- **El pedido**: las skills declaradas en el manifiesto global que el binario empotra, más las herederas
  (`Retiradas.Herederas`) de las declaradas con nombre retirado, y ninguna más.
- **El árbol del global**: `~/.agents/skills` entero y las entradas de skills de kitlegal en el directorio de skills de
  cada host del ámbito global, comparados por ruta, tipo, contenido y destino de enlace.
- **Una consulta**: una orden de un applet distinto de `skills`.
- **Una línea de aviso más**: la salida de error de la invocación es la de la misma orden sin skills instaladas con
  exactamente una línea de aviso más (la de FR-006, la de FR-009 o la de hoy), en el sitio de la línea de hoy, o
  ninguna más. Donde esa orden deja vacía la salida de error (con `--json` y sin error), es exactamente esa línea
  (`\A…\n\z`) o nada; con `--dry-run`, cuya descripción escribe el kernel, o con un error sin `--json`, no. Así se lee
  «la salida de error es exactamente la línea …» en este spec.

- **FR-001**: **Qué invocaciones pueden reinstalar.** Solo una invocación de la CLI —por el primer argumento o por el
  nombre de un applet (multicall)— que cumple las tres condiciones: (1) su análisis termina bien y decide ejecutar;
  (2) no lleva `--dry-run` ni `--describe`; (3) no es del applet `skills` ni del verbo `mcp serve` ni de sus llamadas.
  El código con que termine el verbo no importa: también reinstala la que sale con 4.
- **FR-002**: **Las invocaciones que no pueden reinstalar dan lo de hoy.** La ayuda (`kitlegal --help`, la de un applet
  y la de un verbo) y `kitlegal version` no dan ninguna línea; `kitlegal skills …` tampoco; `kitlegal mcp serve` da la
  línea de hoy una vez por arranque y sus llamadas ninguna; una invocación con `--describe` o cuyo análisis falla
  (exit 2) da la línea de hoy. En todas, nada escrito (FR-013). `--dry-run` lo rige
  FR-009.
- **FR-003**: **Cuándo reinstala.** Una invocación de FR-001 reinstala si y solo si se cumplen a la vez: (a) **no hay
  manifiesto local**: el aviso de hoy, con su regla de búsqueda (FR 070 de H19), examinaría el manifiesto global, y la
  versión del binario tiene forma SemVer; (b) **la regla de versiones** de FR-004; (c) **el plan sale sin conflictos**,
  con FR-005; (d) **toma el candado**, con FR-010. Si falla cualquiera, da la línea de hoy y nada escrito (FR-013).
- **FR-004**: **La regla de versiones.** Cada versión comparada es menor o igual que la del binario en precedencia
  SemVer 2.0.0, y al menos una es menor. Para ordenar, la `v` inicial no cuenta, una prerelease es anterior a su
  versión final (`0.2.0-rc.1` < `0.2.0`) y los metadatos de construcción no ordenan (`0.1.0+abc` y `0.1.0` tienen la
  misma precedencia, así que esa pareja no reinstala). Si una versión comparada o la del binario no tiene forma SemVer,
  no reinstala. Un binario más viejo que alguna versión comparada no reinstala.
- **FR-005**: **El plan.** Es el de `kitlegal skills install -g` sin `--host` con el pedido, sobre el ámbito global: la
  misma comprobación de conflictos y de huellas (`Planificar` y `ComprobarConflictos`), sin otra. Un fichero que falta
  no es un conflicto (FR 044 de H19); uno editado o ajeno sí, como una entrada ajena en un host. Una skill de nombre
  retirado que no se puede retirar porque se ha editado («sin retirar») cuenta como tocada: no reinstala. Los hosts son
  los de `install -g` sin `--host`: enlaza en todo host cuyo directorio existe en ese momento y quita del manifiesto el
  que ya no existe. Las skills empotradas que no están en el pedido no se instalan. Tras reinstalar, el árbol del global
  es el mismo que deja `kitlegal skills install -g <las skills del pedido>` del mismo binario sobre el mismo estado; con
  los binarios v1 y v2 del arnés de H19, cuyo pedido son todas las empotradas, el mismo que deja `V2 skills install -g`
  directo, y `V2 skills doctor -g` no da ningún hallazgo.
- **FR-006**: **La línea cuando reinstala.** Exactamente `aviso: kitlegal <versión del binario> ha actualizado las
  skills globales, que eran de <versión que nombra>; se aplican en la próxima conversación`, con las dos versiones tal
  como están (con su `v`, si la llevan), en la salida de error.
- **FR-007**: **La línea en cualquier otro caso.** La línea de hoy, sin cambiar un byte, o ninguna si hoy no hay
  ninguna.
- **FR-008**: **Una sola línea por invocación**, en la salida de error, nunca en la estándar: la de FR-006, la de
  FR-009 o la de hoy, nunca dos. Sale en el mismo punto de la invocación en que sale hoy la línea de hoy (tras el
  análisis y antes de lo que escriba el verbo), así que el orden de la salida de error respecto de los mensajes del
  verbo no cambia.
- **FR-009**: **`--dry-run`.** Una invocación que cumple FR-001 salvo por llevar `--dry-run` (sin `--describe`, con el
  análisis bien) da, si se cumplen FR-003 (a), (b) y (c), exactamente `aviso: kitlegal <versión del binario> se
  actualizarían las skills globales, que eran de <versión que nombra>; se aplican en la próxima conversación`; si no,
  la línea de hoy. En los dos casos no toma el candado, no prueba el permiso y nada escrito (FR-013); como
  `install --dry-run`, admite la sonda de enlace que la planificación crea y retira (`Enlazador.Disponible`), que no
  queda.
- **FR-010**: **El candado.** Es el fichero `skills-global.lock` del directorio que da `cache.Directorio()`
  (`KITLEGAL_CACHE_DIR` o `~/.cache/kitlegal`), y nunca nada en `~/.agents/skills`. Se intenta tomar solo cuando se
  cumplen FR-003 (a), (b) y (c), y se crea en exclusiva. Si el directorio no existe, se crea con los permisos con que la
  caché crea el suyo. Si `cache.Directorio()` falla, o no se pueden crear el directorio o el candado, no reinstala y da
  la línea de hoy. Si el candado existe y su fecha de modificación es de hace menos de 10 minutos, no reinstala, no
  espera y da la línea de hoy, sin tocarlo; si es de hace 10 minutos o más, se sustituye y la invocación sigue con
  FR-003. El candado que toma una invocación se borra al terminar su reinstalación, salga bien o mal. De dos o más
  invocaciones a la vez, como mucho una reinstala; las demás dan la línea de hoy o, si la miran cuando ya terminó,
  ninguna.
- **FR-011**: **Un fallo a mitad.** Si una escritura de la reinstalación falla (también por falta de permiso: no hay
  prueba previa), la invocación da la línea de hoy —la del estado anterior a intentarlo—, borra el candado y deja un
  estado que `kitlegal skills install -g` completa sin conflictos (FR 044 de H19). La invocación siguiente reinstala si
  se siguen cumpliendo FR-003, o da lo de hoy en ese estado.
- **FR-012**: **Lo que no cambia de ninguna invocación.** La salida estándar, byte a byte, y el código de salida son los
  mismos que los de la misma orden sin skills instaladas, reinstale o no y falle o no la reinstalación; un fallo de la
  reinstalación nunca es un error de la invocación. Ni la reinstalación ni su decisión abren red.
- **FR-013**: **Qué escribe la reinstalación, y solo cuando reinstala.** El árbol del global, el candado y el
  directorio de la caché que crea para tomarlo. En cualquier otro caso —sin reinstalación, con `--dry-run`, con un plan
  con conflictos, con el candado tomado— **nada escrito**: el árbol del global no cambia y el directorio de la caché
  queda como lo deja la misma orden sin skills instaladas (sin candado y sin un directorio que esa orden no crearía); lo
  que escribe la propia orden, como el grafo del mundo, no cuenta. La sonda de enlace que la planificación crea y
  retira no queda.
- **FR-014**: **El ámbito local** nunca se reinstala: un manifiesto local viejo da la línea de hoy, sin ` -g`; un local
  al día con un global viejo no da ninguna línea ni reinstala.
- **FR-015**: **El servidor MCP** nunca reinstala: `kitlegal mcp serve` da la línea de hoy, una vez por arranque, y la
  expectativa de `h21-mcp-protocolo.txtar` no cambia.
- **FR-016**: **Lo que el hito no toca**: `kitlegal skills install`, `list` y `doctor` (sus argumentos, salidas y
  códigos), la forma del manifiesto `kitlegal.json` y los hosts siguen igual; el candado no está en `~/.agents/skills`,
  así que no cambia lo que ven.
- **FR-017**: **`README.md` y `CHANGELOG.md`.** El párrafo «Actualizar» de `README.md` (el de `brew upgrade` y `scoop
  update`) dice que las skills globales se ponen al día solas en la primera consulta, salvo que se hayan tocado, y que
  las skills nuevas del producto y las locales siguen pidiendo `kitlegal skills install`, con `-g` o sin él.
  `CHANGELOG.md` lo dice en «Unreleased» > «Cambiado» y declara Windows sin probar (la CI es solo Ubuntu).
- **FR-018**: **Los e2e de H19 que cambian, y solo ellos.** `internal/app/testdata/script/h19-skills-aviso.txtar`,
  `h19-skills-aviso-sin-aviso.txtar` y `h19-skills-conflictos-dentro.txtar`, en sus casos globales, y sus copias de
  `specs/009-h19-instalar-sin-clonar/aceptacion/` pasan a esperar la reinstalación. En
  `h19-skills-aviso-sin-aviso.txtar`, los tres controles seguidos sobre el mismo `home-v1` (`vacio`, `agents-vacio` y
  `skills-vacio`) reinstalan la v1 entre uno y otro, porque tras la primera reinstalación los otros dos quedarían
  mudos. Las expectativas locales de esos guiones y la de `h21-mcp-protocolo.txtar` no cambian. Tocar
  `internal/app/testdata/` y las copias congeladas lo hace una tarea `[datos]` que las declara. Ningún otro guion ni
  otra suite congelada cambia.
- **FR-019**: **Unitarios**: del orden SemVer de FR-004 (prereleases, metadatos de construcción y versiones sin forma)
  y de un fallo a mitad de la escritura que repara `kitlegal skills install -g` (FR-011).

### Key Entities

- **Manifiesto global** (`~/.agents/skills/kitlegal.json`): la versión de nivel superior y las skills declaradas, con
  sus versiones, huellas y entradas de host; su forma no cambia (FR-016).
- **El pedido**: las skills que se reinstalan (definición de arriba).
- **El candado** (`<cache.Directorio()>/skills-global.lock`): fichero creado en exclusiva cuya fecha de modificación
  decide si está tomado (menos de 10 minutos) o huérfano.
- **Las dos líneas nuevas**: la de FR-006 y la de FR-009, de una línea cada una.

### Quién consume cada salida, cuánto y hasta cuándo

- **La línea de FR-006** la leen el agente y la persona en la salida de error de una consulta. Ninguna `SKILL.md` la
  pide ni la trata (comprobado: ver Contexto); la ve el agente en cualquiera de las órdenes que una skill lanza por
  pregunta (de una a varias `kitlegal boe articulo`, `cita`, `plazos`, `territorio`). Sale **una vez por actualización
  del binario**: en la primera consulta tras ella; en la siguiente las versiones coinciden y no sale nada (FR-003 b,
  escenario 2 de US1). Mide una línea de 120 bytes, con su salto, con versiones como `v0.10.1` y `v0.9.0`, y no depende del uso
  acumulado: cientos de normas y miles de bloques consultados no entran en ella, solo dos versiones. La reinstalación
  examina las skills del pedido (cuatro hoy) y sus ficheros, no la caché ni el grafo.
- **La línea de FR-009** la ve quien pide `--dry-run`, una vez por esa orden; deja de darse cuando la reinstalación se
  hace (en la primera consulta sin `--dry-run`) o cuando la persona toca lo instalado.
- **La línea de hoy** no cambia (FR-007): con algo tocado o con el candado tomado sigue saliendo en cada consulta, como
  en H19, hasta que la persona ejecuta `kitlegal skills install -g` tras resolver los conflictos o el candado caduca
  (10 minutos); ese contrato es de H19 y el hito lo deja igual.
- **El estado del disco** tras reinstalar lo leen `install`, `list` y `doctor` como si lo hubiera dejado `install -g`
  (FR-005, FR-016).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Con `kitlegal skills install -g` de la v1 y el binario v2 en el `PATH`, la primera consulta deja
  `~/.agents/skills/kitlegal.json` en la v2 y su salida de error es exactamente 1 línea, la de FR-006; la segunda
  consulta tiene 0 líneas en la salida de error. Lo hace cumplir el guion e2e de aceptación del hito, que ejecuta `make
  ci` (`TestEntregaDelHito`) y sale en rojo si la primera tiene otra cosa que esa línea (`\A…\n\z`) o la segunda
  alguna.
- **SC-002**: Con un fichero editado, uno ajeno, una entrada ajena en un host o una retirada «sin retirar», la salida de
  error es exactamente 1 línea, la de hoy, y el árbol del global no cambia en ningún byte. Lo hace cumplir el e2e de
  `make ci`, que sale en rojo si la línea es otra o si el árbol difiere del de antes de la consulta.
- **SC-003**: En el 100 % de los casos e2e del hito, la salida estándar y el código de salida son, byte a byte, los de
  la misma orden sin skills instaladas, y la salida de error es la de esa orden con una línea de aviso más o ninguna
  más (definición en «Functional Requirements»). Lo hace cumplir cada guion e2e del hito en `make ci`, que compara con
  esa orden la salida estándar, el código y la salida de error y sale en rojo ante la primera diferencia.
- **SC-004**: El candado de hace menos de 10 minutos da la línea de hoy y no reinstala; el de hace más de 10 minutos
  (`touch -t`) se sustituye y reinstala; sin directorio de caché, se crea y reinstala; tras cualquier reinstalación,
  acabe bien o mal, 0 candados en `$KITLEGAL_CACHE_DIR`; de dos invocaciones lanzadas a la vez sobre la misma
  instalación global v1, reinstala como mucho 1. Lo hace cumplir el e2e del candado en `make ci`, que sale en rojo si
  el reciente reinstala, si el huérfano no reinstala o si el candado queda; y una comprobación de `make ci` (dónde
  vive y cómo fuerza la concurrencia es del plan) que sale en rojo si las dos invocaciones dan la línea de FR-006.
- **SC-005**: La reinstalación y su decisión abren 0 conexiones de red. Lo hace cumplir `TestArquitectura` (subprueba
  sin red, `internal/arch_test.go`) en `make ci`, que sale en rojo si el código del aviso o de la reinstalación alcanza
  un paquete de red; y los e2e corren con los proxies cerrados (arnés de H19).
- **SC-006**: Los casos excluidos de FR-002, FR-014 y FR-015 (`mcp serve`, `--describe`, análisis fallido, ayuda,
  ámbito local) y los de FR-004 (binario más viejo, metadatos de construcción) no escriben nada (FR-013): 0 cambios en
  el árbol del global y 0 entradas en la caché que no deje la misma orden sin skills instaladas. Lo hacen cumplir sus
  e2e en `make ci`, que salen en rojo si el árbol cambia o si la caché tiene un candado u otra entrada que esa orden
  no deja.
- **SC-007**: `make ci` en verde, con los unitarios de FR-019 y con los únicos e2e de H19 cambiados que nombra FR-018.

## Criterios del hito (literales)

**Aceptación** (del fichero del hito): con `kitlegal skills install -g` de la v1 y el binario v2 en el `PATH`, la
primera consulta deja `~/.agents/skills/kitlegal.json` en la v2 y dice la línea nueva, y la segunda no dice nada. Con
un fichero editado, ni una cosa ni la otra, y el árbol queda como estaba. (US1.1, US1.2, US2.1; SC-001, SC-002.)

**Controles** (del fichero del hito):

- **e2e (`testscript`, con los binarios v1 y v2 de H19).** En cada uno se compara la salida estándar y el código de
  salida, byte a byte, con los de la misma orden sin skills instaladas, y se comprueba que la salida de error es
  exactamente una línea (`\A…\n\z`), o ninguna. Donde la misma orden sin skills ya escribe en la salida de error
  (`--dry-run`, un error sin `--json`), se lee como «una línea de aviso más» (definición en «Functional Requirements»;
  SC-003); «nada escrito» y «no escribe», como en FR-013 (SC-006):
  - **reinstala**: el árbol del global queda igual al de `V2 skills install -g` directo, y `V2 skills doctor -g` no da
    hallazgos. La segunda invocación no escribe nada en la salida de error (US1.1 a US1.3; FR-005);
  - **un fichero que falta**: reinstala (US1.4);
  - **un fichero editado, uno ajeno, una entrada ajena en un host o una retirada «sin retirar»**: la línea de hoy, y el
    árbol del global no cambia (US2.1, US2.2; SC-002);
  - **sin permiso de escritura en `~/.agents/skills`**: la línea de hoy, el árbol sin cambios y el candado borrado. Se
    hace con `chmod` sobre ese directorio, con la caché escribible, devolviendo los permisos al final, como
    `h19-instalador-rechazos.txtar` (US4.4; FR-011);
  - **un binario más viejo que lo instalado, o con metadatos de construcción**: la línea de hoy y nada escrito (US2.3,
    US2.4; FR-004);
  - **una skill declarada en una versión anterior a la del manifiesto**: reinstala, y la línea nombra la versión que
    nombraría la de hoy (US1.5);
  - **un manifiesto que declara un nombre retirado**: queda la heredera y la retirada se va (US1.6);
  - **el ámbito local**: un local viejo solo avisa, y un local al día con un global viejo ni avisa ni reinstala (US3.1,
    US3.2; FR-014);
  - **`mcp serve`**: solo avisa (US3.3; FR-015);
  - **`--describe` y un análisis fallido** (exit 2): la línea de hoy, sin escribir. **La ayuda**: ninguna línea, como
    hoy (US3.4; FR-002);
  - **una orden que termina con exit 4**: reinstala (US1.7);
  - **el candado**, en `$KITLEGAL_CACHE_DIR/skills-global.lock`, que el arnés fija dentro de `$WORK`: tomado y
    reciente da la línea de hoy; huérfano de más de 10 minutos (con `touch -t`), reinstala; sin directorio de caché
    todavía, lo crea y reinstala (US4.1 a US4.3; SC-004). De dos a la vez, como mucho una reinstala (US4.6; SC-004);
  - **`--dry-run`**: dice «se actualizarían» y no escribe (US3.5; FR-009).
- **Unitarios** del orden SemVer (prereleases, metadatos de construcción y versiones sin forma) y de un fallo a mitad
  de la escritura que repara `kitlegal skills install -g` (FR 044 de H19) (FR-019; US4.5).
- **Los e2e de H19 que cambian**, declarados aquí como hizo H41.4 con los suyos (FR-018):
  - `internal/app/testdata/script/h19-skills-aviso.txtar`, `h19-skills-aviso-sin-aviso.txtar` y
    `h19-skills-conflictos-dentro.txtar`, en sus casos globales;
  - sus copias de `specs/009-h19-instalar-sin-clonar/aceptacion/`.

  Pasan a esperar la reinstalación. En `h19-skills-aviso-sin-aviso.txtar`, los tres controles seguidos sobre el mismo
  `home-v1` (`vacio`, `agents-vacio` y `skills-vacio`) reinstalan la v1 entre uno y otro, porque tras la primera
  reinstalación los otros dos quedarían mudos. Las expectativas locales y la de `h21-mcp-protocolo.txtar` no cambian.
  Tocar `internal/app/testdata/` y las copias congeladas lo hace una tarea `[datos]` que las declara.

## Assumptions

- **El `chmod` del caso sin permiso** quita el permiso de escritura a `~/.agents/skills` y a lo que contiene (como el
  `chmod -R u+w` con que `h19-instalador-rechazos.txtar` los devuelve): con solo el directorio de arriba sin permiso,
  una retirada dentro de un subdirectorio escribible podría salir bien antes del primer fallo, y el control pide el
  árbol sin cambios.
- **Los binarios v1 y v2 del arnés** (`$KITLEGAL_V1_BIN`, `v0.1.0`; `$KITLEGAL_V2_BIN`, `v0.2.0`;
  `specs/009-h19-instalar-sin-clonar/contracts/arnes-e2e.md`) empotran las mismas skills, así que el pedido de una
  instalación v1 son todas las empotradas de v2 (FR-005). Los estados con un nombre retirado, con metadatos de
  construcción y con una skill en otra versión que el manifiesto se construyen en el plan con lo que da el arnés.
- **«La primera consulta»** es la primera invocación de FR-001 tras la actualización; las invocaciones de FR-002 antes
  de ella no cuentan ni cambian nada.
- **Los documentos de `specs/009-h19-instalar-sin-clonar/`** (salvo las copias congeladas de FR-018) no se reescriben:
  lo que cambia de su contrato lo dicen este spec y el ADR 0060 (ADR 0044, punto 7, según el ADR 0060).
- **El estado del ADR 0060** («propuesta») no lo cambia el run: queda aceptado cuando una persona fusiona.

## Fuera de alcance

- Reinstalar en el ámbito local o desde el servidor MCP.
- Volver a una versión anterior.
- Reinstalar con conflictos o forzar, aunque sea un solo fichero.
- Instalar skills que no se habían pedido (las empotradas nuevas que no heredan de ninguna declarada).
- Abrir red.
- Cambiar la salida estándar o el código de salida de cualquier invocación.
- Cambiar `kitlegal skills install`, `list`, `doctor`, el manifiesto o los hosts.
- Recargar las skills de la conversación en curso.
- Cualquier cambio en el plugin, la extensión o los instaladores.
- Probar en Windows: la CI es solo Ubuntu, y el `CHANGELOG.md` lo declara sin probar.
- Cambiar el texto o las condiciones de la línea de hoy, o que deje de repetirse en cada consulta cuando algo está
  tocado (contrato de H19).
- Esperar a que se libere un candado tomado, o reintentar dentro de la misma invocación.
- Reescribir `spec.md`, `contracts/aviso.md` u otros documentos de `specs/009-h19-instalar-sin-clonar/` distintos de
  sus copias congeladas de `aceptacion/`, o cambiar el estado del ADR 0060 o el texto del ADR 0019.
- Cambiar ninguna `SKILL.md`, `data/`, evals o la web.
