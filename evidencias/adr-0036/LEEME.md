# Evidencia del ADR 0036

El texto de una sentencia que trae una persona (ADR 0036, «Enmienda», 2026-10-06). Es la entrada de la eval de H23
en la que la persona pega el texto de una sentencia: un run ni lo descarga ni lo escribe de memoria. Como todo
`evidencias/`, un run del workflow no la escribe.

No la escribió el paso `grabar_datos`, y el binario no la pide: kitlegal no consulta el CENDOJ (ADR 0036, enmienda
del 2026-10-07). Es también la ficha con la que se prueba `cita cotejar`. `manifiesto.json` da la huella y el tamaño de cada fichero.

| Fichero | Qué es |
|---|---|
| `ecli-es-ts-2023-3144-fragmento.txt` | El encabezamiento y el fallo de la sentencia 1088/2023, de 4 de julio, de la Sala de lo Civil del Tribunal Supremo (ROJ `STS 3144/2023`, `ECLI:ES:TS:2023:3144`): la que `docs/JURISPRUDENCIA.md` §3 resuelve por sus tres vías |

## Cómo se hizo

1. **El documento.** Jorge lo descargó del buscador del CENDOJ con su navegador el 2026-10-06: un PDF de nueve
   páginas. No se versiona: el manifiesto guarda su huella.
2. **El fragmento.** Del texto del PDF, la ficha del encabezamiento hasta «Tipo de Resolución», el nombre del
   tribunal, la sala y el número de la sentencia, la fecha y el fallo entero. Lo omitido —los antecedentes y los
   fundamentos— va marcado con `[…]`. No lleva «Resoluciones del caso», que nombra otras resoluciones.
3. **Lo que cambia respecto del PDF.** Las líneas de cada párrafo, que el PDF parte por su maquetación, van
   unidas. Nada más: ni una palabra, ni la numeración del fallo, que salta del 2.º al 4.º en el original. Se
   comprobó sin modelo que cada línea del fragmento está en el texto del PDF, sin contar los blancos.

## Por qué solo esto

Con las grabaciones de los tests, es lo único del CENDOJ que hay en el repositorio, y se mantiene en lo mínimo
(ADR 0036, «En contra, y asumido»). El nombre del fichero lleva el ECLI y no «STS 1088/2023»: con ese número hay
dos sentencias, la de número de resolución 1088/2023, que es esta, y la de ROJ `STS 1088/2023`, que es otra, de 9
de febrero de 2023. Al pedir «la STS 1088/2023» llegó primero la segunda.
