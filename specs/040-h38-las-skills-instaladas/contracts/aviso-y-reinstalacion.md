# Contrato: el aviso que reinstala las skills globales (H38; ADR 0060)

Sustituye, para lo que dice, a §3 y §4 de `specs/009-h19-instalar-sin-clonar/contracts/aviso.md`; §1, §2 y §5 de aquel
siguen valiendo (dónde se busca, qué se compara y que nunca escribe en la salida estándar). «La línea de hoy» es la
de aquel contrato tal como está en `main` (spec, «Definiciones»).

## §1 · El permiso de cada invocación

El kernel lo calcula en `resolverApplet` (`internal/app/main.go`), en el punto en que hoy busca el aviso: después de
`cli.Analizar` y antes de atender la decisión, termine como termine el análisis (FR-008).

1. Si la invocación no busca el aviso (`buscaElAviso` falso: toda ayuda y el applet `skills`), no se llama al avisador.
   Tampoco en `version` ni en lo que falla antes de resolver un applet con su verbo, que no llegan aquí, ni en las
   llamadas de herramienta, que se resuelven con el registro sin avisador.
2. Si el análisis falló: `SoloAvisar`.
3. Si la decisión es describir (`--describe`): `SoloAvisar`.
4. Si los argumentos del verbo analizado son un `servidor` (`mcp serve`): `SoloAvisar`, con o sin `--dry-run`.
5. Si la decisión es ejecutar con `--dry-run`: `Ensayar`.
6. En cualquier otro caso: `Reinstalar`. El código con que termine después el verbo no cuenta (FR-001).

Por el primer argumento o por el nombre del applet (multicall) es la misma: las dos llegan a `resolverApplet`.

## §2 · La decisión

`Reinstalacion.Avisar(permiso)` (`internal/core/instalacion`), en este orden; el primer paso que decide termina:

1. **Hoy.** `linea, hay := Aviso(Disco, Home, Version, Empotradas)`. Sin línea de hoy, ninguna línea y nada más (el
   binario de desarrollo, sin manifiesto, versiones iguales, `HOME` inservible, local ilegible: como en H19).
2. **Solo avisar.** Con `SoloAvisar`, la línea de hoy.
3. **Ámbito (FR-003 a, FR-014).** Si el manifiesto que compara el aviso es el local, la línea de hoy.
4. **Versiones (FR-004).** Si la del binario o alguna comparada (la del manifiesto global y la de cada skill que
   declara y el binario empotra) no tiene forma SemVer, o alguna comparada es mayor que la del binario en precedencia
   SemVer 2.0.0, o ninguna es menor, la línea de hoy.
5. **Plan (FR-003 c, FR-005).** `Planificar(Disco, Enlazador, pedido, Empotradas, Retiradas, Version)` con el pedido
   de data-model §4. Si devuelve un error —un conflicto o un fallo del disco
   o de la sonda— o alguna skill del plan sale «sin retirar», la línea de hoy.
6. **Ensayo (FR-009).** Con `Ensayar`, la línea de FR-009 con la versión que nombra la de hoy. No toma el candado ni
   escribe; la sonda de enlace de `Planificar` se crea y se retira, como en `install --dry-run`.
7. **Candado (FR-010).** `liberar, tomado := Candado()`. Sin tomarlo, la línea de hoy. Tomado, `liberar` se llama al
   terminar este paso y los dos siguientes, pase lo que pase.
8. **Otra vez, bajo el candado (research.md D3).** Los pasos 1, 3, 4 y 5 sobre el disco de ahora. Sin línea de hoy,
   ninguna línea (otra invocación ya reinstaló). Si alguno no se cumple, la línea de hoy de ahora.
9. **Aplicar (FR-011).** `Aplicar(plan, Escritor)` con el plan del paso 8. Con un error, la línea de hoy del paso 8
   (la del estado anterior a intentarlo). Sin error, la línea de FR-006 con la versión que nombra la de hoy del paso 8.

Nada de esto abre red ni devuelve un error: lo que falla se queda en la línea de hoy (FR-012; research.md D11).

## §3 · Las líneas, con sus bytes

Una sola por invocación, en la salida de error, con su salto (FR-008), y nunca en la salida estándar. Ejemplos con los
binarios del arnés de H19 (`v0.2.0` sobre una instalación `v0.1.0`):

```text
aviso: kitlegal v0.2.0 ha actualizado las skills globales, que eran de v0.1.0; se aplican en la próxima conversación
aviso: kitlegal v0.2.0 se actualizarían las skills globales, que eran de v0.1.0; se aplican en la próxima conversación
aviso: las skills instaladas son de kitlegal v0.1.0 y este binario es kitlegal v0.2.0; ejecuta: kitlegal skills install -g
```

Miden 119, 122 y 123 bytes con el salto (research.md, «Lo medido en esta sesión»); con `v0.10.1` y `v0.9.0`, la
primera mide 120. Solo varían las dos versiones: ninguna lleva nada que crezca con el uso.

## §4 · El candado

`cache.TomarCandado(nombre, caducidad)` (`internal/cache/candado.go`), que la composición llama con
`"skills-global.lock"` y `10 * time.Minute`:

| Situación | Resultado | Queda en el disco |
|---|---|---|
| `cache.Directorio()` falla (`KITLEGAL_CACHE_DIR` vacía, o una ruta que no es un directorio) | no tomado | nada |
| el directorio no existe | se crea con 0o700 y se sigue | el directorio |
| el directorio no se puede crear | no tomado | nada |
| no hay candado | se crea en exclusiva (0o600, vacío): tomado | el candado, hasta `liberar` |
| hay uno de hace menos de 10 minutos | no tomado, sin tocarlo y sin esperar | el candado ajeno, igual |
| hay uno de hace 10 minutos o más | se retira y se crea otra vez en exclusiva: tomado | el candado, hasta `liberar` |
| la creación, la retirada o la segunda creación fallan por otra causa | no tomado | lo que hubiera |

`liberar` retira el fichero. Nunca se crea nada en `~/.agents/skills`, así que `install`, `list` y `doctor` no lo ven
(FR-016).

`TomarCandado` es lo único que `internal/cache` exporta de más, y este contrato lo aprueba: entra como
`"func TomarCandado"` en la lista cerrada `superficieDelContrato` de `internal/cache/superficie_test.go`, que
`TestSuperficieExportada` compara con lo exportado (research.md D4). Lo que no se diseña de la sustitución del huérfano, y por qué, en research.md D5.

## §5 · Qué escribe, y cuándo (FR-013)

- **Reinstala** (pasos 7 a 9): el árbol del global (lo que `Aplicar` escribe), el candado, que se retira al terminar, y
  el directorio de la caché si no existía.
- **Cualquier otro caso**, `Ensayar` y `SoloAvisar` incluidos: nada. La sonda de enlace que `Planificar` crea y retira
  en los pasos 5 y 8 no queda.

## §6 · Quién lee cada salida, cuánto y hasta cuándo

| Salida | Quién la lee | Cuántas veces | Tamaño con meses de uso | Se apaga |
|---|---|---|---|---|
| Línea de FR-006 | el agente y la persona, en la salida de error de una consulta | una por actualización del binario, en la primera consulta | 119-120 bytes cada vez, sin nada por norma ni por bloque consultado: con cientos de normas y miles de bloques en la caché mide lo mismo | en la consulta siguiente: las versiones ya son iguales (paso 1) |
| Línea de FR-009 | quien pide `--dry-run` | una por esa orden | 122 bytes, igual | cuando una consulta sin `--dry-run` reinstala, o cuando algo se toca (pasa a la de hoy) |
| Línea de hoy | el agente y la persona | en cada consulta mientras algo está tocado o el candado está tomado (contrato de H19, que el hito no cambia) | 123 bytes por consulta | al ejecutar `kitlegal skills install -g` tras resolver los conflictos; con el candado, al terminar quien lo tiene o al caducar (10 minutos) |
| El global reinstalado | `install`, `list` y `doctor` | cuando se ejecuten | el de `install -g` del pedido: 18 ficheros y 81 433 bytes hoy en 4 skills | — |

Ninguna `SKILL.md` lee ni trata estas líneas (spec, «Contexto»). Una pregunta lanza de una a varias órdenes
(`kitlegal boe articulo`, `cita`, `plazos`, `territorio`), y cada una pasa por §2: con el global al día, el paso 1
lee el manifiesto y compara versiones, como en H19, y termina; `Planificar` (que lee y huella como mucho los 18
ficheros declarados de las skills del pedido, no la caché ni el grafo) solo corre si alguna versión es menor, es
decir, en la primera consulta tras una actualización, o en cada consulta mientras lo instalado está tocado, que es
cuando la línea de hoy ya se repetía en H19.
