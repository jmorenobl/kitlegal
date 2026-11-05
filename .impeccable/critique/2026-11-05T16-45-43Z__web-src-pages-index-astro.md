---
target: critique web/src/pages/index.astro
total_score: 20
max_score: 28
na_heuristics: 7,9,10
p0_count: 0
p1_count: 2
target_identity: "file:/Users/jorge/Projects/kitlegal/web/src/pages/index.astro"
target_fingerprint: "sha256:19bacdaa45ba29da2672d19c5d2d41ce5194510aee17f98473bbace0c9f4304c"
target_path: /Users/jorge/Projects/kitlegal/web/src/pages/index.astro
timestamp: 2026-11-05T16-45-43Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05, tras clarify, layout, distill y typeset

Método: dos subagentes aislados (A: revisión de diseño · B: detector). Nielsen 20/28 (n/a: 7 y 9; la 10 aparte): 1=3, 2=3, 3=3, 4=3, 5=3, 6=3, 8=2. Con la 10 (3), 23/32 (72 %). La anterior: 19/28 y 22/32; sube la 4 (la insignia editorial deja la forma de sello, el pie en Inter).

Especificidad: las pruebas son propias (díptico con la Ley 30/1992 detectada, CitaArticulo con bloque, fecha y huella, notas fechadas, «Lo que no hará: CENDOJ»); el esqueleto sigue siendo de landing (rail de pasos, franja a dos columnas, tabla, rejilla 2×2, contacto partido, CTA centrado, pie a cuatro columnas). El cierre no termina con una prueba.

Detector: fuentes .astro limpios (exit 0, solo reglas de regex). Página construida:
- 1280: 10 tiny-text + 10 wide-tracking (estampa de CitaArticulo, falso positivo: citation-stamp de DESIGN.md), 3 line-length (alerta del art. 21; blockquote de dos casos, 86-87 car.), 1 unbalanced-heading (h3 del silencio, 20 % en la última línea pese a text-pretty).
- 390: 2 tiny-text + 2 wide-tracking (estampa). Desaparecen el straggler «detrás.» y la viuda del contrato menor.
- Medidas: 10.737 px a 390 (antes 11.978) y 6.734 a 1440 (antes 6.806); casos 1.393 px a 390 (antes 2.640); cabeceras del díptico en y=844/848 a 1440×900 (caben, unos 30 px sobre el pliegue; a 1280×800 no). Sin desbordamiento. Notas a 390: zona táctil 23×34 px.

## Problemas prioritarios
- [P1] El secreto profesional, al pie de las garantías en 14 px con icono atenuado, tras el enlace «Cómo se comprueba»: es el segundo freno de un despacho y se lee como letra pequeña. Arreglo: darle el peso de «Ninguna sentencia sin su documento» y llevar el enlace al final. layout.
- [P1] En móvil la prueba llega tarde y la página no acaba: díptico en y=1132 tras los cinco pasos; comparativa ~2.000 px; pie 1.458 px. Arreglo: pasos tras el díptico o plegados en móvil; comparativa móvil con kitlegal abierta y el resto plegado; pie a dos columnas. adapt.
- [P2] Redundancia en la tarjeta roja (insignia, etiqueta de ejemplo, alerta y pie dicen lo mismo, seis elementos en rojo); la frase de vigencia repetida literal en la comparativa y en la garantía; «¿Qué le falta para tu despacho?» en la franja y como H2 del contacto. distill.
- [P2] El aviso «No incluyas datos de clientes» queda lejos del botón de correo (y ~700 px en móvil); los otros mailto no lo llevan. harden.
- [P2] Incoherencias: «Instalar» en la cabecera; dos estilos de enlace en las tarjetas de casos; iconos de una misma función en ámbar y en azul; «CITA DEL BOE» en mono y mayúsculas siendo etiqueta de la página; hover en filas de tabla no pulsables. polish.

## Personas
- Jordan: «Claude Code», «Codex», «Antigravity» en la primera pantalla; «sha256», «bloque a21» sin saber que hay explicación plegada; no sabe qué pide instalar.
- Riley: «Por qué» desaparece al abrirse; sin vuelta desde la fuente a la celda; comprobar #fuente-N en Safari y Firefox.
- Casey: 10.737 px; enlaces en línea de 17 px de alto; comparativa y pie largos.
- Marta (ChatGPT): la promesa del H1 se estrecha por partes («sin probar», «hoy, Claude», Claude Pro); «Lo que no hará: CENDOJ» puede leerse como «no sirve para jurisprudencia».

## Menores
- El px-1 de los <mark> del díptico separa la puntuación («seis meses ,»).
- A 1440 la tarjeta roja acaba ~190 px antes que la verde; la columna izquierda del contacto queda vacía ~450 px.
- Casos y contacto con el mismo fondo.
- «Tres garantías» con un cuarto elemento al pie.
- «Software libre» en la cabecera de la tabla repite la fila de precio; v0.9.0 en la cabecera.
