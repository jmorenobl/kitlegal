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
- **Decisiones que esta entrada aplica**: el ADR 0060, que va en la misma propuesta y cambia el contrato del aviso de H19
  (`specs/009-h19-instalar-sin-clonar/contracts/aviso.md`, §5; FR-070 a FR-077).
- **Lo comprobado antes de escribirla**, leyendo el código, sin probarlo con la app:
  - el binario del plugin y de la extensión corre como `kitlegal mcp serve`, tiene versión SemVer y miraría
    `~/.agents/skills/kitlegal.json`, que la app no lee (ADR 0035). Por eso `mcp serve` no reinstala;
  - la comprobación de huellas ya existe: es la de `kitlegal skills install`, `ComprobarConflictos` en
    `internal/core/instalacion/plan.go`, y el hito la reutiliza, sin escribir otra.
- **Entrega**:
  - **La reinstalación global** (ADR 0060, punto 1). Una orden de la CLI distinta de `skills`, de la ayuda y de
    `mcp serve` reinstala lo que declara el manifiesto global, en los hosts de siempre, cuando se cumplen a la vez estas
    condiciones:
    - no hay manifiesto local;
    - el global es de una versión anterior en orden SemVer;
    - `kitlegal skills install -g` saldría sin conflictos, con su misma comprobación y su mismo plan;
    - el binario puede escribir y toma el candado.
  - **Las dos líneas** (ADR 0060, punto 2).
    - Si reinstala: `aviso: kitlegal <versión> ha actualizado las skills globales, que eran de <versión anterior>; se aplican en la próxima conversación`.
    - En cualquier otro caso, la de hoy, sin cambiar un byte.
    - Con `--dry-run`, la primera dice «se actualizarían» y no escribe nada.
  - **El candado** (ADR 0060, punto 6): un fichero creado en exclusiva junto al manifiesto global. El spec fija desde
    cuándo se ignora uno huérfano.
  - **El orden SemVer**, solo para decidir si se reinstala. Una versión que no se puede ordenar no reinstala.
  - **`README.md`**: el párrafo «Actualizar» dice que las skills globales se ponen al día solas en la primera consulta,
    salvo que se hayan tocado; las locales siguen pidiendo `kitlegal skills install`. **`CHANGELOG.md`** lo dice en
    «Cambiado».
- **Fuera**:
  - reinstalar en el ámbito local o desde `mcp serve`;
  - volver a una versión anterior;
  - reinstalar con conflictos o forzar, aunque sea un solo fichero;
  - abrir red;
  - cambiar la salida estándar o el código de salida de cualquier orden;
  - cambiar `kitlegal skills install`, `doctor`, el manifiesto o los hosts;
  - recargar las skills de la conversación en curso;
  - cualquier cambio en el plugin, la extensión o los instaladores.
- **Controles**:
  - **e2e (`testscript`, con los binarios v1 y v2 de H19)** de las salidas:
    - reinstala;
    - un fichero editado, que da el aviso de hoy y no toca nada;
    - sin permiso de escritura, que da el aviso de hoy;
    - un binario más viejo que lo instalado, que da el aviso de hoy y no vuelve atrás;
    - el ámbito local, que solo avisa;
    - `mcp serve`, que solo avisa;
    - el candado tomado;
    - `--dry-run`, que no escribe.

    En cada una, la salida estándar y el código de salida, byte a byte iguales a los de la misma orden sin
    skills instaladas.
  - **Unitarios del orden SemVer** (prereleases y versiones sin forma) y de que un fallo a mitad de la escritura deja
    lo que repara `kitlegal skills install -g` (FR-044).
  - **Los e2e de H19 que cambian**, declarados aquí, como hizo H41.4 con los suyos. Son
    `internal/app/testdata/script/h19-skills-aviso.txtar` (global), `h19-skills-aviso-sin-aviso.txtar` (global) y
    `h19-skills-conflictos-dentro.txtar` (global), con sus copias de `specs/009-h19-instalar-sin-clonar/aceptacion/`:
    pasan a esperar la reinstalación. Las expectativas locales y la de `h21-mcp-protocolo.txtar` no cambian. Tocar
    `internal/app/testdata/` lo hace una tarea `[datos]`.
- **Aceptación**: con `kitlegal skills install -g` de la v1 y el binario v2 en el `PATH`, la primera consulta deja
  `~/.agents/skills/kitlegal.json` en la v2 y dice la línea nueva, y la segunda no dice nada. Con un fichero editado, ni
  una cosa ni la otra.
- **Dependencias**: ninguna.
