---
target: critique web/src/pages/index.astro
total_score: 18
max_score: 28
na_heuristics: 7,9,10
p0_count: 0
p1_count: 2
target_identity: "file:/Users/jorge/Projects/kitlegal/web/src/pages/index.astro"
target_fingerprint: "sha256:19bacdaa45ba29da2672d19c5d2d41ce5194510aee17f98473bbace0c9f4304c"
target_path: /Users/jorge/Projects/kitlegal/web/src/pages/index.astro
timestamp: 2026-11-05T14-35-24Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05, tras colores de estado, distill, typeset y menores

Método: dos subagentes aislados (A: revisión de diseño · B: detector). Nielsen 18/28 (n/a: 7 y 9; la 10 aparte): 1=3, 2=3, 3=3, 4=2, 5=3, 6=2, 8=2. Con la 10 (3), 21/32 (66 %). La anterior: 19/28 y 22/32; baja la 6.

Especificidad: las dos primeras pantallas son de kitlegal (díptico, CitaArticulo, «Cómo se lee la cita»). Desde la comparativa, lenguaje de landing SaaS: tabla con la columna propia resaltada, rejilla 2×2 de tarjetas, CTA centrado, todo sobre el mismo lavanda. La carcasa oscura de terminal de DESIGN.md no aparece.

Detector: fuente limpia (exit 0; los .astro solo pasan por expresiones regulares). En la página construida:
- 1280: 10 tiny-text y 10 wide-tracking (text-estampa de CitaArticulo, falsos positivos: citation-stamp de DESIGN.md), 3 line-length (alerta del art. 21, ~86 car.; blockquote de dos casos, 86-87), 1 unbalanced-heading (h3 del silencio).
- 390: 10 + 10 de la estampa, 1 unbalanced-heading (h3 del contrato menor). Sin line-length.
- Desaparecen nested-cards y el straggler de la tarjeta de precio.
- Altura: 11.923 px a 390 (antes 14.009) y 6.483 a 1440 (antes 7.962). Cabecera fija: 65 px a 390 (antes 126) y 81 a 1440. Sin desbordamiento horizontal.

## Problemas prioritarios
- [P1] Promesa que choca para quien usa ChatGPT: la meta descripción dice «No cambies de IA» (audiencias.ts) y la columna de la comparativa «En el asistente que ya usas», mientras la línea del hero y la fila de precio dicen Claude y «como Claude Pro». Arreglo: meta descripción sin «No cambies de IA»; columna «En tu asistente (hoy, Claude)». clarify.
- [P1] Móvil: la comparativa ocupa ~3.000 px y las llamadas de nota miden 5-6 × 12-14 px, con cuerpo de 9 px; las fuentes quedan plegadas lejos. Arreglo: zona táctil ≥24 px y cuerpo ≥11 px en Nota; tarjeta móvil más compacta. adapt, distill.
- [P2] Bloque de garantías sin jerarquía: sin encabezado aunque el menú lleva a #pilares; «Tres garantías que da el programa» va como hermano de «El secreto profesional», que no es una de las tres. Arreglo: h2 con las tres garantías en paralelo (la de sentencias destacada), el secreto aparte, el enlace al cierre. layout.
- [P2] Desvíos de DESIGN.md fuera del alcance de esta pasada: mono en los títulos del pie y el © (Pie.astro), en la insignia «CITA DEL BOE» y en los números de nota; el px-1 de los <mark> separa la puntuación («seis meses ,»); el identificador del aviso se parte «[BOE-A-|1992-26318]». typeset, polish.
- [P2] Hero a 1440: lg:items-end hunde los pasos y deja un hueco arriba a la derecha; con pt-28, las tarjetas del díptico quedan bajo el pliegue a 900 px de alto. layout.

## Personas
- Jordan: «Claude Code», «bloque a21» y «sha256» sin glosa fuera del <details>; que hace falta la app de escritorio y un plan de pago no está junto al botón.
- Riley: «No lee el BOE, así que no sabe si una norma sigue en vigor» y «En su chat» sobre el generalista van sin nota (los asistentes buscan en la web); «versión del 2 de octubre de 2015» parece texto desfasado.
- Casey: 14 pantallas; «Instalar» en la cabecera móvil lleva a una instalación de escritorio.
- Marta (despacho de dos con ChatGPT): «La IA que ya usas» y diez segundos después «sin probar»; la fila de precio le sugiere Claude Pro; el secreto profesional, en un bloque sin título.

## Menores
- Las tarjetas de casos cierran con dos estilos de enlace distintos; hay cuatro estilos de enlace en la página.
- hover:bg-papel/70 en las filas de una tabla estática.
- Columna izquierda del contacto vacía bajo el botón a 1440.
- «Software libre» en insignia y otra vez en «0 €, software libre (EUPL-1.2)».
- h3 con viuda en dos tarjetas de casos (text-balance o text-pretty).
