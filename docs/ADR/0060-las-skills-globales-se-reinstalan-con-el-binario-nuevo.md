# 0060 · Las skills globales se reinstalan con el binario nuevo, si nadie las ha tocado

- **Estado**: propuesta (2026-11-21); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-21
- **Hito**: H38. Se escribe con su entrada, antes del run y fuera de él, en una sesión interactiva. Cambia el contrato
  del aviso de versión de H19 (`specs/009-h19-instalar-sin-clonar/contracts/aviso.md`, §5; FR-070 a FR-077), como pide
  el ADR 0044, punto 7. Precisa el ADR 0019 en lo que dice del aviso.

## Contexto y problema

El 2026-10-21, `brew upgrade` dejó el binario en la v0.8.0 y `~/.agents/skills` en la v0.7.0 (`docs/USO.md`). El
2026-11-21 pasó lo mismo con otro equipo: binario v0.9.0 y skills v0.8.0. El aviso de H19 funcionó las dos veces,
pero lo vio el agente y no la persona. Tampoco dice que la conversación ya cargó las skills viejas.

El aviso es hoy de solo lectura (`internal/core/instalacion/aviso.go`). Su contrato dice que nunca escribe en la salida
estándar, nunca cambia el código de salida ni la salida estándar y nunca abre red. No dice que no escriba en disco,
pero hoy no lo hace, y el ADR 0019 lo da por supuesto. La comparación que haría falta para reinstalar sin riesgo ya
existe: `kitlegal skills install` comprueba todos los conflictos antes de escribir (`ComprobarConflictos`), con las
huellas de cada fichero del manifiesto `kitlegal.json`.

Hay tres hechos que limitan qué se puede hacer:

- **Dos binarios pueden ver el mismo `~/.agents/skills`.** El que corre dentro del plugin y de la extensión de la app
  (`kitlegal mcp serve`) tiene versión SemVer, mira `~/.agents/skills/kitlegal.json` y puede ser más viejo que el de
  Homebrew. La app no lee ese directorio (ADR 0035).
- **El aviso no ordena versiones** (FR-077): avisa ante cualquier diferencia.
- **Dentro del sandbox de Claude Code**, una orden lanzada por el agente no puede escribir en `~/.agents`.

## Opciones consideradas

1. **Dejar el aviso como está y mejorar su texto.** Rechazada: el aviso ya funciona, y lo que falla es que nadie
   ejecute la orden.
2. **Reinstalar ante cualquier diferencia de versión, desde cualquier binario.** Rechazada: el binario del plugin
   volvería atrás lo que instaló Homebrew, y los dos se pisarían.
3. **Reinstalar solo en el ámbito global, solo desde la CLI, solo hacia una versión más nueva y solo si nadie ha tocado
   lo instalado. En todos los demás casos, el aviso de hoy.** Elegida.

## Decisión

1. **Cuándo se reinstala.** Una orden de la CLI distinta de `skills`, de la ayuda y de `mcp serve` reinstala las skills
   globales cuando se cumplen a la vez estas condiciones:
   - no hay manifiesto local, que hoy tapa al global y lo sigue tapando;
   - el manifiesto global es de una versión anterior a la del binario, en orden SemVer;
   - `kitlegal skills install -g` saldría sin conflictos, con su misma comprobación: ningún fichero editado ni ajeno.
     Un fichero que falta no es un conflicto, como en FR-044;
   - el binario puede escribir y toma el candado de la reinstalación.

   Reinstala lo mismo que haría esa orden: las skills que declara el manifiesto, no todas las empotradas, y en los
   hosts de siempre. Nunca toca un ámbito local.
2. **Qué se dice.** Si reinstala, escribe una línea en la salida de error:
   `aviso: kitlegal <versión> ha actualizado las skills globales, que eran de <versión anterior>; se aplican en la próxima conversación`.
   En cualquier otro caso escribe la línea de hoy, sin cambios: si el binario es más viejo o la versión no se puede
   ordenar, si hay conflictos, si no puede escribir, si el candado está tomado, si la escritura falla, y siempre en el
   ámbito local.
3. **Lo que no cambia del contrato de H19.** La salida estándar, byte a byte, y el código de salida de la orden son los
   mismos. No se abre red. Un fallo de la reinstalación nunca es un error de la orden: se queda en el aviso de hoy.
   `--dry-run` y `--describe` no escriben nada. Con `--dry-run`, la línea dice que las skills «se actualizarían», con
   las mismas versiones.
4. **Lo que cambia.** El aviso puede escribir en `~/.agents/skills` y en los enlaces de sus hosts, solo en el caso del
   punto 1. Deja de valer, solo para esa decisión, que el aviso no ordena versiones (FR-077). La línea de hoy se sigue
   dando ante cualquier diferencia.
5. **Una vez.** No hace falta ninguna marca: cuando la reinstalación sale bien, la orden siguiente ve la misma versión y
   no hace nada. En el servidor MCP no se reinstala nada; el aviso sigue siendo una vez por arranque.
6. **El candado.** Es un fichero que se crea en exclusiva junto al manifiesto global y se borra al terminar. Si existe,
   el binario no espera y da el aviso de hoy. Si se queda huérfano porque el proceso murió, el spec fija desde cuándo se
   ignora.

## Consecuencias

**A favor**

- `brew upgrade`, `scoop update` o el instalador dejan las skills globales al día en la primera consulta, sin que nadie
  ejecute nada. La línea lo dice a la persona y al agente.
- Lo que la persona ha tocado no se pisa nunca: se queda en el aviso de hoy.
- El binario del plugin no vuelve atrás lo que instaló otro gestor.

**En contra, y asumido**

- La conversación en curso sigue con las skills viejas; la línea lo dice.
- Dentro del sandbox de Claude Code no se reinstala, y la persona sigue dependiendo del aviso.
- Una orden de solo lectura escribe en disco en un caso. Lo acota el punto 1 y lo prueban sus e2e.
- Cambian tres e2e globales de H19 y sus copias congeladas: dejan de esperar el aviso y esperan la reinstalación.
