# Contrato: el guion de aceptación

La primera tarea (`[aceptacion]`) escribe `specs/038-h41-4-las-skills-del/aceptacion/skills-instaladas.txtar` con el
texto de §2, byte a byte; el workflow lo congela y lo activa como
`internal/app/testdata/script/h41-4-skills-instaladas.txtar` (research V11), que ejecuta `TestEntregaDelHito`.

## 1. Qué afirma y qué no

- Lo instala `kitlegal skills install --dir destino --json`, el binario que construye el test con las skills
  empotradas de `skills/`.
- Por skill, dos `grep` sobre su `SKILL.md` instalado:
  - el frontmatter abre con `---`, `name: <skill>` y `description: >-`, y la clave sigue con el texto de la variante
    (cada patrón es el texto con `regexp.QuoteMeta` y cada espacio cambiado por `( |\n  )`: un espacio o un corte con
    su sangría de dos espacios, que el bloque `>-` vuelve a leer como un espacio), seguido de `metadata:` o, en
    `consulta-juridica`, del cierre `---` (FR-001, FR-002);
  - la clave tiene al menos dos líneas sangradas (FR-002).
- No afirma dónde se corta cada línea, que el spec no fija (research D5), ni nada del cuerpo, de las referencias o de
  los límites: eso lo comprueban los tests de contracts/controles.md.
- Hoy falla por una aserción (`:22: no match`, el primer `grep`); con un bloque `>-` de una sola línea, en el segundo
  `grep` de la skill (research V9).

## 2. El guion, byte a byte

```text
# skills-instaladas.txtar
# Lo que instala `kitlegal skills install` de las cuatro skills del producto
# (contracts/aceptacion.md de H41.4): en cada SKILL.md, el frontmatter abre con
# `name:` y, detrás, la clave `description: >-`, cuyo texto, en más de una
# línea sangrada, se lee como el de la variante producto-descripciones-cortas de
# H41.3; detrás va `metadata:`, o el cierre del frontmatter en
# consulta-juridica. Cada patrón es el texto de la variante con cada espacio
# como `( |\n  )`: un espacio, o un corte de línea con su sangría, que el bloque
# `>-` vuelve a leer como un espacio. Dónde se corta cada línea no lo fija el
# spec, y el guion no lo afirma.
#
# Cubre: US1 · FR-001, FR-002 · SC-001
#
# Que la description que lee skills.LeerFrontmatter sea, byte a byte, la de la
# variante, sus límites y que nada más cambie lo comprueban
# TestDescripcionesDelProducto, TestConsultaJuridicaDesdeLaBase,
# TestSkillsDelProductoDesdeControl y TestSkillsDelRepositorio.

exec kitlegal skills install --dir destino --json

# --- consulta-juridica (US1, 1 y 3) ---
grep '\A---\nname: consulta-juridica\ndescription: >-\n  Punto( |\n  )de( |\n  )entrada( |\n  )de( |\n  )toda( |\n  )pregunta( |\n  )legal( |\n  )en( |\n  )España,( |\n  )de( |\n  )cualquier( |\n  )materia:( |\n  )dice( |\n  )qué( |\n  )capacidad( |\n  )de( |\n  )kitlegal( |\n  )la( |\n  )resuelve( |\n  )—normativa,( |\n  )jurisprudencia,( |\n  )plazos( |\n  )y( |\n  )territorio—( |\n  )y( |\n  )con( |\n  )qué( |\n  )skill,( |\n  )y( |\n  )lo( |\n  )que( |\n  )no( |\n  )cubre( |\n  )lo( |\n  )dice( |\n  )sin( |\n  )dar( |\n  )de( |\n  )memoria( |\n  )plazos,( |\n  )órganos( |\n  )ni( |\n  )vías\.( |\n  )Se( |\n  )usa( |\n  )al( |\n  )empezar( |\n  )toda( |\n  )pregunta( |\n  )legal,( |\n  )antes( |\n  )que( |\n  )las( |\n  )demás( |\n  )skills( |\n  )de( |\n  )kitlegal,( |\n  )también( |\n  )si( |\n  )no( |\n  )nombra( |\n  )norma( |\n  )o( |\n  )pide( |\n  )qué( |\n  )hacer,( |\n  )a( |\n  )quién( |\n  )reclamar( |\n  )o( |\n  )qué( |\n  )escrito( |\n  )presentar\.\n---\n' destino/consulta-juridica/SKILL.md
grep '\A---\nname: consulta-juridica\ndescription: >-\n  [^\n]+\n  \S' destino/consulta-juridica/SKILL.md

# --- jurisprudencia (US1, 1 y 3) ---
grep '\A---\nname: jurisprudencia\ndescription: >-\n  Sentencias( |\n  )y( |\n  )jurisprudencia( |\n  )de( |\n  )los( |\n  )tribunales( |\n  )españoles,( |\n  )sin( |\n  )citar( |\n  )ninguna( |\n  )de( |\n  )memoria\.( |\n  )Se( |\n  )usa( |\n  )siempre( |\n  )que( |\n  )la( |\n  )pregunta( |\n  )nombre,( |\n  )pida( |\n  )o( |\n  )cite( |\n  )una( |\n  )sentencia( |\n  )\(STS,( |\n  )ECLI,( |\n  )ROJ\),( |\n  )pregunte( |\n  )si( |\n  )existe( |\n  )o( |\n  )qué( |\n  )dice,( |\n  )pida( |\n  )jurisprudencia( |\n  )sobre( |\n  )una( |\n  )materia( |\n  )o( |\n  )traiga( |\n  )una( |\n  )resolución,( |\n  )aunque( |\n  )el( |\n  )agente( |\n  )crea( |\n  )conocerla\.( |\n  )kitlegal( |\n  )no( |\n  )consulta( |\n  )el( |\n  )CENDOJ:( |\n  )prepara( |\n  )la( |\n  )búsqueda( |\n  )y( |\n  )coteja( |\n  )el( |\n  )documento\.( |\n  )Se( |\n  )activa( |\n  )antes( |\n  )de( |\n  )cita_preparar( |\n  )o( |\n  )cita_cotejar\.\nmetadata:\n' destino/jurisprudencia/SKILL.md
grep '\A---\nname: jurisprudencia\ndescription: >-\n  [^\n]+\n  \S' destino/jurisprudencia/SKILL.md

# --- normativa (US1, 1 y 3) ---
grep '\A---\nname: normativa\ndescription: >-\n  Consulta( |\n  )y( |\n  )cita( |\n  )normativa( |\n  )consolidada( |\n  )del( |\n  )BOE( |\n  )de( |\n  )cualquier( |\n  )materia:( |\n  )lee( |\n  )con( |\n  )kitlegal( |\n  )índice( |\n  )y( |\n  )artículos( |\n  )y( |\n  )cita( |\n  )identificador( |\n  )y( |\n  )bloque\.( |\n  )Se( |\n  )usa( |\n  )siempre( |\n  )que( |\n  )la( |\n  )respuesta( |\n  )dependa( |\n  )de( |\n  )lo( |\n  )que( |\n  )dice( |\n  )una( |\n  )norma( |\n  )—un( |\n  )artículo,( |\n  )ley( |\n  )o( |\n  )real( |\n  )decreto,( |\n  )el( |\n  )plazo( |\n  )o( |\n  )requisito( |\n  )que( |\n  )fija—( |\n  )o( |\n  )de( |\n  )lo( |\n  )que( |\n  )decía( |\n  )en( |\n  )una( |\n  )fecha,( |\n  )aunque( |\n  )el( |\n  )agente( |\n  )crea( |\n  )saberlo\.( |\n  )Se( |\n  )activa( |\n  )antes( |\n  )de( |\n  )boe_articulo( |\n  )u( |\n  )otra( |\n  )herramienta( |\n  )boe_…:( |\n  )qué( |\n  )pedir( |\n  )y( |\n  )cómo( |\n  )citar( |\n  )lo( |\n  )dice( |\n  )la( |\n  )skill\.\nmetadata:\n' destino/normativa/SKILL.md
grep '\A---\nname: normativa\ndescription: >-\n  [^\n]+\n  \S' destino/normativa/SKILL.md

# --- plazos-y-territorio (US1, 1 y 3) ---
grep '\A---\nname: plazos-y-territorio\ndescription: >-\n  Calcula( |\n  )con( |\n  )kitlegal( |\n  )hasta( |\n  )qué( |\n  )día( |\n  )se( |\n  )puede( |\n  )recurrir( |\n  )una( |\n  )resolución( |\n  )administrativa,( |\n  )cuándo( |\n  )vence( |\n  )un( |\n  )plazo( |\n  )del( |\n  )procedimiento( |\n  )común( |\n  )\(LPAC\)( |\n  )o( |\n  )el( |\n  )del( |\n  )contencioso( |\n  )y( |\n  )cuándo( |\n  )hay( |\n  )silencio,( |\n  )con( |\n  )los( |\n  )festivos( |\n  )que( |\n  )cuentan,( |\n  )e( |\n  )identifica( |\n  )el( |\n  )territorio( |\n  )de( |\n  )un( |\n  )municipio:( |\n  )INE,( |\n  )provincia,( |\n  )comunidad,( |\n  )DIR3( |\n  )y( |\n  )boletines\.( |\n  )No( |\n  )cuenta( |\n  )plazos( |\n  )de( |\n  )otras( |\n  )materias\.( |\n  )Se( |\n  )usa( |\n  )al( |\n  )preguntar( |\n  )cuándo( |\n  )vence( |\n  )un( |\n  )plazo( |\n  )o( |\n  )un( |\n  )recurso,( |\n  )o( |\n  )qué( |\n  )boletines( |\n  )tiene( |\n  )un( |\n  )municipio\.\nmetadata:\n' destino/plazos-y-territorio/SKILL.md
grep '\A---\nname: plazos-y-territorio\ndescription: >-\n  [^\n]+\n  \S' destino/plazos-y-territorio/SKILL.md
```

## 3. Uso

Lo lee el mantenedor cuando `TestEntregaDelHito` sale en rojo: testscript da la línea del guion y el patrón que no
casa, una vez por ejecución de `make ci`. No depende del uso del kit.
