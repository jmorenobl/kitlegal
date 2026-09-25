# Evidencia de H6

Material de origen del que salen `data/territorio/` y la verificación del DIR3 (ADR 0017, ADR 0018). Desde el
workflow 2.0.0 esta carpeta la escribe el paso `grabar_datos`; la de H6 es anterior y se trajo el 2026-09-25 desde la
copia local `~/.local/share/kitlegal/evidencias/h6/`, comprobando cada huella contra las que registran
`specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md`, `gates/tarea-T001.md` y `gates/tarea-T003.md`.
`manifiesto.json` da, por fichero, su huella SHA-256, su tamaño, la fila de `docs/SOURCES.md` de la que sale, la fecha
y la dirección.

| Fichero | Fuente | Qué es |
|---|---|---|
| `diccionario26.xlsx` | `ine.municipios` | Relación de municipios del INE: código, dígito de control, provincia y comunidad (`data/territorio/municipios.yaml`) |
| `cod_ccaa.htm`, `cod_provincia.htm` | `ine.codigos-territoriales` | Tablas de códigos del INE, de donde salen los nombres de comunidades y provincias |
| `rel.xls` | `mpt.rel` | Volcado de municipios del Registro de Entidades Locales (2026-09-21): el número de inscripción del que se deriva el DIR3 |
| `eatimes.xls` | `mpt.rel` | Volcado de entidades de ámbito territorial inferior al municipio (2026-09-22) |
| `pag-fichas/ficha-L01….html` | `pag.directorio` | Las siete fichas de unidad orgánica del Punto de Acceso General con las que se verificó la derivación contra DIR3 real (2026-09-24). Fuente: Punto de Acceso General, administracion.gob.es |
| `pag-fichas/pag-resultados.txt` | — | Registro de esa consulta, escrito por el guion de H6 |

Las tres descargas del INE se perdieron del directorio temporal de la sesión de H6; las de aquí se descargaron de nuevo
el 2026-09-25 y tienen exactamente la misma huella que las originales del 2026-09-20.

No se versionan, porque no son evidencia de ningún dato: la primera tanda de consultas al PAG (2026-09-23), en la que
cinco de siete páginas eran la misma copia cacheada, ni los guiones de un solo uso con los que se consultó y se leyó.
