# 0024 · La web del proyecto: kitlegal.es, en `web/`, sin ninguna cita escrita a mano

- **Estado**: aceptada
- **Fecha**: 2026-09-27
- **Hito**: transversal (distribución, antes de H7). Cambia la dirección que publica el binario en su identificación
  (FR-006 de H2) y el `$id` de los esquemas de `schemas/` (FR-110 de H4).

## Contexto y problema

kitlegal no tenía web. Quien llegaba desde un buscador encontraba el repositorio; quien administra un sitio y veía en
sus registros `kitlegal/x.y (+https://ventanillalegal.es/bot)` no encontraba nada en esa dirección, y el `$id` de
cada esquema de `schemas/` apuntaba a otra dirección que tampoco resolvía.

Hay una maqueta de marca (Stitch: `DESIGN.md` y una portada en HTML) con dos audiencias, despachos y ciudadanía. Su
contenido no se podía publicar tal cual: citaba el Estatuto de los Trabajadores con un identificador que no existe
(`BOE-A-2015-12700`, que el BOE no encuentra) y metía texto inventado entre las comillas de su art. 49.1.d;
parafraseaba entre comillas otros tres artículos; y prometía compatibilidad con MCP, Cursor y Windsurf, fuentes que
el binario no consulta, «cero alucinación» y «100 % local». Una web de una herramienta cuya regla es «nada sin cita»
no puede hacer lo que la herramienta prohíbe.

Faltaba decidir dónde vive la web, con qué se construye, en qué dirección se publica y qué garantiza su contenido.

## Opciones consideradas

**Dónde.**

1. **Un repositorio aparte.** Rechazada: la web lee `data/` (municipios), `schemas/` y la última versión publicada;
   en otro repositorio serían copias que se separan del original (la maqueta ya decía 8.131 municipios donde
   `data/territorio/municipios.yaml` tiene 8.132).
2. **`docs/`**, el directorio que GitHub Pages sabe publicar sin flujo. Rechazada: `docs/` es la documentación del
   proyecto (ADR, hoja de ruta), y con un flujo de Actions Pages publica cualquier directorio.
3. **`pages/`.** Rechazada: nombra dónde se aloja, no lo que es, y con Astro daría `pages/src/pages/`.
4. **`web/`**, en este repositorio.

**Con qué.**

1. **HTML escrito a mano.** Rechazada: varias páginas con la misma cabecera y el mismo pie, y datos del repositorio
   que tienen que llegar al HTML al construir, no a mano.
2. **Hugo.** Es Go y no necesita Node, y se podría fijar en `tools/` como las demás herramientas. Rechazada: la
   maqueta es un diseño de componentes (insignias de vigencia, citas, tarjetas), y las plantillas de Go son la peor
   forma de expresarlo; la ventaja de no tener Node se consigue igual dejando la web fuera de `make ci`.
3. **Astro**: HTML estático sin JavaScript por omisión, componentes, Tailwind y la API de Node al construir para
   leer el repositorio y los sobres.

**Dirección.**

1. **`jmorenobl.github.io/kitlegal`.** Rechazada: la autoridad que acumula una dirección es de quien la tiene, y
   mudarse después cuesta lo acumulado.
2. **`ventanillalegal.es`**, la que ya publicaba el binario. Rechazada como dirección principal: el nombre del
   proyecto es kitlegal (nombre único) y `kitlegal.es` es del proyecto.
3. **`kitlegal.es`**, con `ventanillalegal.es` redirigida allí.

## Decisión

- **La web vive en `web/`** y es un proyecto Astro que produce HTML estático. Se publica en **https://kitlegal.es**
  con GitHub Pages, desde `.github/workflows/web.yml`: construye y comprueba en cada propuesta de cambio que toca la
  web o lo que la web lee, y despliega en `main` y al terminar cada release.
- **Ninguna frase de una norma se escribe a mano.** Cada cita de la web se declara en `web/src/data/citas.yaml`
  (norma, bloque y fragmentos) y su texto sale del sobre que devolvió el binario (`kitlegal boe articulo … --json`),
  versionado en `web/src/data/sobres/`. La construcción falla si un fragmento no está literal en su sobre, si los
  fragmentos no siguen el orden del texto, si un sobre no es de éxito o no es de esa norma y ese bloque, y si sobra o
  falta un sobre. La página muestra con cada cita la norma, el bloque, la fecha de consulta, la fecha de la versión,
  la huella y el enlace al BOE, y el pie da la fuente de los datos como piden las condiciones de reutilización del
  BOE. Los avisos de vigencia son los del sobre de `kitlegal boe metadatos`. `make web-citas` regenera los sobres
  con el binario de `make build`, y es la única orden de la web que toca la red.
- **Ninguna cifra se escribe a mano**: el número de municipios sale de `data/territorio/municipios.yaml` y la versión,
  de la API de releases (en local, de la última etiqueta).
- **Una URL por audiencia**: `/` para despachos y `/ciudadania/` para ciudadanía, cada una con su título, su
  descripción y sus ejemplos. El selector de la maqueta, que cambiaba el texto con JavaScript, es navegación entre
  las dos.
- **La web publica lo que el binario ya enlaza**: `/bot/` explica el rastreador (qué es, cómo se comporta, cómo
  limitarlo en `robots.txt`) y `/schemas/<fichero>` sirve cada esquema de `schemas/` en la dirección de su `$id`.
- **El binario pasa a `kitlegal.es`**: la identificación es `kitlegal/x.y (+https://kitlegal.es/bot)` y los `$id`
  de `schemas/` son `https://kitlegal.es/schemas/…`. Las grabaciones de `testdata/` y de `internal/*/testdata/`
  conservan la identificación con la que se grabaron: registran lo que se envió, y la reproducción no la compara.
  `ventanillalegal.es` se redirige a `kitlegal.es` conservando la ruta, para las versiones hasta la 0.2.0.
- **Sin nada de terceros en la página**: fuentes servidas desde la propia web (Fontsource), iconos en SVG en línea,
  ni scripts ni estilos en línea, y una política de seguridad de contenido que lo impone. Las imágenes para compartir
  y el icono de Apple se generan al construir.
- **Fuera de `make ci`**: la web necesita Node y pnpm, y comprobar el producto no. `make web`, `make web-dev` y
  `make web-citas` son sus órdenes; el flujo de la web usa `make web`. pnpm no instala una versión hasta que lleva
  una semana publicada (`minimumReleaseAge`), y Dependabot espera lo mismo (`cooldown`).

## Consecuencias

- Cambiar una cita de la web es cambiar `citas.yaml` y ejecutar `make web-citas`; escribir el texto a mano no
  construye.
- El binario se identifica con una dirección que resuelve, y los esquemas publicados se pueden descargar por su
  `$id`. Para quien valide contra los esquemas por su `$id`, el cambio de dirección es un cambio incompatible de la
  próxima versión.
- La persona tiene que activar Pages con origen «GitHub Actions» y el dominio `kitlegal.es` en el repositorio,
  apuntar el DNS de `kitlegal.es` a GitHub Pages, verificar el dominio en la cuenta y redirigir `ventanillalegal.es`.
- El repositorio tiene un segundo ecosistema de dependencias (npm, en `web/`), aislado: no entra en `make ci`, en el
  binario ni en la release.
