---
target: critique web/src/pages/index.astro
total_score: 23
max_score: 36
na_heuristics: 7
p0_count: 1
p1_count: 3
target_identity: "file:/Users/jorge/Projects/kitlegal/web/src/pages/index.astro"
target_fingerprint: "sha256:19bacdaa45ba29da2672d19c5d2d41ce5194510aee17f98473bbace0c9f4304c"
target_path: /Users/jorge/Projects/kitlegal/web/src/pages/index.astro
timestamp: 2026-11-04T23-22-36Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05

Nielsen 23/36 (7 n/a): 1=3, 2=2, 3=3, 4=2, 5=2, 6=3, 8=2, 9=3, 10=3.

Especificidad: núcleo propio (díptico memoria/cita, CitaArticulo) dentro de un envoltorio SaaS genérico (hero con foto de ambiente, píldora con punto que late, métricas, pilares con icono en caja, tabla con columna resaltada, franja oscura de CTA). El hero promete citas y no enseña ninguna.

Detector: fuente exit 0 (4 avisos design-system-font-size: h1 2.75rem Portada:78; text-[10px] Portada:168, Portada:219, ComparativaMercado:169). Página construida exit 2: tiny-text/wide-tracking (falso positivo: estampa 11px +0.06em), undersized-ui-text 10px (real), kicker-above-heading 6, icon-tile-stack 4, nested-cards 3, pulsing-dot 1 (Portada:59, sin motion-reduce), line-length 5.

## Problemas prioritarios
- [P0] Afirmaciones legales y de mercado escritas a mano: SinergiaVentanilla:53 (RD 486/1997 en aulas), :103 (Ley 19/2013 un mes), :109 (art. 118 LCSP); ComparativaMercado:36 (precios sin fuente), :237 («ya sancionan», «elimina de raíz»). Quitar o pasar por citas.yaml; fechar fuentes; reformular.
- [P1] Instalar no es acción del hero: los dos CTA son anclas; en móvil instalar está a ~13k px. CTA secundario a /instalar/ y franja instalar/escribir tras el díptico.
- [P1] Comparativa: Antigravity «Nativa» (:88) contra «Sin probar» del Pie:30; min-w-[760px] ilegible en 390px, sin caption/scope/tabindex. Tarjetas por criterio en móvil.
- [P1] Ventanilla ~2.400 px en mitad del embudo de despachos, más píldora del hero y «Portal cívico» en el menú; registro activista. Una banda sobria cerca del pie.
- [P2] Jerga y tono SaaS: «a pelo», «sin tools», «Motor abierto de certeza jurídica», «lo mejor de ambos mundos», «huella SHA-256».

## Menores
8.132 escrito a mano (ComparativaMercado:126); caja roja NORMA DEROGADA en la tarjeta buena sin etiqueta; h1 fuera de escala; entradillas >75 caracteres; casos con enlace solo en 2 de 4; instalación repite el host 3 veces; ❌ emoji y «Software Libre»; píldoras rounded-full contra DESIGN.md.
