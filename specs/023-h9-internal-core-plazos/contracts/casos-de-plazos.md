# Tabla de casos de `plazos` (FR-082, SC-005)

Cada caso, con la fecha esperada calculada a mano y su razonamiento, independiente del código. La comprueba
`TestCasosDePlazos` (`internal/app/plazos_casos_test.go`) llamando al applet como lo llama la skill, con los municipios
por su nombre y los ficheros empotrados; los casos 19 y 20 son de cobertura por materia y los comprueba el mismo test
contra `data/cobertura.yaml`. Los días de la semana, de `cal` (research V25); los festivos locales, del BOCM grabado
(research V9); los nacionales, de las resoluciones del BOE (research S3). Los guiones de aceptación usan solo los casos
1, 2, 3, 15, 17 y 18 (research D27). Diecinueve fechas y dos coberturas: al menos once, como pide SC-005. El caso 21
lo añadió la revisión final del hito (`gates/supuestos.md`, «corrector_revision»).

Abreviaturas: L = Leganés (`28074`), M = Madrid (`28079`), T = Tordesillas (`47165`), B = Brunete, V = El Vellón,
A = Vielha e Mijaran (`25243`), en el territorio de Arán.

| # | Caso | Llamada | Fecha | Avisos | Razonamiento |
|---|---|---|---|---|---|
| 1 | Eval (a) | `reposicion 2026-09-09 L L` | 2026-10-13 | — | Un mes (art. 124.1 LPAC) desde el día siguiente (art. 30.4): vence el mismo día del mes siguiente, el viernes 9 de octubre, fiesta local de Leganés (`BOCM-20251212-34`). Se prorroga al primer hábil (art. 30.5): el 10 y el 11 son sábado y domingo y el 12 es fiesta nacional (`BOE-A-2025-21667`): martes 13. Festivos saltados: 9 (local, L) y 12 (nacional) |
| 2 | Eval (b) | `reposicion 2026-09-09 T T` | 2026-10-09 | `faltan-festivos-locales` | Un mes: viernes 9 de octubre, que no es fiesta nacional ni de Castilla y León. kitlegal no tiene los locales de Tordesillas: la fecha es la más temprana posible y lo avisa (FR-026) |
| 3 | Eval (e) | `contencioso 2026-06-15 M` | 2026-09-15 | — | Dos meses (art. 46.1 LJCA) de fecha a fecha darían el 15 de agosto; el tramo pasa por agosto, que no corre (art. 128.2 LJCA): el vencimiento pasa al mismo día del mes siguiente, el martes 15 de septiembre, que no es inhábil del art. 182.1 LOPJ en Madrid |
| 4 | Art. 30.6, festivo del interesado | `alzada 2026-09-09 L M` | 2026-10-13 | — | Un mes (art. 122.1): viernes 9 de octubre, fiesta local de Leganés y no de Madrid (`05-15`, `11-09`); con los dos municipios, inhábil (art. 30.6); 10 y 11, fin de semana; 12, nacional: martes 13. Con solo Madrid sería el 9 |
| 5 | Art. 30.6, festivo de la sede | `reposicion 2026-10-09 L M` | 2026-11-10 | — | Un mes: lunes 9 de noviembre, fiesta local de Madrid y no de Leganés; inhábil por el art. 30.6: martes 10. Con solo Leganés sería el 9 |
| 6 | Último día en sábado | `reposicion 2026-09-03 L L` | 2026-10-05 | — | Un mes: sábado 3 de octubre; prórroga (art. 30.5): domingo 4, lunes 5 |
| 7 | Cambio de año hacia 2027 | `generico 2026-12-21 L L --duracion 10 --unidad dias-habiles` | 2027-01-07 | `faltan-festivos-locales` | Diez días hábiles desde el martes 22 (art. 30.2): 22, 23, 24 (1 a 3); 25, Navidad, nacional; 26 y 27, fin de semana; 28, 29, 30, 31 (4 a 7); 1 de enero, nacional; 2 y 3, fin de semana; 4, 5 (8 y 9); 6, Epifanía, nacional; 7 (10). Leganés no tiene locales cargados de 2027: avisa |
| 8 | Municipio «no comunicado» antes de su corrección | `reposicion 2026-08-14 B B`, con la principal de 2026 sola | 2026-09-14 | `faltan-festivos-locales` | Un mes: lunes 14 de septiembre. Con la principal (`BOCM-20251212-34`) sola, Brunete es `no-comunicado`: el 14 cuenta como hábil y avisa |
| 9 | El mismo, con la corrección | `reposicion 2026-08-14 B B` | 2026-09-15 | — | La corrección `BOCM-20251229-19` da a Brunete el 15 de mayo y el 14 de septiembre: el 14 es inhábil y se prorroga al martes 15 |
| 10 | Entidad menor | `reposicion 2026-09-09 V V` | 2026-10-09 | — | Un mes: viernes 9 de octubre, fiesta de El Espartal, entidad de El Vellón, en su propia línea; los de El Vellón son el 2 y el 5 de octubre. El 9 cuenta como hábil y no hay aviso por él (FR-024) |
| 11 | Silencio | `silencio 2026-07-09 L L --duracion 3 --unidad meses` | 2026-10-14 | — | Tres meses desde la entrada (art. 21.3.b) contados desde el día siguiente: vencen el viernes 9 de octubre, fiesta local; prórroga al martes 13 (art. 30.5). El silencio se produce al día siguiente del vencimiento (arts. 24.1 y 24.4; research D8): miércoles 14 |
| 12 | Días hábiles con festivos | `generico 2026-10-07 L L --duracion 5 --unidad dias-habiles` | 2026-10-16 | — | Desde el jueves 8 (1); 9, local; 10 y 11, fin de semana; 12, nacional; 13, 14, 15 y 16 (2 a 5). Festivos saltados: 9 y 12 |
| 13 | Días naturales | `generico 2026-09-24 L L --duracion 15 --unidad dias-naturales` | 2026-10-13 | — | Quince días naturales desde el viernes 25 de septiembre: el 9 de octubre (art. 30.2, párrafo segundo), local; prórroga (art. 30.5): 10, 11, 12 inhábiles, martes 13 |
| 14 | Mes sin día equivalente | `generico 2026-01-31 L L --duracion 1 --unidad meses` | 2026-03-02 | — | Febrero de 2026 no tiene 31: vence el último día, el sábado 28 (art. 30.4); prórroga: domingo 1 de marzo, lunes 2 |
| 15 | Contencioso con el art. 183 | `contencioso 2026-10-27 M` | 2026-12-28 | `dias-del-art-183-lopj` | Dos meses de fecha a fecha: domingo 27 de diciembre, sin agosto; inhábil (art. 182.1): lunes 28, que no es fiesta en Madrid. El tramo pasa por los días del 24 de diciembre al 6 de enero (art. 183): se cuentan y avisa |
| 16 | Contencioso con el 31 de diciembre | `contencioso 2026-10-31 M` | 2026-12-31 | `dias-del-art-183-lopj` | Dos meses: jueves 31 de diciembre, que no es inhábil: el art. 182.1 LOPJ grabado, en la redacción de la LO 14/2022, no enumera el 24 ni el 31 de diciembre, y no es fiesta en Madrid. El tramo pasa por los días del art. 183: se cuentan y avisa. El plan daba el lunes 4 de enero de 2027, con `faltan-festivos-locales`, por un art. 182.1 que lo grabado no confirma (research S4; T005, `gates/supuestos.md`) |
| 17 | Año anterior a 2025 | `generico 2024-12-30 L L --duracion 3 --unidad dias-habiles` | 2025-01-03 | `faltan-festivos-nacionales-y-autonomicos`, `faltan-festivos-locales` | Martes 31 de diciembre de 2024 (1), sin festivos cargados de 2024; miércoles 1 de enero de 2025, nacional (`BOE-A-2024-21316`); jueves 2 y viernes 3 (2 y 3). Avisa de que faltan los nacionales, los autonómicos y los locales de 2024 |
| 18 | Mes que cruza a 2027 sin festivo decisivo | `reposicion 2026-12-14 L L` | 2027-01-14 | `faltan-festivos-locales` | Un mes: jueves 14 de enero de 2027, que no es fiesta nacional ni de Madrid; Leganés no tiene locales de 2027: avisa (US2.2) |
| 19 | Eval (c), cobertura | — | — | — | La aplicación de los tributos y su revisión las regula aparte la DA 1ª.2.a LPAC y las rige la LGT (`BOE-A-2003-23186`, art. 97): `data/cobertura.yaml` lleva la entrada con esa norma |
| 20 | Eval (d), cobertura | — | — | — | Los sancionadores de tráfico y seguridad vial los regula aparte la DA 1ª.2.c LPAC y los rige el texto refundido de Tráfico (art. 1): `data/cobertura.yaml` lleva la entrada con el identificador que da la búsqueda grabada |
| 21 | Fiesta sustituida en una parte | `reposicion 2025-11-26 A A` | 2025-12-26 | `faltan-festivos-locales`, `festivo-de-una-parte` | Un mes: viernes 26 de diciembre de 2025, fiesta de Cataluña que la nota 2 de `BOE-A-2024-21316` sustituye en el territorio de Arán por el 17 de junio: no es fiesta de toda la comunidad y cuenta como hábil (FR-011), con el aviso de una parte. En Arán es el último día; en el resto de Cataluña la fecha real sería el lunes 29, y la del aviso es la más temprana. kitlegal no tiene los locales de Cataluña |

Los casos 1 y 4 comparten la fecha y prueban cosas distintas: el 1, la prórroga con un solo municipio; el 4, el art.
30.6. Con los casos 4 y 5 cae el mutante de un solo municipio sea cual sea el que deje, y con el 3, el de agosto
contado, que daría el lunes 17 de agosto (FR-089). El caso 8 lo compone
el test con los festivos que deriva `internal/source/bocm` de la resolución principal de 2026 sola (research D23,
D30).

**Festivo de una parte** (FR-011): `TestFestivoDeUnaParte` (`internal/app/plazos_test.go`) toma de los ficheros
empotrados el primer festivo con `parte` que cae en día laborable, y un municipio de su comunidad de
`data/territorio/`; con un plazo genérico de un día hábil desde la víspera, la fecha es ese día (cuenta como hábil) y
lleva el aviso `festivo-de-una-parte` con el día y la parte. Si lo grabado no trae ninguno, el test falla: el hito lo
da por hecho (S1).
