# Contrato: formato de evals, `evals/boe-legislacion/`, grabaciones y comprobación sin red

FR-060 a FR-066, FR-072, FR-074, FR-075, US4-3 a US4-7, SC-003 (composición), SC-009, SC-010. Decisiones: research.md
D10 (formato), D11 (grabación con manifiesto y reutilización de H4), D12 (comparación), D14 (preparación de caché).
El job que ejecuta las evals con un modelo está en [job-de-evals.md](./job-de-evals.md).

## 1. Formato común (`schemas/eval.yaml.json`)

Un fichero YAML por eval en `evals/<skill>/`, nombre `<nn>-<descripción>.yaml` (`nn` de dos cifras, descripción en
minúsculas con guiones).

```yaml
# Eval positiva: debe activar la skill, leer el bloque y citarlo.
pregunta: "¿qué dice el art. 21 de la Ley 39/2015?"
activa: true
comandos:
  - applet: boe
    norma: BOE-A-2015-10565
    bloque: a21
citas:
  - norma: BOE-A-2015-10565
    bloque: a21
```

```yaml
# Eval de no activación: pregunta ajena a la normativa del BOE.
pregunta: "¿Cómo invierto una lista enlazada en Go?"
activa: false
```

Esquema (JSON Schema 2020-12, `$id` `https://ventanillalegal.es/schemas/eval.yaml.json`):

- raíz `object`, `additionalProperties: false`, `required: [pregunta, activa]`;
- `pregunta`: `string`, `minLength: 1`; `activa`: `boolean`; `reproduce`: `string` con `^[a-z0-9]+(-[a-z0-9]+)*$`;
- `if` `activa` es `true` `then` `required: [comandos, citas]`; `else` `not` `anyOf` `[required comandos, required citas]`
  (FR-061);
- `comandos`: `array`, `minItems: 1`, `items` `oneOf` de las tres formas de data-model §6.1, cada una
  `additionalProperties: false`:
  - bloque: `required: [applet, norma, bloque]`;
  - consulta de norma: `required: [applet, verbo, norma]`, `verbo` `enum` `[indice, metadatos, analisis]`;
  - búsqueda: `required: [applet, verbo, terminos]`, `verbo` `const` `buscar`, `terminos` `array` `minItems: 1` de
    `string` `minLength: 1`;
  - `applet` con `^[a-z0-9]+(-[a-z0-9]+)*$`, `norma` con `^BOE-A-[0-9]{4}-[0-9]{1,9}$`, `bloque` con
    `^[A-Za-z0-9][A-Za-z0-9.-]{0,63}$`;
- `citas`: `array`, `minItems: 1`, `items` `{norma, bloque}` con los mismos patrones, `additionalProperties: false`.

`internal/evals.LeerEval(nombre string, contenido []byte) (Eval, error)`: lectura YAML común de data-model («Lectura de
documentos YAML»: `yaml.Node`, rechazo de toda clave repetida con sus dos líneas, `(*yaml.Node).Decode`, normalización
a tipos JSON) y validación contra el esquema; todo error empieza por el nombre del fichero (US4-4). Una clave repetida
no puede quedarse con el último valor en silencio: analizar a `yaml.Node` no la detecta (research.md V43), y por eso la
rechaza el recorrido del lector. La `Eval` devuelta lleva en `Fichero` el `nombre` con el que se leyó (data-model §6).

`internal/evals.LeerConjunto(dir string) (Conjunto, error)` (`internal/evals/conjunto.go`) lee **todos** los ficheros
del directorio antes de devolver nada y separa los bien formados de los mal formados (FR-071: el job falla nombrando cada
fichero mal formado antes de ejecutar ninguna eval). Solo lee y comprueba el formato: las reglas del conjunto no son
suyas, sino de `ComprobarConjuntoDeBoeLegislacion` (§2).

```go
type Conjunto struct {
	Evals       []Eval              // las bien formadas, en orden de nombre de fichero, cada una con su Fichero
	MalFormados []FicheroMalFormado // las que no se pueden leer como eval, en orden de nombre de fichero
}

type FicheroMalFormado struct {
	Fichero string // nombre de la entrada dentro del directorio
	Error   error  // por qué no es una eval; su texto empieza por Fichero
}
```

1. Lista las entradas de `dir` en orden de nombre. Toda entrada es un fichero de eval: una que no es un fichero
   regular, o cuyo nombre no tiene la forma `<nn>-<descripción>.yaml` de arriba (`^[0-9]{2}-[a-z0-9]+(-[a-z0-9]+)*\.yaml$`),
   es un `FicheroMalFormado` con ese motivo, nunca una entrada que se salta.
2. Lee cada fichero y lo pasa a `LeerEval` con su nombre: si no se puede leer o `LeerEval` devuelve un error, es un
   `FicheroMalFormado` con ese error; si no, su `Eval` va a `Evals`.
3. El `error` queda para un `dir` que no se puede listar (no existe, no es un directorio o no se puede leer): nombra
   `dir` y va con un `Conjunto` vacío. Un fichero mal formado nunca es el `error`, y un directorio vacío da un `Conjunto`
   vacío sin error: cuántas evals hacen falta lo dice `ComprobarConjuntoDeBoeLegislacion`.

Quien lee un conjunto decide qué hace con los mal formados, con esta misma firma: `TestEvalsDelRepositorio/formato`
(§2) y `PrepararSesion` (contrato del job §3.2, paso 1) fallan nombrando cada uno con su error; `EscribirInforme`
(contrato del job §3.3, paso 1) los lleva a `ficheros_mal_formados` y juzga las sesiones con las bien formadas.

| Test | Qué fija |
|---|---|
| `TestLeerConjunto` | sobre un `t.TempDir()` que el propio test llena desde constantes: `bien-formadas` (`01-…` y `02-…` válidas: las dos en `Evals`, en ese orden y con su `Fichero`, `MalFormados` vacía y sin error); `con-mal-formadas` (`01-…` válida, `02-…` sin `pregunta` y `03-…` con `activa` repetida: solo la 01 en `Evals`; la 02 y la 03 en `MalFormados`, cada una con un error que empieza por su nombre; sin error); `entradas-que-no-son-evals` (`01-…` válida, un subdirectorio `02-subdirectorio.yaml` y un fichero `notas.txt`: los dos en `MalFormados`, nombrados, y la 01 en `Evals`); `directorio-inexistente` (error que nombra el directorio y `Conjunto` vacío) |
| `TestLeerEval` | válidas (las tres formas de comando, con y sin `reproduce`, de no activación); inválidas nombrando el fichero: sin pregunta, sin `activa`, positiva sin citas, no activación con comandos, comando con `verbo` y `bloque` a la vez, cita sin identificador `BOE-A-…` válido, bloque `.a1`, clave desconocida, y **clave repetida** —`activa` dos veces en la raíz (la primera `true` y la segunda `false`) y `bloque` dos veces dentro de una cita—, cada una nombrando la clave y sus dos líneas |
| `TestEsquemaDeEval` | el esquema compila; sus patrones de norma y bloque cubiertos por `TestGramaticasCoincidenConBoe` |

## 2. `evals/boe-legislacion/` (FR-062 a FR-066)

Doce ficheros: diez positivos de materias distintas y dos de no activación. Donde el identificador lo fija la pausa de
grabación, la tabla lo escribe como «‹id de …›»; la tarea que escribe las evals lo copia de la búsqueda grabada.

| Fichero | Pregunta | Comandos esperados | Citas esperadas |
|---|---|---|---|
| `01-lpac-articulo-21.yaml` | ¿qué dice el art. 21 de la Ley 39/2015? | bloque `BOE-A-2015-10565` `a21` | `BOE-A-2015-10565` `a21` |
| `02-lcsp-contrato-menor.yaml` | ¿Qué debe incluir el expediente de un contrato menor según el artículo 118 de la Ley de Contratos del Sector Público? | bloque `BOE-A-2017-12902` `a1-30` | `BOE-A-2017-12902` `a1-30` |
| `03-lrbrl-atribuciones-del-pleno.yaml` | ¿Qué atribuciones tiene el Pleno del ayuntamiento según el artículo 22 de la Ley reguladora de las Bases del Régimen Local? | bloque `BOE-A-1985-5392` `a22` | `BOE-A-1985-5392` `a22` |
| `04-lgt-prescripcion.yaml` | ¿En cuántos años prescribe el derecho de la Administración a liquidar una deuda tributaria según el artículo 66 de la Ley General Tributaria? | bloque ‹id de la Ley 58/2003› `a66` | ‹id› `a66` |
| `05-trlrhl-impuestos-municipales.yaml` | ¿Qué impuestos pueden exigir los ayuntamientos según el artículo 59 del texto refundido de la Ley reguladora de las Haciendas Locales? | bloque ‹id del RDLeg 2/2004› `a59` | ‹id› `a59` |
| `06-irpf-rendimientos-del-trabajo.yaml` (`reproduce: boe-fiscal`) | ¿Qué rendimientos se consideran rendimientos íntegros del trabajo según el artículo 17 de la ley del IRPF? | bloque ‹id de la Ley 35/2006› `a17` | ‹id› `a17` |
| `07-lrjsp-principio-de-legalidad.yaml` | ¿Qué dice el artículo 25 de la Ley 40/2015 sobre el principio de legalidad en la potestad sancionadora? | bloque ‹id de la Ley 40/2015› `a25` | ‹id› `a25` |
| `08-ltaibg-plazo-de-resolucion.yaml` | ¿En qué plazo hay que resolver una solicitud de acceso a la información pública según el artículo 20 de la Ley 19/2013? | bloque ‹id de la Ley 19/2013› `a20` | ‹id› `a20` |
| `09-constitucion-articulo-140.yaml` | ¿Qué dice el artículo 140 de la Constitución? | bloque ‹id de la Constitución› `a140` | ‹id› `a140` |
| `10-et-vacaciones.yaml` | ¿Cuántos días de vacaciones anuales reconoce el artículo 38 del Estatuto de los Trabajadores? | bloque ‹id del RDLeg 2/2015› `a38` | ‹id› `a38` |
| `11-no-activa-programacion.yaml` | ¿Cómo invierto una lista enlazada en Go? | — (`activa: false`) | — |
| `12-no-activa-acuerdo-entre-amigos.yaml` | Reescribe en un tono más cercano esta frase de un acuerdo entre amigos para compartir coche: «Las partes se turnarán el uso del vehículo en fines de semana alternos». | — (`activa: false`) | — |

- **Comandos esperados según el protocolo.** Las diez positivas nombran el artículo en la pregunta, y de cada una se
  espera solo su bloque. El índice que el paso 3 de la skill lee para copiar el id (D23) está entre las consultas
  necesarias de toda norma de las evals (data-model §7.1) pero no se espera, porque una invocación que no es un comando
  esperado no cambia `pasa` (data-model §10.2). En nueve de las diez normas el artículo N tiene el id `aN`; en la Ley
  9/2017 no, y el artículo 118 es `a1-30` (research.md V64 (2)), así que en la 02 el bloque esperado solo sale de la
  entrada del índice, como manda el paso 3. La 05 y la 07 nombran el artículo desde T041 y las otras seis desde T043
  (research.md D23, «Las positivas nombran el artículo»): preguntadas por materia, con un índice sin rúbricas, medían
  si el modelo de las sesiones recuerda el número del artículo, que cambia de una sesión a otra, y no el protocolo.
  Ninguna espera `buscar`, porque todas las normas están en `references/normas.md` y el protocolo toma de ahí el
  identificador (US1-2).
- **Bloques candidatos.** `a21`, `a1-30` (artículo 118 de la LCSP) y `a22` están en índices grabados en H4. `a66`,
  `a59`, `a17`, `a25`, `a20`, `a140` y `a38` son los ids que la gramática del BOE da a esos artículos y los confirma el
  índice grabado en la pausa; si uno no está, la persona corrige el bloque en el manifiesto y en esta tabla antes de
  aprobar (S3 de research.md).
- `06` reproduce `articulo BOE-A-2006-20764 a17` e `indice BOE-A-2006-20764` de los usos documentados de `refs/boe.py`
  (líneas 8-9 y 760-761), y lo declara con `reproduce: boe-fiscal` (FR-064): el bloque es su comando esperado, y el
  índice, que desde T043 no se espera, sigue entre sus consultas necesarias y lo lee el paso 3 de la skill.
- Materias distintas (FR-062): cada positiva cita una norma que ninguna otra cita.

`internal/evals.ComprobarConjuntoDeBoeLegislacion(evals []Eval, normas map[string]NormaConocida) []DefectoDelConjunto`
(`internal/evals/conjunto.go`) aplica las reglas de data-model §6.3, salvo la de revisión (`sin municipio`), a las evals
bien formadas que devolvió `LeerConjunto` (§1); no lee ficheros, y `LeerConjunto` no la llama.

```go
// Lo que las reglas del conjunto necesitan de una norma de data/normas.yaml; el mapa va por identificador.
type NormaConocida struct {
	Abreviatura string
	Materias    []string
}

type DefectoDelConjunto struct {
	Regla   string // nombre de la regla en la tabla de data-model §6.3 (p. ej. «materias distintas», «normas conocidas»)
	Mensaje string // qué se incumple, nombrando los ficheros de eval y las normas implicadas
}
```

Devuelve un defecto por cada incumplimiento, en el orden de la tabla de §6.3, y ninguno si se cumplen todas. Recibe las
normas como un mapa propio y no como `[]skills.Norma` porque `conjunto.go` llega en el paso 2 del orden de implementación
y `internal/skills/normas.go` en el 6: `TestEvalsDelRepositorio` (paso 7) construye el mapa con `skills.LeerNormas` sobre
`data/normas.yaml`, y `TestConjuntoDeEvals`, con normas sintéticas.

`TestEvalsDelRepositorio` (`internal/evals`), subtests, sobre `evals/boe-legislacion/` y `data/normas.yaml` de la raíz
(`../../evals/boe-legislacion` y `../../data/normas.yaml` desde el directorio del paquete, research.md V46):

| Subtest | Qué fija |
|---|---|
| `formato` | al menos las doce evals bien formadas en `evals/boe-legislacion/`; y `LeerConjunto` sin error y con `MalFormados` vacía sobre **cada** directorio `evals/<skill>/` de la raíz, el de `boe-legislacion` entre ellos, porque el formato común es el de las evals de cualquier skill (FR-060); si alguno tiene mal formados, falla nombrando cada fichero con su directorio y su error. Que se lean todos los directorios lo fija `TestFormatoDeLasEvalsDeCadaSkill`, sobre un `evals/` que el test escribe en un `t.TempDir()` |
| `conjunto` | `ComprobarConjuntoDeBoeLegislacion` sobre las `Evals` de `formato` y el mapa de `data/normas.yaml`: ningún defecto, salvo los de la regla `normas conocidas`, que presenta el subtest siguiente |
| `normas-conocidas` | los defectos de la regla `normas conocidas` de esa misma llamada: ninguno (toda norma de cita o comando está en `data/normas.yaml`, FR-020) |
| `grabado` | §5.2 sobre el árbol real (SC-010) |

`TestConjuntoDeEvals` fija `ComprobarConjuntoDeBoeLegislacion` con evals y normas sintéticas escritas como constantes:
un conjunto que cumple todas las reglas no da ningún defecto, y un subtest por regla, sobre una copia que incumple solo
esa (nueve positivas, dos positivas que citan la misma única norma, sin `reproduce`…), da exactamente un defecto con esa
`Regla` y un `Mensaje` que nombra los ficheros implicados.

## 3. Grabación (tarea `[datos]` con pausa)

### 3.1 Manifiesto `testdata/evals/grabaciones.json`

Vive bajo `testdata/` de la **raíz**, junto a las grabaciones de H5 (`testdata/evals/boe.legislacion-consolidada/`),
porque, fuera de `internal/source/`, es la única ubicación de material de prueba en la que un fichero nuevo hace pausar
al workflow (`clasificar_datos`; research.md V36 y D11), y en esa pausa graba la persona. Bajo `internal/evals/testdata/`
la tarea del manifiesto no pausaría: nadie grabaría y la tarea del esquema de normas se quedaría sin rangos grabados.

Una entrada por norma de `data/normas.yaml` (data-model §7.2). Contenido de partida (las búsquedas y prefijos se
confirman o corrigen en la pausa):

| `busqueda` | `titulo_empieza_por` | `bloques` |
|---|---|---|
| procedimiento administrativo común | `Ley 39/2015,` | `a21` |
| contratos del sector público | `Ley 9/2017,` | `a1-30` |
| bases del régimen local | `Ley 7/1985,` | `a22` |
| ley general tributaria | `Ley 58/2003,` | `a66` |
| texto refundido haciendas locales | `Real Decreto Legislativo 2/2004,` | `a59` |
| impuesto renta personas físicas | `Ley 35/2006,` | `a17` |
| régimen jurídico del sector público | `Ley 40/2015,` | `a25` |
| transparencia acceso información pública buen gobierno | `Ley 19/2013,` | `a20` |
| constitución española | `Constitución Española` | `a140` |
| estatuto de los trabajadores | `Real Decreto Legislativo 2/2015,` | `a38` |

La búsqueda de la primera entrada, `procedimiento administrativo común`, es la misma que grabó H4
(`internal/source/boe/testdata/grabaciones.json`): el arnés la sirve desde las grabaciones de H4 y no la vuelve a grabar
(§3.2, paso 1), de modo que ningún fichero de H5 lleva el nombre del de H4 (§3.4).

`internal/evals.LeerManifiesto(contenido []byte) (Manifiesto, error)` (`internal/evals/grabaciones.go`) es el único
lector del manifiesto: `TestGrabarEvals` (§3.2), `TestIdentificadoresDeLasNormas` (contrato de normas y referencias §6) y
`TestManifiestoDeGrabaciones` le pasan el contenido del fichero, y ninguno decodifica el JSON por su cuenta.

```go
// Manifiesto es el manifiesto de grabación de H5 (data-model §7.2).
type Manifiesto struct {
	Fuente string                 `json:"fuente"`
	Normas []EntradaDelManifiesto `json:"normas"`
}

// EntradaDelManifiesto es una norma que se graba: con qué se busca, por qué prefijo se reconoce su resultado, qué
// bloques se graban de ella y para qué eval y requisito.
type EntradaDelManifiesto struct {
	Busqueda         string   `json:"busqueda"`
	TituloEmpiezaPor string   `json:"titulo_empieza_por"`
	Bloques          []string `json:"bloques"`
	Para             string   `json:"para"`
}
```

Comprueba, en este orden, y devuelve como error el primer defecto; ninguno se descarta:

1. **Documento.** Lo decodifica con `encoding/json/v2`, `json.Unmarshal(contenido, &m, json.RejectUnknownMembers(true))`:
   un único valor JSON, sin nada detrás; un miembro que no es un campo de arriba (los nombres se comparan distinguiendo
   mayúsculas) o una clave repetida en cualquier objeto son un error que la nombra con su ruta JSON (research.md V55). No
   se usa `encoding/json`: ante una clave repetida se queda con el último valor en silencio, también con
   `DisallowUnknownFields`, que es como lee H4 sus datos de prueba (V55).
2. **Fuente.** `fuente` es `boe.NombreDeLaFuente` (`boe.legislacion-consolidada`); si no, el error nombra el valor.
3. **Normas.** `normas` tiene al menos una entrada.
4. **Entrada.** En cada una, `busqueda`, `titulo_empieza_por` y `para` no están vacíos; cada id de `bloques` es válido
   para `boe.ValidarBloque` (la gramática `BLOQUE` de data-model, sin copiar su expresión) y ninguno se repite en la
   entrada. `bloques` puede faltar o estar vacía: de una norma que ninguna eval cita, como una que solo nombra
   `SKILL.md`, se graban la búsqueda, el índice y los metadatos, pero ningún bloque.
5. **Norma repetida.** Dos entradas cuyos `titulo_empieza_por` son iguales, o uno empieza por el otro, son la misma norma
   repetida. El manifiesto no lleva identificadores: cada entrada resuelve su norma como el único resultado de su
   búsqueda cuyo título empieza por su prefijo (§3.2), y un mismo título solo empieza por dos prefijos si uno de ellos
   empieza por el otro. La regla recoge así, antes de grabar nada y aunque las búsquedas sean distintas, toda pareja de
   entradas que podría grabar dos veces la misma norma. El prefijo se compara byte a byte (`strings.HasPrefix`), igual
   que al elegir el resultado.

Los errores de los puntos 2 a 5 nombran la entrada por su posición (desde 1) y su `titulo_empieza_por`, y el campo o el
bloque; el de norma repetida, las dos entradas.

`TestManifiestoDeGrabaciones` (`internal/evals/grabaciones_test.go`) fija `LeerManifiesto` con un subtest por caso:

| Subtest | Contenido | Qué fija |
|---|---|---|
| `valido` | sintético: dos entradas, una con dos bloques y otra sin `bloques` | sin error; `Manifiesto` con la fuente y las dos entradas en su orden, con sus campos |
| `clave-desconocida` | sintético: una entrada con un miembro `norma` | error que nombra `norma` |
| `clave-repetida` | sintético: `busqueda` dos veces en una entrada, con valores distintos | error que nombra `busqueda`, en lugar de quedarse con el segundo valor |
| `datos-tras-el-valor` | sintético: un segundo objeto detrás del manifiesto | error |
| `otra-fuente` | sintético: `fuente` `boe.otra` | error que nombra `boe.otra` |
| `sin-normas` | sintético: `normas` vacía | error |
| `busqueda-vacia`, `prefijo-vacio`, `para-vacio` | sintético: el campo vacío en la segunda entrada | error que nombra la entrada 2 y el campo |
| `bloque-mal-formado`, `bloque-repetido` | sintético: `.a1`; `a21` dos veces | error que nombra la entrada y el bloque |
| `prefijo-repetido` | sintético: dos entradas con `Ley 39/2015,` y búsquedas distintas | error que nombra las dos entradas, por posición y prefijo |
| `prefijo-de-otro-prefijo` | sintético: `Ley 39/2015,` y `Ley 39/2015, de 1 de octubre` | error que nombra las dos entradas |
| `repositorio` | el manifiesto real, `grabaciones.json` de `testdata/evals/` de la raíz (`../../testdata/evals/grabaciones.json` desde el directorio del paquete, research.md V46) | sin error y con al menos diez entradas (plan.md, obligación 12) |

Los contenidos sintéticos son constantes del propio test y llegan con `grabaciones.go`, en el paso 3 del orden de
implementación (plan.md). El subtest `repositorio` llega en el paso 6, no antes: el manifiesto real lo crea el paso 4 y
el 5 es una tarea `[datos]` sin código, y en el paso 3 ese subtest rompería `make ci` por leer un fichero que aún no
existe.

### 3.2 Arnés y orden

- `internal/evals/grabacion_test.go`, `//go:build grabacion`, `TestGrabarEvals`. Falla antes de pedir nada si
  `KITLEGAL_RECORD` no vale `1` o si `LeerManifiesto` (§3.1) devuelve un error sobre el manifiesto.
- `scripts/grabar-evals.sh`:

```bash
#!/usr/bin/env bash
# Graba las respuestas del BOE que necesitan las evals de las skills y la verificación de data/normas.yaml, según
# testdata/evals/grabaciones.json, en testdata/evals/boe.legislacion-consolidada/. Toca la red: lo ejecuta una persona
# en la pausa de la tarea [datos] del manifiesto (contracts/evals-y-grabaciones.md §3.3 de H5). Ningún objetivo de
# make, gancho, tarea ni flujo lo ejecuta.
set -euo pipefail
cd "$(dirname "$0")/.."

KITLEGAL_RECORD=1 go test -tags=grabacion -count=1 -v -run '^TestGrabarEvals$' ./internal/evals/
```

- Qué hace `TestGrabarEvals`, por este orden:
  1. crea una caché temporal y, antes de pedir nada a la red, intenta servir en esa caché desde las **grabaciones de
     H4** con `Preparar` (§5.1, solo el conjunto H4) **toda** consulta de cada entrada: primero la búsqueda
     `buscar <busqueda>` y, una vez resuelto el identificador, sus metadatos, su índice y cada bloque. Solo lo que da una
     falta se pide a la red y se graba: lo que H4 ya grabó no se vuelve a pedir ni a grabar (FR-074: «se reutilizan las
     grabaciones de H4»), y por eso ninguna grabación nueva tiene el nombre de una de H4 (§3.4). La búsqueda
     `procedimiento administrativo común` de la primera entrada coincide con una de H4 y se sirve así (§3.1);
  2. compone `boe` con `httpx.New(httpx.ConFuente(boe.NombreDeLaFuente),
     httpx.ConRaizDeGrabacion("../../testdata/evals"), httpx.ConIntervalo(boe.IntervaloEntrePeticiones))` y esa caché
     —la raíz es `testdata/evals/` relativa al directorio del paquete, como `../../schemas` en `internal/app`; `httpx`
     exige que exista, y existe porque la tarea del manifiesto la creó, y escribe bajo ella
     `boe.legislacion-consolidada/` (research.md V46)—, y ejecuta en proceso con `app.Main` (la composición de §5.1,
     con este cliente en lugar de `httpx.Replay`), por cada entrada del manifiesto: `boe buscar <busqueda> --json`; elige el único resultado cuyo título empieza por el prefijo, con `strings.HasPrefix` (0 o más de
     1 → falla nombrando la entrada y los títulos); `boe metadatos <id> --json`, `boe indice <id> --json` y
     `boe articulo <id> <bloque> --json` por cada bloque. Todo código distinto de 0 falla nombrando la entrada y la orden;
  3. registra con `t.Logf` cada identificador resuelto y su título, que la persona copia a `data/normas.yaml` y a las
     evals.
- Es la aplicación de `kitlegal boe buscar` contra la fuente real que exige FR-023: la búsqueda y la elección salen del
  applet y de la respuesta, no de memoria.

### 3.3 Procedimiento de la pausa

La tarea `[datos]` escribe solo el manifiesto. En su pausa, una persona:

1. graba con `scripts/grabar-evals.sh`; si una búsqueda no devuelve la norma entre sus diez resultados, o un bloque
   termina en 3, corrige esa búsqueda, ese prefijo o ese bloque en el manifiesto y vuelve a grabar;
2. revisa en el registro de la orden cada identificador y título resueltos, y en el diff las grabaciones nuevas,
   comprobando que ninguna tiene el nombre de un fichero de `internal/source/boe/testdata/boe.legislacion-consolidada/`
   salvo `GET_https_www.boe.es_robots.txt.json` (§3.4);
3. confirma en la rama las grabaciones de `testdata/evals/boe.legislacion-consolidada/` y, si lo corrigió, el manifiesto,
   y aprueba.

La fila de `docs/SOURCES.md` de la fuente no cambia (spec, *Fuera de alcance*): el ritmo y los términos son los de H4.

Desde la tarea siguiente, ninguna línea de tarea contiene la ruta completa del manifiesto ni del directorio de
grabaciones de H5 (ni de un directorio que los contenga), por la regla del extractor de rutas del workflow (plan,
obligación 3).

### 3.4 Sin solape

`TestGrabacionesSinSolape`: los nombres de fichero de los conjuntos H4 y H5 solo coinciden en
`GET_https_www.boe.es_robots.txt.json`.

## 4. Consultas necesarias

`internal/evals.ConsultasNecesarias(conjunto []Eval) []Consulta`, data-model §7.1. `TestConsultasNecesarias` fija el
conjunto sin repetidos y el origen de cada consulta sobre una eval con las tres formas de comando, citas repetidas y dos
normas.

## 5. Preparación de caché y comprobación sin red (FR-074, FR-075)

### 5.1 Preparación

`internal/evals.Preparar(dirCache string, grabaciones []string, consultas []Consulta) ([]Falta, error)`, sin contexto:
cada consulta es una invocación entera de `app.Main`, la raíz de composición del binario, que abre el suyo con el plazo
de `--timeout` y no admite el de quien llama; un contexto que no llegara a las invocaciones sería una cancelación a
medias, que es lo que `contextcheck` rechaza en `make ci` (research.md D14 y V59):

1. copia en un temporal los conjuntos de `grabaciones` en ese orden (H4, `internal/source/boe/testdata/boe.legislacion-consolidada/`,
   y después H5, `testdata/evals/boe.legislacion-consolidada/`);
2. monta el registro con el valor cero de `app.Registro` y `registro.Registrar(app.AppletBoe(app.DependenciasDeBoe{…}))`,
   con `Cliente: func(*slog.Logger) (*httpx.Cliente, error) { return httpx.Replay(<temporal>) }` y
   `Cache: []cache.Opcion{cache.ConDirectorio(dirCache)}`;
3. ejecuta cada consulta en proceso con
   `app.Main([]string{"kitlegal", "boe", <verbo>, <args…>, "--json"}, &registro, &salida, &errores, <version>, <commit>, <fecha>)`
   (`internal/app/main.go` 127; research.md V26), con `salida` y `errores` búferes nuevos por consulta. No se usa
   `app.Arrancar`: recibe una función `construir func() (*Registro, error)` y no un registro ya montado
   (`internal/app/main.go` 78-83);
4. toda consulta con código distinto de 0 es una `Falta` con su eval, su punto (comando esperado, índice o metadatos de
   la norma, bloque de la cita) y el mensaje de la salida de error (que nombra la petición sin grabación).

El `error` es para lo que impide preparar, y no es una falta de lo grabado: un conjunto de grabaciones que no se puede
copiar, un temporal que no se puede crear o un applet que `Registrar` rechaza. Se devuelve con su causa y nombrando el
directorio o el applet; nunca se convierte en una lista de faltas vacía.

No modifica `internal/source/boe` ni el kernel: compone lo que ya exportan (spec, *Fuera de alcance*).

### 5.2 Comprobación sin red

`internal/evals.ComprobarSinRed(dirCache string, consultas []Consulta, opciones ...cache.Opcion) ([]Falta, error)`
(sin contexto, por lo mismo que `Preparar`; las opciones solo las usa el test para adelantar el reloj de la caché):
registro montado como en §5.1, con `boe` sobre
`httpx.Replay` de un **directorio vacío** (cualquier petición que se escapara fallaría en lugar de salir a la red) y la
misma caché más `opciones`; cada consulta con
`app.Main([]string{"kitlegal", "boe", <verbo>, <args…>, "--offline", "--json"}, &registro, …)`; código distinto de 0 →
`Falta` nombrando la eval y el comando, o la eval, la norma y lo que falta. El `error`, como en §5.1: solo lo que impide
comprobar (el directorio vacío que no se puede crear, el registro que no se monta).

`TestEvalsDelRepositorio/grabado` = `Preparar` + `ComprobarSinRed` sobre `evals/boe-legislacion/` y las dos grabaciones.
`TestPrepararYComprobar` (`internal/evals/preparar_test.go`, sin etiqueta) usa una copia de las grabaciones de H4, que
bastan para sus casos (metadatos, índice y `a21` de `BOE-A-2015-10565`), y la eval sintética `01-lpac-articulo-21.yaml`
(bloque y cita `BOE-A-2015-10565` `a21`), que el test escribe desde constantes en un `t.TempDir()`, como
`TestPrepararDirectorioDeSesion` (contrato del job §9): en el paso del plan que lo introduce todavía no existen las
grabaciones de H5 ni las evals del repositorio. Subtests, el completo y los negativos (SC-010):

| Subtest | Qué se retira | Mensaje exigido |
|---|---|---|
| `completo` | nada | ni faltas ni error |
| `sin-metadatos-de-un-bloque-esperado` | los metadatos de `BOE-A-2015-10565` | nombra `01-lpac-articulo-21.yaml` y `boe articulo BOE-A-2015-10565 a21` |
| `sin-indice-de-una-norma` | el índice de `BOE-A-2015-10565` (01 no espera `indice`) | nombra `01-lpac-articulo-21.yaml`, `BOE-A-2015-10565` e `indice` |
| `sin-bloque-de-una-cita` | el bloque `a21` | nombra la eval y la cita |
| `consulta-caducada` | (reloj de la caché adelantado 301 s con `cache.ConReloj` en la comprobación) | nombra `metadatos` y la eval |

## 6. Comparación de una sesión con su eval (FR-072)

`internal/evals.Juzgar(eval Eval, sesion Sesion, skill string) ResultadoDeEval`, data-model §10.2. Mecánica, sin modelo.

| Test | Qué fija |
|---|---|
| `TestInterpretarInvocacion` | `scripts/boe articulo N B`, `/ruta/kitlegal boe articulos N B1 B2`, banderas antes y después del verbo, `--timeout 5s` y `--timeout=5s`, `--asunto x`, `--describe` y `--dry-run` sin consulta, también `--describe=true`; con consulta con `--describe=false` y `--dry-run=0` y con un valor que el analizador rechaza (`--describe=quizá`); la ayuda sin consulta y fuera de los argumentos (`--help`, `-h` delante del verbo y `--help=false`, que Kong atiende igual); `kitlegal version` ignorada, applet desconocido ignorado (US4-7) |
| `TestExtraerCitas` | la forma fija con redacción alrededor; `a85bis.` antes del corchete; la forma legible dentro de los corchetes, delante del identificador, con las tres citas reales de las pruebas de red de T030, cada una con su pareja (`[Real Decreto Legislativo 2/2015, BOE-A-2015-11430, bloque a38]`, `[Constitución Española, BOE-A-1978-31229, bloque a140]` sola tras una transcripción y `[art. 20.1 de la LTAIBG, BOE-A-2013-12887, bloque a20]`); la anidada `[art. 118.2, LCSP [BOE-A-2017-12902, bloque a1-30]]`, una sola cita; no cuentan la forma, con o sin la forma legible dentro, sin sus dos corchetes en la misma línea o sin «bloque», con texto detrás del id ni con el identificador pegado a una letra o una cifra; varias citas en una línea (research D23, «La forma legible dentro de los corchetes») |
| `TestJuzgar` | un subtest por caso, con estos nombres (inventario de tests de plan.md): `pasa`; falla por activación (`positiva-no-activada`, y `no-activa-pero-activada`: eval de no activación en la que la skill se activó, US4-5); falla por **sesión sin terminar** aunque todo lo demás coincida, con `motivos` que nombra la eval y por qué no terminó (`sesion-sin-terminar-no-activa`: eval de no activación, sin activación, código 124 y sin mensaje `result`, que no puede pasar en vacío; `sesion-sin-terminar-positiva`: la sesión de la eval 01 con los comandos y la cita esperados, código 0 y último mensaje `result` con `subtype: error_max_turns`); `bloque-leido-con-codigo-4`: falla por bloque solo leído con código 4 (US4-2); `bloque-leido-sin-codigo`: falla por bloque solo leído por una invocación sin código, la que el tope dejó sin terminar (data-model §9, regla 6), que tampoco va a `fuera_de_lo_grabado` ni a `otras_fallidas`; `articulos-satisface-un-bloque`: `kitlegal boe articulos` satisface un bloque esperado (US4-7); `metadatos-no-satisface-indice` (US4-7); `otro-bloque-no-satisface`: la sesión de la eval 01 con `boe articulo BOE-A-2015-10565 a22 --json` y `boe articulos BOE-A-2015-10565 a20 a22 --json`, las dos con código 0, y la cita en la respuesta: comando ausente y `pasa` falso (FR-072, «ese identificador y ese bloque»); `otro-verbo-con-el-bloque-no-satisface`: igual con `boe buscar BOE-A-2015-10565 a21 --json`, que lleva la norma y el bloque pero no los lee; `ayuda-no-satisface`: igual con `boe articulo BOE-A-2015-10565 a21 --help` y `… -h`, sin consulta, que tampoco van a `fuera_de_lo_grabado` ni a `otras_fallidas` (data-model §9); `buscar-verbo-y-terminos-como-palabras`: un `buscar` esperado lo satisface la invocación de `buscar` cuyos argumentos, en minúsculas, contienen cada término como palabra, y no la que solo lo contiene dentro de otra palabra, ni la búsqueda a la que le falta el último de sus términos, ni una de otro verbo (`metadatos`) cuyos argumentos contienen el término (data-model §6.1); `cita-de-otro-bloque` y `cita-de-otra-norma`: una cita de la respuesta con otro bloque u otra norma no cuenta (US4-3); `fuera-de-lo-grabado-con-y-sin-offline`: invocaciones fuera de lo grabado con y sin `--offline` no cambian el resultado (US4-8, FR-076); `red-no-cambia-la-eval`: una invocación con una conexión de clase `red` no cambia `pasa`, que solo lo decide FR-072, y queda en `llegadas_a_la_red` para el veredicto global (data-model §10.2, FR-076); `describe-y-dry-run-no-satisfacen`: la sesión de la eval 01 con `boe articulo BOE-A-2015-10565 a21 --describe` y `boe articulo BOE-A-2015-10565 a21 --dry-run`, las dos con código 0, y la cita en la respuesta: comando ausente, `pasa` falso y ninguna de las dos en `fuera_de_lo_grabado` ni en `otras_fallidas`, porque no tienen consulta (data-model §6.1 y §9); `describe-y-dry-run-falsos-consultan`: la sesión de la eval 01 con `boe articulo BOE-A-2015-10565 a9998 --dry-run=false --offline --json` terminada con código 4 y `boe articulo BOE-A-2015-10565 a21 --describe=false --json` con código 0, y la cita: las dos tienen consulta, así que la primera va a `fuera_de_lo_grabado` y la segunda satisface el comando esperado, y la eval pasa (data-model §9); `otra-fallida`: la sesión de la eval 01 con `boe articulo BOE-A-15-1 a21 --json` terminada con código 2 y `boe articulo BOE-A-2015-10565 a9999 --json` con código 3, además de la lectura de `a21` con código 0 y la cita: las dos primeras en `otras_fallidas`, con `orden` y `codigo`, ninguna en `fuera_de_lo_grabado`, y la eval pasa (data-model §10.2, FR-076); y **SC-009** (`sc-009-otro-bloque`, `sc-009-otra-norma`): la misma sesión pasa con la eval real y falla con la cita esperada cambiada a `a22` y a otra norma |
