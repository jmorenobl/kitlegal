---
target: critique web/src/pages/index.astro
total_score: 21
max_score: 28
na_heuristics: 7,9,10
p0_count: 0
p1_count: 1
target_identity: "file:/Users/jorge/Projects/kitlegal/web/src/pages/index.astro"
target_fingerprint: "sha256:19bacdaa45ba29da2672d19c5d2d41ce5194510aee17f98473bbace0c9f4304c"
target_path: /Users/jorge/Projects/kitlegal/web/src/pages/index.astro
timestamp: 2026-11-05T17-33-08Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05, tras layout, harden, adapt, distill y polish

Método: dos subagentes aislados (A: revisión de diseño · B: detector). Nielsen 21/28 (n/a: 7 y 9; la 10 aparte): 1=3, 2=3, 3=3, 4=3, 5=3, 6=3, 8=3. Con la 10 (3), 24/32 (75 %). La anterior: 20/28 y 23/32; sube la 8 (una vez cada cosa en el díptico, el contacto y el CTA; garantías a la par).

Especificidad: el díptico y CitaArticulo son propios (Ley 30/1992 detectada, «seis meses» frente a «tres meses», huella y versión del BOE); la comparativa también (notas fechadas, «Tarifa no publicada»). De plantilla: CTA centrado, rejilla 2×2 de casos, tarjeta de contacto, pasos 1-5. La personalidad «archivo y terminal» solo vive en el díptico y en las citas.

Detector:
- Fuentes .astro: exit 0, pero solo con expresiones regulares (sin reglas de elemento).
- HTML construido (estático): 75 avisos, falsos positivos de análisis estático (48 cramped-padding sobre bordes que separan, 18 tiny-text de la estampa, 4 repeated-container-text por las dos variantes de la cita de los casos, 5 design-system-* del enlace de salto y del logotipo).
- Página renderizada a 1280: 10 tiny-text + 10 wide-tracking (estampa de CitaArticulo, falso positivo: citation-stamp de DESIGN.md), 3 line-length (alerta del art. 21 y dos blockquote de casos, 86-87 car.), 1 unbalanced-heading (h3 del silencio). A 390: 2 + 2 de la estampa. Sin cambios frente a la anterior.
- Medidas: 9.286 px a 390 (antes 10.737; objetivo de 8.000 aplazado por decisión), 6.537 a 1440 (antes 6.734); cabeceras del díptico en y=844/845 a 1440×900 y a 1280×800. Sin desbordamiento a 390.
- Comprobaciones: los 4 mailto llevan el aviso en body=; el aviso queda a 14 px bajo el botón de correo; a 390 los pasos van tras el díptico y a 1440 junto al titular; las filas de la tabla no cambian con hover.

## Problemas prioritarios
- [P1] En móvil, «Instalar kitlegal» ocupa la primera pantalla y la cabecera, pero se instala en el ordenador (app de escritorio de Claude) y eso solo se dice en la franja, a unos 2.000 px. Arreglo: una línea bajo el botón en móvil («Se instala en el ordenador…») y un mailto «Envíatelo para instalarlo luego». adapt + clarify.
- [P2] La comparativa móvil deja a la vista solo la celda de kitlegal: sin «Tarifa no publicada» ni «Desde 20 $», deja de comparar. Arreglo: una línea con los dos `destacado` ajenos bajo la de kitlegal y plegar solo subtextos y notas. layout.
- [P2] Las garantías son un h3 dentro de #comparativa (4.136 px en móvil), aunque la cabecera las enlaza como sección. Arreglo: sección propia con h2 o adelantar una línea de las dos destacadas a la franja. layout.
- [P2] Lo que da credibilidad es pequeño: alerta del error en 12 px rojo sobre rojo, subtextos de la comparativa en 12 px; notas «Fuente N» de 7×14 px y summary y enlaces de bloque de 20 px a 390. typeset + harden.
- [P3] La tarjeta verde rotula su naturaleza tres veces («Cita del BOE», «Texto consolidado del BOE», «Con fuente verificable»); la roja acaba ~250 px antes que la verde a 1440; casos y contacto con el mismo fondo; columna izquierda del contacto vacía. distill + polish.

## Personas
- Jordan: «skill», «bloque a21», «sha256» sin saber que «Cómo se lee la cita» lo explica (enlace gris de 12 px).
- Riley: la estampa «NORMA DEROGADA» se parte en dos líneas a 390; llamadas a fuentes plegadas por comprobar en Safari; el contacto es solo mailto.
- Casey: el botón de instalar lleva a un paso a paso que no puede hacer en el móvil; la prueba roja y verde ocupa ~1.470 px; cuatro botones iguales en la comparativa.
- Marta (ChatGPT): «sin probar» la deja fuera en la primera pantalla sin un paso siguiente; el precio de Claude está plegado en móvil; «Redactar escritos» al fondo.

## Menores
- El aviso de fotografías del pie no tiene objeto en la portada (no hay fotografías); es decisión cerrada, va en todo el sitio.
- «determinista» en el texto de las garantías es jerga para Marta.
- «pendiente» y «bloqueado» con el mismo ámbar en el contacto.
- El CTA final («Dota a tu despacho…») no retoma la prueba del díptico.
