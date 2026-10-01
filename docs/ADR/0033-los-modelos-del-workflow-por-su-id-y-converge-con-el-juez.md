# 0033 · Los modelos del workflow `hito` se fijan por su id, y `converge` va con el modelo del juez

- **Estado**: aceptada
- **Fecha**: 2026-10-01
- **Hito**: transversal, tras la release v0.3.2 y antes de H20. Workflow `hito` 2.5.0; la constitución no cambia.
  Sustituye la última frase del punto 3 del ADR 0031 (la «segunda señal»); el resto del ADR 0031 sigue vigente.
  Aplica al workflow el criterio de los ADR 0007 (el juez, al menos tan capaz como el autor) y 0031 (el modelo, por su
  id y con un diff).

## Contexto y problema

Antes de H20 se revisó, con los runs de H6 a H7.4, si cada rol del workflow tiene el modelo que le corresponde. Casi
todo se sostiene (abajo, «Lo que se queda como está»). Cuatro cosas no:

1. **`converge` no ha encontrado nunca nada.** Es el único paso que busca lo que falta por construir antes de la
   revisión final, y corría con `modelo_analisis` (`sonnet@high`), el rol de «lectura y transformación». En sus 14
   sesiones (H0 a H7.4) concluyó siempre que no faltaba nada y no escribió una sola tarea en `tasks.md`. Lee poco: 5
   órdenes en H19, entre 10 y 18 llamadas a herramientas en los cinco hitos de H7, entre 2 000 y 6 000 tokens de
   salida. Después, la primera ronda de la revisión final rechazó en 8 de los 10 hitos que guardan sus veredictos por
   ronda (H4 a H7.4; 18 motivos en H19, 11 en H6, 10 en H5). No está medido cuántos de esos motivos habría visto un
   `converge` más profundo; lo que está medido es que el paso no decide nada y que su modelo es menos capaz que el que
   implementa (`opus@xhigh`), al revés de lo que el ADR 0007 pide de quien juzga.
2. **El modelo de un rol cambia sin que nadie lo decida.** Los roles llevaban alias, y a qué id resuelve un alias lo
   decide la versión de Claude Code de quien lanza (ADR 0031, «Contexto»). En los transcripts del proyecto, la última
   respuesta de `claude-opus-5` es del 2026-09-24 a las 14:03 (Claude Code 2.1.278) y la primera de `claude-opus-5-5`,
   de las 14:07 (2.1.280): el cambio cayó entre las rondas a mano de H6. `sonnet` cambió el
   2026-09-28 a las 22:15, con el run de H7.1 en marcha. Ningún diff recoge ninguno de los dos, y comparar un hito con
   el anterior mezcla el cambio de modelo con los del workflow. Un modelo nuevo tampoco hace mejor todo lo que hacía
   el anterior: Sonnet 5.5 dejó de activar la skill en la eval 04 (ADR 0031).
3. **No se podía fijar aunque se quisiera.** La lista cerrada de valores de los inputs `modelo_*` no tenía
   `claude-opus-5-5` ni `claude-sonnet-5-5`.
4. **`scripts/coste-run.sh`, con el que se mide todo esto, daba cifras que no eran.** Tres defectos:
   - casaba el modelo por prefijo, así que Opus 5.5 llevaba el precio de Opus 5 (5/25/0,50 USD por millón de tokens
     de entrada, salida y lectura de caché, en vez de 4/20/0,20): sus sesiones salían 1,8 veces más caras de lo que
     son (351 frente a 194 USD en H7.4, 661 frente a 368 en H7). Es el defecto que el ADR 0031 cerró en el control de
     las evals;
   - la fila «manual (paso.sh)» y cualquier paso que quedó abierto se llevaban todas las sesiones posteriores, también
     las de los runs siguientes: el run de H7 salía con 2 378 USD, de los que 1 638 eran los runs de H7.1 a H7.4;
   - `docs/WORKFLOW.md` seguía diciendo que Fable lee de caché a mitad de precio que Opus. Con Opus 5.5 es al revés
     (0,25 frente a 0,20).

## Lo que se queda como está

- **Implementación (`opus@xhigh`).** De H19 a H7.4, 115 de 120 tareas salieron al primer intento, y ninguna fue a
  cuarentena.
- **Juez adversarial (`fable@xhigh`).** Ha rechazado él solo, con el juez A aprobando, por defectos reales: un test
  que fallaba al llegar al árbol por un enlace simbólico (H5, ronda 3) y `territorio resolver --timeout 1ms` saliendo
  con 4 contra su contrato (H6, ronda 8). La otra familia ve lo que la primera no.
- **Escalada (`fable@xhigh`).** Entre el 2 y el 9 % del coste de un run (H6 a H7.4); actúa solo cuando algo ya falló.
- **Jueces de entrada (`opus@xhigh`) y correctores (`opus@high`).** Desde el ADR 0028 convergen: en H7.1 a H7.4 el
  spec se aprueba en 2 rondas y el plan en 2 (4 en H7.1), sin pendientes. Antes, H7 agotó las 4 de spec y de plan.
- **`clarify_integrar` y `analyze` (`sonnet@high`).** Transforman y contrastan artefactos, y lo que escriben lo juzga
  después un juez (`juez_spec`, `juez_tasks`).

## Opciones consideradas

**Para `converge`**: dejarlo en `modelo_analisis`; pasarlo a `modelo_juez`; darle un rol propio; ponerlo en Fable.
Se elige `modelo_juez`: es un juicio sobre la implementación, y el criterio del proyecto para un juez ya está escrito.
Un rol propio no añade nada mientras su valor sea el del juez, y nada indica que el paso necesite otra familia: no ha
fallado con Opus, no lo ha probado.

**Para el modelo de cada rol**:

- *Seguir con alias.* El workflow recibe el modelo nuevo de cada familia sin tocar nada, pero lo recibe cuando se
  actualiza Claude Code, a mitad de un run si coincide, y sin medida que lo separe de lo demás.
- *Alias resuelto al lanzar y congelado para el run.* Quita el cambio a mitad de run, pero el modelo sigue cambiando
  entre runs sin diff ni revisión, y obliga al supervisor a abrir sesiones con modelo antes de empezar.
- *Id completo en los valores por defecto.* Se elige: el modelo de un rol es un dato del repositorio, cambia con una
  propuesta de cambio y el primer run posterior mide ese cambio y nada más. Es lo que el ADR 0031 ya decidió para el
  job de evals.

## Decisión

1. **`converge` usa `modelo_juez`.** `modelo_analisis` queda para `clarify_integrar` y `analyze`.
2. **Los valores por defecto de los ocho roles llevan el id completo**: `claude-opus-5-5` donde decía `opus`,
   `claude-fable-5-1` donde decía `fable` y `claude-sonnet-5-5` donde decía `sonnet`, con el mismo esfuerzo. Son los
   ids a los que resuelven hoy los tres alias (Claude Code 2.1.284, comprobado con la orden del punto 4), así que
   ningún rol cambia de modelo con esta decisión salvo `converge`. El repuesto por límite de uso
   (`KITLEGAL_MODELO_FALLBACK`) pasa a `fable=claude-opus-5-5,opus=claude-sonnet-5-5`.
3. **La lista de valores admite `claude-opus-5-5` y `claude-sonnet-5-5`** con sus esfuerzos. Los alias y los ids
   anteriores siguen en ella, para sobrescribir un run con `KITLEGAL_MODELO_<ROL>`.
4. **Detección.** Antes de lanzar un hito, quien lo lanza ejecuta la orden del ADR 0031 con los tres alias:
   `claude -p --model <alias> --output-format json 'Responde solo: ok'`, y mira `modelUsage`. Si un id no es el del
   workflow, subirlo es una propuesta de cambio aparte, y `scripts/coste-run.sh` del primer run con el modelo nuevo
   es su medida. Seguir el alias es aplicar este ADR, no decidir otra vez.
5. **`scripts/coste-run.sh`** lleva los precios de Opus 5.5 y Sonnet 5.5, casa el id exacto (o seguido de la fecha de
   la versión) y deja «sin precio» lo que no conoce, y acota cada run por el comienzo del siguiente.

## Qué mide H20

H20 es el primer run con esta decisión. En su tabla de `scripts/coste-run.sh`, la fila de `converge` dice si el paso
lee más (turnos y tokens de salida, frente a 9–14 turnos y 4 000–6 000 tokens en H7 a H7.4), y `tasks.md` dice si
añadió tareas. Si sigue sin encontrar nada y la primera ronda de la revisión final sigue rechazando por trabajo que
faltaba, el defecto no era el modelo sino lo que se le pide al paso, y eso es otra decisión.

## Consecuencias

**A favor.** El paso que decide si falta algo lo ejecuta un modelo tan capaz como el que implementó. El modelo de cada
rol está en el repositorio y su cambio se revisa y se mide. Las cifras de coste vuelven a ser las de cada run, con el
precio de cada modelo: un paso en Fable 5.1 cuesta entre 1,9 y 2,2 veces lo que en Opus 5.5 con el mismo perfil de
tokens (H7, H7.3 y H7.4), que es el dato con el que se decide dónde va Fable.

**En contra, y asumido.**

- **Un modelo nuevo no llega solo.** Hasta que alguien ejecuta la orden del punto 4 y sube el id, el workflow sigue
  con el anterior. Es lo que se quiere, y depende de una persona, como en el ADR 0031.
- **La «segunda señal» del ADR 0031 desaparece**: la tabla de `scripts/coste-run.sh` ya no muestra a qué resuelve
  `sonnet`, porque los pasos de análisis piden un id. La detección del ADR 0031 (su punto 3) no cambia, y la orden de
  este ADR la extiende a los tres alias.
- **Cada modelo nuevo exige tocar la lista de valores.** Es el precio de que la lista sea cerrada.
- **`converge` consume más.** Pasa de menos de 1 USD por run a lo que cueste una sesión de Opus a `xhigh` que lea el
  código; H20 lo mide.
