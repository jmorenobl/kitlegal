# 0034 · La web habla a quien tiene el asunto: portada de la ciudadanía, despachos, consultas con su cita y un buzón

- **Estado**: aceptada; **enmendada el 2026-11-21**: la página `/plan/` lleva una encuesta anónima que se envía a un
  formulario de Google («Enmienda: la encuesta del plan»). Sigue en pie el buzón, y ninguna otra página envía nada.
- **Fecha**: 2026-10-01
- **Hito**: transversal (tras v0.3.2, antes de H20). Enmienda el ADR 0024 en «Una URL por audiencia» y extiende su
  regla de las citas a lo que la web dice de una norma con otras palabras. No cambia el producto.

## Contexto y problema

La web (ADR 0024) nació con dos portadas, `/` para despachos y `/ciudadania/`, y cuatro páginas en total. El
2026-09-28, con las cuatro indexadas, Search Console no daba ni una impresión en 28 días: Google las tenía y no las
enseñaba para ninguna búsqueda. Para saber cómo busca quien podría usar kitlegal se miró lo que había fuera: el
autocompletado de Google en España con unas 200 búsquedas de arranque, los estudios publicados entre 2024 y 2026 y
lo que ya existe para acercar el BOE a un asistente. Salió esto:

- **La ciudadanía busca lo que le pasa, no la ley**: «mi casero no me devuelve la fianza», «el ayuntamiento no
  contesta a un escrito», «cuánto tiempo tiene Hacienda para reclamar una deuda», «plazo recurso de alzada»,
  «modelo recurso de reposición». Lo que más se repite es el plazo, y después el siguiente paso.
- **Los despachos son pequeños y lo que les frena es la formación y la fiabilidad** (Libro Blanco del CGAE, enero de
  2026: el 78,5 % de la muestra trabaja en despachos de una a tres personas). Buscan «ia para abogados gratis»,
  «… españa», «claude para abogados». Los tribunales han multado por jurisprudencia inventada con IA, no por citar
  normas derogadas; las guías colegiales piden contrastar con fuentes oficiales y vigilar la normativa desfasada.
- **Ninguna de las dos audiencias trabaja en una terminal**, y la web vendía «sin barreras técnicas».
- **La web no respondía a nada que la gente escriba**: su texto contaba lo que hace el producto (texto consolidado,
  vigencia, huella).

Jorge decidió, en sesión interactiva: la web habla **al usuario final** —la ciudadanía y los despachos—, no al
perfil técnico, que «o se lo monta él o actúa como ciudadano», ni a las administraciones públicas, que no son el
usuario tipo y no podrían instalarla; la portada del sitio es la de la ciudadanía; y la de despachos dice que la
herramienta se está haciendo para ellos y les invita a escribir.

Al aplicar los cambios, `make web-citas` trajo además una redacción nueva del artículo 36 de la Ley de Arrendamientos
Urbanos, del 30 de septiembre de 2026 (Real Decreto-ley 26/2026), con un apartado 7 que no existía cuatro días antes y
el aviso `consolidacion-no-finalizada`. La web publicada seguía citando la versión de 2019 y no mostraba ningún aviso
con sus citas.

## Opciones consideradas

**Qué audiencia ocupa `/`.**

1. **Despachos**, como hasta ahora. Rechazada: quien llega buscando un problema suyo y encuentra «para despachos» se
   va, y un abogado entiende una portada escrita para cualquiera.
2. **Una portada neutra que reparte** a dos páginas. Rechazada: una página más sin nada que una búsqueda pueda
   encontrar.
3. **La ciudadanía**, y despachos en `/despachos/`.

**Cómo responder a lo que la gente busca.**

1. **Solo retocar los textos de las portadas.** Rechazada: una portada no puede responder a «plazo recurso de
   alzada» y a «mi casero no me devuelve la fianza» a la vez.
2. **Páginas generadas por plantilla para cada artículo de `data/normas.yaml`.** Rechazada: serían el texto del BOE
   sin la pregunta, que es lo que el BOE ya publica mejor.
3. **Una página por consulta, escrita**: la pregunta como se busca, la respuesta corta y cada punto con la cita que
   lo sostiene.

**Cómo saber qué necesita un despacho.**

1. **Medir el uso.** Descartada por el ADR 0027: kitlegal no envía datos de uso a ningún servidor, tampoco con
   consentimiento.
2. **Un formulario.** Rechazada: la web no tiene servidor, su política de seguridad de contenido prohíbe enviar
   formularios y uno de un tercero metería a un tercero en la página.
3. **Un buzón**, `info@kitlegal.es`, con una petición concreta.

## Decisión

- **Dos audiencias, las dos de usuario final.** `/` es la portada de la ciudadanía y `/despachos/`, la de los
  despachos; `/ciudadania/` redirige a `/`. La web no tiene nada para el perfil técnico ni para las administraciones
  públicas, y su presencia en directorios de skills o de servidores MCP no es una vía de la web.
- **Una página por consulta, en `/consultas/<slug>/`**, declarada en `web/src/data/consultas.ts`: la pregunta tal
  como se escribe en un buscador, la respuesta corta y los puntos que conviene saber. **Cada afirmación va con la
  cita que la sostiene, mostrada al lado**: lo que la web dice de una norma con otras palabras solo dice lo que
  dice el fragmento que enseña. Lo que la norma citada no dice va en «Lo que esta página no te dice», que habla de
  la página y no afirma nada de otra norma. Sin fragmento que lo sostenga, un punto no entra: se pide el artículo
  con `make web-citas` o no se escribe.
- **Cada cita declara la versión del bloque con la que se escribió** (`version` en `citas.yaml`, la `fecha_version`
  del sobre). Si `make web-citas` trae otra, la construcción falla hasta que alguien relee lo que la web dice de ese
  bloque: que los fragmentos sigan estando literales no garantiza que la explicación siga siendo completa.
- **La cita muestra los avisos de vigencia de su sobre**, como hace el agente antes de citar.
- **La web no promete lo que el producto no hace.** Dice qué asistentes hacen falta (Claude Code, Codex,
  Antigravity) y que con los chats del navegador y del móvil todavía no funciona; la portada de despachos dice que
  hoy no consulta ni comprueba jurisprudencia y que no redacta escritos; ninguna página ofrece recurrir una multa de
  tráfico ni lo que el ADR 0027 deja fuera de la base, y las consultas sobre la Ley 39/2015 citan su disposición
  adicional primera.
- **Un buzón, `info@kitlegal.es`**, en la portada de despachos —con lo que más ayuda saber, lo que hoy no hace y la
  petición de no enviar datos de clientes ni de asuntos—, en `/instalar/` y en el pie. No hay formulario ni nada de
  terceros.
- **`/instalar/` es una guía paso a paso** para quien no ha abierto nunca una terminal, con las skills en global
  (`kitlegal skills install -g`), y deja los detalles —hosts, otras formas de instalar, atestación— al final.

## Consecuencias

- Una reforma de un artículo citado detiene la construcción de la web en el siguiente `make web-citas`, por el
  fragmento que ya no es literal o por la versión, y obliga a releer la consulta. Entre dos `make web-citas`, la web
  publicada puede citar una versión que ya no es la vigente: comprobarlo cada noche, como `scripts/verify-sources.sh`
  comprueba el producto, queda anotado en `docs/USO.md`.
- Añadir una consulta es añadir una entrada a `consultas.ts` y, si hace falta, sus citas a `citas.yaml`; la página,
  su imagen para compartir, su entrada en `/llms.txt` y en el sitemap salen solas.
- `/ciudadania/`, que Google ya tenía indexada, pasa a ser una redirección con su canónica en `/`.
- `docs/SEO.md` deja de proponer los directorios para desarrolladores.
- Lo que cuenten los despachos por correo entra en `docs/USO.md`, la bitácora con la que se reprioriza el roadmap
  (ADR 0013). Que ninguna de las dos audiencias trabaje en una terminal es una entrada de esa bitácora, no una
  decisión de este ADR: el servidor MCP y el plugin siguen en el backlog de distribución.

## Enmienda: la encuesta del plan (2026-11-21)

**Contexto.** La persona que mantiene el proyecto quiere que juristas que no usan Claude revisen el plan —lo hecho, lo
que viene y en qué orden— y digan si es correcto y qué necesitan, y poder contar las respuestas: cuántos piden una
cosa u otra, desde qué tipo de despacho y en qué áreas. El buzón sirve para conversar con alguien, no para contar
respuestas. Una encuesta que no pide nombre, correo ni ningún dato que identifique a nadie no recoge datos personales
ni datos de uso del producto: lo que el ADR 0027 prohíbe es que kitlegal envíe a un servidor las preguntas, los
asuntos o el uso de quien lo usa, y eso no cambia.

**Decisión.**

- **`/plan/` cuenta el plan en lenguaje llano** y termina con una encuesta anónima. El estado y el orden de cada hito
  salen de `docs/roadmap/` al construir, y lo que kitlegal hace hoy, de `data/cobertura.yaml`; los textos llanos
  viven en `web/src/data/plan.ts`, y la construcción falla si un hito pendiente no tiene el suyo.
- **La encuesta es un formulario HTML sin JavaScript que se envía a un formulario de Google** solo cuando la persona
  pulsa «Enviar». La página no carga nada de Google: ni scripts, ni marcos, ni fuentes. La política de seguridad de
  contenido de esa página admite enviar el formulario al origen `https://docs.google.com` y a ningún otro; las demás
  páginas siguen con `form-action 'none'`.
- **Es anónima.** No pide nombre ni correo, el formulario de Google no recoge el correo ni pide iniciar sesión, y los
  campos libres piden no incluir datos de clientes ni de asuntos. La página dice que Google recibe el envío como
  cualquier web que se visita y que kitlegal solo ve las respuestas.
- **Las preguntas tienen una sola fuente**, `web/src/data/encuesta.ts`. `web/scripts/encuesta.mjs` genera con ellas
  el guion de Google Apps Script que crea el formulario, y lo que el guion devuelve —la dirección de envío, el campo de
  cada pregunta y la huella de las preguntas— va en `web/src/data/encuesta-formulario.json`. Si las preguntas cambian
  después de crear el formulario, la huella deja de coincidir y la construcción falla. Sin formulario conectado, la
  página se publica sin la encuesta.
- Esto precisa la opción 2 de este ADR («Un formulario. Rechazada») y su decisión «No hay formulario ni nada de
  terceros», y la consecuencia «Sin nada de terceros en la página» del ADR 0024, solo para `/plan/`: allí el tercero
  aparece en el envío, nunca al cargar la página.

**Consecuencias.**

- Las respuestas llegan a una hoja de cálculo de Google de la persona, que se descarga para analizarlas. Lo que digan
  entra en `docs/USO.md`, como lo que llega por correo (ADR 0013).
- Un cambio en `docs/roadmap/` vuelve a construir y publicar la web (`.github/workflows/web.yml`), y una entrada nueva
  en la cola necesita su texto llano en `web/src/data/plan.ts`.
