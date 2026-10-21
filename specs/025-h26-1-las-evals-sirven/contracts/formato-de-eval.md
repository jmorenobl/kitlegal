# Contrato: `lecturas` y `sin_servir` en el formato de las evals

Quién lo consume: quien escribe una eval (una persona o una tarea de un hito), `LeerConjunto` y, por él, la preparación
de las sesiones, el juicio y `TestEvalsDelRepositorio`. Requisitos: FR-006 a FR-014, FR-017, FR-018.

## 1. El esquema (`schemas/eval.yaml.json`)

Dos propiedades nuevas en el objeto raíz, que sigue con `additionalProperties: false`, y dos definiciones:

```json
"lecturas": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/lectura" } },
"sin_servir": { "type": "array", "minItems": 1, "items": { "$ref": "#/$defs/lectura-sin-servir" } }
```

```json
"lectura": {
  "type": "object", "additionalProperties": false, "required": ["norma", "bloque", "motivo"],
  "properties": {
    "norma": { "$ref": "#/$defs/norma" }, "bloque": { "$ref": "#/$defs/bloque" },
    "motivo": { "type": "string", "pattern": "\\S" }
  }
},
"lectura-sin-servir": {
  "oneOf": [
    { "type": "object", "additionalProperties": false, "required": ["norma", "bloque"],
      "properties": { "norma": { "$ref": "#/$defs/norma" }, "bloque": { "$ref": "#/$defs/bloque" } } },
    { "type": "object", "additionalProperties": false, "required": ["norma", "verbo"],
      "properties": { "norma": { "$ref": "#/$defs/norma" }, "verbo": { "enum": ["indice", "metadatos"] } } }
  ]
}
```

`"\\S"` exige algún carácter que no sea blanco: el validador del módulo busca la coincidencia en cualquier punto de la
cadena (research V21). Ninguna otra regla del esquema cambia. Lo modifica una tarea `[datos]` (`schemas/`).

## 2. Lo que rechaza `LeerEval` detrás del esquema

Cada uno hace de la eval un fichero mal formado de `LeerConjunto`; el error empieza, como todos los de `LeerEval`, por el
nombre del fichero (FR-008, FR-014):

| La eval lleva | Error (fragmento) |
|---|---|
| una entrada de `lecturas` cuyo bloque —misma norma y mismo bloque— es un comando de bloque, con o sin `fecha` | `lecturas: BOE-A-2015-10566 a25: ya es un comando esperado de la eval` |
| una entrada de `lecturas` cuyo bloque es una de sus `citas` | `lecturas: BOE-A-2015-10566 a25: ya es una cita esperada de la eval` |
| un bloque de `sin_servir` que es una de sus `citas` | `sin_servir: BOE-A-2015-10565 a124: es una cita esperada de la eval` |

El motivo vacío o solo de blancos, otro `verbo`, una entrada con `bloque` y `verbo` a la vez, o sin `motivo`, los
rechaza el esquema, con la ruta de la entrada (`/lecturas/0/motivo`, `/sin_servir/0`…). Un bloque de `sin_servir` que es
un comando de bloque de la misma eval se admite (FR-016).

## 3. Las evals del hito

### 3.1 `evals/legal-core/11-plazo-reposicion-sin-leer-el-precepto.yaml` (FR-017)

```yaml
# Eval de plazo (e): la pregunta de la 05 con la lectura del art. 124 de la LPAC sin servir (sin_servir), como con el
# BOE inalcanzable. El paso 10 pide leer el precepto en la fecha de inicio: la sesión lo intenta y la lectura sale con
# 4 o 5. La respuesta tiene que dar igual la fecha de plazos en la línea ⚠ PLAZO:, con la cita de cada precepto de
# data.preceptos, y decir aparte que no pudo leer esa redacción (ADR 0043, punto 6; FR-017, FR-029). Sin citas ni
# redaccion_fechada: lo que cuenta es la línea del plazo y la lectura intentada (FR-016, FR-033).
pregunta: "Me notificaron el 9 de septiembre de 2026 una resolución del Ayuntamiento de Leganés y vivo en Leganés, ¿hasta cuándo puedo presentar el recurso de reposición?"
activa: true
comandos:
  - applet: territorio
    verbo: resolver
    municipio: Leganés
  - applet: plazos
    verbo: calcular
    tipo: reposicion
  - applet: boe
    norma: BOE-A-2015-10565
    bloque: a124
    fecha: "2026-09-09"
sin_servir:
  - norma: BOE-A-2015-10565
    bloque: a124
plazo:
  fecha: "2026-10-13"
  avisos: []
  razonamiento: >-
    <el de la 05, tal cual>
```

### 3.2 `evals/legal-core/12-cobertura-seguridad-social-sin-leer-la-da-primera.yaml` (FR-018)

```yaml
# Eval de cobertura (f): un recurso contra una reclamación de cuotas de la Seguridad Social, el ejemplo del paso 7, que
# la DA 1ª.2.b de la LPAC deja fuera del procedimiento común, con la DA 1ª sin servir (sin_servir). La línea ⚠ NO
# CUBIERTO: no afirma lo que dice ningún precepto y va aunque falle la lectura: la de la entrada sin norma de
# data/cobertura.yaml, con la LPAC, y ningún plazo (ADR 0043, punto 7; FR-018, FR-030).
pregunta: "Vivo en Leganés y el 9 de septiembre de 2026 me notificaron una reclamación de cuotas de la Seguridad Social. ¿Hasta cuándo puedo recurrirla?"
activa: true
sin_servir:
  - norma: BOE-A-2015-10565
    bloque: daprimera
prohibidos:
  - applet: plazos
    verbo: calcular
no_cubierto:
  norma: BOE-A-2015-10565
```

### 3.3 `lecturas` de las evals 14, 15 y 16 de `boe-legislacion` (FR-011)

Se añade la clave detrás de `citas`; nada más cambia en ellas. Los motivos dicen por qué la eval provoca la lectura y
no afirman nada de lo que dice el bloque:

```yaml
# 16-lrjsp-legalidad-por-materia.yaml
lecturas:
  - norma: BOE-A-2015-10566
    bloque: a26
    motivo: "Junto al art. 25 que pide la eval, en su mismo capítulo: la exploración que provoca la pregunta, vista en los jobs de la #175 y la #178 (ADR 0043, clase b)."
  - norma: BOE-A-2015-10566
    bloque: a27
    motivo: "Junto al art. 25 que pide la eval, en su mismo capítulo: la exploración que provoca la pregunta, vista en los jobs de la #175 y la #178 (ADR 0043, clase b)."
```

```yaml
# 15-irpf-rendimientos-por-materia.yaml
lecturas:
  - norma: BOE-A-2006-20764
    bloque: a18
    motivo: "Junto al art. 17 que pide la eval: la exploración que provoca la pregunta, vista en los jobs de la #175 y la #178 (ADR 0043, clase b)."
```

```yaml
# 14-trlrhl-impuestos-por-materia.yaml
lecturas:
  - norma: BOE-A-2004-4214
    bloque: a15
    motivo: "De la misma norma que el art. 59 que pide la eval: la exploración que provoca la pregunta, vista en los jobs de la #175 y la #178 (ADR 0043, clase b)."
```

## 4. El manifiesto de grabación (`internal/evals/testdata/grabaciones.json`, FR-010)

Tres entradas al final de `normas`, en una tarea `[datos]` que no lleva otro trabajo:

```json
{"busqueda": "régimen jurídico del sector público", "titulo_empieza_por": "Ley 40/2015,", "bloques": ["a26", "a27"], "para": "lecturas de la eval 16 de boe-legislacion (H26.1, FR-010, FR-011; ADR 0043, clase b): los arts. 26 y 27 de la Ley 40/2015, con el id de bloque que da su índice grabado. La búsqueda, los metadatos y el índice ya están grabados en H5: solo los bloques son nuevos"},
{"busqueda": "impuesto renta personas físicas", "titulo_empieza_por": "Ley 35/2006,", "bloques": ["a18"], "para": "lecturas de la eval 15 de boe-legislacion (H26.1, FR-010, FR-011; ADR 0043, clase b): el art. 18 de la Ley 35/2006, con el id de bloque que da su índice grabado. La búsqueda, los metadatos y el índice ya están grabados en H5: solo el bloque es nuevo"},
{"busqueda": "reguladora de las haciendas locales", "titulo_empieza_por": "Real Decreto Legislativo 2/2004,", "bloques": ["a15"], "para": "lecturas de la eval 14 de boe-legislacion (H26.1, FR-010, FR-011; ADR 0043, clase b): el art. 15 del texto refundido de la Ley reguladora de las Haciendas Locales, con el id de bloque que da su índice grabado. La búsqueda, los metadatos y el índice ya están grabados en H5: solo el bloque es nuevo"}
```

Tras su commit, `grabar_datos` ejecuta `TestGrabarEvals` (`//go:build grabacion`, `internal/evals/grabacion_test.go`)
con `KITLEGAL_RECORD=1`, que pide con `internal/httpx` a `boe.legislacion-consolidada` (fila revisada el 2026-09-13 en
`main`) solo los cuatro bloques —lo demás lo sirven H4 y H5— y deja las cuatro respuestas
`GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_<norma>_texto_bloque_<bloque>.json` en
`internal/evals/testdata/boe.legislacion-consolidada/`. No hay material de origen de ningún fichero de `data/`: nada va a
`evidencias/`.
