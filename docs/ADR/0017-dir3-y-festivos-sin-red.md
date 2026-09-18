# 0017 · El DIR3 de los ayuntamientos y los festivos locales no se piden en red: se congelan en `data/`

- **Estado**: aceptada
- **Fecha**: 2026-09-18
- **Hito**: pieza previa a H6; enmienda el alcance y los controles de H6 y el alcance de H9 en `docs/ROADMAP.md` §4.
  No toca el contrato `Applet` (ADR 0005), ni el sobre (ADR 0006), ni `internal/httpx`

## Contexto y problema

H6 prometía «DIR3 desde el inventario público (formato, licencia y forma de actualización a verificar en tarea
`[datos]`, con fila en `docs/SOURCES.md`)» y H9 «formato a verificar en tarea `[datos]`» para los festivos locales.
Las dos frases daban por supuesto lo mismo: que hay un fichero público que un cliente automatizado puede descargar, y
que solo faltaba mirar su formato. La investigación de fuentes (issue #37) comprobó que para el DIR3 eso es falso y que
para los festivos de Madrid está prohibido:

- El conjunto nacional de DIR3 del Catálogo Nacional de Datos Abiertos (`E05188501`) está **despublicado**: responde
  `HTTP 404` y sus distribuciones RDF no resuelven. El portal web que lo sustituye está tras Imperva/Incapsula.
- El área de descargas del CTT de Administración Electrónica sí publica `Listado Unidades EELL.xlsx` (~2 MB), pero tras
  el desafío de un WAF (F5 ASM): a un cliente no interactivo —`curl`, `internal/httpx`— le devuelve un cuerpo de 0
  bytes. Además su contenido son unidades orgánicas, no una correspondencia 1:1 con municipios del INE.
- Los servicios web de DIR3 (SOAP y REST) existen, pero solo dentro de la **Red SARA** y con credenciales de organismo
  autorizado. FACe tampoco sirve: no ofrece descarga masiva, sus servicios exigen certificado X.509 y no devuelve el
  código INE del municipio.
- La única vía pública, automatizable y descargable limpiamente sin credenciales ni WAF es el **Registro de Entidades
  Locales** (`rel_municipios_espana.xls`, ~2 MB), que trae todos los municipios con su número de inscripción registral
  (`01PPMMMDC`), del que se sigue —por el ENI y sus normas técnicas— el DIR3 del ayuntamiento (`L01PPMMMDC`).
- Para los festivos, el `robots.txt` de `datos.comunidad.madrid` es `User-agent: *` / `Disallow: /`. El dato está
  publicado en CC BY y su API CKAN funciona, pero el sitio prohíbe el acceso automatizado: `internal/httpx` no puede
  pedirlo, ni en ejecución ni en una grabación.

El problema de fondo no es el formato de ningún fichero. Es que «la fuente es pública» y «la fuente se puede consultar
desde el binario» son dos cosas distintas, y el roadmap las trataba como una.

## Decisión

1. **Ningún applet pide DIR3 ni festivos locales por red.** No hay base de DIR3 ni de `datos.comunidad.madrid` en
   `internal/httpx`, ni fila de fuente consultable en ejecución en `docs/SOURCES.md` para ninguna de las dos. Su fila
   documenta el origen del fichero congelado y la fecha en que se generó.
2. **Los dos datos entran como fichero versionado**, en `data/territorio/` y `data/festivos/`, generados por una tarea
   `[datos]` fuera de la ejecución normal, a partir de descargas públicas: la relación de municipios del INE (que
   aporta el dígito de control oficial), el REL (que aporta el número de inscripción) y el conjunto histórico de
   festivos de la Comunidad de Madrid.
3. **La derivación del DIR3 desde el número de inscripción es una hipótesis, no un dato.** La tarea `[datos]` la
   verifica contra DIR3 real antes de fijar el fichero, en una muestra que incluya los casos donde puede romperse.
   Cada fila lleva su `source`. Donde la derivación no se verifica, la fila **no entra**: `territorio` lo declara en
   `cobertura` y nunca devuelve un código calculado como si fuera registral, igual que no da por inexistente un
   boletín que no tiene configurado.
4. **Una fuente tras un WAF, un CAPTCHA, un certificado o la Red SARA no es una fuente automatizable**, aunque el dato
   sea público y reutilizable. El criterio no es si el fichero se puede obtener, sino si se puede obtener sin simular a
   una persona; hacerlo sería cruzar la frontera humana por el otro lado (ADR 0004). Cuando solo queda esa vía, o hay
   una alternativa pública —aquí, el REL— o el dato entra congelado con su procedencia, o no entra.

## Consecuencias

- H6 deja de depender de la red para el DIR3: `territorio resolver` responde con el repositorio y nada más, y el mismo
  camino sirve para `--offline`. El dato envejece con el repositorio y se refresca repitiendo la tarea `[datos]`; los
  municipios cambian un puñado de veces al año y el INE publica las modificaciones de forma continua.
- La verificación del punto 3 es trabajo real dentro de H6, no un trámite: si la derivación falla en más casos de los
  esperados, la correspondencia deja de ser una regla y pasa a ser una tabla, y eso se decide con la muestra delante.
- En el grafo (ADR 0014), un `Organo` con id natural DIR3 solo se crea con la fila verificada detrás. Un municipio sin
  DIR3 verificado no crea `Organo`; el grafo prefiere el hueco a la invención.
- `scripts/verify-sources.sh` no puede vigilar estas dos fuentes, porque no se piden en red. Lo que envejece aquí no lo
  detecta el CI nightly: lo detecta quien repite la tarea `[datos]`.
- La investigación que sostiene esto está en el issue #37, con la fecha de cada comprobación. Este ADR recoge lo que
  cambia de decisión; el detalle de URLs, campos, licencias y cabeceras vive en `docs/SOURCES.md`.
