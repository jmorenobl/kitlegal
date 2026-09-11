# kitlegal: documentos semilla

Tres documentos de diseño para arrancar el proyecto. Léelos en este orden.

**Enfoque (2026-09-11).** El producto son las skills; el binario Go es la capa de herramientas deterministas que usan. Todo es genérico para cualquier municipio de España y se valida primero en la Comunidad de Madrid con Leganés; los demás territorios llegan después como datos. Ver la constitución (principios VIII y IX) y los ADR 0005 y 0006; ante conflicto, prevalecen sobre estos documentos.

1. `mapa-sistema-legal-skills.md` — Qué es el sistema legal español (jerarquía, competencias, jurisdicción, vía administrativa), dónde está cada tipo de información (fuentes con semáforo de automatizabilidad) y el catálogo de skills agénticas con sus comandos.
2. `kitlegal-estructura-y-ecosistema.md` — Estructura del monorepo (skills, packs y datos; binario multicall `kitlegal` como capa de herramientas), buenas prácticas, packs por vertical y superficies de distribución.
3. `kitlegal-grafo.md` — El grafo legal (mundo + asunto): modelo sobre ELI, almacenamiento SQLite, alimentación automática por uso, consulta, validación (`graph check`) y detección de anomalías.

## Estado de partida

- No existe código previo. Nada de lo descrito está implementado.
- La única pieza existente es la skill `boe-fiscal` (Python, `scripts/boe.py` + `references/boe_api.md` + `references/normas_fiscales.md`), que sirve de patrón y se porta a Go como applet `boe`.
- Nombre fijado: `kitlegal` para repo, módulo (`github.com/jmorenobl/kitlegal`), binario y directorios (`.kitlegal/`, `~/.cache/kitlegal/`).

## Primer hito

`kitlegal boe articulo BOE-A-2015-10565 a21` funcionando en Go con caché SQLite, y la skill `boe-fiscal` migrada para invocar el binario en lugar de `boe.py`, con el mismo comportamiento que hoy. Todo lo demás se construye encima.

## Decisiones ya tomadas

- Skills primero: cada hito entrega o mejora una skill medible con evals; solo va a Go lo que exige determinismo o verificabilidad.
- Skills sin código: `SKILL.md` + `references/` generadas desde `data/*.yaml`; `scripts/` son symlinks al binario.
- Genericidad territorial: nada se particulariza para un municipio; datos de municipio desde registros nacionales, lo territorial por comunidad o boletín, y cobertura declarada fuera del territorio de validación.
- Toda salida en JSON con `fuente`, `url`, `fecha_consulta`, `hash`. Sin eso no hay cita.
- No se abstrae `internal/cli` a librería externa hasta que tres applets repitan el patrón.
- Solo se automatizan fuentes públicas; presentar escritos, notificaciones y cualquier acción con identidad requieren firma humana (exit code 6).
- CENDOJ: solo resolución de ECLI y metadatos; nunca ingesta masiva.
- El grafo del asunto (`.kitlegal/` en cwd) nunca se sincroniza ni exporta por defecto.

## Pendiente de verificar antes de codificar

- Identificadores BOE de la tabla de leyes vertebrales (`kitlegal boe buscar`).
- Endpoints actuales de BDNS (swagger) y nombres de los feeds Atom de PLACSP.
- Relación de municipios del INE, inventario DIR3 y festivos locales de la Comunidad de Madrid (formato, licencia, actualización).
- Términos de uso de PETETE (DGT), DYCTEA (TEAC) y buscador HJ del TC.
