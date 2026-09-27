# 0009 · Genericidad territorial: sirve para cualquier municipio, se valida primero en uno

- **Estado**: aceptada
- **Fecha**: 2026-09-11
- **Hito**: transversal (entra en H7)

> Nota (2026-09-27): el ADR 0027 cambia de quién es el municipio que guarda `.kitlegal/config.yaml`: ya no
> es el de la persona que usa el kit, sino los del asunto —el del interesado y el de la sede del órgano, los
> que mira el art. 30.6 LPAC—, porque un despacho lleva asuntos de muchos municipios. Lo demás no cambia.

## Contexto y problema

Los documentos de diseño estaban escritos desde Leganés, y no solo en los ejemplos. El mecanismo de
extensión era un YAML curado a mano por ayuntamiento (`entidades/leganes.yaml`), la fase del pack
fiscalizador se titulaba «sobre Leganés» y su aceptación dependía de casos de Leganés. Además daba por
supuestas cosas que solo valen en la Comunidad de Madrid:

- Que el boletín local es el BOCM. Madrid es uniprovincial y no tiene BOP; un municipio de una
  comunidad multiprovincial, como Tordesillas, necesita el boletín autonómico **y** el provincial.
- Que las ordenanzas se obtienen con un crawler por ayuntamiento, cuando muchos municipios pequeños no
  tienen sede propia.
- Que los festivos son un fichero por año, cuando los locales cambian por municipio y afectan al
  cálculo de plazos (art. 30.6 LPAC).
- Que el Catastro OVC y la normativa tributaria estatal cubren todo el territorio, cuando País Vasco y
  Navarra tienen catastros y haciendas forales.
- Que los umbrales de anomalías pueden calibrarse con los casos de un solo municipio, cuando en uno
  pequeño la concentración de adjudicatarios es alta por naturaleza.

Con ese diseño, cada municipio nuevo exige trabajo manual y los resultados fuera de Madrid son
silenciosamente incorrectos.

## Opciones consideradas

1. **Particularizar primero y generalizar después.** Rápido para Leganés, pero la generalización
   posterior obliga a rehacer herramientas y skills, y lo particular tiende a quedarse.
2. **Genérico y completo desde el principio**: configurar las 19 comunidades y ciudades autónomas y
   todos los boletines antes de entregar nada. Correcto, pero contrario al criterio lean: mucho trabajo
   de verificación antes de que nadie use nada.
3. **Genérico en el diseño, local en la implementación**: las herramientas y las skills no conocen
   ningún municipio; los datos de municipio salen de registros nacionales; lo territorial se configura
   por comunidad o por boletín, y solo se rellena primero el territorio de validación. Fuera de él, la
   salida declara su cobertura.

## Decisión

Se adopta la **opción 3**, como principio IX de la constitución:

- Ninguna herramienta, dato versionado ni skill lleva un caso especial para un municipio concreto.
- Los datos de cada municipio (código INE, provincia, comunidad, DIR3) salen de registros públicos
  nacionales. Lo que varía por territorio (boletines, festivos locales, órganos de control externo y de
  transparencia, plataformas de contratación, régimen foral) se configura por comunidad o por boletín en
  `data/territorio/` y `data/boletines/`. No existe configuración curada por ayuntamiento.
- El municipio de la persona usuaria vive en su `.kitlegal/config.yaml`, nunca en el repositorio ni en
  un pack.
- Territorio de validación: **Comunidad de Madrid**, municipio de referencia **Leganés**. Se implementa
  primero. Los demás territorios (BOCYL y resto de boletines autonómicos, BOP de comunidades
  multiprovinciales, régimen foral…) llegan en la fase 6 del roadmap como datos y adaptadores, sin
  tocar las skills.
- Fuera del territorio cubierto, la salida declara la cobertura **dentro de `data`** (el sobre de salida
  no cambia): qué está disponible y qué no. Ninguna skill concluye «no existe» a partir de un resultado
  sin cobertura completa.
- Las fuentes nacionales (PLACSP, BDNS, BORME, TEU, presupuestos de entidades locales, INE) funcionan
  para cualquier municipio desde el primer día; solo lo territorial espera a su fase.
- Las ordenanzas se buscan en el boletín donde la ley obliga a publicarlas (art. 70.2 LRBRL; art. 17.4
  TRLRHL para las fiscales), no con un crawler por ayuntamiento.
- Toda herramienta con dimensión territorial se prueba con una matriz: un municipio cubierto (Leganés),
  uno no cubierto (Tordesillas) y, cuando aplique, uno foral.
- Las reglas de anomalías relativas comparan con municipios de tamaño parecido (padrón del INE) y nunca
  se calibran sobre uno solo.

## Consecuencias

**A favor**

- Un municipio nuevo no requiere trabajo: lo nacional funciona ya y lo territorial se hereda de su
  comunidad cuando esta se configura.
- Los resultados fuera del territorio cubierto son incompletos pero honestos, nunca silenciosamente
  incorrectos.
- Añadir una comunidad es un cambio de datos revisable, no una rama de código.
- Leganés sigue siendo donde se valida con casos reales, sin contaminar el diseño.

**En contra, y asumido**

- Hace falta la pieza `territorio` (H7) antes de las skills municipales, y cada applet territorial
  carga con la matriz de pruebas y el campo de cobertura.
- Mientras solo esté cubierta la Comunidad de Madrid, en el resto de España los boletines, las
  ordenanzas y los festivos locales devuelven cobertura parcial.
- La promesa comercial de «entidades curadas por municipio» desaparece: los datos de municipio son
  públicos y se generan, no se curan.
