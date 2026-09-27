---
name: seo-semanal
description: >-
  Informe semanal de visibilidad de kitlegal.es en Google, en tres líneas y con acciones concretas, a partir de los
  datos de Google Search Console (servidor MCP `gsc`, .mcp.json). Úsala cuando se pida el informe SEO, el resumen
  semanal, «cómo va la web en Google», qué páginas suben o bajan, qué consultas tienen impresiones sin clics o qué
  URL no está indexada. No es parte del producto: es una herramienta de quien mantiene la web (docs/SEO.md).
---

# Informe semanal de kitlegal.es en Google

Quien lee este informe no sabe de SEO y no quiere aprenderlo: quiere saber en un minuto si la web se ve más o menos
que la semana pasada, por qué, y qué hacer. El informe tiene siempre la misma forma (abajo) y cabe en una pantalla.
Todo lo que dice sale de los datos que devuelven las herramientas del servidor MCP `gsc` en esta misma conversación;
ninguna cifra se estima, se recuerda ni se inventa. Si una herramienta falla o no hay datos, el informe lo dice en
lugar de rellenar el hueco.

## Antes de empezar

1. Comprueba que las herramientas `gsc` están disponibles (`get_capabilities` o `list_properties`). Si no lo están, no
   sigas: di que el servidor MCP no está configurado y remite a `docs/SEO.md`, que explica cómo se pone en marcha
   (cuenta de servicio de Google, clave en `~/.config/kitlegal/gsc-service-account.json`, permiso en Search Console).
2. Con `list_properties`, toma la propiedad de kitlegal.es. Puede ser una propiedad de dominio (`sc-domain:kitlegal.es`)
   o de prefijo (`https://kitlegal.es/`); si hay las dos, usa la de dominio. Si no aparece ninguna, la cuenta de
   servicio no tiene permiso todavía: dilo y remite a `docs/SEO.md`, paso 3.
3. Fija las dos ventanas. Search Console publica los datos con dos o tres días de retraso, así que la semana «actual»
   son los 7 días que terminan hace 3 días, y la anterior, los 7 días previos. Escribe las dos fechas en el informe.

## Datos que se leen

Pide siempre lo mismo, en este orden, y guarda los resultados para redactar:

| Qué                                   | Herramienta                                                    | Detalle                                                             |
| ------------------------------------- | -------------------------------------------------------------- | ------------------------------------------------------------------- |
| Totales de las dos semanas            | `compare_search_periods` (o dos `get_performance_overview`)    | clics, impresiones, CTR y posición media de cada ventana            |
| Páginas, esta semana y la anterior    | `get_search_analytics` con dimensión `page`, una vez por ventana | hasta 25 filas por ventana                                          |
| Consultas de esta semana              | `get_search_analytics` con dimensión `query`                   | hasta 50 filas, ordenadas por impresiones                           |
| Página y consulta juntas              | `get_search_by_page_query`                                     | solo para las 3 consultas del punto «oportunidades»                 |
| Sitemap                               | `list_sitemaps_enhanced` o `get_sitemaps`                      | que `sitemap-index.xml` esté enviado, sin errores, y cuándo se leyó |
| Indexación de las páginas de la web   | `check_indexing_issues` o `inspect_url_enhanced` por URL       | las URL del sitemap publicado (ver abajo)                           |

Las URL de la web son las que lista el sitemap publicado: `https://kitlegal.es/sitemap-0.xml` (lo enlaza
`sitemap-index.xml`). Si no puedes leerlo, usa las rutas de `web/src/data/paginas.ts` y `web/src/data/audiencias.ts`
del repositorio, más la portada `/`. No inspecciones URL que no estén en el sitemap.

Si la propiedad es nueva y los totales de las dos ventanas son cero, no hay nada que comparar: emite solo la línea 3
(indexación) y di que las líneas 1 y 2 empezarán a tener datos cuando Google lleve unas semanas mostrando la web.

## Cómo se decide cada línea

**Línea 1, la tendencia.** Compara los clics y las impresiones de las dos semanas. Con menos de 20 clics semanales, una
variación es ruido: dilo así («pocos datos aún») y no hables de subidas ni bajadas. Con más, nombra la página que más
sube y la que más baja en clics (no en posición), y la consulta que más impresiones trae.

**Línea 2, las oportunidades.** Son las consultas con muchas impresiones y pocos clics, que es donde la web aparece y
nadie entra. Criterio: al menos 30 impresiones en la semana y CTR por debajo de la mitad del CTR medio de la
propiedad, o posición media entre 4 y 15. Ordena por impresiones y quédate con las tres primeras. Para cada una,
mira con `get_search_by_page_query` qué página de kitlegal.es sale en esa consulta: la acción es casi siempre que el
título y la descripción de esa página respondan a la consulta tal como la escribe la gente, o crear una página que la
responda si ninguna lo hace. Nunca propongas tocar una cita legal del contenido: las citas de la web se comprueban
contra el binario al construirla y no son material de SEO.

**Línea 3, la indexación.** Cuenta cuántas URL del sitemap están indexadas y cuáles no. Una URL no indexada es la única
urgencia real del informe: escribe la URL, el motivo que da Google (por ejemplo «descubierta, sin indexar» o
«rastreada, sin indexar») y la acción («pedir indexación en Search Console» si es nueva; «revisar que enlaza desde
la portada» si lleva más de tres semanas). Si el sitemap tiene errores o no se ha enviado, va aquí también.

## Forma del informe

Escribe exactamente esto, en español, sin secciones adicionales ni explicaciones de método:

```
kitlegal.es · semana del <día> al <día> (frente a <día>–<día>)

1. Tendencia: <clics> clics (<±n%>) y <impresiones> impresiones (<±n%>). Sube <página>; baja <página>. La
   consulta que más nos muestra es «<consulta>».
2. Oportunidades: «<consulta>» (<impresiones> impresiones, <CTR>, posición <n>) sale con <página>; «…»; «…».
3. Indexación: <n> de <total> páginas indexadas. <URL no indexada: motivo>. Sitemap <ok / problema>.

Qué hacer esta semana (máximo 3):
- <acción concreta, una frase, con la página o consulta a la que se refiere>
- …
```

Reglas de redacción:

- Porcentajes redondeados a enteros; CTR con un decimal; posición con un decimal. Sin decimales de más.
- Si una línea no tiene datos suficientes, se escribe igualmente con la frase «pocos datos aún» y nada más.
- Las acciones son cosas que una persona puede hacer en menos de una hora, sobre una página o consulta concreta. No
  propongas «mejorar el SEO», «crear más contenido» ni nada sin una URL o una consulta al lado.
- No adjuntes tablas ni exportes datos salvo que se pidan. Si se piden, entonces sí: la tabla completa de páginas o
  consultas, tal como la devuelve la herramienta.

## Lo que esta skill no hace

- No cambia nada en Search Console (no envía sitemaps ni pide indexaciones): lo propone y lo hace la persona.
- No edita la web. Si una acción es un cambio en `web/`, propónla y deja que la persona decida; la web tiene sus propias
  reglas (ADR 0024) y `make web` la comprueba.
- No interpreta datos de otras herramientas ni de otras webs. Para auditorías de fondo (técnica, contenido, datos
  estructurados, visibilidad en buscadores con IA) está el plugin `claude-seo` (`/seo audit https://kitlegal.es`).
