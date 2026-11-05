---
target: critique web/src/pages/index.astro
total_score: 19
max_score: 28
na_heuristics: 7,9,10
p0_count: 0
p1_count: 3
target_identity: "file:/Users/jorge/Projects/kitlegal/web/src/pages/index.astro"
target_fingerprint: "sha256:19bacdaa45ba29da2672d19c5d2d41ce5194510aee17f98473bbace0c9f4304c"
target_path: /Users/jorge/Projects/kitlegal/web/src/pages/index.astro
timestamp: 2026-11-05T14-43-07Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05, tras las garantías con jerarquía y las notas tocables

Método: dos subagentes aislados (A: revisión de diseño · B: detector). Nielsen 19/28 (n/a: 7 y 9; la 10 aparte): 1=3, 2=3, 3=3, 4=2, 5=3, 6=3, 8=2. Con la 10 (3), 22/32 (69 %). La anterior, en la misma rama: 18/28; sube la 6 (las notas abren las fuentes plegadas y son tocables).

Especificidad: el díptico y la comparativa con notas fechadas son propios del producto. El resto sigue el esqueleto de landing (rail de pasos, rejilla 2×2 de casos, contacto a dos columnas, CTA centrado); la personalidad de archivo de DESIGN.md solo vive dentro de las citas.

Detector: fuente limpia (exit 0). Página construida:
- 1280: 10 tiny-text + 10 wide-tracking (estampa de CitaArticulo, falsos positivos: citation-stamp de DESIGN.md), 3 line-length (alerta del art. 21; blockquote de dos casos, 86-87 car.), 1 unbalanced-heading (h3 del silencio).
- 390: los 20 de la estampa, 1 straggler («detrás.» en la tarjeta móvil «Dónde se usa»), 1 unbalanced-heading (h3 del contrato menor).
- Altura 11.938 px a 390 y 6.806 a 1440; cabecera 65 y 81 px; sin desbordamiento. Notas a 390: 11 px y zona táctil de 24×34 px (WCAG 2.5.8).

## Problemas prioritarios
- [P1] La respuesta «de memoria» no dice a la vista que es un ejemplo: el aviso es sr-only y la leyenda «La misma consulta, dos respuestas» sugiere una salida real. Arreglo: etiqueta visible «Respuesta de ejemplo». clarify.
- [P1] A 1440×900 el díptico queda bajo el pliegue (empieza en y≈880); el rail de pasos con lg:items-end deja un hueco arriba a la derecha. Arreglo: menos pt y rail arriba, o los pasos bajo el díptico. layout.
- [P1] Longitud y repetición: la idea de la vigencia sale cinco veces; los casos ocupan 2.640 px en móvil con cuatro citas completas. distill.
- [P2] Mono y verde fuera de su papel: la insignia editorial «Norma derogada · plazo erróneo» tiene la forma del aviso del BOE; títulos y © del pie en mono; el verde del titular de la tarjeta y del resaltado de la cita (CitaArticulo, decisión cerrada). typeset.
- [P2] Garantías: anidadas en la comparativa como h3 mientras la cabecera las trata como sección; «Tres garantías» con un cuarto elemento (el secreto) al pie; el enlace deja media caja vacía a 1440; gavel también en «Contratación pública». layout.

## Personas
- Jordan: «Metadatos del BOE · sha256» antes de la cita en la tarjeta verde; «Claude Code», «Codex», «Antigravity» y «Agent Skills» sin explicar.
- Riley: respuesta mala sin procedencia; «desde Claude en el móvil o en la web» sin enlace; la columna kitlegal sin nota.
- Casey: 14 pantallas sin índice; «Instalar» lleva a una instalación de escritorio; el aviso de derogación se parte con el borde roto.
- Marta (ChatGPT): «la IA que ya usas» frente a «sin probar» y «como Claude Pro»; nada dice qué puede hacer hoy.

## Menores
- El px-1 de los <mark> separa la puntuación («seis meses ,»).
- Casos y contacto seguidos con el mismo fondo; py-24/32 en todas las secciones.
- La jerarquía de títulos empieza con h2 en las tarjetas del díptico.
- Dos pesos de enlace al pie de las tarjetas de casos.
- Chip v0.9.0 en la cabecera.
