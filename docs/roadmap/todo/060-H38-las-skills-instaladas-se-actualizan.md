---
id: H38
fase: "1"
---

# H38 · Las skills instaladas se actualizan con el binario (ADR 0060)

- **Objetivo**: que actualizar el binario con el gestor de paquetes no deje las skills globales en la versión anterior.
  El aviso de H19 lo ve el agente, no la persona, y no dice que la conversación ya cargó las skills anteriores.
- **Evidencia de partida**:
  - el 2026-10-21, `brew upgrade` dejó el binario en la v0.8.0 y `~/.agents/skills` en la v0.7.0 (`docs/USO.md`);
  - el 2026-11-21, en otro equipo, el binario era de la v0.9.0 y las skills de la v0.8.0, y el aviso salió en cada
    consulta.
- **Decisiones que esta entrada aplica**: el ADR 0060, que va en la misma propuesta. Cambia, del contrato del aviso de
  H19 (`specs/009-h19-instalar-sin-clonar/`), sus §3 y §4 de `contracts/aviso.md`, FR-071, FR-077 y SC-013. Lo demás
  del contrato sigue igual.
- **Lo comprobado antes de escribirla**, leyendo el código, sin probarlo con la app:
  - el binario del plugin y de la extensión corre como `kitlegal mcp serve`, tiene versión SemVer y miraría
    `~/.agents/skills/kitlegal.json`, que la app no lee (ADR 0035). Por eso el servidor no reinstala;
  - la comprobación de huellas ya existe: es la de `kitlegal skills install`, con `Planificar` y `ComprobarConflictos`
    en `internal/core/instalacion/plan.go`. El hito la reutiliza, sin escribir otra;
  - `Planificar` retira las skills de `skillsRetiradas` (`internal/app/retiradas.go`), así que el pedido tiene que
    llevar sus herederas;
  - `cli.Preliminar` no sabe si se pidió `--dry-run`, así que cuando el análisis falla no se puede reinstalar.
- **Entrega**:
  - **Qué invocaciones reinstalan** (ADR 0060, punto 1). Solo una invocación de la CLI que cumpla las tres condiciones:
    - que su análisis termine bien y decida ejecutar;
    - que no lleve `--dry-run`, `--describe` ni la ayuda;
    - que no sea del applet `skills` ni del verbo `mcp serve` y sus llamadas. Las llamadas ya usan
      `registro.sinAvisador()`, así que no avisan.

    La invocación multicall por el nombre de un applet cuenta como CLI. El código con que acabe el verbo no importa:
    también reinstala la que sale con 4.
  - **Cuándo reinstala** (ADR 0060, punto 1). Tienen que cumplirse a la vez:
    - **no hay manifiesto local**, que tapa al global como hoy;
    - **la regla de versiones**: hoy el aviso compara la versión del manifiesto global y la de cada skill declarada que el
      binario empotra (`versionQueDifiere`). Todas tienen que ser menores o iguales que la del binario en precedencia
      SemVer, y al menos una menor. La `v` inicial no cuenta y una prerelease es anterior a su versión final. Los
      metadatos de construcción no ordenan: `0.1.0+abc` frente a `0.1.0` no reinstala. Una versión sin forma SemVer
      tampoco reinstala;
    - **el plan sale sin conflictos**: `Planificar`, sin `--host`, con un pedido que son las skills declaradas que el
      binario empotra más las herederas (`Retiradas.Herederas`) de las declaradas con nombre retirado. Las empotradas
      nuevas que no heredan de ninguna no se instalan. Un fichero que falta no es un conflicto (FR-044); uno editado o
      ajeno sí lo es. Una skill retirada que no se puede retirar porque se ha editado («sin retirar») cuenta como tocada:
      no reinstala. Los hosts son los de `Planificar`: enlaza en todo host cuyo directorio existe ahora y quita del
      manifiesto el que ya no existe. Una entrada ajena en un host es conflicto;
    - **se puede escribir**: el binario puede escribir y toma el candado.
  - **Las dos líneas** (ADR 0060, punto 2), en la salida de error y una sola por invocación.
    - Si reinstala: `aviso: kitlegal <versión del binario> ha actualizado las skills globales, que eran de <versión que nombraría la línea de hoy>; se aplican en la próxima conversación`.
    - En cualquier otro caso, la línea de hoy, sin cambiar un byte.
    - Con `--dry-run`, si se cumplen las tres primeras condiciones, la misma línea con «se actualizarían» en lugar de
      «ha actualizado». No toma el candado, no prueba el permiso y no escribe nada. Como `install --dry-run`, admite la
      sonda de enlace que `Planificar` crea y retira (`Enlazador.Disponible`).
  - **El candado** (ADR 0060, punto 6). Es `skills-global.lock`, en el directorio de caché de kitlegal
    (`internal/cache`), creado en exclusiva y borrado al terminar, salga bien o mal. Si existe y su fecha de
    modificación tiene menos de 10 minutos, no se reinstala y sale la línea de hoy. Si tiene más, se sustituye. No va en
    `~/.agents/skills`, así que no cambia lo que ven `install`, `list` ni `doctor`.
  - **Un fallo a mitad** deja lo que repara `kitlegal skills install -g` (FR-044). El candado se borra, y la invocación
    siguiente reintenta si se siguen cumpliendo las condiciones, o da la línea de hoy.
  - **`README.md` y `CHANGELOG.md`.** El párrafo «Actualizar» de `README.md` dice que las skills globales se ponen al
    día solas en la primera consulta, salvo que se hayan tocado. Las skills nuevas del producto y las locales siguen
    pidiendo `kitlegal skills install`, con `-g` o sin él. `CHANGELOG.md` lo dice en «Cambiado».
- **Fuera**:
  - reinstalar en el ámbito local o desde el servidor MCP;
  - volver a una versión anterior;
  - reinstalar con conflictos o forzar, aunque sea un solo fichero;
  - instalar skills que no se habían pedido;
  - abrir red;
  - cambiar la salida estándar o el código de salida de cualquier invocación;
  - cambiar `kitlegal skills install`, `list`, `doctor`, el manifiesto o los hosts;
  - recargar las skills de la conversación en curso;
  - cualquier cambio en el plugin, la extensión o los instaladores.
  - probar en Windows: la CI es solo Ubuntu, y el CHANGELOG lo declara sin probar.
- **Controles**:
  - **e2e (`testscript`, con los binarios v1 y v2 de H19).** En cada uno se compara la salida estándar y el código de
    salida, byte a byte, con los de la misma orden sin skills instaladas, y se comprueba que la salida de error es
    exactamente una línea (`\A…\n\z`), o ninguna:
    - **reinstala**: el árbol del global queda igual al de `V2 skills install -g` directo, y `V2 skills doctor -g` no da
      hallazgos. La segunda invocación no escribe nada en la salida de error;
    - **un fichero que falta**: reinstala;
    - **un fichero editado, uno ajeno, una entrada ajena en un host o una retirada «sin retirar»**: la línea de hoy, y
      el árbol del global no cambia;
    - **sin permiso de escritura**: la línea de hoy y el árbol sin cambios. Se hace con `chmod`, devolviendo los
      permisos al final, como `h19-instalador-rechazos.txtar`;
    - **un binario más viejo que lo instalado, o con metadatos de construcción**: la línea de hoy y nada escrito;
    - **una skill declarada en una versión anterior a la del manifiesto**: reinstala, y la línea nombra la versión que
      nombraría la de hoy;
    - **un manifiesto que declara un nombre retirado**: queda la heredera y la retirada se va;
    - **el ámbito local**: un local viejo solo avisa, y un local al día con un global viejo ni avisa ni reinstala;
    - **`mcp serve`**: solo avisa;
    - **`--describe`, la ayuda y un análisis fallido** (exit 2): la línea de hoy, sin escribir;
    - **una orden que termina con exit 4**: reinstala;
    - **el candado**: tomado y reciente da la línea de hoy; huérfano de más de 10 minutos, reinstala;
    - **`--dry-run`**: dice «se actualizarían» y no escribe.
  - **Unitarios** del orden SemVer (prereleases, metadatos de construcción y versiones sin forma) y de un fallo a mitad
    de la escritura que repara `kitlegal skills install -g` (FR-044).
  - **Los e2e de H19 que cambian**, declarados aquí como hizo H41.4 con los suyos:
    - `internal/app/testdata/script/h19-skills-aviso.txtar`, `h19-skills-aviso-sin-aviso.txtar` y
      `h19-skills-conflictos-dentro.txtar`, en sus casos globales;
    - sus copias de `specs/009-h19-instalar-sin-clonar/aceptacion/`.

    Pasan a esperar la reinstalación. En `h19-skills-aviso-sin-aviso.txtar`, los tres controles seguidos sobre el mismo
    `home-v1` (`vacio`, `agents-vacio` y `skills-vacio`) reinstalan la v1 entre uno y otro, porque tras la primera
    reinstalación los otros dos quedarían mudos. Las expectativas locales y la de `h21-mcp-protocolo.txtar` no cambian.
    Tocar `internal/app/testdata/` y las copias congeladas lo hace una tarea `[datos]` que las declara.
- **Aceptación**: con `kitlegal skills install -g` de la v1 y el binario v2 en el `PATH`, la primera consulta deja
  `~/.agents/skills/kitlegal.json` en la v2 y dice la línea nueva, y la segunda no dice nada. Con un fichero editado,
  ni una cosa ni la otra, y el árbol queda como estaba.
- **Dependencias**: ninguna.
