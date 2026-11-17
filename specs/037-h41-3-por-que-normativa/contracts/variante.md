# Contrato: la variante `producto-descripciones-cortas`

FR-001, FR-002, FR-033; SC-001. Research D1 a D4; data-model §1.

## 1. Dónde y qué forma tiene

`laboratorio/producto-descripciones-cortas/`, fuera de `skills/`, de `evals/` y de lo que empotra el binario, con un
fichero por skill de `skills/` y nada más:

```text
laboratorio/producto-descripciones-cortas/
├── consulta-juridica.txt
├── jurisprudencia.txt
├── normativa.txt
└── plazos-y-territorio.txt
```

El texto de la `description` de una skill es su fichero sin el salto de línea final; cada fichero es una sola línea
terminada en `\n`. Nada más de ninguna skill vive aquí: ni cuerpos, ni referencias, ni otras claves del frontmatter
(FR-001). Lo usa solo el arnés, con `INSTALACION=producto-descripciones-cortas` (contracts/arnes.md §3); ninguna skill del
producto cambia (FR-033).

## 2. Los cuatro textos

Contados en runas (research V2). Cada fichero lleva exactamente esta línea y su salto de línea.

### `consulta-juridica.txt` — 417 (hoy 705)

```text
Punto de entrada de toda pregunta legal en España, de cualquier materia: dice qué capacidad de kitlegal la resuelve —normativa, jurisprudencia, plazos y territorio— y con qué skill, y lo que no cubre lo dice sin dar de memoria plazos, órganos ni vías. Se usa al empezar toda pregunta legal, antes que las demás skills de kitlegal, también si no nombra norma o pide qué hacer, a quién reclamar o qué escrito presentar.
```

- **Conserva**: punto de entrada de toda pregunta legal en España, de cualquier materia (su capacidad); qué capacidad de
  kitlegal la resuelve —normativa, jurisprudencia, plazos y territorio— y con qué skill; lo que no cubre, sin plazos,
  órganos ni vías de memoria; se usa al empezar toda pregunta legal, antes que las demás skills de kitlegal; cuando la
  pregunta no nombra norma; cuando pide qué hacer, a quién reclamar o qué escrito presentar.
- **Deja**: «sobre una situación legal» (queda «pregunta legal»); los ejemplos de materia («un seguro, un contrato, una
  multa, un recurso contra una administración, una herencia, una sentencia o un artículo de una ley»); «dice qué tipo de
  pregunta es»; «qué cubre la base y qué no»; «también cuando el agente cree saber la respuesta».

### `jurisprudencia.txt` — 420 (hoy 842)

```text
Sentencias y jurisprudencia de los tribunales españoles, sin citar ninguna de memoria. Se usa siempre que la pregunta nombre, pida o cite una sentencia (STS, ECLI, ROJ), pregunte si existe o qué dice, pida jurisprudencia sobre una materia o traiga una resolución, aunque el agente crea conocerla. kitlegal no consulta el CENDOJ: prepara la búsqueda y coteja el documento. Se activa antes de cita_preparar o cita_cotejar.
```

- **Conserva**: sentencias y jurisprudencia de los tribunales españoles, sin citar ninguna de memoria (su capacidad); se
  usa siempre que la pregunta nombre, pida o cite una sentencia, con las formas STS, ECLI y ROJ; pregunte si existe o qué
  dice; pida jurisprudencia sobre una materia; traiga una resolución; aunque el agente crea conocerla; kitlegal no consulta
  el CENDOJ, prepara la búsqueda y coteja el documento; se activa antes de `cita_preparar` o `cita_cotejar`.
- **Deja**: «o resuma»; «o un auto»; los ejemplos literales («STS 1088/2023, de 4 de julio», «la sentencia del Supremo
  sobre…», «qué dice la jurisprudencia sobre la cláusula suelo»); «el texto o el PDF»; «una sentencia solo se cita con su
  documento delante»; «para que la persona la busque con su navegador»; «qué pedir, qué decir de la sentencia que no está
  delante y cómo citar la que sí lo dice la skill».

### `normativa.txt` — 426 (hoy 887)

```text
Consulta y cita normativa consolidada del BOE de cualquier materia: lee con kitlegal índice y artículos y cita identificador y bloque. Se usa siempre que la respuesta dependa de lo que dice una norma —un artículo, ley o real decreto, el plazo o requisito que fija— o de lo que decía en una fecha, aunque el agente crea saberlo. Se activa antes de boe_articulo u otra herramienta boe_…: qué pedir y cómo citar lo dice la skill.
```

- **Conserva**: consulta y cita normativa consolidada del BOE de cualquier materia (su capacidad); lee con kitlegal índice
  y artículos y cita identificador y bloque; se usa siempre que la respuesta dependa de lo que dice una norma —un
  artículo, ley o real decreto, el plazo o requisito que fija— o de lo que decía en una fecha; aunque el agente crea
  saberlo; se activa antes de `boe_articulo` u otra herramienta `boe_…`, y qué pedir y cómo citar lo dice la skill.
- **Deja**: «(Boletín Oficial del Estado)» desarrollado; la lista de materias («procedimiento administrativo, contratación
  pública, régimen local, tributos, transparencia, relaciones laborales, seguros…»); «o procedimiento», «dónde se regula
  una materia»; «la del acto o de los hechos»; «sin leer la norma con kitlegal no hay cita»; la frase de cómo se nombra la
  norma («estatal o autonómica», número y año, las abreviaturas LPAC, LCSP, LRBRL, LGT, TRLRHL, LRJSP e identificador
  BOE-A-…). La pregunta de la eval 04 nombra «el artículo 66 de la Ley General Tributaria», que cubre «un artículo, ley»
  (research D2).

### `plazos-y-territorio.txt` — 427 (hoy 879)

```text
Calcula con kitlegal hasta qué día se puede recurrir una resolución administrativa, cuándo vence un plazo del procedimiento común (LPAC) o el del contencioso y cuándo hay silencio, con los festivos que cuentan, e identifica el territorio de un municipio: INE, provincia, comunidad, DIR3 y boletines. No cuenta plazos de otras materias. Se usa al preguntar cuándo vence un plazo o un recurso, o qué boletines tiene un municipio.
```

- **Conserva**: calcula con kitlegal hasta qué día se puede recurrir una resolución administrativa, cuándo vence un plazo
  del procedimiento común (LPAC) o el del contencioso y cuándo hay silencio, con los festivos que cuentan (su capacidad);
  identifica el territorio de un municipio: INE, provincia, comunidad, DIR3 y boletines; no cuenta plazos de otras
  materias; se usa al preguntar cuándo vence un plazo o un recurso, o qué boletines tiene un municipio.
- **Deja**: «de los municipios» y «régimen» en la lista del territorio; «hoy … y la interposición del contencioso» en la
  cobertura; la lista de lo que la ley regula aparte («tributos, multas de tráfico, Seguridad Social, extranjería»); y,
  de «Se usa…», «hasta cuándo se puede recurrir» (lo dice ya la primera frase), «cuándo se produce el silencio», «qué
  comunidad o provincia», «si un municipio es foral» y «qué norma prevalece».

Total: 1690 runas de 1700; cada una por debajo de 450; ninguna con `<`, así que ninguna con una etiqueta XML. La
tercera persona y lo que dice cada texto no se juzgan en este hito (FR-001; ADR 0046, punto 4).

## 3. La comprobación (FR-002)

`descripcionesCortas(raiz)` (research D3). Las skills son los subdirectorios de `<raiz>/skills/`; la variante, las
entradas de `<raiz>/laboratorio/producto-descripciones-cortas/`. Un defecto por línea, en este orden: por cada skill, en
orden de nombre, el que le falte o, si está, el de su longitud y el de su etiqueta; después, por cada entrada de la
variante que no es `<skill>.txt` de una skill, en orden de nombre; y al final el de la suma. Cada línea empieza por la
ruta de la variante:

| Caso | Defecto |
|---|---|
| a una skill le falta su fichero | `laboratorio/producto-descripciones-cortas: falta la descripción de <skill>` |
| una entrada no es `<skill>.txt` de una skill de `skills/` | `laboratorio/producto-descripciones-cortas: «<entrada>» no es la descripción de ninguna skill de skills/` |
| una tiene más de 450 runas | `laboratorio/producto-descripciones-cortas: la descripción de <skill> tiene <n> caracteres (máximo 450)` |
| las que hay suman más de 1700 | `laboratorio/producto-descripciones-cortas: las descripciones suman <n> caracteres (máximo 1700)` |
| una tiene una etiqueta XML | `laboratorio/producto-descripciones-cortas: la descripción de <skill> tiene la etiqueta XML «<etiqueta>» (no admite ninguna)` |

«Caracteres» son runas, como en los límites de la `description` del producto (ADR 0046, punto 2), y la etiqueta es la
primera que encuentra `skills.PrimeraEtiquetaXML`, la definición del punto 3 (research D4). Una entrada repetida no
tiene forma de darse: un directorio no tiene dos con el mismo nombre (research D1). Un directorio de la variante o de
`skills/` que no se puede leer no es un defecto: es el error de la función, defecto `inesperado` (ADR 0023).

Ejemplos, en un árbol con las cinco skills `consulta-juridica`, `jurisprudencia`, `normativa`, `plazos-y-territorio` y
`quinta`, y la variante de §2 sin `plazos-y-territorio.txt`, con `normativa.txt` de 451 runas que empieza por
`<ejemplo>` y con `otra.txt`:

```text
laboratorio/producto-descripciones-cortas: la descripción de normativa tiene 451 caracteres (máximo 450)
laboratorio/producto-descripciones-cortas: la descripción de normativa tiene la etiqueta XML «<ejemplo>» (no admite ninguna)
laboratorio/producto-descripciones-cortas: falta la descripción de plazos-y-territorio
laboratorio/producto-descripciones-cortas: falta la descripción de quinta
laboratorio/producto-descripciones-cortas: «otra.txt» no es la descripción de ninguna skill de skills/
```

Con cuatro de 450, 450, 400 y 401, el único defecto es `… las descripciones suman 1701 caracteres (máximo 1700)`.

## 4. Controles

`TestDescripcionesCortas` (`internal/evals/laboratorio_descripciones_test.go`, en `make ci`): la variante del
repositorio, sin ningún defecto y con una descripción por skill de `skills/`; y, sobre árboles que el
test monta en `t.TempDir()`, cada fila de la tabla de §3 con su línea exacta: falta una, sobra una (`otra.txt`), 451
runas, 1701 en total, `<ejemplo>` y una quinta skill sin descripción. Falla con la primera línea que difiere
(SC-001).

## 5. Uso, de fuera adentro

Los cuatro textos los lee el modelo de cada sesión de la fila de remedio una vez, en el listado de skills del arranque:
160 sesiones del sondeo. Ninguna `SKILL.md` del producto los lee: 0 invocaciones por pregunta. Tamaño fijo, 1690 runas
(1732 bytes), que no crece con el uso del producto —cientos de normas y miles de bloques consultados no lo cambian—. La
comprobación de §3 la lee quien hace `make ci`, solo en rojo: una línea por defecto, como mucho una por cada entrada y
dos por skill, más la de la suma; deja de darse al corregir la variante, y vuelve a darse si un hito añade una skill a
`skills/` sin su descripción. La variante deja de usarse cuando el ADR que lea el sondeo decide.
