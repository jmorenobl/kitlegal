# Jurisprudencia: qué se puede consultar y por dónde

Propuesta de hito para el backlog (grupo «fuentes»), con el estado de cada vía verificado el **2026-09-13**.
No es una decisión cerrada: es el trabajo previo para que, cuando la jurisprudencia entre en una fase, no haya
que volver a investigar los términos de uso.

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
| **CENDOJ** | consulta individual **para uso particular**; prohibidos uso comercial, descarga masiva y elaboración de bases de datos sin licencia del CGPJ; `Crawl-delay: 5` | ROJ + ECLI | 🔴 |

Aviso legal del CENDOJ, literal (`poderjudicial.es/search/indexAN.jsp`):

> «El usuario de la base de datos podrá consultar los documentos siempre que lo haga para su uso particular.
> No está permitida la utilización de la base de datos para usos comerciales, ni la descarga masiva de
> información. La reutilización de esta información para la elaboración de bases de datos o con fines
> comerciales debe seguir el procedimiento y las condiciones establecidas por el CGPJ a través de su Centro de
> Documentación Judicial.»

Condiciones del BOE, literales (`boe.es/informacion/aviso_legal/index.php#reutilizacion`): «Las presentes
condiciones permiten la reutilización de los documentos sometidos a ellas para fines comerciales y no
comerciales», citando «Fuente de los datos: Agencia Estatal Boletín Oficial del Estado», sin desnaturalizar el
sentido de la información y mencionando la fecha de la última actualización.

## 3. Lo que no se puede hacer, y por qué

**No existe forma de resolver un ECLI del Tribunal Supremo sin interrogar el buscador del CENDOJ.** Verificado:

- Las URLs de documento llevan un **id interno opaco** (`/search/AN/openDocument/{hash}/{fecha}`,
  `/search/contenidos.action?…reference={n}…`) que no se deriva del ECLI ni del ROJ.
- El ECLI y el ROJ **no están en el índice de texto libre**, así que la ruta de búsqueda por GET
  (`/search/sentencias/{texto}/{pagina}/{base}`, que sí existe y responde) devuelve cero resultados para ellos.
  Solo aparecen como campos del **formulario avanzado**.
- No hay resolutor nacional, ni API pública, ni fichero de mapeo publicado por el CGPJ o Justicia. El estándar
  español **no define URL persistente**: las Conclusiones del Consejo sobre el ECLI (BOE `DOUE-Z-2019-70039`)
  prevén la URL como **metadato del registro**, no como algo deducible del identificador.

Lo que sí es determinista y no toca la red: **la equivalencia ECLI ↔ ROJ**. La ficha oficial española dice que
el número final del ECLI «se corresponde con el número correlativo del identificador nacional ROJ», de modo que
`ECLI:ES:TS:2026:3505` **es** `STS 3505/2026`.

## 4. Hito propuesto

Reutiliza el 80 % de H4 (`httpx`, caché, sobre, exit codes, `--describe`, grabación de fixtures). Lo nuevo es
el parseo del sumario, que tiene XSD, y el descomponedor de ECLI.

1. **`kitlegal boe sumario <fecha>`** sobre `/datosabiertos/api/boe/sumario/{aaaammdd}`, con el sobre de
   siempre. Filtrando por el departamento `TRIBUNAL CONSTITUCIONAL` y por la sección del TC salen todas sus
   sentencias publicadas; en «III. Otras disposiciones» aparecen las del TS que anulan disposiciones.
2. **`kitlegal cita resolver ECLI:ES:TC:AAAA:N`** → la sentencia con su `BOE-A-…`, fecha, número y URL oficial,
   resuelta contra el índice del BOE.
3. **`ECLI:ES:TS`** → reconoce el identificador, **da su ROJ equivalente** (que es determinista) y sale con
   código informativo explicando que la reutilización del CENDOJ exige licencia del CGPJ, con el enlace al
   buscador oficial para la consulta humana. Mismo patrón de frontera humana que rige para las sedes
   electrónicas: la herramienta no fabrica una URL que no puede sostener.

Si algún día se quiere jurisprudencia del Supremo de verdad, el camino **no es técnico**: es solicitar al
CENDOJ la **licencia tipo de reutilización** (anual, no exclusiva) por el procedimiento que su propio aviso
legal señala. Es una decisión de negocio, no de código.

## 5. Pendiente de verificar antes de fijar nada en `data/` o en código

- El código exacto de la sección del TC en el sumario: el XSD no enumera los códigos de sección.
- Que la API del sumario responde 200 con `Accept: application/json` (documentado, no probado en vivo).
- Si el buscador del CENDOJ tiene hoy CAPTCHA: lo afirman fuentes de terceros de 2026, no está verificado. Si
  lo tiene, **no se sortea**: es una medida técnica de protección.
- La cifra que circula de que el CENDOJ considera «descarga masiva» unas 100 descargas diarias no aparece en
  ningún texto oficial.

## 6. Nota sobre la caché y el grafo

Si alguna vez se consultara una fuente cuyos términos prohíben «la elaboración de bases de datos», habría que
tener presente que kitlegal **guarda por diseño todo lo que ve**: la caché (`~/.cache/kitlegal/cache.db`) y,
desde H7, el grafo del mundo (`world.db`, con su tabla `texts`). Una consulta puntual repetida en el tiempo
construye una base de datos. Excluir una fuente de la caché y del grafo sería una excepción al diseño general y
necesitaría su propio ADR.
