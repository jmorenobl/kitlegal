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
timestamp: 2026-11-05T18-43-34Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05, tras bolder, colorize y polish (rama web/portada-fuerza-visual)

Método: dos subagentes aislados (A: revisión de diseño · B: detector). Nielsen 20/28 (n/a: 7 y 9; la 10 aparte): 1=3, 2=3, 3=3, 4=3, 5=3, 6=3, 8=2. Con la 10 (3), 23/32. La anterior: 21/28 y 24/32; baja la 8 por la densidad de la tarjeta verde y los huecos (roja ~250 px más corta, contacto).

Especificidad: núcleo propio (díptico, columna de kitlegal en noche, citas verificadas); cáscara genérica (hero H1 + foto a la derecha, foto de escritorio intercambiable).

Detector: fuentes .astro limpias; dist/index.html 79 avisos, todos falsos positivos (53 cramped-padding de separadores, 18 tiny-text del citation-stamp, 4 repeated-container-text, 3 design-system-font por el sufijo «Variable», 1 design-system-color). Inyectado a 1280: 18 del sello, 3 line-length (alerta del art. 21, citas LCSP y LPACAP 24.1, 86-87 car.), 1 unbalanced-heading (h3 del silencio). A 390: 2 del sello. Sin desbordamiento de 320 a 1440. Alto: 9.405 px a 390, 6.601 a 1440. Cabeceras del díptico en y=858 a 1440×900, 858 a 1280×800, 820 a 1024×768. A 390 no se pide la foto. Contrastes AA (sobre-noche 7,2:1; sello-tinta 4,8:1). Foco: contorno de 2 px azul; durante ~150 ms la transición de color lo deja casi invisible en los botones oscuros.

## Problemas prioritarios
- [P1] La prueba queda bajo el pliegue: a 1440×900 la foto ocupa la mitad derecha y el díptico empieza en y≈825 (no se ve a 1280×800 ni 1024×768). Decisión de la pasada: la foto vuelve al hero; alternativa, tarjeta verde compacta en su lugar. layout.
- [P1] Tarjeta verde sobrecargada: tres rótulos («Cita del BOE», «Texto consolidado del BOE», «Con fuente verificable», este en azul de enlace sin serlo), dos bloques de metadatos; el chip «NORMA DEROGADA» se parte a 390. distill.
- [P2] La fila de pasos entre la consulta y las tarjetas corta la lectura consulta → respuesta y su título con regla noche parece una sección. layout.
- [P2] Huecos: ~250 px bajo la tarjeta roja a 1440 (items-start), ~220 px bajo el botón de contacto; la franja «sin terminal» en 14 px tenue. layout.
- [P2] La comparativa móvil no compara: los ajenos plegados en cuatro acordeones iguales. adapt.

## Personas
- Jordan: «bloque a21», «sha256», «programa determinista» sin glosar.
- Casey: CTA principal «Instalar» en móvil para algo de escritorio; chip partido; 9.405 px.
- Riley: a 1024 el paso 5 en tres líneas; notas de 11 px difíciles de pulsar.
- Marta (ChatGPT): «La IA que ya usas» frente a «sin probar» en 14 px; Claude Pro en la fila Precio a ~1.700 px.

## Menores
- Sombras teñidas en las tarjetas del díptico y shadow-lg en la foto (DESIGN pide contención).
- Correo cinco veces en la página.
- El icono de «Lo que no hará» en tenue (intencionado).

## Juicio aparte frente a df6163bc
Sí, por poco: más contenida y adulta (sin etiquetas en mono sobre cada sección, garantías con jerarquía, ritmo de fondos limpio, mitad de altura). df hacía mejor la foto (más grande y nítida; en la rama estaba escalada desde 480 px, corregido después en 4d3c276a), los pasos con icono y flecha y la banda final con botón ámbar macizo.
