# Evidencia del ADR 0039: los festivos locales de la Comunidad de Madrid, el BOCM frente al catálogo CKAN

Material del cotejo que sostiene el ADR 0039, hecho el 2026-10-17 en una sesión interactiva a petición de Jorge, fuera
de cualquier run (ADR 0032). Como todo `evidencias/`, un run del workflow no la escribe. Las huellas de cada fichero
están en `manifiesto.json`.

## Qué hay

- `festivos_locales_historicos.json`: el conjunto histórico de festivos locales del catálogo CKAN de
  `datos.comunidad.madrid` (fila `comunidad-madrid.festivos-locales` de `docs/SOURCES.md`). Lo descargó Jorge con su
  navegador el 2026-10-17, porque el `robots.txt` del sitio es `Disallow: /` para cualquier agente. CC BY 4.0, «Fuente:
  Comunidad de Madrid». 10 096 filas, de 1998 a 2025; campos `año`, `municipio_codigo` (los tres dígitos del municipio
  en el código INE), `municipio_nombre`, `entidad_codigo` (0, el municipio; otro, una entidad menor, como Cerceda o
  Mataelpino en El Boalo), `entidad_nombre` y `fecha_festivo`.
- `bocm/`: el XML de las siete disposiciones del Boletín Oficial de la Comunidad de Madrid que declaran o corrigen las
  fiestas laborales de ámbito local de 2025 y 2026, pedidas el 2026-10-17 a
  `https://www.bocm.es/boletin/CM_Orden_BOCM/AAAA/MM/DD/BOCM-AAAAMMDD-N.xml`, a 10 s entre peticiones, como pide su
  `robots.txt`. En el run de H9 las vuelve a grabar `grabar_datos` en `evidencias/h9/`; estas son las del cotejo.

| Año | Disposición | Qué hace |
|---|---|---|
| 2025 | `BOCM-20241213-23` | Resolución de 3 de diciembre de 2024: las fiestas locales de 2025; seis municipios «no comunicado» |
| 2025 | `BOCM-20241227-32` | Resolución de 16 de diciembre de 2024: Belmonte de Tajo, Montejo de la Sierra, Patones y Talamanca de Jarama |
| 2025 | `BOCM-20250203-16` | Resolución de 21 de enero de 2025: Cabanillas de la Sierra, La Hiruela y Majadahonda |
| 2025 | `BOCM-20250217-24` | Resolución de 4 de febrero de 2025: Daganzo de Arriba |
| 2026 | `BOCM-20251212-34` | Resolución de 2 de diciembre de 2025: las fiestas locales de 2026; catorce municipios «no comunicado» |
| 2026 | `BOCM-20251229-19` | Resolución de 15 de diciembre de 2025: ocho de los catorce y un cambio en Coslada |
| 2026 | `BOCM-20260108-18` | Resolución de 22 de diciembre de 2025: los seis restantes |

Con las tres de 2026, ningún municipio queda «no comunicado». Las correcciones se buscaron con un buscador web por el
texto con que se titulan («modifica parcialmente la Resolución de …»), porque el buscador del BOCM está cerrado a los
agentes (`Disallow: /search/`) y la página de la Comunidad sobre el calendario laboral solo enlaza las principales.
No se encontró ninguna otra de esos dos años; una posterior al 2026-10-17 no está aquí.

## El cotejo de 2025

Se leyeron las líneas «Municipio: fechas.» de las cuatro disposiciones de 2025, aplicadas en el orden en que se
publicaron, y se compararon con las filas de 2025 del fichero CKAN, casando por el nombre sin artículos ni
preposiciones (lo hizo un guion de un solo uso, fuera del repositorio y sin versionar, que no es parte del producto; en
el run, el cruce se hace por código INE).

- **Coinciden** 172 entidades.
- **Faltan en el fichero** los seis municipios que añadieron las correcciones: Belmonte de Tajo, Cabanillas de la
  Sierra, La Hiruela, Montejo de la Sierra, Patones y Talamanca de Jarama.
- **El fichero no aplica dos correcciones**: Daganzo de Arriba, 1 de septiembre donde la de 4 de febrero de 2025 dice el
  8; Majadahonda, 14 de septiembre donde la de 21 de enero de 2025 dice el 15.
- **El fichero da otra fecha sin ninguna corrección detrás**: Fuente el Saz de Jarama, 6 de septiembre donde la
  resolución principal dice «8 y 9 de septiembre».
- **Una línea no se lee sin interpretarla**: «Valdeolmos-Alalpardo: 15 de mayo y 11 de agosto (Valdeolmos) y 25 de
  agosto (Alalpardo)». El fichero la lee como 15 de mayo y 11 de agosto en Valdeolmos y 15 de mayo y 25 de agosto en
  Alalpardo; es una lectura razonable, pero es una lectura (ADR 0039, punto 3).

De 2026 no hay cotejo posible: el fichero no trae ese año.
