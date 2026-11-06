# Los municipios

El detalle del territorio de un municipio (pasos 1 a 4 de `SKILL.md`): qué hacer si falta el municipio, qué trae
`data`, la plantilla con que se presenta, qué dice cada aspecto de la cobertura y qué hacer cuando `territorio` o
`plazos` no salen con 0. La forma de las tres líneas de cobertura es la de «Cómo se presenta el territorio» del cuerpo.

## El municipio que falta

- Toma el municipio de la conversación, por su nombre o su código INE. Si no lo dice, pregúntalo y espera la
  respuesta: no lo deduzcas de otros datos ni sigas con un municipio de ejemplo, tampoco con el de quien usa el kit.
- Si solo se nombra una provincia o una comunidad, pregunta también por el municipio: `kitlegal territorio` resuelve
  municipios.
- Pasa el nombre o el código tal como los dio la persona; no conviertas de memoria un nombre en un código ni al revés.

## Lo que trae `data`

- Con el código 0, `data` trae siempre `municipio`, `codigo_ine`, `provincia`, `comunidad`, `dir3`, `regimen`,
  `boletines` y `cobertura`, y cada dato lleva su `source`.
- Ningún dato de territorio —comunidad, provincia, régimen, DIR3, boletines— se da por sabido ni se escribe de
  memoria: todos salen de `data` en esta conversación. Ningún municipio ni ninguna comunidad tiene un trato propio.
- `territorio` y `plazos` responden con lo que lleva dentro el binario, sin consultar ninguna fuente.

## La plantilla del territorio

Cada dato sale de `data`, con el nombre y el código tal como los da. Si `regimen.valor` es `foral`, dilo (regla 4), y
di de cuándo son los datos: `fecha_consulta`.

```text
Municipio: <nombre>, código INE <codigo> (dígito de control <digito_de_control>)
Provincia: <nombre de la provincia>
Comunidad o ciudad autónoma: <nombre de la comunidad>, régimen <común o foral, según regimen.valor>
DIR3 del ayuntamiento: <código, o «no verificado»>
Boletines: <código> — <nombre> (<nivel>), uno por cada entrada de `boletines`
boletin_autonomico: <valor>
boletin_provincial: <valor>
dir3: <valor>
```

## Qué dice cada aspecto de la cobertura

| Aspecto | Valores | Qué dice |
|---|---|---|
| `boletin_autonomico` | `configurado`, `no-configurado` | si kitlegal tiene declarado el boletín de la comunidad |
| `boletin_provincial` | `configurado`, `no-configurado` | si kitlegal tiene declarado el de la provincia |
| `dir3` | `verificado`, `no-verificado` | si el DIR3 del ayuntamiento está en la correspondencia verificada |

- Lo `no-configurado` o `no-verificado` dilo con sus palabras: **no** significa que no exista (regla 5).
- No nombres ningún boletín que el applet no haya devuelto, aunque creas conocerlo; la clase de boletín de la
  jerarquía normativa no suple un nivel `no-configurado`. El nivel autonómico es el de la `comunidad` de `data`; los
  boletines concretos, solo los de `boletines`.
- Con `dir3` en `no-verificado`, no compongas el DIR3 a partir del código INE: la plantilla dice «no verificado».

## Cuando no sale con 0

- **Código 2 (`argumentos`) con candidatos**: `data.mensaje` los da todos, cada uno como `<código INE> <nombre>
  (<provincia>)`. Ofrécelos, pregunta a cuál se refiere la persona y vuelve al paso 2 con el código que elija.
- **Código 2 sin candidatos**: la consulta no forma un código INE válido o su dígito de control no es el oficial. Díselo
  y pide el nombre o el código; no corrijas tú el código.
- **Código 3 (`no-encontrado`)**: no está en la relación de municipios que lleva kitlegal. Dilo así, sin concluir que
  no exista, y pide otra forma del nombre o su código INE.
- En `plazos`, el 2 o el 3 de un municipio se tratan igual, con los candidatos en `data.mensaje` tras el municipio. El
  3 con un `data.mensaje` que empieza por «el plazo de tipo» no es de un municipio: es el plazo sin calcular del paso 9.
- Con cualquier otro código o `clase`, di que no se pudo resolver y no suplas sus datos.
