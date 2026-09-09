# kitlegal: estructura del proyecto, buenas prácticas y ecosistema "Ventanilla"

Estado de partida (importante para quien arranque el proyecto): **no existe código previo**. Lo único implementado es la skill `boe-fiscal` (Python, `scripts/boe.py` contra la API de Legislación Consolidada del BOE), que se toma como patrón y se porta a Go. Nombre único para repo, módulo, binario y directorios: `kitlegal`.

Premisas de diseño:
- El binario Go es multicall: un único ejecutable, `os.Args[0]` o el primer argumento decide el applet (`boe`, `placsp`, `bdns`…). CLI con Kong (`github.com/alecthomas/kong`) y un paquete propio `internal/cli` con las convenciones para agentes: `--json`, `--timeout`, `--offline`, `--dry-run`, exit codes estables. Esas convenciones se escriben dentro de kitlegal; no se extraen a una librería aparte hasta que haya al menos tres applets que las repitan.
- Las skills siguen el estándar abierto Agent Skills (`SKILL.md` con frontmatter `name`/`description`, `references/`, `scripts/`), lo que las hace portables a Claude Code, Claude.ai, Codex, Cursor, etc.
- Los dominios `ventanilla{legal,fiscal,laboral,mercantil}.es` son **verticales** sobre un mismo núcleo, no cuatro productos distintos.

---

## 1. Estructura de directorios (monorepo)

```
kitlegal/
├── go.mod                        # module github.com/jmorenobl/kitlegal
├── Makefile                      # build, test, lint, skills-sync, release
├── .goreleaser.yaml              # binarios darwin/linux/windows + checksums
├── README.md
├── LICENSE                       # ver §6 (open-core)
│
├── cmd/
│   └── kitlegal/
│       └── main.go               # multicall: os.Args[0] decide el applet (boe, placsp, bdns…)
│
├── internal/                     # no importable desde fuera del módulo
│   ├── cli/                      # Kong + flags globales (--json, --timeout, --offline, --dry-run) + exit codes
│   ├── app/                      # registro de applets, dispatch multicall
│   ├── core/
│   │   ├── cita/                 # parser "art. 21.1 Ley 39/2015" → {norma, bloque}; resolver a ELI
│   │   ├── plazos/               # días hábiles, festivos, LPAC 30-31, LJCA, LTAIBG
│   │   ├── competencia/          # heurística Estado/CCAA/Local
│   │   ├── schema/               # tipos Go de salida + JSON Schema generado
│   │   └── ids/                  # BOE-A-…, ELI, ECLI, CELEX, código INE, NIF: parse + validate
│   ├── source/                   # UN paquete por fuente = adaptador
│   │   ├── boe/                  # legislación consolidada, sumario, BORME, TEU (misma API)
│   │   ├── placsp/               # Atom + CODICE XML
│   │   ├── bdns/                 # REST
│   │   ├── catastro/             # OVC SOAP/JSON
│   │   ├── ine/                  # Tempus
│   │   ├── datosgob/             # CKAN
│   │   ├── eurlex/               # SPARQL Cellar
│   │   ├── congreso/             # Open Data
│   │   ├── dgt/                  # PETETE (HTML, ratelimited)
│   │   ├── teac/                 # DYCTEA (HTML)
│   │   ├── tc/                   # HJ (HTML)
│   │   ├── ecli/                 # resolutor, sin CENDOJ masivo
│   │   ├── boletin/              # motor genérico por YAML (BOCM, DOGC, BOJA, BOP…)
│   │   └── sede/                 # crawler de sedes/portales municipales (ordenanzas, transparencia)
│   ├── httpx/                    # cliente HTTP: retries, rate limit por host, UA identificable, robots.txt
│   ├── cache/                    # SQLite (modernc.org/sqlite, sin cgo) con TTL por fuente
│   ├── store/                    # DB local para sync incremental (PLACSP, BDNS, BORME)
│   ├── analysis/                 # reglas de anomalías: fraccionamiento, único licitador, cruces beneficiario/adjudicatario/administrador
│   ├── render/                   # json / table / markdown; --json manda
│   └── docgen/                   # plantillas de escritos → DOCX/MD (capa acción)
│
├── pkg/                          # API pública estable para terceros (opcional, pequeña)
│   └── legalkit/                 # tipos de dominio + cliente Go de alto nivel
│
├── skills/                       # LAS SKILLS (estándar Agent Skills). Una carpeta por skill.
│   ├── legal-core/
│   │   ├── SKILL.md
│   │   ├── references/
│   │   │   ├── jerarquia_normativa.md
│   │   │   ├── leyes_vertebrales.md      # generado desde data/normas.yaml
│   │   │   ├── recursos_y_plazos.md
│   │   │   └── reglas_de_cita.md
│   │   └── scripts/
│   │       └── plazos -> ../../../bin/kitlegal   # symlink al multicall (o wrapper sh)
│   ├── cita-verificada/
│   ├── boe-legislacion/
│   ├── boe-fiscal/                # tu skill actual, migrada (scripts/boe.py → kitlegal boe)
│   ├── boletines-autonomicos/
│   ├── bop-y-edictos/
│   ├── ordenanzas-locales/
│   ├── eurlex/
│   ├── tramitacion-parlamentaria/
│   ├── jurisprudencia/
│   ├── doctrina-administrativa/
│   ├── contratacion-publica/
│   ├── subvenciones/
│   ├── presupuestos-y-cuentas/
│   ├── transparencia-portal/
│   ├── entidades-y-registros/
│   ├── datos-abiertos/
│   ├── redaccion-escritos/
│   └── seguimiento-expedientes/
│
├── packs/                        # agrupaciones de skills por vertical (ver §4)
│   ├── legal/pack.yaml
│   ├── fiscal/pack.yaml
│   ├── laboral/pack.yaml
│   ├── mercantil/pack.yaml
│   └── fiscalizador/pack.yaml
│
├── data/                         # datos de referencia versionados, fuente de verdad
│   ├── normas.yaml               # id BOE ↔ nombre corto ↔ ELI ↔ materia ↔ vertical
│   ├── organos.yaml              # códigos DIR3, INE, NIF de entidades
│   ├── festivos/2026.json
│   ├── boletines/                # bocm.yaml, dogc.yaml, boja.yaml… (motor genérico)
│   └── entidades/                # leganes.yaml, comunidad-madrid.yaml…
│
├── schemas/                      # JSON Schema de cada salida (generado desde internal/core/schema)
│   ├── norma.json
│   ├── bloque.json
│   ├── licitacion.json
│   ├── convocatoria.json
│   └── cita.json
│
├── testdata/                     # fixtures grabados (respuestas reales de cada API) → tests offline
│   ├── boe/
│   ├── placsp/
│   └── …
│
├── evals/                        # evaluación de skills: tareas + respuestas y citas esperadas
│   ├── boe-legislacion/
│   └── contratacion-publica/
│
├── mcp/                          # servidor MCP que expone los mismos applets como tools
│   └── server.go                 # kitlegal mcp serve --stdio | --http
│
├── plugin/                       # empaquetado como plugin (Claude Code marketplace, etc.)
│   ├── .claude-plugin/plugin.json
│   └── marketplace.json
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── ADR/                      # decisiones (por qué no CENDOJ, por qué SQLite, por qué multicall)
│   ├── SOURCES.md                # tabla de fuentes + semáforo + TOS + fecha de última verificación
│   └── CONTRIBUTING.md
│
└── scripts/                      # dev tooling, no runtime
    ├── skills-sync.sh            # regenera references/*.md desde data/*.yaml y symlinks
    ├── verify-sources.sh         # smoke test contra APIs reales (CI nightly)
    └── record-fixtures.sh
```

Punto clave: **las skills no llevan código**. `scripts/` de cada skill apunta al mismo binario; el conocimiento (protocolo de razonamiento, identificadores, reglas) vive en `SKILL.md` y `references/`, y se **genera** desde `data/*.yaml` para que no haya dos verdades.

---

## 2. Cómo se organiza el multicall

```
kitlegal <applet> <verbo> [args] [--json] [--describe] [--dry-run]

applets (uno por fuente o por capacidad core):
  boe        buscar | indice | articulo | articulos | metadatos | analisis | sumario | vigilar | eli
  borme      sumario | empresa | administrador
  edictos    buscar
  boletin    sumario | buscar          (--boletin bocm|dogc|…)
  sede       ordenanzas | transparencia (--entidad leganes)
  eurlex     buscar | celex | nim
  congreso   iniciativa | votaciones
  ecli       resolver
  tc         buscar
  dgt        consulta | buscar
  teac       buscar
  placsp     sync | licitaciones | organo | adjudicatario | anomalias
  bdns       convocatorias | concesiones | beneficiario
  catastro   rc | direccion
  ine        tabla | serie
  datosgob   buscar | dataset
  presupuesto descargar | comparar
  cita       resolver | validar
  plazos     calcular
  competencia quien-regula
  escrito    generar
  expediente listar | vencimientos | add
  vigilar    run                       (orquesta: boe vigilar + placsp sync + bdns + edictos → diff)
  mcp        serve
  skills     list | install | doctor   (gestiona packs en ~/.claude/skills o donde toque)
```

Aliases por `os.Args[0]`: un symlink `boe -> kitlegal` hace que `boe articulo …` funcione, así el `SKILL.md` puede seguir diciendo `scripts/boe articulo BOE-A-2015-10565 a21` como ahora.

Convenciones de salida (en `internal/core/schema`), obligatorias en todos los applets:

```json
{
  "ok": true,
  "fuente": "boe.legislacion-consolidada",
  "url": "https://www.boe.es/…",
  "fecha_consulta": "2026-09-09T10:12:00+02:00",
  "hash": "sha256:…",
  "data": { … }
}
```

Exit codes estables (definidos en `internal/cli`): 0 ok · 2 args · 3 no encontrado · 4 fuente no disponible · 5 rate-limited/TOS · 6 requiere identidad humana.

---

## 3. Buenas prácticas (Go + skills + fuentes públicas)

**Go**
- `internal/source/<x>`: interfaz común `Source` con `Name()`, `Fetch(ctx, req) (Result, error)`, `TTL()`, `Terms()` (URL de TOS + fecha revisada). Nada de HTTP fuera de `httpx`.
- Sin cgo: `modernc.org/sqlite`. Un binario estático es la gracia del multicall.
- Contextos y `--timeout` en todo; `--offline` que responde solo desde caché (útil en agentes que corren en sandbox sin red).
- Fixtures grabados en `testdata/` con un modo `KITLEGAL_RECORD=1`; los tests unitarios no tocan la red. Un job nightly de CI (`verify-sources.sh`) sí lo hace y abre issue si una API cambió.
- Versionado semántico del binario **y** de cada skill (`version:` en frontmatter, o `metadata.version`). El `SKILL.md` declara la versión mínima de `kitlegal` que necesita.
- `--describe` en cada applet emite el JSON Schema de entrada/salida → de ahí se generan automáticamente las `tools` del servidor MCP y la tabla de comandos de cada SKILL.md. Una definición, tres consumidores.

**Skills**
- SKILL.md corto (< 300 líneas): protocolo de razonamiento + tabla de comandos + reglas. Lo pesado va a `references/` y se carga bajo demanda (progressive disclosure).
- `description` escrito para el trigger, no para el humano: incluye los términos que un usuario usaría ("IRPF", "licitación", "ordenanza", "recurso de alzada", identificadores BOE…). Optimizar el triggering con evals (preguntas que deben activar la skill y preguntas que no).
- Reglas invariantes en **todas** las skills: nunca inventar contenido legal; cada afirmación con cita resuelta por `cita`; distinguir ley/reglamento; señalar variación autonómica; nunca ejecutar acción con identidad.
- `references/*.md` generados, con cabecera `<!-- generado desde data/normas.yaml, no editar -->`.
- Evals por skill en `evals/`: 10-20 preguntas reales con respuesta esperada y citas esperadas. Corren en CI con un modelo barato.

**Fuentes públicas**
- `docs/SOURCES.md` es el contrato legal del proyecto: por fuente, licencia de reutilización (la mayoría de datos del sector público español van bajo la Ley 37/2007 + RD 1495/2011, reutilizables con atribución; CENDOJ es la excepción notable), TOS, rate limit acordado, fecha de última revisión.
- User-Agent explícito (`kitlegal/1.2 (+https://ventanillalegal.es/bot)`) con página que explique qué hace el bot y cómo contactarte.
- Rate limit por host en `httpx`, respeto de `robots.txt`, backoff exponencial, y **nada** de paralelismo agresivo contra PETETE/DYCTEA/HJ.

---

## 4. Ecosistema "Ventanilla": del kit a los cuatro dominios

### 4.1 Modelo mental

```
                     ventanillalegal.es  (paraguas: marca, docs, registro de skills, cuenta)
                                 │
        ┌────────────────┬───────┴────────┬────────────────┐
  ventanillafiscal   ventanillalaboral   ventanillamercantil   (verticales = packs + producto)
        │                    │                  │
        └──────────── kitlegal (núcleo Go + skills, open source) ─────────────┘
                                 │
              fuentes públicas: BOE · PLACSP · BDNS · BORME · DGT · TEAC · INE · EUR-Lex …
```

- **Núcleo** (`kitlegal`): un solo repo, un solo binario, todas las fuentes. Open source. Es lo que da credibilidad y adopción.
- **Packs** (`packs/*.yaml`): selección de skills + `references/` específicas + evals por vertical. Un pack es lo que un usuario instala con `kitlegal skills install fiscal`.
- **Verticales** (dominios): cada uno = pack + producto hospedado (chat/agente con la skill precargada, como ya es Ventanilla Fiscal) + documentación + casos de uso. Ventanilla Fiscal migra a consumir el núcleo en vez de su loop propio.
- **Paraguas** (`ventanillalegal.es`): catálogo/registro de skills, docs del binario, página del bot, blog, y el punto de entrada para agentes ("añade `https://ventanillalegal.es/mcp` a tu cliente").

### 4.2 Qué contiene cada pack

| Pack | Skills núcleo | Skills/references propias del vertical | Fuentes clave |
|---|---|---|---|
| **legal** (base, todos lo heredan) | legal-core, cita-verificada, boe-legislacion, eurlex, jurisprudencia, doctrina-administrativa, redaccion-escritos | procedimiento administrativo, contencioso, civil general | BOE, EUR-Lex, ECLI, TC |
| **fiscal** | + boe-fiscal (existente) | normas_fiscales.md, calendario del contribuyente, modelos AEAT, DGT/TEAC | BOE, PETETE, DYCTEA, AEAT |
| **laboral** | + `laboral-normas` (ET, LGSS, LPRL, LISOS, LRJS), `convenios` (REGCON: buscador de convenios colectivos), `seguridad-social` (bases y tipos de cotización, SEPE) | tablas de cotización, indemnizaciones, plazos laborales (20 días hábiles despido…) | BOE, REGCON, SEPE, Seg. Social, TSJ social vía ECLI |
| **mercantil** | + `borme`, `entidades-y-registros`, `concursal` (Registro Público Concursal), `contabilidad` (PGC), `competencia` (CNMC) | LSC, Código de Comercio, TRLC, modelos de cuentas, cláusulas | BORME, RPC, CNMC, Catastro, BDNS/PLACSP para due diligence |
| **fiscalizador** (tu kit cívico) | + contratacion-publica, subvenciones, presupuestos-y-cuentas, transparencia-portal, bop-y-edictos, ordenanzas-locales, seguimiento-expedientes | entidades/*.yaml, reglas de anomalías | PLACSP, BDNS, Rendición de Cuentas, BOCM/BOP |

`pack.yaml` mínimo:

```yaml
name: fiscal
version: 0.3.0
requires: { kitlegal: ">=1.2" }
extends: legal
skills: [boe-fiscal, doctrina-administrativa]
references:
  - data/normas.yaml#materia=tributario
entities: []            # el fiscalizador sí lista leganes, comunidad-madrid…
evals: evals/fiscal/
```

### 4.3 Superficies de distribución (todas desde el mismo código)

1. **Skills instalables** (`kitlegal skills install <pack>`): copia `skills/` al directorio del cliente (Claude Code, Codex, Cursor…) y deja el binario en `PATH`. También como zip descargable desde cada dominio.
2. **Plugin de marketplace** (`plugin/`): `marketplace.json` con un plugin por pack. Un solo `/plugin marketplace add jmorenobl/kitlegal`.
3. **Servidor MCP** (`kitlegal mcp serve`): local por stdio, o remoto en `ventanillalegal.es/mcp` con auth. Aquí encaja la capa de pago: el MCP remoto elimina la instalación y añade caché compartida, cuotas y logs de citas.
4. **Producto hospedado** por vertical (lo que hoy es ventanillafiscal.es): agente con el pack precargado, UI y memoria de expedientes.
5. **Librería Go** (`pkg/legalkit`) para quien quiera embeber.

### 4.4 Gobernanza y calidad, lo que hace que sea un "ecosistema" y no cuatro repos

- Registro de skills en `ventanillalegal.es/skills/` con índice JSON (`index.json`: nombre, versión, pack, hash, fuentes, fecha de verificación). Terceros pueden publicar skills que cumplan el contrato (schema de salida + reglas de cita + evals verdes).
- `kitlegal skills doctor`: comprueba versión del binario, integridad (hash) de cada skill, y lanza `verify-sources` para las fuentes que usa.
- Nightly `verify-sources` publicado como página de estado ("BOE ✅, PETETE ⚠️ formulario cambiado 2026-09-07"). Esto solo ya es un servicio valioso para toda la comunidad legaltech.
- Changelog por skill; versiones de datos (`data/normas.yaml`) etiquetadas por fecha para reproducibilidad de una respuesta ("con datos del 2026-09-01").

---

## 5. Migración desde lo que ya tienes

1. Crear el repo con `cmd/kitlegal` e `internal/cli` (Kong + convenciones); portar `boe.py` a `internal/source/boe` (es el más maduro y el que valida el patrón). Mantener el `SKILL.md` de `boe-fiscal` idéntico salvo `scripts/boe.py` → `scripts/boe`.
2. Extraer `references/normas_fiscales.md` a `data/normas.yaml` con campo `vertical: fiscal` y generar el .md.
3. Añadir `cita`, `plazos` y `legal-core`.
4. `placsp`, `bdns`, `borme` → pack `fiscalizador` funcionando en Leganés.
5. Ventanilla Fiscal pasa a llamar al binario (o al MCP) en vez de su loop propio con google-genai; el loop se queda solo como orquestador.
6. Publicar `ventanillalegal.es` con docs, registro e índice de packs; laboral y mercantil arrancan como packs "mínimos" (normas + citas) y crecen con fuentes propias.

---

## 6. Licencia y modelo (para decidir pronto, cambia la estructura)

- **Open-core recomendado**: núcleo + skills bajo Apache-2.0 o MIT (adopción, contribuciones, credibilidad jurídica de "puedes auditar cómo cito"); MCP remoto, productos hospedados, packs premium (p. ej. `entidades/*.yaml` curados de 8.000 municipios, reglas de anomalías avanzadas) y soporte, de pago bajo los dominios.
- Marca única: "Ventanilla" como familia, `kitlegal` como herramienta. Registrar la marca antes de abrir el repo.
- Aviso legal en todo output: "información, no asesoramiento; citas verificables en origen". Especialmente en laboral y fiscal, donde el usuario final actúa sobre la respuesta.
