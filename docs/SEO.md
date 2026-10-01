# Visibilidad de kitlegal.es: Search Console y las herramientas de SEO

Este documento es para quien mantiene la web y no sabe de SEO. Explica qué hay montado en el repositorio para medir
y mejorar la visibilidad de https://kitlegal.es en Google y en los buscadores con IA, cómo se pone en marcha una sola
vez, y qué hacer cada semana. Nada de esto es parte del producto: son herramientas de agente para el mantenimiento
de la web (ADR 0024).

## Qué hay

| Pieza                                     | Dónde                                                                | Para qué                                                                                                                                                                        |
| ----------------------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Servidor MCP `gsc`                        | `.mcp.json` (Claude Code lo ofrece al abrir el repositorio)          | Da al agente los datos de Google Search Console: clics, impresiones, consultas, páginas, indexación y sitemaps. Es [mcp-gsc](https://github.com/AminForou/mcp-gsc), software libre, y corre en el equipo. |
| Skill `/seo-semanal`                      | `.agents/skills/seo-semanal/` (enlace en `.claude/skills/`)          | El informe de cada semana, en tres líneas y con acciones concretas. Lee los datos con `gsc`.                                                                                     |
| Plugin `claude-seo`                       | `.claude/settings.json` (`extraKnownMarketplaces`, `enabledPlugins`) | Auditorías de fondo: técnica, contenido, datos estructurados, visibilidad en buscadores con IA. Es [claude-seo](https://github.com/AgriciDaniel/claude-seo), MIT.                |
| `/llms.txt`, `robots.txt`, sitemap, JSON-LD | `web/src/pages/`, `web/src/lib/estructurados.ts`                   | Lo que la web ya dice de sí misma a Google y a los agentes de IA. Se genera al construir; no se edita a mano.                                                                     |

Lo que sale de Search Console tarda: una propiedad nueva no muestra datos hasta pasados unos días, y una web nueva no
tiene tendencias que leer hasta la tercera o cuarta semana. Las primeras semanas lo único que importa es la línea 3
del informe (indexación).

## Puesta en marcha, una sola vez

Hacen falta cuatro cosas: una clave de Google para que el agente lea Search Console, permiso para esa clave en la
propiedad de kitlegal.es, `uv` en el equipo, y abrir Claude Code en el repositorio. Unos 20 minutos.

### 1. La clave: una cuenta de servicio de Google

Una cuenta de servicio es un usuario de Google que no es una persona: tiene un correo propio y una clave en un fichero,
y con esa clave el servidor MCP lee Search Console sin abrir el navegador ni pedir contraseñas.

1. Entra en la consola de Google Cloud (console.cloud.google.com) con la cuenta de Google que tiene kitlegal.es en
   Search Console. Crea un proyecto (por ejemplo, `kitlegal-web`) o usa uno que ya tengas.
2. En **APIs y servicios → Biblioteca**, busca «Google Search Console API» y pulsa **Habilitar**.
3. En **IAM y administración → Cuentas de servicio → Crear cuenta de servicio**: nombre `kitlegal-gsc`, sin roles (no
   necesita ninguno en el proyecto), **Listo**.
4. Abre la cuenta recién creada, pestaña **Claves → Agregar clave → Crear clave nueva → JSON**. Se descarga un fichero.
5. Guárdalo donde `.mcp.json` lo espera y protégelo:

   ```sh
   mkdir -p ~/.config/kitlegal
   mv ~/Downloads/kitlegal-web-*.json ~/.config/kitlegal/gsc-service-account.json
   chmod 600 ~/.config/kitlegal/gsc-service-account.json
   ```

   Ese fichero es una contraseña: no va al repositorio (`.gitignore` lo rechaza por si acaso) ni a ningún chat.
   Apunta el correo de la cuenta de servicio, que aparece en la consola y dentro del fichero (`client_email`), con la
   forma `kitlegal-gsc@<proyecto>.iam.gserviceaccount.com`.

### 2. El programa que lanza el servidor: `uv`

`.mcp.json` lanza el servidor con `uvx`, que descarga el servidor y su Python la primera vez y lo cachea. En macOS:

```sh
brew install uv
```

En Linux o Windows, `pipx install uv` o el instalador de https://docs.astral.sh/uv/. Comprueba con `uvx --version`.

### 3. El permiso en Search Console

1. Entra en Search Console (search.google.com/search-console), propiedad de kitlegal.es.
2. **Configuración → Usuarios y permisos → Añadir usuario**: pega el correo de la cuenta de servicio, permiso
   **Completo**, **Añadir**.

Con «Completo» el servidor lee todo y puede gestionar sitemaps. Si más adelante una herramienta de `gsc` devuelve un
error de permisos, sube el permiso a «Propietario» (Search Console lo permite para cuentas de servicio como
propietario delegado).

### 4. Abrir Claude Code en el repositorio

1. Abre Claude Code en la raíz del repositorio. La primera vez pregunta si confías en la carpeta y si quieres usar el
   servidor MCP del proyecto (`gsc`): sí a las dos.
2. `/mcp` debe listar `gsc` como conectado. Si no, `/mcp` muestra el error: casi siempre es que falta `uvx` en el
   `PATH` o que el fichero de la clave no está donde dice `.mcp.json`.
3. `/plugin` debe listar `claude-seo` instalado desde el marketplace `agricidaniel-claude-seo`. La primera vez, ejecuta
   `/seo setup`, que comprueba sus dependencias (Python 3.10+; ofrece instalar Chromium para páginas que dependen de
   JavaScript, que kitlegal.es no necesita).
4. Prueba: `/seo-semanal`. Con una propiedad nueva dirá que aún no hay datos y solo informará de la indexación; eso
   es lo esperado.

## El primer día en Search Console

Estas tres cosas se hacen a mano, en la web de Search Console, y solo una vez:

1. **Sitemaps**: el sitemap de la web es `https://kitlegal.es/sitemap-index.xml` (lo declara `robots.txt`). Si en
   **Sitemaps** figura `sitemap.xml`, bórralo y envía `sitemap-index.xml`; `sitemap.xml` no existe y da 404.
2. **Inspección de URL**: pega cada página de la web (la portada, `/despachos/`, `/instalar/`, `/consultas/` y cada
   consulta, `/bot/`) y pulsa **Solicitar indexación**. Google tardaría semanas en llegar solo; así tarda días. Lo
   mismo cada vez que se añade una consulta.
3. **`/llms.txt`**: comprueba en el navegador que https://kitlegal.es/llms.txt responde. Lo genera la web en cada
   despliegue; si da 404, el despliegue de `main` no ha terminado o ha fallado (`.github/workflows/web.yml`).

## Cada semana

Los lunes, en Claude Code, en el repositorio: `/seo-semanal`. El informe tiene tres líneas y hasta tres acciones; con
leerlo y hacer las acciones basta. Lo que mide cada línea:

1. **Tendencia**: clics e impresiones frente a la semana anterior. Un clic es una visita desde Google; una impresión,
   una vez que la web apareció en los resultados. Con menos de 20 clics a la semana, las variaciones son ruido y el
   informe lo dice.
2. **Oportunidades**: consultas en las que la web aparece (muchas impresiones) y nadie entra (pocos clics). Suele
   arreglarse con el título y la descripción de la página que sale para esa consulta, o creando una página que la
   responda.
3. **Indexación**: cuántas páginas del sitemap tiene Google y cuáles no. Una página que no está indexada no puede
   aparecer; es lo único urgente que puede salir en el informe.

Una vez al mes, o cuando cambie la web de forma visible, una auditoría de fondo con el plugin:
`/seo audit https://kitlegal.es`. Su informe es largo; pídele al agente las tres cosas más importantes y haz esas.

## Lo que de verdad da visibilidad

Las herramientas miden; la visibilidad la dan tres cosas, por orden de efecto. Las dos últimas ya están hechas.

1. **Una página por consulta.** La gente no busca la ley: busca lo que le pasa («mi casero no me devuelve la fianza»,
   «plazo recurso de alzada», «cuánto tiempo tiene Hacienda para reclamar una deuda»). Cada consulta tiene su URL en
   `/consultas/<slug>/` (`web/src/data/consultas.ts`), con la pregunta como se escribe en el buscador, la respuesta
   corta y cada punto con la cita del BOE que lo sostiene, tal como la devuelve el binario (ADR 0024 y 0034: ninguna
   cita a mano y ninguna afirmación sin su cita al lado). Hay seis; con veinte o treinta, la web deja de ser una
   portada y pasa a ser una referencia. Es el trabajo que más visibilidad da y el único que no se acaba nunca:
   `/seo-semanal` dirá qué consultas piden página, y hasta que haya datos sirve el autocompletado de Google (las
   sugerencias de una búsqueda son las búsquedas que la gente hace).
2. **`/llms.txt`**: el índice de la web para los agentes de IA. Hecho (#72); se genera desde la versión publicada y
   lista las consultas.
3. **Datos estructurados (JSON-LD)**: `SoftwareApplication` y `SoftwareSourceCode` en las dos portadas, migas en las
   demás páginas y `FAQPage` en cada consulta. Hecho (`web/src/lib/estructurados.ts`).

Dos públicos, los dos de usuario final (ADR 0034): la ciudadanía, en la portada, y los despachos, en `/despachos/`.
La web no habla al perfil técnico —que se lo monta solo o llega como ciudadano— ni a las administraciones públicas,
así que los directorios de skills y de servidores MCP y las listas para desarrolladores no son una vía de la web.
Las consultas traen a los dos públicos; el título y la descripción de `/despachos/` llevan lo que un abogado escribe
(«IA para abogados», «España», «gratis», el asistente con el que funciona). Search Console dirá en dos meses qué
consultas responden.

## Si algo falla

- `/mcp` dice que `gsc` no arranca: `uvx --python ">=3.11" mcp-search-console` en una terminal enseña el error real.
  Lo habitual es que falte `uv` o que `GSC_CREDENTIALS_PATH` no apunte al fichero.
- `list_properties` devuelve una lista vacía: la cuenta de servicio no tiene permiso en la propiedad (paso 3), o la
  propiedad de Search Console está en otra cuenta de Google.
- Un error de cuota de la API: Search Console limita las consultas por día; el informe semanal está muy por debajo,
  pero una auditoría del plugin con muchas URL puede agotarla. Espera al día siguiente.
- Para cambiar la ruta del fichero de la clave, edita `GSC_CREDENTIALS_PATH` en `.mcp.json`; `${HOME}` se expande.
