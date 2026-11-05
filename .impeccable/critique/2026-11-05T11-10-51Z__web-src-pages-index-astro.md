---
target: critique web/src/pages/index.astro
total_score: 22
max_score: 32
na_heuristics: 7,9
p0_count: 0
p1_count: 2
target_identity: "file:/Users/jorge/Projects/kitlegal/web/src/pages/index.astro"
target_fingerprint: "sha256:19bacdaa45ba29da2672d19c5d2d41ce5194510aee17f98473bbace0c9f4304c"
target_path: /Users/jorge/Projects/kitlegal/web/src/pages/index.astro
timestamp: 2026-11-05T11-10-51Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05, tras la rama web/portada-critica-impeccable

Método: dos subagentes aislados (A: revisión de diseño · B: detector). Nielsen 22/32 (n/a: 7 y 9): 1=3, 2=3, 3=3, 4=2, 5=3, 6=3, 8=2, 10=3. La anterior fue 23/36 (64 %); esta, 69 %.

Especificidad: el duelo de memoria frente a cita, en el hero, la comparativa con notas fechadas y las tarjetas de cita son propias; la tira de 5 pasos, los 3 pilares, las tarjetas de instalación 1 y 2 y la banda oscura siguen siendo de plantilla.

Detector: la fuente sale con exit 0 y sin hallazgos. La página construida sale con exit 2 y 30 avisos:
- 13 tiny-text y 13 wide-tracking: casi todos la estampa mono de 11px de DESIGN.md (falsos positivos), salvo la etiqueta sans «Lo detecta kitlegal antes de citar» (Portada:159), que usa text-estampa en Inter.
- 4 line-length: tres por max-w-prose con Inter (86-89 caracteres) y el aviso de 12px de la comparativa (ComparativaMercado:266).
Desaparecen el texto de 10px, los antetítulos, las cajas de icono, las tarjetas anidadas y el punto que late.

## Problemas prioritarios
- [P1] La jurisprudencia inventada, el miedo nº 1, no aparece en el hero ni en las garantías: solo en un aviso de 12px y en «Lo que hoy no hace». Cambiar el pilar «Territorio real» por uno de sentencias, con el texto aprobado de audiencias.ts («No busca sentencias: solo cita la que traes…»), y subir el aviso a text-sm.
- [P1] Redundancia y longitud (~8.500 px): la tira de 5 pasos, los pilares y la sección de instalar repiten el método que ya enseña el duelo. Quitar o fundir la tira; «Ver cómo funciona» no debe saltar el duelo.
- [P2] «Bajo presupuesto» puede leerse como «barato»: mejor «Tarifa no publicada» o «Precio a consultar».
- [P2] El cierre se diluye: Ventanilla Legal va tras la llamada final, el texto final es cauto y «Leer la documentación» lleva a GitHub.
- [P2] La alerta del duelo parafrasea a mano el plazo de tres meses (audiencias.ts:251): recortarla y dejar que la cita lo pruebe.

## Menores
- La cabecera móvil centra un párrafo de 6 líneas.
- La flecha del enlace de garantías se queda sola.
- Nada indica que la navegación móvil se desplaza.
- La tarjeta izquierda del duelo deja un hueco de ~230 px.
- La mitad derecha del hero queda vacía a 1440.
- «Sin descargar nada» va junto a un icono de descarga.
- «En ChatGPT no funciona» choca con «Sin probar» del pie.
- Los contenedores alternan max-w-6xl y max-w-7xl.
- El verde de vigente se usa en la lista de peticiones.
- El «2» en ámbar.
- Una cabecera de la tabla usa text-[11px] en vez de text-estampa.
