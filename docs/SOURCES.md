# Fuentes

Las fuentes públicas externas que consulta kitlegal: de dónde salen los datos, con qué licencia, bajo qué términos de
uso y a qué ritmo se piden. Cualquier cambio en este fichero es una decisión humana (constitución, capa 3): cada fila la
revisa una persona, que anota en «Revisado» el día en que leyó los términos de uso y el `robots.txt`. Mientras nadie lo
ha hecho, «Revisado» dice `pendiente` y la fila es una propuesta: ninguno de sus datos está comprobado. Si los términos
de uso o el `robots.txt` prohíben el acceso automatizado, la fuente no se usa y el hito que la introduce se detiene.

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
