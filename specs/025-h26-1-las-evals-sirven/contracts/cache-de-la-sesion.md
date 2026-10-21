# Contrato: lo que sirve la caché de una sesión de evals

Quién lo consume: la skill dentro de cada sesión del job, del sondeo y de la reconstrucción de la medida del juez, que
lee del BOE con `kitlegal boe …` o con sus herramientas y recibe 0 si la caché lo sirve y 4 o 5 si no; y la persona que
lee el informe, que ve en `fuera_de_lo_grabado` lo que no se sirvió. Requisitos: FR-001 a FR-006, FR-010, FR-011,
FR-013; SC-003 a SC-005, SC-011, SC-018.

## 1. Qué consultas se preparan

La caché de una sesión de la eval *E* de una skill se prepara, como hoy, con `Preparar` sobre `UnionDeGrabaciones()`,
en proceso y sin red, con estas consultas, en este orden y sin repetir ninguna invocación:

1. **Las necesarias de las evals de la skill** (`ConsultasNecesarias`): los comandos esperados, el `indice` y los
   `metadatos` de las normas de sus comandos y citas, el bloque de cada cita y, desde este hito, cada bloque de sus
   `lecturas`; cada bloque, sin fecha y con `--fecha 9999-12-31`.
2. **Lo que la skill lleva a leer** (`anotarLosDatos(datosDelRepositorio)`): cada bloque de `data/plazos.yaml` y de
   `data/cobertura.yaml`, sin fecha y con `--fecha 9999-12-31`, y el `indice` y los `metadatos` de cada norma de esos
   dos ficheros y de `data/normas.yaml`. Es lo mismo para las tres skills.
3. **Menos el `sin_servir` de *E*** (`sinLoQueNoSeSirve`): las que consultan un bloque de su `sin_servir` (las dos
   lecturas, sin fecha y con `9999-12-31`) o la consulta de norma que nombra, aunque las pidan otra eval, `lecturas` o
   `data/`. Las sesiones de las demás evals las tienen.

Lo que no se prepara, la sesión no lo tiene: con `--offline`, 4; sin `--offline`, en el job, 4 o 5 (research V3, S1).
La lectura de un bloque con `--fecha 9999-12-31` deja sus versiones, de las que se sirve después la del bloque en
cualquier fecha (research V3).

La base de la medida del juez (`baseVigente`) prepara 1, sin las `lecturas`, y 2, sin 3: es una para todas las
sesiones de la skill, y solo la tienen `boe-legislacion` y `jurisprudencia`, sin ninguna eval con `sin_servir`
(research D4). Las `lecturas` son lo que las sesiones de los informes versionados leyeron sin tenerlo, con 4 o 5, y la
base no las sirve: toda invocación de `boe` de la sesión de un caso etiquetado que terminó así en su informe sigue
terminando con 4 o 5 sobre la base (`TestBaseDeLaReconstruccion`).

## 2. Lo que se sirve hoy con el `data/` del repositorio

17 bloques y 19 normas: 72 consultas de lo que la skill lleva a leer, todas grabadas (research V1). Las de `lecturas`
son 8 más en las sesiones de `boe-legislacion`, con sus cuatro bloques grabados por `grabar_datos` (research V2). Las de la
11 y la 12 de `legal-core` quitan 2 cada una de sus sesiones.

## 3. Si algo de 1 o 2 no está grabado

`Preparar` da una `Falta` por cada consulta que no termina en 0, como hoy, con una línea por cada origen, y la sesión no
se abre (`PrepararSesion` devuelve las faltas; el repartidor la cuenta como no preparada). Ejemplo con `a128` cambiado a
`a129` en una copia de `data/plazos.yaml`:

```text
data/plazos.yaml: lo que la skill lleva a leer boe articulo BOE-A-1998-16718 a129: «boe articulo BOE-A-1998-16718 a129» terminó con código 1: … no hay grabación de GET https://www.boe.es/datosabiertos/api/legislacion-consolidada/id/BOE-A-1998-16718/texto/bloque/a129 …
```

`TestEvalsDelRepositorio` sale en rojo con cualquier consulta de 1 o 2 sin grabar (FR-003).

## 4. Invocaciones por pregunta y tamaño

- La skill las pide en la sesión: `legal-core` lee `daprimera` una vez por pregunta de una materia regulada aparte (paso
  7) y cada precepto de `data.preceptos`, de uno a tres, en la fecha de inicio (paso 10); `boe-legislacion` lee el índice
  de la norma de la pregunta y de uno a tres bloques. La caché las sirve sin red, sin pedir nada a nadie.
- El arnés prepara una vez por sesión: 72 consultas de `data/` más las de las evals de su skill, en proceso, del orden de
  un milisegundo cada una (research V1: tres preparaciones con 150 consultas en 0,31 s).
- El tamaño crece con `data/`, no con el uso: cada sesión empieza de cero, con una caché nueva en su directorio, y no
  hereda nada de otra. Los meses de uso de quien tiene kitlegal instalado, con sus cientos de normas y miles de bloques en
  `~/.cache/kitlegal/`, no entran en ninguna sesión. Con los preceptos que añada H9.1, unas decenas de consultas más.
- Deja de darse cuando el precepto sale de `data/` o la lectura sale de `lecturas`; y `sin_servir`, cuando la eval deja
  de llevarlo.
