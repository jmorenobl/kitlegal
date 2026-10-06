# Contrato: el informe del job con el juez, y la salida del sondeo

Lo que cambia en `informe.json`, en `informe.md` y en la salida de `make evals-sondeo` (FR-012 a FR-014, FR-030 a
FR-037, FR-043, FR-044, FR-060 a FR-062, FR-070, FR-075, FR-076, FR-104). El voto y la regla, en
[juez-y-voto.md](./juez-y-voto.md); la medida, en [medida-del-juez.md](./medida-del-juez.md).

## 1. Qué respuestas se juzgan

Con una skill que tiene juez, `EscribirInforme` juzga, después de juzgar cada sesión sin modelo y antes de componer el
informe:

| Grupo | Respuestas | Cuenta en umbrales |
|---|---|---|
| De cada modo | Las del modelo que decide, de las series que pide el plan en ese modo, cuya eval activa la skill, terminadas, legibles y medidas: las mismas que forman el `total` de `sin_activar:<modelo>:<modo>` (54 por modo con las evals de hoy) | sí, en los de su modo |
| De la eval sin binario ni servidor | Las del modelo que decide, de la serie que pide el plan sin modo, cuya eval activa la skill, terminadas, legibles y medidas (3) | no (FR-013) |

No se juzgan las del modelo informativo, la sesión de la prueba de red, ni las sesiones sin medir, sin terminar o
ilegibles (FR-012). Los votos van por grupos, en ese orden —modo orden, modo herramienta, eval sin binario ni
servidor—, y el reloj de cada modo va de que empieza su primer voto a que termina el último.

`EscribirInforme` recibe con qué votar en `InformeAEscribir`: el votante (`Votar`), el id del modelo del juez y la
versión de Claude Code de sus votos, cuántas respuestas se votan a la vez y el reloj (`Ahora`; sin él, `time.Now`,
como en `esperarLaDecision`). Con una skill con juez y sin votante, es un error que impide escribir el informe. Con
una skill sin juez, el votante no se llama.

## 2. `umbrales` (FR-030 a FR-037)

Con los dos modos, doce elementos en `boe-legislacion`, diez con `decide: true`, en este orden; `[]` en `legal-core`.
Los de las respuestas existen si la skill tiene juez (research D13).

| # | `nombre` | `medida` | `total` | `umbral` | `decide` |
|---|---|---|---|---|---|
| 1, 4 | `sin_activar:<modelo>:<modo>` | como hoy | respuestas del grupo | 0 | sí |
| 2, 5 | `afirma_lo_no_leido:<modelo>:<modo>` | respuestas del grupo marcadas en la clase | respuestas del grupo | 0 | sí |
| 3, 6 | `cuenta_su_proceso:<modelo>:<modo>` | respuestas del grupo con sí en su primer voto | respuestas del grupo | 0 | no |
| 7 | `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` | `defectos.sin_marcar` de la medida versionada | `defectos.casos` (212) | 0 | sí |
| 8 | `medida_del_juez:afirma_lo_no_leido:correctos_marcados` | `correctos.marcados` | `correctos.casos` (47) | 0 | sí |
| 9, 10 | `duracion_de_las_sesiones:<modo>` | como hoy | — | el objetivo del job (900) | sí |
| 11, 12 | `duracion_del_juez:<modo>` | segundos de los votos del grupo, redondeados hacia arriba | — | 900 | sí |

- Los de una clase salen de `clases.yaml`: su nombre, su `umbral` y su `decide`. Los dos de la medida, por cada clase
  que decide. Ninguno suma dos modos ni cuenta la eval sin binario ni servidor.
- Una respuesta sin juzgar sigue en el `total` y no en la `medida`: el veredicto ya es `fallo` por §4.
- `comparacion` es `"<="` en todos. `umbrales` no lleva `expresiones_prohibidas:…` ni `redaccion_no_leida:…`, de
  ningún modelo (FR-034).

Un elemento (325 B; los doce, unos 3,6 KB, fijos, una vez por job):

```json
{"nombre":"afirma_lo_no_leido:claude-sonnet-5-5:orden","descripcion":"Respuestas de claude-sonnet-5-5 en el modo orden que el juez marca en afirma_lo_no_leido con sus tres votos, sobre sus respuestas juzgadas en las evals que activan la skill","medida":0,"total":54,"comparacion":"<=","umbral":0,"cumple":true,"decide":true}
```

## 3. `juez` en `informe.json` (FR-060, FR-061)

Clave nueva de la raíz, detrás de `umbrales`: `null` si la skill no tiene juez.

| Clave | Qué lleva |
|---|---|
| `modelo`, `version_de_claude_code` | Los fijados para el juez, tal como se recibieron |
| `respuestas` | Una entrada por respuesta juzgada con **algún voto afirmativo** en cualquier clase, también nulo, en orden de sesión. Las demás no están |
| `respuestas[].sesion` | El nombre de la sesión |
| `respuestas[].clases[]` | Una por clase, en el orden de `clases.yaml`: `clase`, `marcada` y `votos` |
| `…votos[]` | Todos los votos de la respuesta, en su orden, también los nulos: `voto` (1 a 3), `nulo`, los campos de `esquema.json` para esa clase tal como los dio el juez (`motivo`, `respuesta`, `frase` y, si la clase lo tiene, `precepto`) y `frase_en_la_respuesta` |
| `sin_juzgar` | Una entrada por respuesta sin juzgar, en orden de sesión: `sesion` y `motivo` |

`marcada`, en una clase que decide, es la de la regla de los tres votos; en una que solo se publica, que su primer
voto dice sí con su frase. Una respuesta sin juzgar va solo en `sin_juzgar`, sin los votos que sí llegaron, aunque
alguno dijera sí: no es una respuesta juzgada.

Una entrada con los tres votos de «sí, sí y no» (1.415 B):

```json
{"sesion":"15-irpf-rendimientos-por-materia-claude-sonnet-5-5-02","clases":[{"clase":"afirma_lo_no_leido","marcada":false,"votos":[{"voto":1,"nulo":false,"motivo":"La respuesta dice de qué trata el art. 7, que ninguna herramienta devolvió.","respuesta":"si","frase":"Remite a otros preceptos de la propia ley, como el art. 7 sobre rentas exentas y la disposición adicional decimoctava.","precepto":"Art. 7 de la Ley 35/2006","frase_en_la_respuesta":true},{"voto":2,"nulo":false,"motivo":"Añade a la remisión al art. 7 la materia de ese artículo, que no está en los textos.","respuesta":"si","frase":"como el art. 7 sobre rentas exentas","precepto":"Art. 7 de la Ley 35/2006","frase_en_la_respuesta":true},{"voto":3,"nulo":false,"motivo":"La remisión está en el texto devuelto y la respuesta no expone la regla del artículo remitido.","respuesta":"no","frase":"","precepto":"","frase_en_la_respuesta":false}]},{"clase":"cuenta_su_proceso","marcada":false,"votos":[{"voto":1,"nulo":false,"motivo":"La respuesta no cuenta comprobaciones ni nombra herramientas.","respuesta":"no","frase":"","frase_en_la_respuesta":false},{"voto":2,"nulo":false,"motivo":"No hay ninguna frase sobre el proceso del asistente.","respuesta":"no","frase":"","frase_en_la_respuesta":false},{"voto":3,"nulo":false,"motivo":"No cuenta lo que hizo ni lo que va a hacer.","respuesta":"no","frase":"","frase_en_la_respuesta":false}]}]}
```

**Tamaño.** No crece con el uso del kit: lo acotan las 111 respuestas juzgadas (54, 54 y 3). Un voto con sus dos
clases ocupa unos 0,8 KB (799 B de media en los 1.189 votos versionados; research M3). Con `cuenta_su_proceso` entre 12
y 30 de cada 54, son de 24 a 60 respuestas con un voto: de unos 20 a unos 60 KB. El máximo, todas marcadas y cada voto
repetido por nulo, 111 × 6 × 0,8 KB, unos 530 KB. Cada job lo escribe de nuevo; una respuesta sin votos afirmativos no deja
nada.

## 4. Motivos del veredicto (FR-007, FR-035)

En la raíz, en el orden de hoy, con estos añadidos:

| Motivo | Dónde va | Texto |
|---|---|---|
| Una clase que decide con alguna marcada en un modo | con los de los umbrales | `umbral afirma_lo_no_leido:<modelo>:<modo>: <n> de <t> (<p> %), y tiene que ser ≤ 0,0 %: <sesión>: «<frase 1>» · «<frase 2>» · «<frase 3>»`, y detrás `; <sesión>: …` por cada otra marcada |
| Respuestas sin juzgar | detrás del de las sesiones sin medir | `de la ejecución, no de la skill: el juez dejó <n> respuestas sin juzgar: <sesión> (<motivo>), …` |
| Los votos de un modo pasan de 900 s | con los de la duración | `de la ejecución, no de la skill: duracion_del_juez:<modo>: <s> s, y tiene que ser ≤ 900 s` |

Los dos últimos empiezan por el prefijo de la ejecución, que los distingue sin modelo de los de la skill, como el de
una sesión sin medir (H7.3 FR 043). El primero, con una respuesta marcada, ocupa unos 400 B (394 B el de la sesión
del ejemplo de §3 con tres frases; research M4) y crece otro tanto por cada marcada más; deja de darse en el primer job
sin marcadas en ese modo. El de las respuestas sin juzgar, 143 B con una, deja de darse en el primer job en que todos
los votos llegan a darse. Una marca en la eval sin binario ni servidor no
da ningún motivo (FR-013). El veredicto sigue siendo `fallo` si hay algún motivo.

## 5. El instrumento sin medir (FR-043)

Antes de abrir ninguna sesión, el job comprueba la medida ([medida-del-juez.md](./medida-del-juez.md) §2). Si no
corresponde o no se cumple, no abre ninguna sesión, ni de evals ni del juez, y escribe el informe con:

- `veredicto`: `fallo`;
- `motivos`: uno por cada cosa que no coincide, con el prefijo de la ejecución: `de la ejecución, no de la skill: el
  instrumento no está medido: <qué>` (174 B el de la versión de Claude Code; research M4);
- `umbrales`: `[]`; `juez`: con su modelo y su versión, `respuestas` y `sin_juzgar` vacías; `tasas`, `evals`,
  `sesiones_sin_medir`, `fuera_de_lo_grabado` y `red`: `[]`;
- lo demás, como en cualquier informe: la cabecera con lo recibido y `sin_python`; y, como no se leen las evals ni
  ninguna sesión, `ficheros_mal_formados`, `modelos_de_sesion` y `versiones_de_claude_code` vacías, y la duración y
  los reintentos a 0. `informe.md` es el documento de siempre, con «ninguna» o «ninguno» en cada sección sin filas.

Así el informe final no da por medido ningún umbral del juez: sus controles salen como «el job no publica ese umbral».
Deja de darse cuando una persona versiona una medida que corresponde y se cumple.

## 6. Lo que sale del informe (FR-034, FR-062, FR-070)

- De la raíz, `expresiones_prohibidas_por_modelo`. De cada eval, `expresiones_prohibidas`. De `umbrales`, los de §2.
- De los motivos de cada sesión, el de una expresión. `Juzgar` no mira la lista: `pasa` ya no depende de ella.
- `scripts/workflow/informe.sh` lee la primera clave con un valor por omisión cuando falta (research V10): no cambia.

## 7. `informe.md`

- Fuera la sección «Expresiones prohibidas por modelo» y la columna «Expresiones prohibidas» de la tabla de sesiones.
- Detrás de «Umbrales», la sección «Juez»: con juez, el modelo y la versión, la tabla de los votos de `respuestas`
  (Sesión, Clase, Voto, Nulo, Respuesta, Frase, En la respuesta, Precepto, Motivo, Marcada) o «ninguna», y la de
  `sin_juzgar` (Sesión, Motivo) o «ninguna»; sin juez, «La skill no tiene juez.».

## 8. La salida del sondeo (FR-075, FR-076)

En lugar de la línea del recuento de expresiones y su 5 %, con una skill con juez (432 B las cinco líneas del
ejemplo; una línea más por cada respuesta sin juzgar):

```text
Juez (claude-opus-5-5), sobre las respuestas de las evals que activan la skill:
- marcadas en afirma_lo_no_leido, con sus tres votos: 0 de 9 (0,0 %).
- con sí en cuenta_su_proceso, con un voto: 2 de 9 (22,2 %).
La medida versionada del juez no corresponde a lo que hay: la versión de Claude Code de sus votos es 2.1.289 y la de este equipo es 2.1.290. Estos recuentos no son los de un juez medido.
Respuestas sin juzgar: ninguna.
```

- Una línea por clase de `clases.yaml`. Si la medida corresponde, la cuarta línea es `La medida versionada del juez
  corresponde a lo que hay, con el Claude Code de este equipo (<versión>).`
- La versión del equipo es la que declaran los transcripts de las sesiones del sondeo, que se abren con el mismo
  `claude` (research D12). Si ninguna la declara, la línea lo dice y el sondeo sigue.
- El sondeo juzga con el votante de [juez-y-voto.md](./juez-y-voto.md) §4, el modelo del juez de la definición del job
  y el `claude` de su `PATH`. No comprueba la medida para decidir nada, no lista votos ni frases, no escribe informe y
  sale con 0 sean cuales sean los recuentos (H7.3 FR 066). Con una skill sin juez: `La skill no tiene juez.`

## 9. Tests, en `make ci`

| Test | Casos | Requisito |
|---|---|---|
| `TestInformeConElJuez` (`informe_test.go`), con sesiones sintéticas y un votante de salidas grabadas | 1 marcada en un modo: `fallo`, el umbral de ese modo con `cumple: false`, el del otro con `true`, y el motivo con la sesión y sus tres frases; 0 marcadas: se cumple y el veredicto es aprobado; `cuenta_su_proceso` con 3: `decide: false`, mismo veredicto y las tres en `juez.respuestas`; sí, sí y no: sin marcar, con sus tres votos y sus dos frases; todos no: sin entrada; un voto que no llega: `sin_juzgar`, `fallo` y el motivo de la ejecución; la eval sin binario ni servidor marcada con tres síes: en `juez.respuestas` con sus tres votos y, frente a las mismas sesiones sin esa marca, los mismos `umbrales`, los mismos motivos y la misma sesión; 901 s del juez en un modo: `fallo` de la ejecución, y 900: se cumple; una skill sin juez: `umbrales` `[]`, `juez` `null` y el votante sin llamadas | FR-012 a FR-014, FR-030, FR-031, FR-033, FR-035, FR-037, FR-060, FR-061, FR-104; SC-004 |
| `TestUmbralesDelInforme` (`umbrales_test.go`) | Los doce elementos en su orden, con sus nombres, sus descripciones y su `decide`; ninguno de los dos que salen | FR-030 a FR-034 |
| `TestEjecucionSinMedir` (`ejecucion_test.go`) | Con una medida que no corresponde: el informe de §5, con 0 llamadas a quien abre las sesiones y 0 al votante | FR-043, FR-104; SC-004 |
| `TestJuzgarSinLaLista` (`juzgar_test.go`) | Con la lista del repositorio en la carpeta, las dos respuestas de la eval sin binario ni servidor del 2026-10-04 y las del calibrado de H7.4 —las 36, 11 y 9 que la lista marca, aplicada con `ExtraerExpresionesProhibidas` a la `respuesta` de cada entrada, en los informes de H7.1, H7.2 y H7.3, exigidas como premisa; las dos de la eval que retiró H7.2, con una eval que solo declara lo que su entrada dice de ella—: ningún resultado lleva la clave `expresiones_prohibidas` ni un motivo de la lista | FR-070, FR-111; SC-011 |
| `TestJuicioDelSondeo`, `TestSalidaDelSondeo` (`sondeo_test.go`) | Las líneas de §8 con votos grabados; el sondeo sale con 0 con marcadas; la medida que no corresponde al Claude Code del equipo | FR-075, FR-076 |
