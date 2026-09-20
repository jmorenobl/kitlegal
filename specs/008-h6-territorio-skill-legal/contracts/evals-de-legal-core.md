# Contrato: el formato común de eval que H6 amplía y las evals de `legal-core`

El formato de eval es **común a todas las skills** desde H5 y aquí se extiende de forma compatible hacia atrás: las
evals de `boe-legislacion` siguen siendo válidas sin tocar un byte y sus reglas del conjunto siguen en verde
(FR-084, SC-015).

## 1. `schemas/eval.yaml.json`: los tres cambios

Fichero **existente** bajo `schemas/`: se modifica en una tarea `[datos]` con pausa humana (FR-085). Esa tarea lleva
consigo el único cambio de código que la forma del esquema impone —la expectativa de `TestLeerEval` que §1.3 explica—,
como excepción razonada del tipo D16 (research.md D28).

### 1.1 Cuarta variante de `comandos`

```json
"comando-territorio": {
  "type": "object",
  "additionalProperties": false,
  "required": ["applet", "verbo", "municipio"],
  "properties": {
    "applet":    { "$ref": "#/$defs/applet" },
    "verbo":     { "const": "resolver" },
    "municipio": { "type": "string", "minLength": 1 }
  }
}
```

Entra como un `$ref` más del `oneOf` de `comandos`. No colisiona con las tres existentes: todas declaran
`additionalProperties: false` y `municipio` no es propiedad de ninguna, y la variante de consulta de norma tiene su
`verbo` en un enumerado cerrado que no incluye `resolver` (research.md V28).

### 1.2 Esperado de territorio

```json
"territorio": {
  "type": "object",
  "additionalProperties": false,
  "minProperties": 1,
  "properties": {
    "comunidad":  { "type": "string", "minLength": 1 },
    "provincia":  { "type": "string", "minLength": 1 },
    "boletines":  { "type": "array", "minItems": 1, "items": { "type": "string", "minLength": 1 } },
    "cobertura":  { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/aspecto-de-cobertura" } }
  }
}
```

`aspecto-de-cobertura` es un enumerado cerrado con las seis combinaciones de los tres aspectos y sus valores
(`boletin_autonomico: configurado`, `boletin_autonomico: no-configurado`, `boletin_provincial: …`, `dir3: verificado`,
`dir3: no-verificado`), derivado del vocabulario de [data-model.md](../data-model.md) §2.4. Un test exige que el
enumerado del esquema y el vocabulario del applet digan lo mismo, como hace H5.1 con los códigos de aviso.

### 1.3 La regla del esperado verificable

El bloque `if/then/else` pasa de

```json
"then": { "required": ["comandos", "citas"] }
```

a

```json
"then": {
  "required": ["comandos"],
  "anyOf": [{ "required": ["citas"] }, { "required": ["territorio"] }]
}
```

y el `else` gana `territorio` a la lista de lo que una eval de no activación no puede declarar. Sigue siendo cierto
que **toda eval que afirme contenido de norma declara `citas`**: quien lo afirma es quien las declara, y las once
evals de `boe-legislacion` que hoy las traen siguen validando sin cambio (compatibilidad hacia atrás, FR-084).

**Lo que este cambio arrastra.** Una eval activa con `comandos` y sin `citas` ni `territorio` sigue siendo inválida,
pero **el defecto deja de ser uno**: el nodo `anyOf` tiene causas, así que el lector común desciende a las hojas de sus
dos ramas y las une (research.md V44). El caso `positiva-sin-citas` de `TestLeerEval`, que compara el mensaje con
`EqualError`, pasa en esta misma tarea a llamarse `positiva-sin-esperado-verificable` y a declarar el mensaje que el
esquema nuevo produce; si la composición del mensaje exigiera tocar `internal/evals/formato.go`, va también aquí
(research.md D28). No se relaja la comparación: el mensaje es parte del contrato que la tarea cambia.

## 2. Go: `internal/evals`

| Cambio | Dónde | Qué |
|---|---|---|
| Campo `Municipio` | `formato.go`, `ComandoEsperado` | Un campo más del `struct` plano, etiqueta `municipio` |
| Campo `Territorio` | `formato.go`, `Eval` | El esperado de §1.2 |
| `formaDelComando` | `formato.go` | **Un único sitio** que decide la variante; lo consumen el juicio, el texto del comando y las consultas necesarias, en lugar de los tres `switch` por verbo de hoy (research.md D21, V29) |
| `satisface` | `juzgar.go` | Una invocación satisface un comando de territorio si su applet es `territorio`, su verbo `resolver` y su argumento, plegado con `territorio.Plegar`, es el municipio esperado |
| `textoDelComando` | `juzgar.go` | `territorio resolver <municipio>` |
| `ConsultasNecesarias` | `consultas.go` | Un comando de territorio **no genera consulta**: no hay nada que grabar (FR-043) |
| `ExtraerTerritorio` | `territorio.go` (nuevo, hermano de `citas.go` y `avisos.go`) | Saca de la respuesta lo declarado, por forma fija |
| `TerritorioEncontrado` / `TerritorioAusente` | `juzgar.go`, `ResultadoDeEval` | Reparto, con su motivo |
| Condición de `Pasa` | `juzgar.go` | Gana `len(TerritorioAusente) == 0` |
| Columnas del informe | `informe.go` | «Territorio encontrado» y «Territorio ausente», como las de avisos |

**Forma fija de lo esperado** (constitución, capa 1: nada de juicio de modelo ni de similitud):

| Esperado | Se da por encontrado si la respuesta contiene |
|---|---|
| `comunidad`, `provincia` | el nombre tal cual, plegado con `territorio.Plegar` como los nombres de municipio (sin distinguir mayúsculas ni diacríticos): un solo pliegue en el árbol, el del dominio (research.md D10, D27) |
| `boletines` | el código del boletín (`BOCM`), exacto y como palabra |
| `cobertura` | la clave del aspecto y su valor, en la forma `<aspecto>: <valor>`, con tolerancia a los espacios y a las mayúsculas y exacta en las dos palabras |

Igual que la cita se compara por identificador y el aviso por su etiqueta: lo que se mide es una regla escrita de la
skill, no una redacción.

## 3. Reglas del conjunto

`ComprobarConjuntoDeBoeLegislacion` pasa a `ComprobarConjunto(evals, normas, reglas)` con dos juegos (research.md
D22). Las diez reglas de `boe-legislacion` **no cambian**.

`ReglasDeLegalCore()`:

| Regla | Exige |
|---|---|
| `tamaño` | Al menos tres evals bien formadas |
| `cubierto` | Al menos una eval activa cuyo comando de territorio resuelve un municipio del territorio configurado y cuyo esperado declara sus boletines |
| `no cubierto` | Al menos una eval activa sobre un municipio de una comunidad sin configuración, cuyo esperado declara los aspectos de cobertura no configurados |
| `no activación` | Al menos una eval con `activa: false` |
| `esperado verificable` | Toda eval activa declara `citas` o `territorio` |

`TestEvalsDelRepositorio` gana el subtest que las aplica a `evals/legal-core/`. El subtest `formato`, que ya recorre
todas las carpetas de `evals/`, cubre el formato desde el primer fichero (research.md V26).

**Cuándo entra cada cosa** (plan, «Orden de implementación»): `ComprobarConjunto` con sus dos juegos de reglas y
`TestConjuntoDeEvals` sobre evals sintéticas, en el paso 18; el subtest `conjunto-legal-core`, en el paso 19 **con las
tres evals que lo hacen pasar**, primero el subtest y después los ficheros, porque un control que exige tres evals no
puede entrar antes que ellas sin quedar en rojo ni después sin pasar en vacío (obligación 12 del plan). El subtest
`cobertura-del-esquema` entra en el paso 18: el enumerado del esquema existe desde el paso 17 y el vocabulario del
applet desde el paso 9.

## 4. Las evals de `legal-core`

Mínimo del hito (FR-080 a FR-082), escritas **antes** que `SKILL.md` (FR-083):

| Fichero | Qué mide |
|---|---|
| `evals/legal-core/01-territorio-municipio-cubierto.yaml` | La pregunta de la aceptación sobre el municipio de referencia: activa, comando `territorio resolver`, esperado con comunidad, provincia y boletín |
| `evals/legal-core/02-territorio-municipio-no-cubierto.yaml` | La misma pregunta sobre un municipio de una comunidad sin configuración: esperado con comunidad, provincia y los dos aspectos `no-configurado` |
| `evals/legal-core/03-no-activa-…​.yaml` | Una pregunta que no es de territorio ni de derecho: `activa: false` |

Ninguna declara `citas`: `legal-core` no afirma contenido de norma (FR-069). Los municipios concretos viven aquí y en
los fixtures, nunca en el código ni en `data/` (FR-024).

## 5. El job

`.github/workflows/evals.yml` pasa a una **matriz de skills** (`boe-legislacion`, `legal-core`) con
`fail-fast: false`, un informe por skill y la misma ejecución del flujo (research.md D23, S6), y su filtro de rutas
gana `internal/core/*` —`data/territorio/…` ya lo cubre el patrón `data/*` que tiene (V31)—. El umbral no cambia: una
eval pasa cuando al menos 2 de sus 3 sesiones pasan, y el informe lleva el commit evaluado y el identificador del
modelo (ADR 0016, SC-015).

**La prueba de red sigue siendo solo de `boe-legislacion`**: su texto invoca el applet `boe` para provocar dos
invocaciones fuera de lo grabado (`internal/evals/preparar.go`, `textoDeLaPruebaDeRed`), y `territorio` no tiene red
que probar —no puede pedir nada por construcción (FR-043)—. El trabajo de la matriz de `legal-core` no la ejecuta, y
eso se decide en el flujo, no en Go: es la misma bandera `prueba_de_red` que ya existe, aplicada a una sola skill.

## 6. Qué lo vigila

| Control | Test | En `make ci` |
|---|---|---|
| Formato de toda eval de toda skill | `TestEvalsDelRepositorio/formato`, `TestLeerEval`, `TestEsquemaDeEval` | sí |
| Compatibilidad hacia atrás | `TestLeerEval` sobre las evals de `boe-legislacion` sin modificar, y `TestEvalsDelRepositorio/conjunto` con sus diez reglas | sí |
| Variante de territorio y su juicio | `TestJuzgar/territorio-*`, `TestFormaDelComando` | sí |
| Que un esperado de territorio ausente impide pasar | `TestJuzgar/territorio-ausente` | sí |
| Que el enumerado de cobertura del esquema coincide con el del applet | `TestEvalsDelRepositorio/cobertura-del-esquema` | sí |
| Reglas del conjunto de `legal-core` | `TestConjuntoDeEvals`, `TestEvalsDelRepositorio/conjunto-legal-core` | sí |
| Que las evals pasan de verdad | job de evals con modelo (ADR 0016) | no (job) |
