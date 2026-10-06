# Contrato: el juez, el voto y la regla

Lo que `internal/evals` hace para juzgar una respuesta con el juez con modelo (FR-001 a FR-014, FR-020 a FR-022,
FR-102, FR-103, FR-107). Vale igual en el job, en la ejecución de la medida y en el sondeo.

## 1. La carpeta del juez de una skill

`evals/<skill>/juez/`, con cinco ficheros. Una skill sin esa carpeta no tiene juez (FR-020): ni votos, ni umbrales del
juez, ni comprobación de la medida. Hoy solo la tiene `boe-legislacion`.

| Fichero | Qué es | De dónde sale |
|---|---|---|
| `clases.yaml` | La declaración de clases | Se escribe en este hito, con su esquema `schemas/juez-clases.yaml.json` (tarea `[datos]`) |
| `rubrica.md` | Las instrucciones del juez | Copia de `evidencias/adr-0037/rubrica.md` |
| `esquema.json` | La forma de la respuesta del juez | Copia de `evidencias/adr-0037/esquema.json` |
| `casos.yaml` | Los casos etiquetados | Copia de `evidencias/adr-0037/casos.yaml` |
| `medida.json` | La medida versionada | Copia de `evidencias/adr-0037/medida.json` |

`clases.yaml` de `boe-legislacion` (132 B sin su comentario de cabecera, fija):

```yaml
clases:
  - nombre: afirma_lo_no_leido
    decide: true
    umbral: 0
  - nombre: cuenta_su_proceso
    decide: false
    umbral: 0
```

`schemas/juez-clases.yaml.json`: un objeto sin más claves que `clases`; `clases`, una lista de al menos un objeto con
exactamente `nombre` (`^[a-z0-9_]+$`), `decide` (booleano) y `umbral` (número entre 0 y 1), los tres obligatorios.

`LeerConjunto` deja de tomar la entrada `juez` por un fichero de eval: si es un directorio, lee de él el juez
(`Conjunto.Juez`); si no lo es, es un fichero mal formado. Son mal formados, con su nombre delante (`juez/clases.yaml:
…`), como una eval (FR-020): `clases.yaml` que no cumple su esquema o repite un nombre; un `esquema.json` cuyas
propiedades no son exactamente las clases declaradas; y cualquiera de los cinco que falta o no se puede leer. Un
fichero mal formado hace fallar `make ci` (`TestEvalsDelRepositorio`, subprueba `formato`) y el job, como hoy.

## 2. Los textos de las herramientas de una sesión (FR-001, FR-002)

`LeerSesion` deja en `Sesion.Textos` los textos que devolvieron las herramientas, en el orden de sus bloques `tool_use`
en el transcript, y nada más:

| Bloque `tool_use` | Cuenta si | `Orden` del texto | `Salida` del texto |
|---|---|---|---|
| De `Bash` | su `input.command` lleva la palabra `kitlegal` | el `command`, tal cual | los textos del `content` de su `tool_result`, unidos por un salto de línea |
| De una herramienta del registro (`boe_articulo`…, con el prefijo que le ponga el agente) | siempre | lo que el informe publica en `invocaciones[].orden` de esa llamada | ídem |

- Vale en los dos modos y en la eval sin binario ni servidor: el modo no cambia la regla. Una orden o una llamada que
  falló deja su texto, con el error. Un `tool_use` sin `tool_result` no deja ninguno; un `tool_result` sin ningún
  texto deja el suyo con la salida vacía, que el mensaje escribe `(sin salida)`.
- No son textos el resultado de `Skill`, de `Read` ni de ninguna otra herramienta, ni las órdenes de Bash que no
  nombran `kitlegal`.
- La traza de `strace` no interviene: el sondeo, que no la tiene, lee los mismos textos.

## 3. El mensaje del voto (FR-001, FR-002, FR-004)

Es el de `prompt_de` de `evidencias/adr-0037/guiones/juez.py`, carácter a carácter. Sus partes, unidas por una línea
en blanco:

```text
<textos_de_las_herramientas>

<texto orden="{Orden}">
{Salida sin blancos en los extremos, o «(sin salida)» si queda vacía}
</texto>

</textos_de_las_herramientas>

<pregunta>
{la pregunta de la eval}
</pregunta>

<respuesta>
{la respuesta a la pregunta de la sesión}
</respuesta>

Responde a las dos preguntas de la rúbrica sobre esta respuesta.
```

Hay un `<texto …>` por texto, en su orden. Sin ninguno, en su lugar va la línea `(ninguna herramienta devolvió ningún
texto)`. El mensaje no lleva nada más: ni `SKILL.md`, ni la eval salvo su pregunta, ni el juicio sin modelo de la
sesión (FR-002).

## 4. La orden del voto (FR-003, FR-004)

Cada voto es un proceso nuevo de `scripts/evals-voto.sh`, que ejecuta con `exec`, con los argumentos de `voto_real`
de `juez.py` y en su orden:

```bash
claude -p --model "$modelo" --tools "" --strict-mcp-config --disable-slash-commands --no-session-persistence \
  --system-prompt "$rubrica" --output-format json --json-schema "$esquema"
```

| Cosa | Valor |
|---|---|
| `$modelo`, `$rubrica`, `$esquema` | Los lee el guion de `../modelo.txt`, `../rubrica.md` (entero, con su salto de línea final, como `juez.py`; research V22) y `../esquema.json` (sin su salto de línea final), que quien vota deja en el directorio del juez |
| Entrada estándar | El mensaje de §3 |
| Directorio de trabajo | `<directorio del juez>/cwd`, vacío |
| Entorno | Solo cuatro variables, como en la validación: `PATH`, `HOME` (el directorio del juez), `CLAUDE_CONFIG_DIR` (`<directorio del juez>/config`) y `CLAUDE_CODE_OAUTH_TOKEN`. Son las que el votante da al guion; `claude` ve además las que bash pone por su cuenta (`PWD`, `SHLVL`, `_`), que el guion ni lee ni quita |
| `claude` | El primero del `PATH` del voto. En el job, el directorio del Claude Code del juez va delante ([job-de-evals.md](./job-de-evals.md) §2); en el sondeo, es el del equipo |
| Tope | 35 s; pasado, el proceso se termina, y a los 5 s se cierran sus tuberías (`exec.CommandContext` con `WaitDelay`) |

- El directorio del juez es un temporal por ejecución, que se retira al acabar. No está en el repositorio.
- El contexto del voto deriva del de las señales del punto de entrada: `SIGINT` y `SIGTERM` cortan los votos abiertos.
- Se piden a la vez, como mucho, tantas **respuestas** como sesiones abre a la vez la skill (su `concurrencia`: 4 en
  `boe-legislacion`). Los votos de una misma respuesta van uno detrás de otro.

## 5. La lectura del voto (FR-007)

De la salida estándar del proceso, como `voto_real`:

1. es un objeto JSON; si no, el voto no llega a darse;
2. si su `is_error` es verdadero, no llega a darse;
3. el juicio es su `structured_output`; si no lo trae, su `result` sin la valla de código que lo envuelva, leído como
   JSON;
4. el juicio cumple `esquema.json`; si no, el voto no llega a darse.

Un voto que no llega a darse deja la respuesta **sin juzgar**, con su motivo, que nombra el voto: `voto <n>: tope de
35 s agotado`, `voto <n>: la sesión del juez terminó con error: <texto>`, `voto <n>: la salida no es JSON (código
<c>): <texto>` o `voto <n>: la respuesta no tiene la forma del esquema: <texto>`. El texto va en una línea y cortado
a 300 caracteres. No se reintenta ni se espera (fuera de alcance del spec): un límite de uso es uno de esos errores.

## 6. La frase está en la respuesta (FR-005)

`normal` y `frase_esta` de `juez.py`:

1. de la frase y de la respuesta se quitan `*`, `_` y el acento grave, estén donde estén;
2. cada serie de blancos se cambia por un espacio y se quitan los de los extremos;
3. la frase está si, sin quedar vacía, es subcadena de la respuesta.

Los blancos son los de `unicode.IsSpace` más U+001C a U+001F (research D8, V6 y S2). Nada más se tolera: mayúsculas,
acentos y puntuación se comparan tal cual. Las formas fijas de `avisos.go` no cambian ni se usan aquí.

## 7. El voto nulo (FR-006)

Un voto es nulo si en alguna clase dice `si` y su frase no está. Se publica como nulo y se pide otro, una vez. El
repetido es el que cuenta, en todas las clases; si lleva también un `si` sin su frase, en esa clase no cuenta como sí
y se publica igual.

## 8. La regla (FR-010, FR-011, FR-014)

```text
para n de 1 a 3:
    pedir el voto n (y su repetición si es nulo); si no llega a darse, la respuesta queda sin juzgar: fin
    si n es 1: cada clase que solo se publica cuenta si este voto dice sí con su frase
    si ninguna clase que decide sigue con todos sus votos en sí con su frase: fin
una clase que decide marca la respuesta si tiene tres votos y los tres dicen sí con su frase
```

| Votos de la clase que decide | Votos pedidos | Marcada |
|---|---|---|
| no | 1 | no |
| sí, no | 2 | no |
| sí, sí, no | 3 | no (se publica con sus dos frases) |
| sí, sí, sí | 3 | sí |
| sí nulo y repetido sí, sí, sí | 4 | sí |

El juicio de la sesión sin modelo no cambia con los votos: ni `pasa`, ni sus motivos, ni la tasa de su serie
(FR-014).

## 9. Tests, en `make ci` y sin ninguna sesión con modelo

| Test (`internal/evals`) | Qué fija | Requisito |
|---|---|---|
| `TestTextosDeLaSesion` (`sesion_test.go`) | Con transcripts escritos en el test: los textos del modo orden y del modo herramienta en su orden; la orden que falla; ninguno de `Skill` ni de un Bash sin `kitlegal`; una sesión sin textos | FR-001, FR-107 |
| `TestMensajeDelVoto` (`juez_test.go`) | El mensaje, byte a byte, con textos y sin ellos; la salida vacía; y que no lleva ni un byte de `SKILL.md`, de la eval salvo su pregunta, ni de los motivos del juicio | FR-001, FR-002, FR-004, FR-107; SC-007 |
| `TestOrdenDelVoto` (`juez_test.go`) | Con el `claude` sustituto: sus argumentos, uno a uno y en su orden; su entrada estándar; su directorio; y sus cuatro variables de entorno y ninguna más, descontadas las que pone una shell (`PWD`, `OLDPWD`, `SHLVL`, `_`) | FR-003, FR-004, FR-107; SC-007 |
| `TestFraseEnLaRespuesta` (`juez_test.go`) | Tabla: frase literal; cruza un salto de línea; pierde un acento grave; pierde `*` y `_`; lleva un espacio de no separación; vacía; cambia una mayúscula, un acento o una coma | FR-005, FR-102 |
| `TestVotoDelJuez` (`juez_test.go`) | Con salidas grabadas: sí con su frase, vale; sí con una frase que no está, nulo y repetido una vez; sí con una frase que solo difiere en blancos y énfasis, vale; y los cuatro motivos de un voto que no llega a darse, con el tope incluido | FR-005 a FR-007, FR-102; SC-002 |
| `TestReglaDeLosVotos` (`juez_test.go`) | Las filas de §8 y la clase que solo se publica, con un votante que cuenta sus llamadas | FR-010, FR-011, FR-103; SC-003 |

Las salidas grabadas son constantes de los tests con la forma de la salida de `claude -p --output-format json`. El
votante de los demás tests es una función que las devuelve: `Votante` es `func(mensaje string) ([]byte, error)`.
