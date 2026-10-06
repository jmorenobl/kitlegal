# Contrato: la medida del juez, sus copias, los casos y su ejecución

Lo que hace que el juez solo decida si está medido (FR-023, FR-024, FR-032, FR-040 a FR-054, FR-105, FR-106, FR-108,
FR-109). El voto y la regla, en [juez-y-voto.md](./juez-y-voto.md).

## 1. La medida versionada

`evidencias/adr-0037/medida.json` (717 B), con su copia en `evals/boe-legislacion/juez/medida.json`. Ya está en `main`:
el hito no la repite ni la escribe (FR-040). De ella se leen:

| Clave | Qué es |
|---|---|
| `clase` | La clase que decide a la que se refiere |
| `fecha` | Cuándo se midió |
| `modelo_del_juez`, `version_de_claude_code` | El id del modelo y la versión de Claude Code de los votos |
| `rubrica.sha256`, `casos.sha256` | Las huellas SHA-256 de `rubrica.md` y de `casos.yaml` |
| `defectos.casos`, `defectos.sin_marcar` | Casos etiquetados como defecto, y cuántos no quedaron marcados |
| `correctos.casos`, `correctos.marcados` | Casos etiquetados como correctos, y cuántos quedaron marcados |

Las demás claves (`skill`, `votos`, `origen`) no se leen. No tiene esquema publicado: una medida que no se puede leer,
o sin alguna de esas claves, no corresponde.

## 2. La comprobación (FR-041)

`comprobarLaMedida` recibe el juez de la skill, el id del modelo del juez fijado y la versión de Claude Code fijada
para sus votos, y devuelve una línea por lo que falla, o ninguna. No usa ningún modelo ni abre ningún proceso.

| Qué compara | Línea |
|---|---|
| La huella de `juez/rubrica.md` con `rubrica.sha256` | `la rúbrica (juez/rubrica.md) no es la de la medida versionada` |
| La huella de `juez/casos.yaml` con `casos.sha256` | `los casos (juez/casos.yaml) no son los de la medida versionada` |
| El modelo fijado con `modelo_del_juez` | `la medida versionada es del modelo <a> y el fijado para el juez es <b>` |
| La versión fijada con `version_de_claude_code` | `la medida versionada es de la versión <a> de Claude Code y la fijada para los votos del juez es <b>` |
| `clase` con las clases que deciden de `clases.yaml` | `la clase <c> decide y la medida versionada es de <d>` |
| `defectos.sin_marcar` y `correctos.marcados` con 0 | `la medida versionada no se cumple: <n> defectos sin marcar` y `…: <n> correctos marcados` |

Una medida que no se puede leer, o sin alguna de las claves de §1, da una sola línea, y ninguna de las demás: `la
medida versionada (juez/medida.json) no corresponde: <motivo>`, con el error de su lectura (`no se puede leer como
JSON: …` o `claves que faltan: …`). Una línea por cada clase que decide y no es la de la medida, en el orden de
`clases.yaml`; y las de los recuentos no cambian con 1 («1 defectos sin marcar»).

Quién la hace y con qué valores fijados:

| Quién | Valores | Si falla |
|---|---|---|
| `make ci` (`TestMedidaVersionada`, subprueba `del-repositorio`) | `MODELO_DEL_JUEZ` y `VERSION_DE_CLAUDE_CODE_DEL_JUEZ` de `.github/workflows/evals.yml`, leídos con `leerDefinicionDelJob` | El test falla con las líneas (FR-042) |
| El job, en una ejecución normal, antes de abrir ninguna sesión | Los que recibe de esas dos variables | El informe de [informe-del-job.md](./informe-del-job.md) §5 y el job en rojo (FR-043) |
| El sondeo | El modelo de la definición del job y la versión de Claude Code del equipo | Lo dice en su salida y sigue (FR-076) |
| La ejecución de la medida | No la hace (FR-043, FR-051) | — |

## 3. Las copias (FR-023, FR-108)

`TestCopiasDelJuez` compara, byte a byte, `rubrica.md`, `esquema.json`, `casos.yaml` y `medida.json` de
`evals/boe-legislacion/juez/` con los de `evidencias/adr-0037/`, y falla nombrando cada fichero que difiere o falta.
La pareja de carpetas es una tabla del test: una skill con `juez/` que no esté en ella lo hace fallar. Ninguna tarea
escribe en `evidencias/`; las copias se escriben una vez, con `cp`, y no se vuelven a tocar (FR-045).

## 4. Los casos (FR-024)

`casos.yaml` (78.214 B): `clase` y 259 `casos`, cada uno con `informe`, `sesion`, `grupo`, `etiqueta` (`defecto` o
`correcto`), `procedencia` (`bitacora`, `derivado` o `lectura`) y, según el caso, `quitado` (`norma` y `bloque`) o
`frase`. Se lee con el lector común de YAML, sin esquema publicado: el spec no lo pide, y un caso que no se puede
resolver es un error de la reconstrucción.

De cada caso, sin escribir nada a mano:

| Dato | De dónde sale |
|---|---|
| La respuesta | `respuesta` de la sesión `sesion` del informe `informe`, tal cual |
| La pregunta | `pregunta` del fichero de eval que nombra esa sesión (`eval`): el de `evals/boe-legislacion/` o, si la eval se retiró, el de `testdata/evals/retiradas/` |
| Los textos | La reconstrucción de §5 con las `invocaciones` de esa sesión |
| En un derivado | Los mismos textos sin el de `quitado` (§5, paso 4) |

**La eval retirada.** Dos de los 259 casos son de una eval que H7.2 retiró, `19-lpac-articulo-21-redaccion-cambiada`,
con el informe de H7.1: su fichero y su grafo previo no están en `main` (research M2, D15). Se restauran, byte a byte
con los de `c4819d1^`, en una tarea `[datos]`:

| Fichero restaurado | Origen (`git show c4819d1^:…`) |
|---|---|
| `testdata/evals/retiradas/19-lpac-articulo-21-redaccion-cambiada.yaml` | `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml` |
| `testdata/evals/retiradas/grafo-previo/lpac-a21-version-anterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json` | `testdata/evals/grafo-previo/lpac-a21-version-anterior/` (el mismo nombre) |

De la eval retirada solo se leen `pregunta` y `grafo_previo`, sin validarla contra el esquema de hoy: no es una eval
del conjunto, no entra en ningún plan y `LeerConjunto` no lee esa carpeta. La grabación es una derivada de la de H4:
`TestGrabacionesDerivadas` de `internal/app` recupera su entrada (`versionDelArticulo21` con la fecha `20151002` y su
párrafo sintético, como en `c4819d1^`) sobre la carpeta nueva.

## 5. La reconstrucción de los textos (FR-024, FR-051)

En proceso, sin red, sin modelo y sin el binario instalado, con lo que `internal/evals` ya usa para preparar una
sesión. Es el arnés de la validación (`evidencias/adr-0037/guiones/arnes_test.go.txt`) con `app.Main` en lugar del
binario:

1. **La base.** `Preparar` llena una caché con `UnionDeGrabaciones()` y las consultas necesarias de las evals de hoy.
   Se rehace si tiene más de 100 s: `buscar` y `metadatos` caducan a los 300 s.
2. **Por sesión**, un directorio de caché nuevo: si su eval tiene grafo previo, `prepararGrafoPrevio` con la carpeta
   de los grafos previos, o con la de la eval retirada; después, los ficheros de la base.
3. **Cada invocación** de la sesión, en su orden y salvo `mcp serve`: sus argumentos son las palabras de su `orden`;
   en una llamada, `<applet>_<verbo>` pasa a `<applet> <verbo>` y gana `--json`; todas ganan `--offline`. Se ejecuta
   con `app.Main` sobre un registro con el applet `boe` (reproducción de un directorio vacío y la caché de la sesión),
   el applet `graph` y la entrega al grafo de esa caché. El texto es su `orden` y lo que escribe en la salida
   estándar, termine con el código que termine.
4. **En un derivado**, de esos textos se quita el del bloque `quitado`: fuera el de cada `boe articulo` de esa norma y
   ese bloque que terminó con 0, y del de cada `boe articulos` que lo pidió y terminó con 0, los elementos de `data`
   con ese `bloque` (si no queda ninguno, fuera el texto). El sobre se vuelve a escribir con sus seis claves en su
   orden, en una línea y con su salto final, como lo escribe el binario. El código es el de la invocación repetida, y
   un derivado al que no se le quita nada es un error.

Las sesiones de los seis informes solo invocaron `boe` (`indice`, `articulo`, `articulos`, `buscar`, `metadatos`,
`analisis`), `graph check` y `mcp serve` (research M6): el registro no necesita más applets.

## 6. El control de derivaciones (FR-109)

`TestGrabacionesDerivadas` de `internal/evals` (`medida_test.go`), en `make ci`, con los casos de la copia:

- cada caso nombra un informe versionado y una sesión que está en él, y su respuesta es, byte a byte, la de esa sesión
  leída aparte;
- cada derivado se diferencia de su sesión solo en el texto quitado: sus textos son los de la sesión, en su orden,
  menos los del bloque `quitado`; en el de un `boe articulos`, los mismos valores JSON menos esos elementos; y se
  quita al menos uno;
- la pregunta de cada caso es la de su eval.

El de `internal/app`, del mismo nombre, sigue con las derivadas de las grabaciones y gana la restaurada (§4).
`internal/app` no puede importar `internal/evals`, que lo importa: por eso son dos (research D16).

## 7. La ejecución de la medida (FR-050 a FR-054)

`make evals-medir-juez SKILL=<skill>` ejecuta `scripts/evals-medir-juez.sh`, y este, `TestMedidaDelJuez` (etiqueta
`evals`), que llama a `medirAlJuez`. La lanza una persona ([job-de-evals.md](./job-de-evals.md) §3).

1. Lee el juez de la skill. No comprueba la medida versionada ni la lee para decidir nada.
2. Reconstruye cada caso (§4, §5). No prepara ni abre ninguna sesión de evals.
3. Vota cada caso con la regla de [juez-y-voto.md](./juez-y-voto.md) §8, cuatro casos a la vez: tres votos por cada
   defecto que se marca y uno por cada correcto que no (683 con los casos de hoy si la medida se cumple; research M2).
4. Escribe en su salida estándar la medida entre dos marcas, como el job escribe `informe.json`:

```text
--- inicio de medida.json ---
{
  "skill": "boe-legislacion",
  "clase": "afirma_lo_no_leido",
  "fecha": "2026-10-06",
  "modelo_del_juez": "claude-opus-5-5",
  "version_de_claude_code": "2.1.289",
  "rubrica": {
    "fichero": "rubrica.md",
    "sha256": "5f1e2115b107d942ccdb4dd9b55f8f9606b1b56b52a2cc3e5a8fdef9692d06ee"
  },
  "casos": {
    "fichero": "casos.yaml",
    "sha256": "4827894aefc4133f99d0af93672585f70c9d3f2b8803af36eb2cec4f1811401a"
  },
  "defectos": {
    "casos": 212,
    "sin_marcar": 0
  },
  "correctos": {
    "casos": 47,
    "marcados": 0
  },
  "origen": "ejecución de la medida del juez del job de evals sobre 0123456789abcdef0123456789abcdef01234567"
}
--- fin de medida.json ---
```

- Sus cuatro claves son las de **lo que hay**: las huellas de las copias que ha leído, y el modelo y la versión que ha
  recibido. No las de la medida versionada (FR-052).
- 656 B la del ejemplo. Una por lanzamiento; no se repite sola, y nada la escribe en el repositorio (FR-054).
- **Falla**, y el job sale en rojo, si un defecto no queda marcado o un correcto queda marcado: imprime la medida, con
  sus recuentos, y una línea por caso, `<informe> <sesión> [sin <norma> <bloque>]: etiquetado <etiqueta> y <marcado |
  sin marcar>: «<frase>» · …`, unos 400 B por caso y como mucho 259. Las frases son las de sus votos que dicen sí con
  su frase en la respuesta: tres en un correcto marcado, y dos, una o ninguna en un defecto sin marcar; sin ninguna,
  la línea termina en `sin marcar`.
- **Con algún caso sin juzgar** (un voto que no llega a darse), falla con esos casos y su motivo, una línea por cada
  uno, `<informe> <sesión> [sin <norma> <bloque>]: sin juzgar: <motivo>`, y **no imprime ninguna medida** ni ninguna
  de sus dos marcas (FR-053).

## 8. Tests, en `make ci`

| Test (`medida_test.go`) | Casos | Requisito |
|---|---|---|
| `TestMedidaVersionada` | `del-repositorio`: la del repositorio corresponde y se cumple. Sobre una copia en un temporal: con la rúbrica, los casos, el modelo fijado o la versión fijada cambiados, uno cada vez, y con cada recuento distinto de 0, la línea que lo nombra: seis de seis | FR-041, FR-042, FR-105; SC-005 |
| `TestCopiasDelJuez` | Las cuatro copias iguales; y, sobre una copia en un temporal, un byte cambiado en cada una, que la nombra | FR-023, FR-108; SC-008 |
| `TestGrabacionesDerivadas` | §6 | FR-024, FR-109; SC-009 |
| `TestEjecucionDeLaMedida` | Sobre el texto que `medirAlJuez` devuelve, que es el que el punto de entrada escribe y su guion imprime. Con un votante que responde según la etiqueta del caso: los 259 bien, con una medida versionada que **no** corresponde en la carpeta del juez (una copia en un temporal con la rúbrica cambiada): vota los 259, pide 683 votos y da la medida con las huellas de la copia y 0 de 212 y 0 de 47; un defecto sin marcar y un correcto marcado: falla con el caso y sus frases; un voto que no llega: falla sin dar medida; y en todos, ninguna llamada a quien abre sesiones de evals | FR-051 a FR-053, FR-106; SC-006 |
| `TestGuionDeLaMedida` | `scripts/evals-medir-juez.sh` con el `go` sustituto, como `TestGuionDelSondeo`: sin la skill o sin una variable obligatoria sale con 1 sin ejecutar nada; imprime `medida.json` entre sus dos marcas si el test lo escribió, y ninguna marca si no; y sale con el código del test | FR-052, FR-053 |

## 9. Uso

| Salida | Quién, cuántas veces | Tamaño | Cuándo deja de darse |
|---|---|---|---|
| Las líneas de `make ci` cuando la medida no corresponde | Quien cambia la rúbrica, los casos, el modelo del juez o su versión; una por `make ci` | Una línea por cosa que no coincide, menos de 200 B | Cuando una persona versiona una medida que corresponde y se cumple |
| La medida impresa | Quien lanzó la ejecución; una por lanzamiento | 656 B; con fallo, además unos 400 B por caso | No se repite sola |
