# 0042 · `plazos` no da la fecha de un plazo con preceptos que no regían el día en que empieza

- **Estado**: aceptada (2026-10-20)
- **Fecha**: 2026-10-20
- **Hito**: sin hito; un cambio puntual de `plazos`, de `data/plazos.yaml` y de `legal-core`, tras el ADR 0041.
  Sustituye, del punto 3 del ADR 0041, la frase de la respuesta «la fecha de `plazos` cuenta con la duración de la
  redacción de hoy»: desde aquí, `plazos` solo da una fecha si cada apartado con que la cuenta decía lo mismo el día
  de inicio. El resto del ADR 0041 sigue vigente: la regla, la fecha de la notificación como la del acto, la lectura
  fechada de cada precepto y que calcular con una redacción anterior queda en el backlog. No toca el contrato
  `Applet` (ADR 0005), ni el sobre (ADR 0006), ni la tabla de resultados (ADR 0023), ni el umbral de ninguna clase
  (ADR 0038).

## Contexto y problema

El ADR 0041 decide que un plazo se rige por la redacción vigente el día en que empieza, y `legal-core` v1.1 lee cada
precepto en esa fecha. `kitlegal plazos calcular` no la mira: cuenta con la duración de `data/plazos.yaml`, escrita con
la redacción de hoy, sea cual sea el inicio. Con el binario de `main` del 2026-10-20:

- `kitlegal plazos calcular reposicion 2014-06-02 Leganés Leganés` da `ok: true` y el 2014-07-02 con el art. 124
  LPAC, que en 2014 no existía, y solo avisa de que faltan festivos. `kitlegal boe articulo BOE-A-2015-10565 a124
  --fecha 2014-06-01` termina con 3: el bloque «no tenía redacción vigente». Lo mismo dan a122, a21, a24 y a30 de la
  LPAC y a20 y a24 de la LTAIBG antes de su vigencia.
- El cómputo también tiene preceptos, y `data/plazos.yaml` no los nombra. El del recurso contencioso cuenta como
  inhábiles los días del art. 182.1 LOPJ en la redacción de la LO 14/2022, que no enumera el 24 ni el 31 de
  diciembre; el aviso de los días del art. 183 LOPJ cita la misma redacción. Antes del 2022-12-23 los dos decían otra
  cosa, y `plazos` contaba un contencioso de 2021 con la de hoy.

Hoy no afecta a ninguna notificación reciente: ningún precepto ha cambiado desde 2022. Afecta a cualquier pregunta con
una fecha de inicio antigua, y afectará a todas cuando cambie uno de esos preceptos.

## Lo que dice la ley

Leído con kitlegal el 2026-10-20:

- **La regla del ADR 0041** —la disposición transitoria tercera de la Ley 39/2015 [BOE-A-2015-10565, bloque
  dttercera] y el art. 2.3 del Código Civil [BOE-A-1889-4763, bloque art2]— ata el régimen de un recurso al día del
  acto y el de un procedimiento al de su inicio, salvo lo que diga el régimen transitorio de la norma que cambió el
  precepto. La disposición transitoria tercera de la Ley 29/1998 [BOE-A-1998-16718, bloque dttercera] es el ejemplo
  de que ese régimen puede llevar la redacción nueva a plazos que ya corrían.
- **El art. 182.1 LOPJ** [BOE-A-1985-12666, bloque acientoochentaydos], leído con `--fecha 2022-12-22`, es la
  redacción de la LO 19/2003, vigente desde el 2004-01-15: «Son inhábiles a efectos procesales los sábados y domingos,
  los días 24 y 31 de diciembre, los días de fiesta nacional…». La de hoy, de la LO 14/2022 [BOE-A-2022-21800],
  vigente desde el 2022-12-23, ya no nombra el 24 ni el 31. **El art. 183** [bloque acientoochentaytres] cambió el
  mismo día: antes declaraba inhábil agosto y no los días del 24 de diciembre al 6 de enero.
- **El art. 128 LJCA** [BOE-A-1998-16718, bloque a128] tiene dos redacciones, la de 1998 y la de la Ley 13/2009
  [BOE-A-2009-17493], vigente desde el 2010-05-04, cuya nota dice «Se modifica el apartado 1». El apartado 2, el de
  agosto, que es el que funda el plazo, dice lo mismo en las dos.
- **El art. 5.1 del Código Civil** [BOE-A-1889-4763, bloque art5], «de fecha a fecha», tiene su redacción desde el
  Decreto 1836/1974, vigente desde el 1974-07-29; el **art. 185 LOPJ** [bloque acientoochentaycinco], la de 1985.

De ahí salen dos cosas. Lo que funda una fecha es el apartado, no el bloque: un cambio de otro apartado no la
cambia. Y si el apartado con que `plazos` cuenta no regía el día de inicio, `plazos` no sabe qué regía: la redacción
de entonces, o la de hoy si el régimen transitorio de la norma que la cambió la llevó a ese plazo. No puede elegir
ninguna de las dos sin decirlo.

## Opciones consideradas

1. **Dar la fecha con un aviso** (fila «Hecho», código 0). Rechazada: el aviso de unos festivos que faltan acompaña a
   una fecha que es la más temprana posible y sirve; una fecha contada con un precepto que no regía no es cota de
   nada, y dada en `data.fecha` es la que una skill o una persona acabaría usando.
2. **Comparar la redacción del bloque entero.** Rechazada: el art. 128 LJCA rechazaría un contencioso de 2008, cuyo
   apartado 2 era el mismo, y cualquier reforma de otro apartado de un precepto rechazaría todos los plazos
   anteriores a ella.
3. **Comprobar solo los preceptos de cada tipo**, los que nombra hoy el fichero. Rechazada: dejaría fuera los del
   cómputo, que fundan la fecha igual —un contencioso de 2021, o un acceso a la información de 2015, con la LTAIBG ya
   vigente y sin el art. 30 LPAC—.
4. **Comprobar cada apartado con que se cuenta la fecha, del tipo y del cómputo, y no dar la fecha si alguno no regía
   el día de inicio.**

## Decisión

Opción 4.

1. **Cada precepto de `data/plazos.yaml` lleva `vigente_desde`**: el primer día, `AAAA-MM-DD`, en que su apartado
   dice lo que dice en la última redacción grabada, con la que está escrito el fichero. El texto de un apartado va de
   la línea que empieza por su número a la del apartado siguiente; sin número, como el art. 183 LOPJ, es el bloque
   entero. Las notas del BOE, al final del bloque, van con el último apartado: una nota nueva lo da por cambiado, que
   es la lectura conservadora.
2. **El dato sale del BOE grabado.** `TestPreceptosDeLosPlazos` lee las redacciones de cada bloque con
   `boe articulo --fecha`, de la última hacia atrás —lo mismo que lee la skill— y exige que `vigente_desde` sea la
   fecha de vigencia de la más antigua de las redacciones seguidas en que el apartado dice lo mismo que hoy. Las
   grabaciones ya traen todas las versiones de cada bloque, y no hizo falta grabar nada.
3. **`data/plazos.yaml` nombra los preceptos de cada cómputo**: los apartados 2 a 6 del art. 30 LPAC en el de la LPAC,
   y los arts. 182.1, 183, 185.1 y 185.2 LOPJ y 5.1 del Código Civil en el contencioso. `plazos.LeerTipos` exige los
   dos cómputos y en cada precepto un `vigente_desde` que sea un día.
4. **`plazos calcular` no da la fecha de un plazo cuyo inicio es anterior al `vigente_desde` de alguno de sus
   preceptos** —los del tipo, el de la unidad en el genérico y los del cómputo—. Sale con 3, `no-encontrado`: lo
   pedido, el plazo de ese día, no está en los datos del applet, como no lo está la redacción de un bloque en un día
   en que no tenía ninguna. Es la fila del ADR 0023, y no hace falta otra. El mensaje dice el plazo, su inicio y por
   qué no se calcula en la primera línea, y en cada una de las siguientes un precepto, con su norma, su bloque, su
   apartado y su `vigente_desde`, en orden y sin repetir. Va después de los argumentos (2) y de los municipios, y no
   cambia nada de lo que se calcula. La fecha que se compara es la de inicio que se pasa, como en el punto 2 del ADR
   0041.
5. **`legal-core` v1.2 lo traslada sin fecha.** Con ese 3, la respuesta no lleva la línea `⚠ PLAZO:` ni ninguna fecha
   de vencimiento, y lleva una línea de forma fija, `⚠ PLAZO SIN CALCULAR:`, con el tipo, el inicio y los preceptos del
   mensaje; no nombra de memoria la norma que regía entonces. Y la frase del punto 3 del ADR 0041 cambia: si la lectura
   fechada de un precepto de `data.preceptos` trae `redaccion_posterior`, `plazos` ya ha comprobado que el apartado con
   que cuenta decía lo mismo ese día, así que lo que cambió es otra parte del bloque. Solo la norma del procedimiento
   que se lee para el silencio, que `plazos` no comprueba, sigue con la frase del régimen transitorio.
6. **Una eval lo mide sin modelo.** `10-plazo-reposicion-anterior-a-la-lpac.yaml`, con una notificación de 2014,
   espera `plazos calcular` y la línea `⚠ PLAZO SIN CALCULAR:` con el identificador de la LPAC, sin ninguna línea
   `⚠ PLAZO:`. Su comando `calcular` lleva `sin_calcular: true`, el único con el que una orden que termina con 3 lo
   cumple. Su clave en el esquema es `plazo_sin_calcular`, y con ella `legal-core` gana el umbral
   `vigencia_sin_declarar:<modelo>:<modo>`, con 0 y decidiendo, como `cobertura_sin_declarar`. Ningún umbral se relaja.

## Consecuencias

- Un plazo con un inicio anterior a la LPAC (2016-10-02), o un contencioso anterior a la LO 14/2022 (2022-12-23),
  deja de recibir una fecha contada con preceptos que no regían. Recibe un 3 que nombra cada uno.
- El dato es el de las grabaciones. Si el BOE publica una redacción nueva de uno de esos apartados, `plazos` no lo sabe
  hasta que se graben otra vez; entonces `TestPreceptosDeLosPlazos` falla hasta que `data/plazos.yaml` diga la
  duración y el `vigente_desde` nuevos.
- Si el régimen transitorio de una reforma lleva la redacción de hoy a plazos que ya corrían, `plazos` no da una
  fecha que podría ser buena. Lo asume: no lee ese régimen, y la respuesta dice que lo decide él.
- Calcular con una redacción anterior, o leer el régimen transitorio de una norma modificadora, sigue en el backlog
  (ADR 0041). Pediría en `data/plazos.yaml` una entrada por redacción.
