# Contrato de `legal-core` v1

Qué lleva `skills/legal-core/SKILL.md` y por qué (FR-040 a FR-048). El borrador, de 248 líneas y 15,9 kB, con su
tabla de comandos regenerada en el prototipo (research V13), es contracts/SKILL-legal-core.prototipo.md; la tarea de
la skill parte de él y regenera la tabla y las referencias con `make skills-sync`. Decisión: research D21.

## 1. Frontmatter

- `description` de 993 caracteres (el límite es 1 024; research V22): gana las preguntas de plazos y recursos de un
  asunto —«hasta cuándo se puede recurrir», «qué plazo hay para un recurso», «cuándo vence un plazo», «cuándo se
  produce el silencio»— y la de lo que la ley regula aparte, y sigue ganando las del territorio (FR-045). Control:
  `sin_activar:<modelo>:<modo>` de `legal-core`, con 0, que decide.
- `kitlegal-applets: territorio plazos`: la tabla de comandos tiene los dos, generada desde `--describe` (FR-046).
- `kitlegal-referencias: leyes_vertebrales jerarquia_normativa recursos_y_plazos cobertura_por_materia` (FR-032).

## 2. El protocolo (FR-040 a FR-044, FR-047)

Los pasos 1 a 6 son los de v0, más cortos. Los nuevos:

| Paso | Qué hace | Requisito |
|---|---|---|
| 7 | Con el acto y la administración, decide si hay plazo que calcular **antes de preguntar nada más**: comprueba la cobertura por materia con `references/cobertura_por_materia.md` y, si el asunto está, lo declara en su forma fija y no pide `plazos`; ante un acto presunto, cita los arts. 122.1 y 124.1 LPAC y el 46.1 LJCA, sin fecha. En los dos casos no pregunta ni municipios ni fecha. Si la conversación no dice la materia, el asunto no está en la lista: no la pregunta ni se detiene por ella, sigue con el 8 y, al responder, avisa de que la fecha no vale para lo que la ley regula aparte | FR-040.1, FR-042, FR-043 |
| 8 | Sitúa el plazo: la fecha de inicio (la pregunta si falta) y los municipios que cuentan, resueltos con `territorio` por los pasos 1 a 4; los pregunta si faltan, nunca el de quien usa el kit | FR-040.2 |
| 9 | Pide el plazo a `plazos` con la fecha de la persona y solo con municipios que `territorio` resolvió en la conversación (paso 8), aunque `plazos` acepte un nombre sin resolver; silencio con la duración de la norma del procedimiento o los tres meses del art. 21.3; variación autonómica del art. 24.6 LTAIBG | FR-040.3, FR-044 |
| 10 | Lee con `boe-legislacion` (`boe articulo`) cada precepto de `data.preceptos` antes de citarlo | FR-040.4 |
| 11 | Responde con la línea del plazo, con una cita por cada precepto de `data.preceptos`, cada una en sus propios corchetes y todas en esa línea, los festivos de `data.festivos` y la línea de cada aviso; la disposición, solo con el aviso de cobertura que la trae | FR-041 |

Una pregunta de plazos o de recursos empieza por el paso 7, y los pasos 1 a 4 solo los recorren las preguntas de
territorio y el paso 8, para los municipios de un plazo que sí se calcula. Es el orden de FR-040 —cobertura antes que
los municipios y que `plazos`, territorio antes que `plazos`, `plazos` antes que el precepto—: un asunto que la ley
regula aparte, o un acto presunto, no necesita municipios ni fecha, porque no se le calcula nada, y la pregunta de
SC-003 sin municipio («¿qué plazo tengo para recurrir una liquidación de la Agencia Tributaria?») se responde con la
declaración, sin preguntar dónde vive la persona. Las reglas de v0 siguen; la 7 gana «ninguna fecha ni ningún plazo» (FR-047), y
entran la 8 (ninguna fecha que no dé `plazos`) y la 9 (lo que la ley regula aparte no recibe el plazo del común).

## 3. Las formas fijas (FR-048)

En `SKILL.md`, tal cual, en bloques `text` (sección «Cómo se presenta un plazo»): la línea `⚠ PLAZO:` —la plantilla y
su forma con los dos preceptos del contencioso, una cita por precepto en sus propios corchetes—, las dos de
`⚠ NO CUBIERTO:` y la cabeza de los cuatro avisos. Las comprueba el juicio sin modelo de contracts/evals-legal-core.md
§2; el texto de cada aviso lo da `data.avisos[].texto` (contracts/applet-plazos.md §5).

## 4. Las referencias generadas (FR-032)

Dos filas nuevas en la tabla de generadores de `internal/skills/referencias.go`:

| Referencia | Datos | Título | Columnas |
|---|---|---|---|
| `references/recursos_y_plazos.md` | `data/plazos.yaml`, con `plazos.LeerTipos` | Recursos y plazos | Tipo · Código · Duración · Cómputo · Precepto |
| `references/cobertura_por_materia.md` | `data/cobertura.yaml`, con `skills.LeerCobertura` y `schemas/cobertura.yaml.json` | Cobertura por materia | Procedimiento · Lo regula aparte · Norma que lo rige |

Cada una con la cabecera `<!-- generado desde data/<fichero>, no editar -->`. Una fila por tipo (el genérico, una por
unidad) y una por entrada: nueve y nueve, unos 2 kB cada una. «Precepto» y «Lo regula aparte» como `art. 124.1
(BOE-A-2015-10565, bloque a124)` o `DA 1ª.2.a (BOE-A-2015-10565, bloque daprimera)`, sin la abreviatura de la norma,
que `data/plazos.yaml` y `data/cobertura.yaml` no llevan: la nombra su identificador (T013, `gates/supuestos.md`);
varios preceptos, separados por «; »; «Norma que lo rige», el nombre y el identificador, o «—» si la entrada no la
trae.
Ninguna lleva texto de ninguna norma.

## 5. Uso, de fuera adentro

| Salida | Quién la pide y cuántas veces | Tamaño | Señales |
|---|---|---|---|
| `SKILL.md` | el agente, una vez por conversación en que se activa | 248 líneas, 15,9 kB (en `main`, 190 y 12,4 kB) | ninguna |
| `recursos_y_plazos.md` y `cobertura_por_materia.md` | la skill, una vez por conversación con una pregunta de plazos: la de cobertura en el paso 7 y la de tipos en el 9 | unos 2 kB cada una; fijas, no dependen del uso | ninguna: cambian con `data/` |
| La respuesta | la persona, una por pregunta | una línea del plazo o una declaración, y una línea por aviso | una por respuesta; no se acumulan |

Por pregunta de plazo, la skill hace: una `territorio resolver` por municipio que cuenta (uno o dos), una
`plazos calcular` por plazo (casi siempre uno) y una `boe articulo` por precepto (uno, dos en el contencioso), con la
activación de `boe-legislacion`. Por pregunta de cobertura, ninguna operación es obligatoria. Nada de eso crece con
meses de uso: ni `territorio` ni `plazos` leen la caché o el grafo, y `boe articulo` lee un bloque.
