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

1. **Cuándo se reinstala.** Reinstala una invocación de la CLI que cumpla tres condiciones:
   - su análisis termina bien y decide ejecutar;
   - no lleva `--dry-run` ni `--describe` (la ayuda no avisa, como hoy);
   - no es del applet `skills` ni del servidor MCP.

   La invocación multicall cuenta como CLI, y el código con que acabe el verbo no importa. Además tienen que cumplirse
   a la vez:
   - no hay manifiesto local, que tapa al global como hoy;
   - las versiones que compara hoy el aviso —la del manifiesto global y la de cada skill declarada que el binario
     empotra— son todas menores o iguales que la del binario en precedencia SemVer, y al menos una es menor. Los
     metadatos de construcción no ordenan, y una versión sin forma SemVer no reinstala;
   - `Planificar` sale sin conflictos, sin `--host`. El pedido son las skills declaradas que el binario empotra más las
     herederas de las declaradas con nombre retirado, y nada más. Un fichero que falta no es un conflicto (FR-044); uno
     editado o ajeno sí, y una retirada que no se puede retirar cuenta como tocada;
   - toma el candado. No hay prueba previa del permiso: si una escritura falla, sale la línea de hoy.
2. **Qué se dice.** Si reinstala, escribe en la salida de error una sola línea:
   `aviso: kitlegal <versión> ha actualizado las skills globales, que eran de <versión que nombraría la línea de hoy>; se aplican en la próxima conversación`.
   En cualquier otro caso escribe la línea de hoy, sin cambios.
3. **Lo que no cambia del contrato de H19.** La salida estándar, byte a byte, y el código de salida de la invocación
   son los mismos. No se abre red. Un fallo de la reinstalación nunca es un error de la invocación: se queda en la línea
   de hoy. Con `--dry-run`, si se cumplen las condiciones salvo las de escritura, la línea dice que las skills «se
   actualizarían» y no se escribe nada, salvo la sonda de enlace que `Planificar` crea y retira, como en
   `install --dry-run`.
4. **Lo que cambia.** Del contrato de H19, sus §3 y §4 de `contracts/aviso.md`, FR-071 (la línea lleva siempre la orden
   que ejecutar), FR-077 (no se ordenan versiones; deja de valer solo para decidir si se reinstala) y SC-013. El aviso
   puede escribir en `~/.agents/skills` y en los enlaces de sus hosts, solo en el caso del punto 1.
5. **Una vez.** No hace falta ninguna marca: cuando la reinstalación sale bien, la invocación siguiente ve la misma
   versión y no hace nada. En el servidor MCP no se reinstala nada, y el aviso sigue saliendo una vez por arranque.
6. **El candado.** Es `skills-global.lock`, en la ruta de `cache.Directorio()`, fuera de `~/.agents/skills`, así que no
   lo ven `install`, `list` ni `doctor`. Si el directorio no existe, se crea; si no se puede, no se reinstala. Se crea en exclusiva y se borra al terminar, salga bien o mal. Si existe
   y tiene menos de 10 minutos, la invocación da la línea de hoy sin esperar; si tiene más, se sustituye.

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
- Una skill nueva del producto, que no hereda de ninguna instalada, sigue pidiendo `kitlegal skills install -g`.
- Windows queda sin probar, porque la CI es solo Ubuntu.
