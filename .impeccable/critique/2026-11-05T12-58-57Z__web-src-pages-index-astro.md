---
target: critique web/src/pages/index.astro
total_score: 19
max_score: 28
na_heuristics: 7,9,10
p0_count: 0
p1_count: 2
target_identity: "file:/Users/jorge/Projects/kitlegal/web/src/pages/index.astro"
target_fingerprint: "sha256:19bacdaa45ba29da2672d19c5d2d41ce5194510aee17f98473bbace0c9f4304c"
target_path: /Users/jorge/Projects/kitlegal/web/src/pages/index.astro
timestamp: 2026-11-05T12-58-57Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05, tras clarify, polish de datos y polish de menores

Método: dos subagentes aislados (A: revisión de diseño · B: detector). Nielsen 19/28 (n/a: 7, 9 y 10): 1=3, 2=3, 3=3, 4=2, 5=3, 6=3, 8=2. Con la 10 puntuada (3), 22/32 (69 %). Igual que la crítica anterior con el mismo criterio.

Especificidad: el contenido es propio del producto (díptico con cita real comprobada al construir, comparativa con notas fechadas, «Todavía no lo hace / Lo que no hará»). El marco visual sigue siendo de landing SaaS: tarjetas blancas con sombra, pasos en cuadrados oscuros, CTA centrado. El sello ámbar y el aire de boletín de DESIGN.md apenas se ven.

Detector: la fuente sale con exit 0 y sin hallazgos. En la página construida:
- 1280: 15 tiny-text y 15 wide-tracking (text-estampa, falsos positivos), 3 line-length (la leyenda «Cómo se lee la cita», ~92 car., y el blockquote de las tarjetas de casos, 86-87), 1 nested-cards (la caja «Detectado antes de citar» dentro de la tarjeta verde), 1 unbalanced-heading.
- 390: 12 + 12 de la estampa, 1 nested-cards, 1 unbalanced-heading, 1 straggler («aparte.» en la tarjeta móvil de precio).

## Problemas prioritarios
- [P1] El H1 promete «La IA que ya usas» y solo funciona en Claude; quien usa ChatGPT lo descubre en el subtítulo y en tres repeticiones de «sin probar». Arreglo: decirlo junto al CTA en una línea y quitar las repeticiones, o nombrar Claude en el titular.
- [P1] La caja «Detectado antes de citar» de la tarjeta verde no se entiende en la primera lectura: el lector no ve por qué se habla de la Ley 30/1992. Arreglo: una línea que la introduzca («Si tu asistente intenta citar la Ley 30/1992, kitlegal lo para») o llevarla al puente entre las dos tarjetas.
- [P2] Los colores de estado (vigente, sello) se usan como color de materia en «Consultas habituales», contra el código que enseña el díptico. Arreglo: etiquetas neutras y color solo para estados.
- [P2] Redundancia y longitud: la garantía de sentencias tres veces, «sin probar» tres veces, instalar cinco; ~14.000 px a 390 y cabecera móvil de ~125 px con la navegación cortada.
- [P3] Mono fuera de su contrato: la consulta del díptico y pies como «Respuesta de memoria» van en JetBrains Mono, que DESIGN.md reserva para datos verificables.

## Personas
- Jordan: «Instalar, sin terminal» junto a «Claude Code»; no hay una captura real de Claude respondiendo.
- Riley: «0 €» frente a «Desde 20 $» compara cosas distintas; el art. 21.2 dice «no podrá exceder de seis meses», así que la respuesta de memoria no es del todo errónea y el matiz no se explica; las tarjetas de casos tienen hover sin ser enlaces.
- Casey: hero y díptico ocupan ~3,5 pantallas; el contacto pasa de los 10.000 px.
- Marta (despacho de dos con ChatGPT): para usarlo tendría que cambiar de asistente; la duda de dónde van los datos del cliente queda en «sigue las condiciones de ese asistente», sin enlace.

## Menores
- La leyenda «Cómo se lee la cita» no lleva max-w-lectura (~92 car.).
- La columna de pasos del hero queda alineada abajo y deja hueco arriba a la derecha.
- «Cuatro preguntas» escrito a mano en la entradilla de casos.
- Dos formulaciones del CTA de instalar.
- El aviso de fotografías del pie sobra en la portada (se queda por decisión: lo usan /instalar/ y /consultas/).
