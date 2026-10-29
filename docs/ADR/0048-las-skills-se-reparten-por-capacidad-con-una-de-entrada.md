# 0048 · Las skills se reparten por capacidad, con una de entrada

- **Estado**: propuesta (2026-10-29); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-10-29
- **Hito**: H32. Decide lo que el ADR 0044 deja a H32 (punto 7): los nombres de las skills por capacidad, y con ellos
  el reparto y el molde de un paquete de materia (puntos 3 y 5). No cambia la regla de `references/` del ADR 0046
  (plano, a un nivel de `SKILL.md`): una referencia a mano puede llevar regiones generadas entre marcas. Cambia `data`
  de `kitlegal skills install` (dos estados y el campo `aviso`), sin ningún código de salida nuevo (ADR 0023).

## Contexto y problema

El ADR 0044 hace de kitlegal un componente legal general: una base por capacidades, lo administrativo como su primer
paquete de materia y ningún caso especial para una materia en el código, en `data/` ni en el protocolo de una skill
(principio X). Las tres skills de `c991808` —`boe-legislacion`, `legal-core` y `jurisprudencia`— no lo eran: dos se
llamaban por su fuente o por un «núcleo», `legal-core` llevaba lo administrativo en su cuerpo y en sus referencias, y
ninguna sabía qué capacidad resolvía una pregunta de otra materia. El punto 7 deja sus nombres a H32, con su ADR.

Al repartirlas había que no arrastrar los tres defectos que quedaban tras H31.1:

- **`cuenta_su_proceso`, por la línea de la comprobación.** En `brazo-P80` (`evidencias/sondeos-20261028/`), las 4
  marcadas de 80 dan el resultado de `kitlegal graph check`; en el job de `cd21c7b` (ejecución 40700995075), 17 de las
  21 marcadas de 120 de `claude-sonnet-5-5`, y en el de `0f9ef2f` (40682099061), las 16 de 120. Se concentra en la eval
  05: 6 de sus 12 respuestas en las dos ejecuciones. Antes de H31 eran el 1,8 % y el 2,8 %, con un cuerpo que decía que
  un bloque sin `version-obsoleta` no da nada a la respuesta (`e1bfb77`, `skills/boe-legislacion/SKILL.md`, líneas 215
  a 220); H31 lo llevó a una referencia que solo se lee con un `version-obsoleta`.
- **`afirma_lo_no_leido`, por reglas de memoria y materia pegada a una remisión.** Las 2 marcadas de `brazo-P80`
  explican «íntegros» con una regla del rendimiento neto y describen el art. 7 de la Ley 35/2006 sin su número; el
  paso 5 lo permitía («o el número, o la materia»; «Puedes avisar de que una materia se regula en otra parte»).
- **Dos formas para la misma cita.** `boe-legislacion` citaba `[BOE-A-2015-10565, bloque daprimera]`; la línea
  `⚠ NO CUBIERTO:` de `legal-core` pedía `(BOE-A-2015-10565)`, la cobertura por materia escribía
  `(BOE-A-2015-10565, bloque daprimera)` y el paso 6 de `legal-core` mandaba citar «como ella cita». Una sesión mezcló
  las dos y dejó fuera la norma que lo rige: `cobertura_sin_declarar` dio 1 de 36 respuestas en las evals 07, 08 y 12,
  y tumbó el job (ejecución 40682099061).

Y sin sitio: los cuerpos tenían 10 995, 10 944 y 10 859 caracteres, con un máximo de 11 000 (ADR 0046).

## Opciones consideradas

1. **Un prefijo común** (`kitlegal-normativa`…). Rechazada: repite la marca y obliga a renombrar `jurisprudencia`.
2. **`plazos` y `territorio` por separado.** Rechazada: todo plazo con festivos locales resuelve antes sus municipios
   con `territorio`, y cada pregunta de plazos cargaría dos skills con el mismo paso, el defecto de una skill demasiado
   estrecha («Design coherent units»).
3. **Un fichero que declare cada paquete, o un generador por materia en Go.** Rechazadas: el primero es una pieza que
   el ADR 0044 (punto 5) no lista; el segundo, un caso especial de materia en el código (principio X).
4. **Una skill de entrada y una por capacidad con herramienta, por su capacidad, y cada materia como paquete.**
   Elegida.

## Decisión

1. **Cuatro skills.** Una de entrada, que no declara applets ni referencias, y tres de capacidad; cada applet de
   capacidad está en exactamente una:

   | Skill | Clase | Capacidad | `kitlegal-applets` |
   |---|---|---|---|
   | `consulta-juridica` | entrada | ninguna: qué tipo de pregunta es, qué capacidad y qué skill la resuelven, qué cubre la base y, en prosa, lo que no cubre | — |
   | `normativa` | capacidad | normativa | `boe graph` |
   | `jurisprudencia` | capacidad | jurisprudencia | `cita` |
   | `plazos-y-territorio` | capacidad | plazos y territorio | `territorio plazos` |

2. **Los nombres**: la capacidad, en español, en minúsculas y con guiones, como sintagma nominal; la de entrada, por lo
   que hace, la consulta. Ninguno contiene `boe`, `cendoj`, `eur`, `anthropic` ni `claude`: H34 añade EUR-Lex a
   `normativa` sin renombrarla. `jurisprudencia` conserva el suyo, que ya es el de su capacidad.
3. **El reparto.** Cada skill de capacidad responde sola: lleva el protocolo, las reglas invariantes y las formas fijas
   que dan sus respuestas, y ningún cuerpo ni referencia suya remite a otra skill. La de entrada no lleva ninguna regla
   que otra necesite para responder. Lo que vale para cualquier materia —la jerarquía normativa, el método— queda en
   una skill de capacidad o en la de entrada; lo de lo administrativo pasa a su paquete.
4. **Los nombres retirados**, con las skills que heredan sus preguntas. El binario los lleva en una lista, y
   `kitlegal skills install` retira la skill del nombre retirado cuyas huellas coinciden con su manifiesto y avisa,
   sin tocarla, de la que no:

   | Nombre retirado | Lo hereda |
   |---|---|
   | `boe-legislacion` | `normativa` |
   | `legal-core` | `plazos-y-territorio` |

5. **El molde de un paquete de materia**, y nada más:
   - su nombre, el valor `vertical: <nombre>` de sus normas en `data/normas.yaml`, un adjetivo («lo <nombre>»); no hay
     fichero que declare el paquete;
   - una referencia por skill de capacidad que la necesite, `skills/<skill>/references/<nombre>.md`, escrita a mano, con
     las tablas que salen de `data/` generadas entre marcas por `make skills-sync`;
   - sus filas de cobertura, las de `data/cobertura.yaml` cuyo precepto es de una norma suya;
   - sus evals, en la carpeta de la skill que las hereda.

   El enlace a cada referencia de un paquete lo escribe `make skills-sync` entre marcas en el `SKILL.md` que la recibe,
   con la condición de su clase («Si la pregunta es de lo <nombre>, lee …»), y la lista de paquetes de la skill de
   entrada, igual: un paquete no edita ningún `SKILL.md` a mano. El primero es lo administrativo: diez normas (LPAC,
   LRJSP, LRJPAC, LTAIBG, LRBRL, LCSP, LGS, LGP, TRLRHL y LJCA), una referencia, `administrativo.md` de
   `plazos-y-territorio`, y las nueve filas de cobertura. Lo prueba una vertical de ensayo de `testdata/` que entra sin
   un `.go` ni un cambio en un `SKILL.md` fuera de las marcas.

**De `c991808` a la entrega:**

| `c991808` | Entrega | Qué pasa |
|---|---|---|
| — | `consulta-juridica` | nueva: la skill de entrada |
| `boe-legislacion` | `normativa` | renombrada; sus ejemplos pasan a la Constitución, que ningún paquete marca |
| `jurisprudencia` | `jurisprudencia` | mismo nombre; sin la remisión a otra skill |
| `legal-core` | `plazos-y-territorio` | renombrada; la cobertura por materia, los recursos y plazos, las leyes vertebrales y lo que solo pide el procedimiento común pasan al paquete de lo administrativo |

## Consecuencias

- Una materia nueva es un paquete de ficheros, no una skill ni código. Lo que un paquete necesite y la base no tenga es
  un hito de la base (ADR 0044).
- Un bloque se cita de una sola forma en todas las skills, `<forma legible> [<identificador>, bloque <id>]`, también en
  las formas fijas y en las tablas generadas; y una forma fija que dos skills comparten es la misma línea en las dos.
  Lo comprueban `forma-de-cita` y `formas-compartidas` en `make ci`.
- Quien instaló con un nombre retirado y vuelve a instalar no se queda con dos protocolos que se activan a la vez: lo
  intacto se retira, y lo que tocó la persona sigue y se avisa con su ruta, con salida 0. Un nombre retirado pedido en
  la invocación sale con 2, como toda skill que el binario no empotra.
- Las evals pasan con su historia a la carpeta de la skill que hereda su pregunta, y `consulta-juridica` trae las
  suyas. Los ADR anteriores conservan los nombres de su fecha.

## Validación

`make ci`: `TestSkillsDelRepositorio` (los presupuestos y la forma de las cuatro skills, `forma-de-cita` y
`formas-compartidas`), `TestVerticalDeEnsayo`, `TestSkillsRetiradas` y los guiones `h32-*` de la suite de aceptación.
El job de evals de la propuesta de cambio mide las cuatro skills, y una persona decide la fusión con un sondeo de 280
sesiones y los límites del ADR 0047.
