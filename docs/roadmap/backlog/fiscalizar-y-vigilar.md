# Fiscalizar y vigilar

ADR 0014, pieza G2; necesita entera «Transparencia y fiscalización municipal», en `transparencia-y-fiscalizacion/`.

- `graph neighbors | path` + reglas de anomalías de contratación en `data/anomalias/*.yaml` (SQL + umbrales): `fraccionamiento`, `licitador-unico`, `concentracion`, `plazo-publicacion`. Las reglas con umbral legal lo toman de la ley; las relativas comparan con municipios de tamaño parecido (tramos de población del padrón del INE). Salida con `nodos`, `aristas`, `verificar_en`. `contratacion-publica` v1. Cada regla con un grafo sintético que la dispara y otro que no; validación del YAML contra schema; SQL de solo lectura (`SELECT`/`WITH`); ninguna regla referencia un municipio u órgano concreto. Todo cambio en `data/anomalias/` es gate humano (constitución, capa 3).
- Cruces entre fuentes + pack `fiscalizador`: `administrador-comun`, `beneficiario-adjudicatario`, `empresa-recien-creada`. `subvenciones` v1. Validación con casos conocidos de Leganés (fixtures anonimizados de aristas, no de personas) más un municipio de otra comunidad.
- `vigilar run` sobre delta + `anomalies mark --confirmada|--descartada`: orquesta `placsp sync`, `bdns`, `borme sumario`, `boletin`, `edictos` y `boe vigilar` sobre el municipio y las entidades de `config.yaml`; anomalías solo sobre aristas nuevas. E2e de dos ejecuciones consecutivas con fixtures distintos; solo emite la diferencia.
- `graph export --format md|jsonld|dot|csv` con anonimización por defecto; la exportación del asunto pide confirmación.

---
