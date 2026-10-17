# 0039 · Los festivos salen del boletín que los declara, con sus correcciones, y se congelan en `data/festivos/`

- **Estado**: aceptada (2026-10-17)
- **Fecha**: 2026-10-17
- **Hito**: pieza previa a H9; enmienda su entrada en `docs/ROADMAP.md`. Sustituye, del ADR 0017, la fuente de los
  festivos locales —«el conjunto histórico de festivos de la Comunidad de Madrid» del catálogo CKAN de
  `datos.comunidad.madrid`, punto 2 de su decisión— y nada más: el DIR3, la regla de que ningún applet pide festivos
  en ejecución y el punto 4 siguen vigentes. No toca el contrato `Applet` (ADR 0005), ni el sobre (ADR 0006), ni
  `internal/httpx`

## Contexto y problema

El ADR 0017 congeló los festivos locales desde el conjunto histórico que la Comunidad de Madrid publica en su catálogo
CKAN, porque el `robots.txt` de `datos.comunidad.madrid` (`User-agent: *` / `Disallow: /`, comprobado de nuevo el
2026-10-17) impide que `internal/httpx` lo pida. Preparando la entrada de H9, una persona descargó ese fichero con su
navegador y se cotejó con lo que publica el Boletín Oficial de la Comunidad de Madrid, que es donde la Dirección
General de Trabajo declara las fiestas laborales de ámbito local de cada año (art. 46 del Real Decreto 2001/1983). El
cotejo (`evidencias/adr-0039/LEEME.md`) da tres cosas:

- **El fichero no sirve para plazos de hoy.** Llega hasta 2025. No trae 2026, que es el año en que corren los plazos
  que se calculan ahora.
- **El fichero no recoge las correcciones.** Cada año la resolución principal sale en diciembre con municipios «no
  comunicado por el Ayuntamiento», y la completan o la cambian resoluciones posteriores que la «modifican
  parcialmente». En 2025, el BOCM y el fichero coinciden en 172 entidades, pero al fichero le faltan los seis municipios
  que añadieron las correcciones (Belmonte de Tajo, Cabanillas de la Sierra, La Hiruela, Montejo de la Sierra, Patones y
  Talamanca de Jarama), conserva la fecha anterior a su corrección en Daganzo de Arriba y Majadahonda, y en Fuente el
  Saz de Jarama da el 6 de septiembre donde el BOCM dice el 8.
- **El BOCM sí se puede pedir.** Su `robots.txt` admite a cualquier agente con `Crawl-delay: 10` y solo cierra rutas de
  administración y de búsqueda, y cada disposición se publica en XML con sus metadatos y su texto
  (`/boletin/CM_Orden_BOCM/AAAA/MM/DD/BOCM-AAAAMMDD-N.xml`). Su aviso legal reserva la propiedad intelectual «de este
  canal, su diseño gráfico y los códigos»; las resoluciones de un órgano administrativo no son objeto de propiedad
  intelectual (art. 13 del texto refundido de la Ley de Propiedad Intelectual, Real Decreto Legislativo 1/1996), y lo
  que se extrae de ellas son fechas.

Con los festivos nacionales y autonómicos pasa lo mismo en otro boletín: los declara cada año la resolución de la
Dirección General de Trabajo que publica en el BOE la relación de fiestas laborales (art. 45.4 del Real Decreto
2001/1983). Es el «calendario laboral oficial» al que se sujetan los calendarios de días inhábiles del art. 30.7 LPAC
y los «festivos a efectos laborales» del art. 182.1 LOPJ. El `robots.txt` de `boe.es` cierra `/diario_boe/xml.php?` y
los PDF, y deja abierto `/diario_boe/txt.php`.

## Decisión

1. **Los festivos se toman del boletín que los declara.** Los nacionales y autonómicos, de la resolución anual de la
   Dirección General de Trabajo en el BOE (`/diario_boe/txt.php?id=BOE-A-…`); los locales de la Comunidad de Madrid,
   de la resolución anual de su Dirección General de Trabajo en el BOCM y de cada resolución que la modifica (el XML de
   la disposición). Ningún resumen, catálogo de datos abiertos ni página de terceros es la fuente.
2. **Se graban con `grabar_datos` y se congelan en `data/festivos/`.** Una tarea `[datos]` declara las disposiciones
   por su identificador, `grabar_datos` las pide por `internal/httpx` con el ritmo del `robots.txt` de cada sitio, y el
   código deriva `data/festivos/` de lo grabado, con la `source` de cada fila: el identificador de la disposición que
   fija esa fecha. Ningún applet los pide en ejecución (ADR 0017, punto 1).
3. **Una corrección gana a lo que corrige, y lo que no se entiende no entra.** Las disposiciones de un año se aplican
   en el orden en que se publicaron. Un municipio «no comunicado» no tiene festivos locales conocidos hasta que una
   corrección los da. Una línea que no se puede leer sin interpretarla —una entidad menor cuyas fechas se reparten con
   paréntesis, por ejemplo— deja esa entidad fuera de cobertura, con aviso, en lugar de adivinar el reparto.
4. **La cobertura es por año y por municipio, y se dice.** Un plazo que pasa por un año, o por un municipio, sin
   festivos locales cargados avisa en `data` de que le faltan, y nunca los da por inexistentes. Los datos dicen hasta
   qué disposición llegan: una corrección publicada después de la grabación no se conoce hasta que se repite la tarea
   `[datos]`.
5. **El fichero del catálogo CKAN no es fuente.** Se conserva en `evidencias/adr-0039/` como material del cotejo que
   sostiene esta decisión, y su fila de `docs/SOURCES.md` deja de tener destino en `data/`.

## Consecuencias

- La Comunidad de Madrid queda cubierta en 2025 y 2026 con las correcciones conocidas el 2026-10-17, y los festivos
  nacionales y autonómicos de todas las comunidades en 2025, 2026 y 2027. Las fiestas locales de 2027 se publicarán en
  diciembre: llegan repitiendo la tarea `[datos]` con la resolución nueva, y hasta entonces un plazo que entra en 2027
  avisa de que le faltan.
- Hay que encontrar las correcciones. El buscador del BOCM está cerrado a los agentes (`Disallow: /search/`) y la
  página de la Comunidad que enlaza el calendario solo lista las resoluciones principales, así que la lista de
  disposiciones de cada año la fija una persona, o una sesión interactiva a petición suya, antes del run. La de 2025 y
  2026 está en `evidencias/adr-0039/LEEME.md`.
- Otras comunidades entran igual, con su boletín, desde el backlog («territorios»), cada una con su fila en
  `docs/SOURCES.md`.
- `scripts/verify-sources.sh` sigue sin vigilar estos datos, porque no se piden en ejecución (ADR 0017).
