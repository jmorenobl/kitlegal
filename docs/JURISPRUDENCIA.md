# Jurisprudencia: qué se puede consultar y por dónde

Propuesta de hito para el backlog (grupo «fuentes»), con el estado de cada vía verificado el **2026-09-13** y,
para el CENDOJ, vuelto a verificar el **2026-10-02** con una prueba a mano (§3). No es una decisión cerrada: es
el trabajo previo para que, cuando la jurisprudencia entre en una fase, no haya que volver a investigar los
términos de uso. Lo que aquí se propone sobre el CENDOJ depende del ADR 0036, que sustituye en parte al 0003 y
está en estado de propuesta.

Ante conflicto, mandan la constitución y `docs/ROADMAP.md`. Lo que aquí se afirma sobre una fuente se comprueba
contra su fila de `docs/SOURCES.md` el día que se implemente: los términos cambian.

## 1. El hallazgo: el BOE publica jurisprudencia

`refs/mapa-sistema-legal-skills.md` clasifica la jurisprudencia como 🔴 porque piensa en el CENDOJ. Pero hay
una parte que llega por una fuente que ya está construida y cuya reutilización está expresamente autorizada:

- **Todas las sentencias del Tribunal Constitucional** se publican en el BOE (sección «T.C.» del sumario y
  departamento `TRIBUNAL CONSTITUCIONAL` de la sección I).
- **Las sentencias del Tribunal Supremo que anulan disposiciones** se publican en «III. Otras disposiciones».
- La API de datos abiertos expone el sumario diario en `/datosabiertos/api/boe/sumario/{aaaammdd}`, con XSD
  publicado: `sumario > diario > seccion[@codigo,@nombre] > departamento > epigrafe? > item{identificador,
  titulo, url_pdf, url_html, url_xml}`. Se navega el árbol; no hace falta filtrar en el servidor.

Y el puente que lo hace útil: **en `ECLI:ES:TC:AAAA:N`, la `N` es el número de sentencia**, que es como el BOE
titula cada una («Sentencia 79/2024»). Es decir, un ECLI del Constitucional **se resuelve contra el índice del
propio BOE**, sin tocar ningún buscador judicial. Verificado sobre `BOE-A-2024-12808`, cuyo texto contiene
`ECLI:ES:TC:2024:79` literal (en el cuerpo, no como metadato estructurado).

## 2. Semáforos

| Fuente | Qué permite | Identificadores | |
|---|---|---|---|
| **BOE datos abiertos** (sumario) | API REST documentada, XML y JSON; reutilización comercial y no comercial autorizada con atribución; `robots.txt` no menciona `/datosabiertos` | `BOE-A-…`, ELI; **no** da ECLI ni ROJ | 🟢 |
| **HJ del Tribunal Constitucional** | solo HTML; `robots.txt` da 404 y **no se localiza aviso legal**; URLs estables por id interno (`/es/Resolucion/Show/{id}`, y `/Resolucion/Api/json|xml/{id}`, no documentadas); sin CAPTCHA observado | ECLI, nº de sentencia, **nº de recurso**, ponente, **nº y fecha de BOE** | 🟡 |
| **Resolutor ECLI de e-Justice** | la URL `https://e-justice.europa.eu/ecli/{ECLI}` es oficial, pero por HTTP devuelve un armazón vacío (probado con un ECLI español, uno neerlandés y uno portugués); el servicio SOAP está declarado **no disponible** por la Comisión | — | 🟡 (enlace de cortesía, no fuente) |
| **CENDOJ: resolver una resolución identificada** | el formulario tiene campos `ECLI`, `ROJ`, número de resolución y número de recurso; una consulta por HTTP simple devuelve los metadatos y la URL del documento (§3); `robots.txt` no excluye `/search/`, `Crawl-delay: 5` | ROJ + ECLI | 🟡 (ADR 0036, propuesta) |
| **CENDOJ: búsqueda por materia, texto íntegro, descarga** | consulta individual **para uso particular**; prohibidos uso comercial, descarga masiva y elaboración de bases de datos sin seguir el procedimiento del CGPJ | ROJ + ECLI | 🔴 |

Aviso legal del CENDOJ, literal (`poderjudicial.es/search/indexAN.jsp`):

> «El usuario de la base de datos podrá consultar los documentos siempre que lo haga para su uso particular.
> No está permitida la utilización de la base de datos para usos comerciales, ni la descarga masiva de
> información. La reutilización de esta información para la elaboración de bases de datos o con fines
> comerciales debe seguir el procedimiento y las condiciones establecidas por el CGPJ a través de su Centro de
> Documentación Judicial.»

El aviso no habla de consultas automatizadas: limita el uso al particular y prohíbe el comercial y la descarga
masiva. `poderjudicial.es/robots.txt`, leído el 2026-10-02: `Crawl-delay: 5` para todos los agentes y, del
buscador, solo excluye `/search_old/`.

Condiciones del BOE, literales (`boe.es/informacion/aviso_legal/index.php#reutilizacion`): «Las presentes
condiciones permiten la reutilización de los documentos sometidos a ellas para fines comerciales y no
comerciales», citando «Fuente de los datos: Agencia Estatal Boletín Oficial del Estado», sin desnaturalizar el
sentido de la información y mencionando la fecha de la última actualización.

## 3. El CENDOJ: qué se ha probado y qué no

### Resolver un ECLI por el formulario funciona con HTTP simple (2026-10-02)

Prueba a mano, dos peticiones separadas 6 segundos, con el agente `kitlegal/0.3 (+https://kitlegal.es/bot)`:

1. `GET https://www.poderjudicial.es/search/indexAN.jsp` → 200 (73 kB) y la cookie de sesión. Con el agente por
   defecto de `curl`, 403.
2. `POST https://www.poderjudicial.es/search/search.action` con la cookie, `X-Requested-With: XMLHttpRequest` y
   los campos `action=query`, `sort=IN_FECHARESOLUCION:decreasing`, `recordsPerPage=10`, `databasematch=AN`,
   `start=1` y `ECLI=ECLI:ES:TS:2023:3144` → 200 (11 kB), **un resultado**: ROJ `STS 3144/2023`, Sala de lo
   Civil, 4 de julio de 2023, resolución 1088/2023, recurso 4703/2019, ponente, el resumen del propio CENDOJ y
   la URL del documento (`/search/AN/openDocument/{hash}/{fecha}`).

Lo que enseña:

- El formulario (`frmBusquedajurisprudencia`) tiene campos propios `ECLI`, `ROJ`, `NUMERORESOLUCION`,
  `NUMERORECURSO`, `PONENTE`, fechas, `JURISDICCION` y `TIPOORGANOPUB`. Sus campos ocultos son parámetros de la
  búsqueda; no lleva ningún token. No hace falta navegador ni ejecutar JavaScript.
- La respuesta no trae CAPTCHA, pero la página carga `captcha.css`: lo hay en algún punto y no se sabe cuándo
  aparece. Si aparece, **no se sortea**: es una medida técnica de protección.
- El ejemplo confirma la equivalencia ECLI ↔ ROJ de más abajo.

Segunda prueba a mano, el 2026-10-03, con la misma sesión y seis segundos entre consultas:

| Consulta | Respuesta |
|---|---|
| `ROJ=STS 3144/2023` | 200, un resultado: la misma sentencia |
| `NUMERORESOLUCION=1088/2023` con `FECHARESOLUCIONDESDE` y `FECHARESOLUCIONHASTA` en `04/07/2023` | 200, un resultado: la misma sentencia |
| `ECLI=ECLI:ES:TS:2023:999999` | 200, 744 bytes: «No se ha encontrado ningún resultado» |

Así que la cita como se escribe en un escrito —«STS 1088/2023, de 4 de julio», con el número de resolución, que no
es el ROJ— se resuelve a su ROJ y su ECLI, y una referencia inventada da cero resultados con una respuesta que se
distingue. El número de resolución solo es único con la fecha: cada órgano numera las suyas.

Sin probar: `NUMERORECURSO`; el filtro por órgano (`TIPOORGANOPUB`), que haría falta para un número de resolución
sin fecha; órganos distintos del Supremo; y a partir de cuántas consultas responde con CAPTCHA o con un bloqueo.

### Lo que sigue sin poder hacerse sin el formulario

- Las URLs de documento llevan un **id interno opaco** (`/search/AN/openDocument/{hash}/{fecha}`,
  `/search/contenidos.action?…reference={n}…`) que no se deriva del ECLI ni del ROJ: no se puede dar un enlace
  a una sentencia sin haberla resuelto antes.
- El ECLI y el ROJ **no están en el índice de texto libre**, así que la ruta de búsqueda por GET
  (`/search/sentencias/{texto}/{pagina}/{base}`, que existía el 2026-09-13) devuelve cero resultados para ellos.
- No hay resolutor nacional, ni API pública, ni fichero de mapeo publicado por el CGPJ o Justicia. El estándar
  español **no define URL persistente**: las Conclusiones del Consejo sobre el ECLI (BOE `DOUE-Z-2019-70039`)
  prevén la URL como **metadato del registro**, no como algo deducible del identificador.

Lo que es determinista y no toca la red: **la equivalencia ECLI ↔ ROJ**. La ficha oficial española dice que
el número final del ECLI «se corresponde con el número correlativo del identificador nacional ROJ», de modo que
`ECLI:ES:TS:2026:3505` **es** `STS 3505/2026`.

### El marco de la reutilización

- **La Ley 37/2007 se aplica a las sentencias**: «Las previsiones contenidas en la presente ley serán de
  aplicación a las sentencias y resoluciones judiciales, sin perjuicio de lo previsto en el artículo 107.10 de
  la Ley Orgánica 6/1985 […] y su desarrollo específico» [BOE-A-2007-19814, bloque `dasegunda`, apartado 2].
  La reutilización es gratuita salvo tarifa limitada a costes marginales (bloque `a7`), las condiciones han de
  ser objetivas, proporcionadas y no discriminatorias (bloque `a4`), y una solicitud se resuelve en veinte días
  (bloque `a10`). Redacción vigente desde el 2021-11-04, leída el 2026-10-02.
- **El CGPJ tiene la competencia** de cuidar de la publicación oficial de las sentencias y potestad
  reglamentaria sobre «publicación y reutilización de las resoluciones judiciales» [BOE-A-1985-12666, bloque
  `aquinientossesenta`, apartados 1.10.ª y 1.16.ª e)].
- **El reglamento que fijaba licencias y precios está anulado.** El Reglamento 3/2010 del CGPJ
  (`BOE-A-2010-17860`) preveía reutilización libre para docencia e investigación sin fin comercial, licencia
  tipo con precio público y autorización caso a caso. El Tribunal Supremo lo anuló el 28 de octubre de 2011
  (Pleno de la Sala Tercera), según la prensa jurídica; la sentencia no se ha leído. No se ha localizado un
  reglamento que lo sustituya.
- **La web del CGPJ no publica el procedimiento.** La página de Jurisprudencia del CENDOJ dice que «gestiona el
  suministro de sentencias […] a diversos reutilizadores, atendiendo igualmente solicitudes que no son
  constitutivas de reutilización, conforme a los criterios establecidos en la Ley de Reutilización de la
  Información del Sector Público». No se han encontrado formulario, condiciones de licencia ni tarifas. Contacto:
  `cendoj@poderjudicial.es`.

Preguntas abiertas, que solo el CGPJ puede contestar (por correo al CENDOJ, o con una solicitud de transparencia
sobre su actividad administrativa): con qué norma exige hoy licencia y precio; si hay una modalidad no
comercial; cuáles son las tarifas y condiciones vigentes; y si resolver un identificador desde el equipo de la
persona cuenta como reutilización.

## 4. Hito propuesto

Reutiliza la infraestructura de H4 (`httpx`, caché, sobre, exit codes, `--describe`, grabación de fixtures). Lo
nuevo es el parseo del sumario, que tiene XSD, el descomponedor de ECLI y el adaptador del formulario.

1. **`kitlegal boe sumario <fecha>`** sobre `/datosabiertos/api/boe/sumario/{aaaammdd}`, con el sobre de
   siempre. Filtrando por el departamento `TRIBUNAL CONSTITUCIONAL` y por la sección del TC salen todas sus
   sentencias publicadas; en «III. Otras disposiciones» aparecen las del TS que anulan disposiciones.
2. **`kitlegal cita resolver ECLI:ES:TC:AAAA:N`** → la sentencia con su `BOE-A-…`, fecha, número y URL oficial,
   resuelta contra el índice del BOE. Su texto se puede leer y citar: es del BOE.
3. **Resolver una resolución identificada de cualquier otro órgano** (ECLI, ROJ, o número de resolución con su
   fecha y órgano) por el formulario del CENDOJ: una consulta por resolución, al ritmo del `robots.txt`, con el
   agente identificable. Devuelve si existe, sus metadatos y la URL oficial. Cero resultados es la fila de «no
   encontrado» (ADR 0023): la cita no se sostiene. No descarga ni lee el texto. Depende del ADR 0036.
4. **La búsqueda por materia la hace la persona.** La skill le prepara la consulta para el buscador (texto,
   operadores, jurisdicción, órgano, fechas); la persona busca en su navegador, elige, y vuelve con el ECLI o el
   ROJ —que se resuelven por el punto 3— o con el texto de la sentencia, que el agente lee y cita. Mismo patrón
   de frontera humana que rige para las sedes electrónicas.
5. **Regla de la skill**: ninguna sentencia se cita sin haberla resuelto, y ninguna se resume ni se caracteriza
   sin haber leído su texto. Una sentencia que el modelo recuerda y que no se resuelve no se cita.

Lo que queda fuera: buscar por materia de forma automática, descargar o guardar el texto de las sentencias del
CENDOJ, y cualquier forma de sortear un CAPTCHA o un bloqueo. Para eso el camino **no es técnico**: es el
procedimiento de reutilización del CGPJ (§3).

Por qué importa: las multas de los tribunales por escritos hechos con IA han sido por jurisprudencia inventada
(`docs/USO.md`), y el benchmark de asistentes legales de observatorio.legal da a la verificación de citas el
mayor peso de su índice (fiabilidad de fuente, 23 de 100; trazabilidad, 9), con 15 de sus 90 preguntas de
jurisprudencia y 10 de referencias falsas. Con el punto 3, una sentencia inventada da cero resultados.

## 5. Pendiente de verificar antes de fijar nada en `data/` o en código

- El código exacto de la sección del TC en el sumario: el XSD no enumera los códigos de sección.
- Que la API del sumario responde 200 con `Accept: application/json` (documentado, no probado en vivo).
- Del formulario del CENDOJ, lo que §3 deja sin probar: `NUMERORECURSO`, el filtro por órgano, otros órganos y
  cuándo aparece el CAPTCHA.
- Si la ruta de búsqueda por GET sirve como enlace para que la persona abra en su navegador una búsqueda ya
  preparada.
- La cifra que circula de que el CENDOJ considera «descarga masiva» unas 100 descargas diarias no aparece en
  ningún texto oficial.

## 6. Nota sobre la caché y el grafo

kitlegal **guarda por diseño todo lo que ve**: la caché (`~/.cache/kitlegal/cache.db`) y, desde H7, el grafo
del mundo (`world.db`, con su tabla `texts`). Una consulta puntual repetida en el tiempo construye una base de
datos, y el aviso del CENDOJ reserva «la elaboración de bases de datos» al procedimiento del CGPJ. El ADR 0003
ya admite que una `Resolucion` entre en el grafo con ECLI, metadatos y URL, y nunca con el texto íntegro; el
hito tiene que decir qué guarda de la respuesta del formulario y durante cuánto, y el resumen del CENDOJ es el
punto dudoso. Excluir una fuente de la caché y del grafo sería una excepción al diseño general y necesitaría su
propio ADR.
