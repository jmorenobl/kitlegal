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
timestamp: 2026-11-05T11-55-25Z
slug: web-src-pages-index-astro
---
# Crítica de la portada (web/src/pages/index.astro), 2026-11-05, tras distill, clarify, layout y polish

Método: dos subagentes aislados (A: revisión de diseño · B: detector). Nielsen 19/28 (n/a: 7, 9 y 10): 1=3, 2=3, 3=3, 4=2, 5=3, 6=3, 8=2. A marcó la 10 como n/a; la anterior la puntuó 3. Con ese mismo criterio queda 22/32 (69 %), igual que antes.

Especificidad: lo de arriba es propio (el díptico con una cita real y el aviso de derogación sacado del sobre, y la comparativa con fuentes fechadas). De «Tres garantías» hacia abajo sigue siendo plantilla: tres tarjetas con icono, las tarjetas «1 / 2» y una banda oscura con dos botones.

Detector: la fuente sale con exit 0 y sin hallazgos. En el escaneo por URL (1280 y 390) quedan:
- 14 tiny-text y 14 wide-tracking: la estampa mono de 11 px de DESIGN.md, son falsos positivos. El caso límite es el subtítulo de columna de ComparativaMercado:194.
- 4 line-length: max-w-prose con Inter a 86-89 caracteres.
- 1 unbalanced-heading: un h3 de consulta sin text-balance.
- 2 straggler en las tarjetas móviles de la comparativa.

En modo fichero salen además 40 cramped-padding y 3 design-system-font, falsos positivos por la CSS no cargada y por el nombre de la familia variable, y #94a3b8 (sobre-noche), que no está en DESIGN.md.

## Problemas prioritarios
- [P1] La tarjeta de kitlegal no se lee como el alivio: abre con un bloque rojo (la derogación detectada) y la respuesta buena, «tres meses», es un subrayado fino al final del artículo, mientras el error de la izquierda va marcado con fondo. Arreglo: dar a la variante destacada de CitaArticulo un fondo vigente simétrico y bajar el rojo del bloque «Lo detecta kitlegal» a una insignia.
- [P1] La fila «Territorio» de la comparativa dice de las editoriales «No hemos encontrado en sus webs cómo lo resuelven» sin nota fechada, y es la fila que menos importa a un despacho. Arreglo: quitarla o cambiarla por un criterio con fuente.
- [P2] Repetición y longitud: lo de las sentencias aparece 4 veces y la vigencia 3; la página mide ~8.200 px en escritorio y ~15.000 en móvil. «Tres garantías» casi no añade nada tras la comparativa, y el párrafo del CENDOJ podría ir a un <details>.
- [P2] Jerarquía de CTA: «Ver cómo funciona», el principal, desplaza unos 100 px; instalar es secundario y aparece con cinco estilos; el botón final es ámbar, cuando DESIGN.md pone el primario en noche. La tira de instalar y correo bajo el díptico repite la cabecera y el contacto.
- [P2] Promesas en «Lo que hoy no hace»: «Lo hará: está en el plan» y «Lo que sí puede llegar…» chocan con «preciso y sin promesas para los despachos».

## Personas
- Jordan: jerga sin explicar (plugin, caché, sha256, bloque a21, DIR3). No queda claro que hace falta un plan de pago del asistente: «Gratis» en el hero y «Desde 20 $» solo en la tabla.
- Casey: página de ~15.000 px y comparativa de 15 bloques. El desvanecido de la navegación móvil existe, pero es sutil.
- Riley:
  - la celda de Territorio sin fuente;
  - el pie habla de fotografías generadas con IA en una portada que ya no tiene ninguna;
  - dos casos van al BOE y dos a una página propia, con el mismo aspecto.
- Marta (abogada de un despacho de dos que usa ChatGPT): «La IA que ya usas» le promete compatibilidad, y que ChatGPT está sin probar lo descubre en la quinta sección. Que no hay jurisprudencia lo ve en el penúltimo bloque.

## Menores
- Campos muertos en `despachos`: insignia, métricas, imagen y antetítulo.
- La caja «Consulta» está repetida en las dos tarjetas del díptico.
- La cabecera de la tabla descuadra por el subtítulo de dos líneas.
- La mitad derecha del hero sigue vacía a 1440.
- «Lo que hoy no hace» usa el azul del agente.
- Ventanilla Legal, ya en el pie, sigue siendo lo primero tras la llamada final.
