# Fuentes

Las fuentes públicas externas que utiliza kitlegal: de dónde salen los datos, con qué licencia, bajo qué términos de
uso y cómo se incorporan al proyecto. Cualquier cambio en este fichero es una decisión humana (constitución, capa 3):
cada fila la revisa una persona, que anota en «Revisado» el día en que leyó los términos de uso y el `robots.txt`.
Mientras nadie lo ha hecho, «Revisado» dice `pendiente` y la fila es una propuesta: ninguno de sus datos está comprobado.

kitlegal distingue dos categorías de fuentes:

1. **Fuentes consultadas en ejecución (red)**: el binario realiza peticiones HTTP vivas a través de `internal/httpx`
   durante la ejecución normal de los applets. Tienen un ritmo mínimo entre peticiones (`Ritmo`), respetan estrictamente
   `robots.txt` y no requieren autenticación ni simular a un usuario. Si los términos de uso o el `robots.txt` prohíben el
   acceso automatizado, la fuente no se usa y el hito que la introduce se detiene.
2. **Fuentes de datos congelados en `data/`**: no se consultan nunca en tiempo de ejecución ni en grabaciones (ADR 0017).
   Son el origen de ficheros estáticos versionados en el repositorio (en `data/territorio/`, `data/festivos/`, etc.),
   generados a partir de descargas públicas mediante tareas offline `[datos]`. En ellas **«Ritmo» no aplica** (el binario no
   emite ninguna petición viva); en su lugar se anota la **fecha del fichero** (fecha oficial de publicación, referencia o
   generación del volcado de origen) y su destino en `data/`. Aunque no se pidan en red durante la ejecución, sus términos
   de uso, licencias y condiciones de acceso se auditan y verifican con el mismo rigor para garantizar la legalidad de la
   reutilización.

## Fuentes consultadas en ejecución (red)

Columnas:

- **Fuente**: el nombre que va en `fuente` del sobre y que nombra las grabaciones.
- **Applet**: el applet que la consulta.
- **Base**: la dirección bajo la que están todas sus peticiones.
- **Licencia**: la de reutilización de los datos.
- **Términos de uso**: la dirección de los términos, entre `<…>`.
- **robots.txt**: el resultado de su revisión.
- **Ritmo**: la separación mínima entre dos peticiones al mismo sitio, como literal de duración de Go.
- **Formato**: en qué formato se piden los recursos y si hace falta autenticación.
- **Revisado**: el día de la revisión, `AAAA-MM-DD`, o `pendiente`.

Cada adaptador ata su fila a sus constantes con un test que falla si divergen: el ritmo es el intervalo con el que pide
y la dirección de los términos y el día de la revisión son los que declara su `Terms()`. Para `boe`,
`TestFuenteCoincideConSources` compara la fila con `internal/source/boe/terminos.go`.

| Fuente | Applet | Base | Licencia | Términos de uso | robots.txt | Ritmo | Formato | Revisado |
|---|---|---|---|---|---|---|---|---|
| `boe.legislacion-consolidada` | `boe` | <https://www.boe.es/datosabiertos/api/legislacion-consolidada> | Ley 37/2007 (reutilización comercial y no comercial permitida; citar «Fuente de los datos: Agencia Estatal Boletín Oficial del Estado», no desnaturalizar el sentido de la información y mencionar la fecha de la última actualización) | <https://www.boe.es/informacion/aviso_legal/index.php> | `User-agent: *` sin `Crawl-delay`; no restringe `/datosabiertos/` (sus `Disallow` son búsquedas dinámicas, PDFs duplicados, edictos judiciales y documentos concretos) | `1s` | XML (bloque) y JSON; sin autenticación | 2026-09-13 |

## Fuentes de datos congelados en `data/`

Estas fuentes corresponden a datos públicos que se incorporan como ficheros estáticos en `data/` mediante tareas
offline `[datos]` y no se piden nunca por red en ejecución ni en grabaciones (ADR 0017). La columna «Ritmo» no aplica
al no existir peticiones vivas del binario; lo que se documenta es la procedencia, la vía de obtención, la fecha de
generación o publicación del fichero de origen y las condiciones jurídicas de reutilización.

Columnas:

- **Fuente**: identificador que nombra la procedencia de los datos (va en `source` de las entidades o del fichero).
- **Destino**: ruta bajo `data/` donde se versionan los datos transformados.
- **Origen**: dirección oficial de descarga o consulta del conjunto de datos.
- **Licencia**: condiciones legales de reutilización y obligaciones de atribución.
- **Términos de uso**: dirección oficial donde constan los términos legales, entre `<…>`.
- **robots.txt / Acceso**: resultado de la comprobación técnica de acceso y `robots.txt`.
- **Fecha del fichero**: fecha oficial de generación, publicación o referencia del archivo descargado (no una petición viva).
- **Formato original**: formato del archivo de origen y características de acceso.
- **Revisado**: el día de la revisión de términos y acceso, `AAAA-MM-DD`.

| Fuente | Destino | Origen | Licencia | Términos de uso | robots.txt / Acceso | Fecha del fichero | Formato original | Revisado |
|---|---|---|---|---|---|---|---|---|
| `ine.municipios` | `data/territorio/` | <https://www.ine.es/daco/daco42/codmun/diccionario26.xlsx> | Creative Commons Reconocimiento ([CC BY 4.0](https://creativecommons.org/licenses/by/4.0/deed.es); reutilización comercial y no comercial permitida; citar «Fuente: Sitio web del INE: www.ine.es» o «Elaboración propia con datos extraídos del sitio web del INE: www.ine.es», mencionar fecha de última actualización y no desnaturalizar la información) | <https://www.ine.es/dyngs/AYU/index.htm?cid=125> | `User-agent: *` permite `/daco/` y `/dyngs/` (solo bloquea `/cgi-bin/`, `/buscar/`, `/Test/`, `/Admin/`, `/testin/`). El INE solo publica la relación completa con dígito de control en `.xlsx` y se actualiza anualmente; entra congelada en `data/territorio/` vía tarea `[datos]` para no añadir dependencias de lectura de hojas de cálculo al binario (ADR 0017) | 2026-02-04 (referencia a 1 de enero de 2026; 8.132 municipios con dígito de control oficial) | Hoja de cálculo Excel OpenXML (`.xlsx`, UTF-8); sin autenticación | 2026-09-18 |
| `ine.codigos-territoriales` | `data/territorio/` | <https://www.ine.es/daco/daco42/codmun/cod_ccaa.htm> y <https://www.ine.es/daco/daco42/codmun/cod_provincia.htm> | Creative Commons Reconocimiento ([CC BY 4.0](https://creativecommons.org/licenses/by/4.0/deed.es); las mismas condiciones y la misma atribución que `ine.municipios`, por ser la misma publicadora) | <https://www.ine.es/dyngs/AYU/index.htm?cid=125> | `User-agent: *` permite `/daco/`, la misma ruta que ya usa `ine.municipios` | 2026-09-20 (fecha de descarga, en la pausa de T002; las tablas de códigos no declaran fecha de referencia) | HTML (`text/html`, ISO-8859-1); tabla de dos columnas, código y literal; sin autenticación | 2026-09-21 |
| `mpt.rel` [^rel-dir3] | `data/territorio/` | <https://registroentidadeslocales.mpt.es/REL/frontend/export_data/file_export/export_excel/municipios/all/all> | Ley 37/2007 y Real Decreto 1495/2011 (reutilización comercial y no comercial permitida; citar «Origen de los datos: Ministerio de Política Territorial y Memoria Democrática», mencionar fecha de última actualización, no desnaturalizar y no atribuir patrocinio) | <https://mptmd.gob.es/portal/footer/nota_legal> | `robots.txt` responde `HTTP 404` (acceso permitido según REP). Descarga directa automatizable con `curl -sL` sin credenciales, sin cookies de sesión y sin WAF bloqueante | 2026-09-21 (fecha del volcado descargado en la pausa `[datos]` de T003; el REL es un registro continuo y no publica fecha de corte) | Binario Microsoft Excel BIFF8 (`.xls`, ~2 MB, OLE Compound Document V2); campos `NUMERO_INSCRIPCION` (`01PPMMMDC`), `DENOMINACION`, etc.; sin autenticación | 2026-09-18 |
| `comunidad-madrid.festivos-locales` | `data/festivos/` | <https://datos.comunidad.madrid/catalogo/dataset/02c712b5-5009-4ffb-b388-3cfb4fca207d/resource/24c290f6-e5bc-4ad2-8476-10e1c3369429/download/festivos_locales_historicos.json> | Creative Commons Reconocimiento ([CC BY 4.0](https://creativecommons.org/licenses/by/4.0/legalcode.es); reutilización comercial y no comercial permitida; citar «Comunidad de Madrid», indicar fecha de extracción o uso, no desnaturalizar y no atribuir patrocinio) | <https://www.comunidad.madrid/servicios/informacion-atencion-ciudadano/aviso-legal-privacidad> | `User-agent: *` / `Disallow: /` con `Crawl-delay: 30` en `datos.comunidad.madrid/robots.txt` (bloqueo total a bots, incluida la API CKAN en `/api/` y `/datastore/`). Al prohibir el acceso automatizado, `internal/httpx` no puede consultarlo ni en ejecución ni en grabaciones. Se descarga fuera de la ejecución y se congela en `data/festivos/` (ADR 0017) | 2026-01-13 (última modificación registrada en catálogo CKAN; serie histórica 1998–2025; fuente legal resoluciones anuales D.G. Trabajo en BOCM) | JSON UTF-8 (`festivos_locales_historicos.json`, raíz `{"data": [...]}`) y CSV ISO-8859-1 (`festivos_locales_historicos.csv`); campos `año`, `municipio_codigo`, `fecha_festivo`; sin autenticación | 2026-09-18 |

[^rel-dir3]: **Por qué la fila es el REL y no el inventario DIR3**: Para la resolución territorial del código DIR3 de los ayuntamientos, la fuente documentada es el Registro de Entidades Locales (REL) y **no** el inventario DIR3, debido a que este último es inaccesible de forma abierta y automatizada por las siguientes razones constatadas en la investigación del issue #37:
- **datos.gob.es**: El conjunto nacional consolidado (`E05188501`) está despublicado (`HTTP 404`), sus distribuciones RDF no resuelven y el portal web alternativo está protegido tras el WAF Imperva/Incapsula.
- **CTT (Centro de Transferencia de Tecnología)**: El archivo `Listado Unidades EELL.xlsx` en el área de descargas de Administración Electrónica está tras el desafío de un WAF (F5 ASM) que devuelve una respuesta vacía de 0 bytes a clientes no interactivos (`curl`, `internal/httpx`). Además, su contenido son unidades orgánicas internas subordinadas, no una correspondencia territorial 1:1 con códigos de municipio INE.
- **Servicios web DIR3 (SOAP y REST)**: Existen, pero están restringidos a Administraciones Públicas a través de la **Red SARA** (requieren autorización previa y credenciales oficiales).
- **FACe**: No dispone de volcado masivo descargable, sus servicios web de consulta masiva exigen certificado electrónico X.509 de proveedor o administración y no devuelven el código INE del municipio (solo tripleta de facturación).

La única vía pública estatal descargable limpiamente sin credenciales, WAF ni Red SARA es el **Registro de Entidades Locales (REL)** del Ministerio de Política Territorial y Memoria Democrática (`rel_municipios_espana.xls`). El REL aporta el `NUMERO_INSCRIPCION` (`01PPMMMDC`), y el DIR3 del ayuntamiento es ese número con la letra `L` delante (`L01PPMMMDC`, el código de la entidad local; las unidades que dependen de ella llevan códigos `LA…`, de otra serie). ADR 0017 dejó esa derivación como hipótesis. H6 la verificó contra DIR3 real el 2026-09-24: las fichas de unidad orgánica del directorio del Punto de Acceso General (`administracion.gob.es`), consultadas dentro de su `robots.txt`, dan a los siete municipios de la muestra —Leganés, dos fusionados, dos forales y dos con entidades locales menores— el código derivado como el de su «Ayuntamiento de …», sin ninguna discrepancia; y la regla se aplica sin discrepancia a los 8.132 municipios del volcado. La muestra, la procedencia de cada código real y los hashes de las fichas están en `specs/008-h6-territorio-skill-legal/gates/verificacion-dir3.md`.
