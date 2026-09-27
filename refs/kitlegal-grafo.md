# kitlegal: el grafo legal

Tercer documento de la serie (1: mapa del sistema legal y skills; 2: estructura del proyecto y ecosistema). Este cubre el grafo: qué representa, cómo se alimenta solo por uso, cómo se consulta, cómo valida y cómo detecta anomalías. Sin monetización.

---

## 1. Idea en una frase

Cada vez que un applet de `kitlegal` toca una fuente, deja un rastro estructurado: nodos con identificadores naturales y aristas con nombre. Con el tiempo el rastro se convierte en dos grafos: el **mundo** (normas, órganos, entidades, resoluciones: público, reproducible) y el **asunto** (hechos, plazos, escritos, citas: privado, en el directorio de trabajo). El asunto apunta al mundo; el mundo nunca apunta al asunto.

Es el mismo papel que GOLEM en Bookwright: memoria estructurada + validación de continuidad. La diferencia es que aquí la ontología base no hay que inventarla: se adopta ELI (European Legislation Identifier, ontología OWL que el BOE ya usa para sus URIs) y se extiende para lo que ELI no cubre (actividad administrativa, entidades, asunto).

---

## 2. Modelo

### 2.1 Tipos de nodo

Regla: todo nodo tiene un **id natural** externo cuando existe. Así dos ejecuciones nunca duplican y el grafo es fusionable con el de cualquier otro usuario.

| Tipo | Id natural | Ejemplo | Origen (applet) |
|---|---|---|---|
| `Norma` | ELI | `eli/es/l/2015/10/01/39` | boe, eurlex, boletin |
| `NormaVersion` | ELI + fecha de consolidación | `eli/es/l/2015/10/01/39/con/20260301` | boe metadatos |
| `Bloque` | ELI + id de bloque BOE | `…/39#a21` | boe articulo/indice |
| `BloqueVersion` | Bloque + fecha_vigencia + hash del texto | `…#a21@20231001:sha256…` | boe articulo |
| `Resolucion` | ECLI | `ECLI:ES:TS:2025:1234` | ecli, tc |
| `Doctrina` | id DGT / TEAC / Consejo de Estado | `DGT:V1234-24` | dgt, teac |
| `Organo` | código DIR3 | `L01280740` (Ayto. Leganés) | territorio, placsp, bdns, sede |
| `Entidad` | NIF | `B12345678` | placsp, bdns, borme |
| `Persona` | nombre normalizado (sin NIF; ver §7) | `persona:garcia-lopez-juan` | borme |
| `Licitacion` | id expediente PLACSP | `placsp:2026/00123` | placsp |
| `Contrato` | id contrato PLACSP | `placsp:contrato:…` | placsp |
| `Convocatoria` | código BDNS | `bdns:812345` | bdns |
| `Concesion` | id concesión BDNS | `bdns:concesion:…` | bdns |
| `Inmueble` | referencia catastral | `rc:1234567AB1234C` | catastro |
| `Publicacion` | id BOE/BORME/boletín | `BOE-B-2026-12345` | boe sumario, edictos, boletin |
| `Municipio` | código INE | `ine:28074` | territorio |
| `Materia` | código materia BOE / EuroVoc | `materia:tributos` | boe |
| **Asunto** | uuid local | `asunto:…` | (privado) |
| `Hecho` | uuid | | (privado) |
| `Parte` | uuid → puede enlazar a Entidad/Organo | | (privado) |
| `Plazo` | uuid | | plazos |
| `Escrito` | uuid + hash del fichero | | escrito |
| `Afirmacion` | uuid (una frase de un escrito que necesita fundamento) | | escrito |
| `Consulta` | uuid + timestamp (cada llamada a un applet) | | todos |

### 2.2 Tipos de arista

Donde ELI tiene nombre, se usa el de ELI. El resto, prefijo `lb:`.

| Arista | De → A | Fuente de la arista |
|---|---|---|
| `eli:amends` / `eli:amended_by` | Norma → Norma | boe analisis |
| `eli:repeals` / `eli:repealed_by` | Norma → Norma | boe analisis |
| `eli:transposes` | Norma → Norma (UE) | eurlex nim |
| `eli:cites` | Norma/Bloque → Norma/Bloque | texto del bloque (remisiones "conforme al artículo…") |
| `eli:has_part` | Norma → Bloque | boe indice |
| `eli:has_version` | Bloque → BloqueVersion | boe articulo |
| `eli:is_about` | Norma → Materia | boe metadatos |
| `lb:desarrolla` | Norma (reglamento) → Norma (ley) | data/normas.yaml + analisis |
| `lb:interpreta` | Resolucion/Doctrina → Bloque | ecli, dgt, teac (cuando el texto cita el artículo) |
| `lb:publica` | Publicacion → Norma/Licitacion/Convocatoria | sumarios, edictos |
| `lb:convoca` | Organo → Licitacion/Convocatoria | placsp, bdns |
| `lb:licita` | Entidad → Licitacion | placsp |
| `lb:adjudica` | Licitacion → Contrato ; Contrato → Entidad | placsp |
| `lb:modifica_contrato` | Contrato → Contrato | placsp |
| `lb:concede` | Convocatoria → Concesion ; Concesion → Entidad | bdns |
| `lb:administra` | Persona → Entidad (con fechas nombramiento/cese) | borme |
| `lb:participa` | Entidad → Entidad (socio, fusión, escisión) | borme |
| `lb:titular_de` | Entidad/Organo → Inmueble | catastro (solo si público) |
| `lb:pertenece_a` | Organo → Municipio/CCAA | territorio (registros INE y DIR3) |
| `lb:competente` | Organo/nivel → Materia | competencia |
| `lb:consulta` | Consulta → cualquier nodo del mundo | todos los applets |
| `lb:en_asunto` | Consulta/Hecho/Parte/Plazo/Escrito → Asunto | privado |
| `lb:es` | Parte → Entidad/Organo/Persona | privado → mundo |
| `lb:fundamenta` | Afirmacion → BloqueVersion / Resolucion / Doctrina | cita |
| `lb:contiene` | Escrito → Afirmacion | escrito |
| `lb:vence` | Plazo → Hecho (acto notificado) con fecha límite | plazos |

Toda arista lleva `observed_at`, `source` (applet y URL) y opcionalmente `valid_from`/`valid_to` (temporalidad: un administrador que cesó sigue en el grafo, con fecha).

---

## 3. Almacenamiento

Dos bases SQLite con el mismo esquema:

```
~/.cache/kitlegal/world.db      # grafo del mundo, compartido entre asuntos
./.kitlegal/case.db             # grafo del asunto, en el directorio de trabajo
./.kitlegal/config.yaml         # nombre del asunto, entidades vigiladas, etc.
./.kitlegal/export/             # markdown/json-ld generados (opcional, ver §6)
```

Esquema (property graph sobre SQLite; suficiente hasta millones de aristas):

```sql
CREATE TABLE nodes (
  id        TEXT PRIMARY KEY,      -- id natural o uuid
  type      TEXT NOT NULL,
  props     TEXT NOT NULL,         -- JSON
  first_seen TEXT NOT NULL,
  last_seen  TEXT NOT NULL
);
CREATE INDEX nodes_type ON nodes(type);

CREATE TABLE edges (
  src        TEXT NOT NULL REFERENCES nodes(id),
  dst        TEXT NOT NULL REFERENCES nodes(id),
  rel        TEXT NOT NULL,
  props      TEXT NOT NULL DEFAULT '{}',
  observed_at TEXT NOT NULL,
  valid_from TEXT, valid_to TEXT,
  source     TEXT NOT NULL,        -- "placsp:sync https://…"
  PRIMARY KEY (src, dst, rel, observed_at)
);
CREATE INDEX edges_src ON edges(src, rel);
CREATE INDEX edges_dst ON edges(dst, rel);

-- texto de bloques: aparte, por tamaño
CREATE TABLE texts (
  hash TEXT PRIMARY KEY, body TEXT NOT NULL, fetched_at TEXT NOT NULL
);

-- FTS para buscar por texto en normas y resoluciones ya vistas
CREATE VIRTUAL TABLE texts_fts USING fts5(body, content='texts', content_rowid='rowid');
```

Notas:
- `props` como JSON con `json_extract` en consultas; cuando una propiedad se consulte mucho (p. ej. `importe` de un contrato) se promociona a columna generada indexada.
- El grafo del asunto solo guarda **ids** de nodos del mundo; al consultar se hace `ATTACH` de `world.db` y se une. Si el asunto se mueve a otra máquina, el binario re-resuelve los ids contra las fuentes.
- Sin cgo: `modernc.org/sqlite`. Una sola dependencia para caché, store y grafo.

---

## 4. Cómo se alimenta (por uso, sin pasos extra)

Cada applet implementa `Emit(ctx) []GraphOp` además de su salida JSON. `internal/graph` aplica las operaciones en una transacción tras cada comando. Flag global `--no-graph` para desactivar; `--asunto <nombre>` o `.kitlegal/` presente para escribir también en el grafo del asunto.

Ejemplos de lo que emite cada applet:

```
kitlegal boe articulo BOE-A-2015-10565 a21
  world: Norma(eli/es/l/2015/10/01/39), Bloque(#a21), BloqueVersion(#a21@fecha:hash),
         edges has_part, has_version, y eli:cites por cada remisión detectada en el texto
  case:  Consulta → lb:consulta → BloqueVersion ; Consulta → en_asunto → Asunto

kitlegal boe analisis BOE-A-2015-10565
  world: eli:amends / amended_by / repeals con las normas relacionadas (nodos stub si no existen)

kitlegal territorio resolver 28074          # cualquier municipio, por nombre o código INE
  world: Municipio(ine:28074), Organo(DIR3 del ayuntamiento); pertenece_a (Organo → Municipio)

kitlegal placsp sync --organo L01280740     # o --municipio 28074, que resuelve el DIR3 con territorio
  world: Organo, Licitacion×N, Contrato×N, Entidad×N; convoca, licita, adjudica, modifica_contrato
         con props: importe_licitacion, importe_adjudicacion, procedimiento, num_licitadores, fechas

kitlegal bdns convocatorias --organo L01280740 --anio 2026
  world: Convocatoria, Concesion, Entidad; convoca, concede (props: importe, fecha, finalidad)

kitlegal borme empresa B12345678
  world: Entidad, Persona×N; administra (valid_from/valid_to), participa

kitlegal plazos calcular --tipo alzada --notificado 2026-09-01
  case:  Plazo(vence 2026-10-01, base "art. 122.1 LPAC") → fundamenta → BloqueVersion(#a122)

kitlegal escrito generar --tipo alzada --hechos hechos.md
  case:  Escrito(hash), Afirmacion×N, contiene, fundamenta → BloqueVersion/Resolucion
         (si una afirmación no puede enlazarse a un nodo del mundo, el escrito sale con [SIN FUNDAMENTO] y exit 3)
```

Detección de remisiones (`eli:cites`) en el texto de los bloques: regex sobre patrones estables del lenguaje legislativo español ("artículo 12.3 de la Ley 39/2015", "apartado anterior", "en los términos del Reglamento…") resueltos por `internal/core/cita`. Precisión antes que cobertura: lo que no se resuelve con certeza no se inserta.

---

## 5. Cómo se consulta

### 5.1 Comandos

```
kitlegal graph show <id>                     # nodo + vecinos, en tabla o --json
kitlegal graph path <id-a> <id-b>            # camino más corto (BFS), útil para "cómo se relacionan"
kitlegal graph neighbors <id> --rel adjudica --depth 2
kitlegal graph query "<SQL>"                 # SQL directo sobre nodes/edges (con world.db attached)
kitlegal graph history <bloque-id>           # versiones de un artículo con diffs
kitlegal graph stats                         # nodos/aristas por tipo, cobertura por fuente, frescura
kitlegal graph check                         # validaciones (§6)
kitlegal graph anomalies --organo <DIR3>     # reglas de anomalías (§7)
kitlegal graph export --format md|jsonld|dot|csv
```

### 5.2 Consultas que un agente hará de forma natural

**Qué ha cambiado en lo que consulté**

```sql
SELECT b.id, v_old.props->>'fecha_vigencia', v_new.props->>'fecha_vigencia'
FROM edges c JOIN nodes v_old ON c.dst = v_old.id            -- lo consultado
JOIN edges hv ON hv.dst = v_old.id AND hv.rel = 'eli:has_version'
JOIN edges hv2 ON hv2.src = hv.src AND hv2.rel = 'eli:has_version'
JOIN nodes v_new ON hv2.dst = v_new.id
JOIN nodes b ON b.id = hv.src
WHERE c.rel = 'lb:consulta' AND v_new.last_seen > v_old.last_seen;
```

**Qué normas dependen de esta (para evaluar impacto de una modificación)**

```
kitlegal graph neighbors eli/es/l/2015/10/01/39 --rel eli:amended_by,lb:desarrolla,eli:cites --depth 2 --incoming
```

**Todo lo que sabemos de una empresa**

```
kitlegal graph show B12345678
  → contratos adjudicados (importe total, órganos), concesiones BDNS, administradores con fechas,
    otras entidades donde esos administradores aparecen, inmuebles (si públicos), publicaciones BORME
```

**Contexto para responder una pregunta legal** (lo usa la skill antes de ir a la fuente):

```
kitlegal graph query "SELECT id FROM nodes WHERE type='Bloque' AND id IN (
  SELECT dst FROM edges WHERE rel='lb:consulta' AND src IN (
    SELECT src FROM edges WHERE rel='lb:en_asunto' AND dst='asunto:…'))"
```
→ "en este asunto ya se han consultado estos artículos"; el agente los reutiliza (desde `texts`) si la versión sigue vigente, y solo va a la red si `graph check` marca cambio.

---

## 6. Validación (`graph check`): la continuidad de GOLEM aplicada a derecho

Se ejecuta al abrir un asunto, antes de generar un escrito, y bajo demanda. Cada regla devuelve severidad, nodo afectado y explicación citable.

| Regla | Severidad | Qué comprueba |
|---|---|---|
| `cita-huerfana` | error | Una `Afirmacion` de un escrito sin arista `fundamenta` → probable alucinación |
| `cita-inexistente` | error | El bloque citado no existe en el índice de la norma (art. 999 de una ley con 120) |
| `version-obsoleta` | warning | `BloqueVersion` fundamentante tiene una versión posterior en el mundo → releer y re-fundamentar |
| `norma-derogada` | error | Cualquier norma citada con `eli:repealed_by` y fecha anterior a hoy |
| `norma-no-vigente` | error | `fecha_vigencia` posterior a la fecha del hecho relevante (aplicación retroactiva) |
| `reglamento-contra-ley` | warning | Se cita solo reglamento cuando el grafo muestra `lb:desarrolla` hacia un artículo de ley no citado |
| `plazo-vencido` | error | `Plazo` con fecha límite anterior a hoy y sin `Escrito` posterior en el asunto |
| `plazo-sin-base` | warning | `Plazo` sin arista `fundamenta` |
| `competencia-dudosa` | info | El órgano al que se dirige el escrito no tiene `lb:competente` sobre la materia del asunto |
| `parte-no-resuelta` | info | `Parte` sin `lb:es` hacia Entidad/Organo (no se ha verificado el NIF/DIR3) |
| `remision-no-seguida` | info | Bloque consultado con `eli:cites` a un bloque nunca consultado en el asunto |
| `fuente-caducada` | info | Nodo del mundo con `last_seen` mayor que el TTL de su fuente |

Salida: JSON y tabla; exit 0 (limpio), 1 (warnings), 2 (errores). `escrito generar` se niega a producir el fichero final con errores salvo `--force`.

> **Nota (2026-09-27, ADR 0023):** los códigos de esta línea quedan sustituidos por el contrato de resultados. `graph
> check` sale con 0 con hallazgos o sin ellos, cada uno en `data` con su clase y su severidad; distinto de 0 solo si no
> ha podido verificar (2 argumentos; 1 si `world.db` no se puede usar, porque es un almacén interno). La negativa de
> `escrito generar` ante hallazgos de severidad `error` es un conflicto: código 7.

---

## 7. Detección de anomalías (`graph anomalies`)

Aquí el grafo deja de ser memoria y pasa a ser instrumento. Todas las reglas son **señales**, no acusaciones: la salida siempre lista las aristas y fuentes que la sustentan para que un humano las verifique. Las reglas se definen en YAML (`data/anomalias/*.yaml`) con una consulta SQL y umbrales, de modo que sean auditables y ampliables sin recompilar.

Las reglas son genéricas: ninguna referencia un municipio u órgano concreto. Las que dependen de un umbral legal (fraccionamiento, plazos) lo toman de la ley; las relativas (licitador único, concentración…) comparan con órganos de municipios de tamaño parecido (tramos de población del padrón del INE), porque en un municipio pequeño la concentración de adjudicatarios es alta por naturaleza. Nunca se calibran sobre un solo municipio.

### 7.1 Contratación

| Regla | Señal | Consulta (idea) |
|---|---|---|
| `fraccionamiento` | Varios contratos menores del mismo órgano al mismo adjudicatario, mismo objeto o CPV, en 12 meses, cuya suma supera el umbral del menor (LCSP art. 118) | agrupar `adjudica` por (organo, entidad, cpv, ventana 365d), sumar importe |
| `licitador-unico` | Procedimientos abiertos con `num_licitadores = 1` en proporción alta para un órgano | ratio por órgano vs. órganos de municipios de tamaño parecido |
| `baja-anomala-inversa` | Adjudicación al precio de licitación (0 % de baja) de forma recurrente | props importe_licitacion ≈ importe_adjudicacion |
| `modificados-recurrentes` | Contrato con `modifica_contrato` que supera +20 % del precio inicial | suma de modificaciones / importe inicial |
| `concentracion` | Un adjudicatario acumula una proporción del importe de un órgano en un ejercicio anómala respecto a municipios de tamaño parecido | Herfindahl por órgano/año vs. grupo de comparación |
| `plazo-publicacion` | Contrato adjudicado cuya publicación en PLACSP excede el plazo legal | fecha_adjudicacion vs fecha_publicacion |
| `empresa-recien-creada` | Adjudicatario con constitución en BORME < 6 meses antes de la licitación | `Entidad.fecha_constitucion` vs `Licitacion.fecha` |

### 7.2 Subvenciones

| Regla | Señal |
|---|---|
| `doble-financiacion` | Misma entidad con concesiones de dos órganos para la misma finalidad y periodo |
| `beneficiario-adjudicatario` | Entidad que recibe subvención y contrato del mismo órgano en el mismo ejercicio |
| `concesion-directa-recurrente` | Concesiones nominativas/directas repetidas a la misma entidad (art. 22.2 LGS) |
| `sin-convocatoria` | Concesión sin `Convocatoria` publicada en BDNS |

### 7.3 Cruces entre fuentes (lo que ninguna fuente ve sola)

| Regla | Señal | Camino en el grafo |
|---|---|---|
| `administrador-comun` | Dos adjudicatarias del mismo órgano comparten administrador en el periodo | Entidad ←administra← Persona →administra→ Entidad, ambas →adjudica← Organo, con `valid_from/to` solapados |
| `puerta-giratoria` | Persona que aparece en el portal de transparencia como cargo del órgano y después como administradora de una adjudicataria | Organo →lb:cargo→ Persona →administra→ Entidad →adjudica← Organo, con fechas ordenadas |
| `domicilio-compartido` | Varias licitadoras del mismo expediente con el mismo domicilio social (BORME) | prop `domicilio` normalizada |
| `ciclo-corto` | Constitución → adjudicación → disolución en < 24 meses | fechas BORME + PLACSP |
| `compromiso-vs-ejecucion` | Partida presupuestaria aprobada con ejecución < 30 % mientras se contrata lo mismo por menores | presupuesto (capítulo/programa) vs contratos por CPV |

### 7.4 Salida

```json
{
  "regla": "administrador-comun",
  "severidad": "media",
  "score": 0.71,
  "nodos": ["B12345678", "B87654321", "persona:garcia-lopez-juan", "L01280740"],
  "aristas": [ {"src":"persona:…","dst":"B12345678","rel":"lb:administra","valid_from":"2023-02-01","source":"borme:…"}, … ],
  "explicacion": "Ambas entidades, adjudicatarias de L01280740 en 2025, comparten administrador entre 2023-02 y hoy.",
  "verificar_en": ["https://www.boe.es/borme/…", "https://contrataciondelsectorpublico.gob.es/…"]
}
```

El `score` es heurístico (peso por regla × solapamiento temporal × importe) y sirve solo para ordenar la revisión humana. Con el tiempo, las reglas que un humano confirme o descarte (`kitlegal anomalies mark <id> --confirmada|--descartada`, guardado en el grafo del asunto) permiten ajustar pesos.

### 7.5 Vigilancia continua

`kitlegal vigilar run` (cron) hace `placsp sync`, `bdns`, `borme sumario`, `boletin`, `edictos` y `boe vigilar` sobre el municipio y las entidades de `.kitlegal/config.yaml` (el municipio de cada persona vive ahí, nunca en el repositorio), aplica las operaciones al grafo y ejecuta `anomalies` **solo sobre el delta**: nuevas aristas de hoy. Emite únicamente anomalías nuevas o cuyo score ha subido. Es la capa de monitorización, y por diseño solo toca fuentes públicas.

---

## 8. Privacidad y límites, en el esquema, no en la documentación

- `Persona`: nunca NIF ni DNI en el grafo del mundo; solo nombre normalizado tal como aparece en BORME/transparencia (ya público) y con `valid_from/to`. Sin datos de personas físicas de otras fuentes.
- Catastro: solo referencia catastral y datos sin titular. La arista `titular_de` solo cuando el titular sea un órgano o entidad y conste en fuente pública.
- El grafo del asunto nunca se exporta ni sincroniza a nada por defecto. `graph export` de un asunto pide confirmación y anonimiza `Parte`/`Hecho` salvo `--completo`.
- CENDOJ: en el grafo solo entran `Resolucion` con ECLI + metadatos + URL; nunca texto íntegro descargado automáticamente.
- Cada nodo y arista conserva `source`: el grafo es siempre trazable a la URL original. Un dato sin fuente no entra.

---

## 9. Exportaciones

- **Markdown con wikilinks** (`.kitlegal/export/*.md`): un fichero por nodo del asunto y por norma/entidad relevante, con enlaces `[[…]]`. Navegable en Obsidian; reutiliza lo que ya hiciste para Bookwright.
- **JSON-LD** con `@context` ELI para el subgrafo de normas: interoperable con EUR-Lex, otros grafos legales y cualquier herramienta RDF si algún día hace falta.
- **DOT/GraphML** para dibujar un subgrafo (`graph neighbors … --format dot | dot -Tsvg`): la imagen de "administrador-comun" vale más que la tabla cuando hay que explicarlo.
- **CSV** de aristas para análisis en pandas/DuckDB.

---

## 10. Orden de implementación

El grafo llega cuando las skills del municipio ya existen (fases 3 y 4 de `docs/ROADMAP.md`); los applets anteriores incorporan `Emit` en ese momento.

1. Tablas `nodes/edges/texts` + `internal/graph` con `Apply([]GraphOp)`. Emiten los applets que ya existen: `boe` (Norma/Bloque/BloqueVersion), `territorio` (Municipio/Organo), `placsp` y `bdns`. (`graph show`, `graph stats`.)
2. Grafo del asunto: `Consulta`, `en_asunto`, `.kitlegal/` en cwd con el municipio de la persona, expedientes. Primeras reglas de `check`: `version-obsoleta`, `plazo-vencido`.
3. `borme` emite sus nodos. `graph neighbors`, `graph path`.
4. Reglas de anomalías de contratación (fraccionamiento, licitador único, concentración), genéricas y relativas a municipios comparables. Validar primero en Leganés contra casos que ya conozcas y después en un municipio de otra comunidad.
5. Cruces entre fuentes (administrador-comun, beneficiario-adjudicatario).
6. `vigilar run` sobre delta.
7. `boe analisis` → aristas ELI. `graph history` con diff de versiones.
8. `escrito` + `Afirmacion` + `fundamenta`; `check` completo bloqueando escritos sin fundamento.
9. Exportaciones.

Lo que deliberadamente queda fuera hasta que haya volumen que lo justifique: triple store, SPARQL, embeddings sobre el grafo, y cualquier sincronización entre máquinas.
