# T015 — sin marcar: la entrega es de una persona y el bucle no puede llegar a su pausa (intento 1)

**Estado**: `[ ]`. El árbol no cambia fuera de `gates/`: ni `testdata/evals/` ni `schemas/normas.yaml.json`. `make ci`
sigue en verde sobre `aa7af30` (T014). Hace falta una persona **con red**, fuera del bucle: abajo, «Qué hace la
persona».

**Intentos 2 y 3**: si `testdata/evals/grabaciones.json` sigue con sus once entradas y `schemas/normas.yaml.json` sin
cambios (`rtk proxy git status --porcelain -- testdata/evals schemas/normas.yaml.json` vacío), no hay nada que hacer
ni que reanalizar: añadid una línea al final de esta nota con el intento y dejad la tarea `[ ]`. Si la persona ya dejó
el material en el árbol, seguid la instrucción «si ya estaba hecha y en verde, limítate a marcarla»: `make ci` en
primer plano y marcar `[X]` en el mismo turno; el commit disparará la pausa de datos, donde se revisan las dos piezas.

## Por qué el ejecutor no puede cerrarla

La mecánica es la de T001 (ver `gates/tarea-T001.md`): `gate_humano_datos` solo se dispara si `clasificar_datos` ve
cambios bajo `testdata/` o `schemas/` en `git diff <base> HEAD`, es decir, **ya commiteados**, y `estado_cierre` solo
commitea con `make ci` en verde. T001 salió porque su parte de ejecutor, un esquema nuevo, no la leía nadie y dejaba
`make ci` en verde. **En T015 las dos piezas tienen ya un lector que las pone en rojo sin las grabaciones**:

| Pieza | Control que la lee | Sin las grabaciones |
|---|---|---|
| Entradas nuevas del manifiesto | `TestIdentificadoresDeLasNormas` (`internal/evals/grabaciones_test.go:301`; corre en `make test` y en `skills-check`): resuelve **toda** entrada del manifiesto reproduciendo `boe buscar` sobre la unión de grabaciones de H4 y H5 | rojo: «no hay grabación de GET …: falta el fichero» |
| Valores nuevos del `enum` de `rango` | `TestEsquemaDeNormas/rangos-grabados` (`internal/skills/normas_test.go:375`): igualdad con el conjunto de rangos de todas las búsquedas grabadas | rojo: el `enum` tiene un valor que ninguna grabación da |

Tampoco vale el camino de H5.1/T008, cuya entrada nueva se resolvía con una búsqueda **ya grabada**: el `enum` actual
es, por ese mismo control, exactamente el conjunto de rangos grabados, y no tiene «Ley Orgánica»; luego ninguna
búsqueda grabada da la LOPJ ni la LOPDGDD, y sus entradas no pueden resolverse sin grabar. Y el ejecutor no tiene
ninguna vía legítima para grabar: ni red ni variable de grabación ni `scripts/grabar-evals.sh` (FR-073, «Cero red en
todo el bucle»), ni memoria («Nada escrito de memoria»).

Conclusión: cualquier cambio del ejecutor en las rutas de T015 deja `make ci` en rojo, no se commitea, y la pausa en la
que la persona debía grabar (plan, paso 15; research D29) no llega nunca. El bucle gastará los intentos 2 y 3 y se
detendrá en `siguiente_tarea` («intentos agotados»). No es un problema de rutas: redelimitar la línea no lo cambia,
porque falta material que solo una persona puede obtener.

### Sonda (este intento)

Sobre un clon desechable de `aa7af30` en `/tmp/kitlegal-t015-sonda`, con dos mutaciones **neutras** (sin ningún dato
legal):

- manifiesto con una entrada más, `{"busqueda": "sonda de una búsqueda sin grabar", "titulo_empieza_por": "Sonda,", …}`:
  `TestIdentificadoresDeLasNormas` **FAIL** — `entrada 12 ("Sonda,"): «boe buscar sonda de una búsqueda sin grabar»
  terminó con código 1: … no hay grabación de GET https://www.boe.es/datosabiertos/api/legislacion-consolidada?… falta
  el fichero …`; `TestManifiestoDeGrabaciones` pasa (la entrada tiene la forma correcta);
- `enum` con un valor más, `"Sonda sin grabación"`: `TestEsquemaDeNormas/rangos-grabados` **FAIL** — `Not equal …
  + (string) (len=20) "Sonda sin grabación"`;
- línea base: con el clon devuelto a `aa7af30`, los tres tests pasan (`ok internal/evals`, `ok internal/skills`).

## Qué hace la persona (fuera del bucle, con red)

Todo en la rama `008-h6-territorio-skill-legal`, con el run detenido. Es el procedimiento de la pausa de H5.1/T008,
ampliado al `enum`.

1. **Manifiesto.** Añadir al final de `normas` de `testdata/evals/grabaciones.json` siete entradas, una por línea y con
   la forma de las demás, **sin `bloques`**: ninguna eval de `legal-core` cita un artículo, y `TestGrabarEvals` graba
   igualmente la búsqueda, los metadatos y el índice de cada norma. Punto de partida **sin verificar** (normas de
   research S7; número y año de `refs/mapa-sistema-legal-skills.md` §1.4, que avisa de que tiene datos de memoria): lo
   único que vale es lo que resuelva la grabación.

   | Norma | `busqueda` propuesta | `titulo_empieza_por` propuesto |
   |---|---|---|
   | LOPJ | `poder judicial` | `Ley Orgánica 6/1985,` |
   | LOPDGDD | `protección datos personales derechos digitales` | `Ley Orgánica 3/2018,` |
   | LJCA | `jurisdicción contencioso administrativa` | `Ley 29/1998,` |
   | LEC | `enjuiciamiento civil` | `Ley 1/2000,` |
   | LGS | `general de subvenciones` | `Ley 38/2003,` |
   | LGP | `general presupuestaria` | `Ley 47/2003,` |
   | Código Civil | `código civil` | el que diga la grabación (norma de 1889: no se supone) |

   `para` de cada una, por ejemplo: `ley vertebral de legal-core (FR-070): identificador, título y rango de <sigla> en
   data/normas.yaml (FR-071, FR-072; T017); los rangos de su búsqueda entran en el enum de schemas/normas.yaml.json
   (D29)`.
2. **Grabar** con `scripts/grabar-evals.sh`. Si una entrada falla con «N resultados de … tienen un título que empieza
   por el prefijo, y tiene que ser uno solo; títulos: […]», corregir la búsqueda o el prefijo **copiando de esos
   títulos** y volver a grabar: `Preparar` sirve antes lo ya grabado, así que cada pasada solo pide lo que falta. El
   registro (`t.Logf`) deja cada identificador y título resueltos: son los que T017 copiará a `data/normas.yaml`.
3. **Restaurar** la grabación reescrita de `robots.txt`
   (`git checkout -- testdata/evals/boe.legislacion-consolidada/GET_https_www.boe.es_robots.txt.json`) y comprobar
   que lo nuevo son solo las siete búsquedas y los metadatos e índices de las siete normas.
4. **`enum` desde la grabación.** `go test -count=1 -run 'TestEsquemaDeNormas/rangos-grabados' ./internal/skills/`:
   el `expected` del fallo es el conjunto de rangos grabados; copiar a `schemas/normas.yaml.json` los que falten en
   `actual`, respetando el orden alfabético actual. Cuentan **todos** los resultados de las siete búsquedas, no solo
   las siete normas: al menos «Ley Orgánica», y lo que traigan los demás resultados.
5. **Comprobar**: `go test -count=1 -run '^(TestManifiestoDeGrabaciones|TestIdentificadoresDeLasNormas|TestGrabacionesSinSolape)$' ./internal/evals/`,
   `go test -count=1 -run '^TestEsquemaDeNormas$' ./internal/skills/` y `make ci` en primer plano, en verde. Si una
   búsqueda no resuelve su norma, no se fuerza: se restaura, se anota aquí y T015 queda rechazada (la pausa lo prevé).
6. **Cerrar**, una de dos:
   - **A (recomendada)**: marcar T015 `[X]` en `tasks.md`, commitear `testdata/evals/`, `schemas/normas.yaml.json` y
     `tasks.md` en un solo commit `feat(H6): T015` con el texto de la tarea, y reanudar con
     `scripts/hito.sh --resume <run>`. `siguiente_tarea` ve T015 `[X]` sin cambios pendientes y pasa a T016. La
     revisión de datos es la de quien graba, como en la salida manual que ofrece `gate_plataforma`.
   - **B**: dejar el material sin commitear, descontar T015 en `gates/tareas-intentos.json` (si el run se detuvo por
     intentos agotados) y reanudar: el intento siguiente verifica, marca y commitea, y ese commit dispara la pausa de
     datos, donde se aprueba.

## Mejora de proceso (fuera del hito, en su propia rama)

El workflow no tiene una pausa **anterior** al implementador para las tareas `[datos]` cuyo material lo produce una
persona y cuyo lector ya existe. Propuesta: una etiqueta (p. ej. `[persona]`) que `siguiente_tarea` detecte como
`[plataforma]` bloqueada y que abra un gate antes de implementar —«deja el material en el árbol y aprueba; el
ejecutor verifica y marca»—, con la pausa de datos de siempre al commitear. Y en el juez de `tasks`, un criterio: toda
tarea `[datos]` «que graba una persona» tiene una parte de ejecutor que deja `make ci` en verde, o lleva esa etiqueta.
Completa la mejora que ya anotaba `gates/tarea-T001.md` («afecta igual a T002, T003 y T015»).

## Notas

- El clon de la sonda queda en `/tmp/kitlegal-t015-sonda`, devuelto a `aa7af30`: borrarlo pedía aprobación en
  headless. No afecta a nada.

## Registro de intentos

- **Intento 2** (2026-09-21): sin material de la persona. `HEAD` sigue en `aa7af30`; `git status --porcelain --
  testdata/evals schemas/normas.yaml.json` vacío; el manifiesto conserva sus once entradas y el `enum` de `rango` no
  tiene «Ley Orgánica». Comprobado que los dos lectores siguen donde dice la tabla (`grabaciones_test.go:301`,
  `normas_test.go:375`). `make ci` en primer plano: «todos los controles en verde». Nada que hacer desde el ejecutor:
  T015 queda `[ ]`. Sigue pendiente lo de «Qué hace la persona».
- **Intento 3** (2026-09-21, el último): sin material de la persona. `HEAD` sigue en `aa7af30`; `git status
  --porcelain -- testdata/evals schemas/normas.yaml.json` vacío; sin stash, sin otro worktree ni rama con material, y
  el clon de la sonda en `/tmp/kitlegal-t015-sonda` limpio en `aa7af30`. El manifiesto conserva sus once entradas, el
  directorio de grabaciones sus treinta y cinco ficheros y el `enum` de `rango` sus nueve valores sin «Ley Orgánica».
  `make ci` en primer plano: «todos los controles en verde». T015 queda `[ ]` y el run se detendrá en
  `siguiente_tarea` por intentos agotados: toca «Qué hace la persona» y después el cierre A (recomendado) o B.

## Resuelto en la pausa (2026-09-22), cierre A

Manifiesto, grabaciones y `enum`, con `make ci` en verde y T015 marcada `[X]`.

**Las 18 entradas resolvieron a la primera**, las siete nuevas incluidas; ninguna necesitó corregir búsqueda ni
prefijo. El prefijo propuesto para el Código Civil («Real Decreto de 24 de julio de 1889») resultó ser el bueno.

| Norma | Identificador | Título que devolvió la grabación |
|---|---|---|
| LOPJ | `BOE-A-1985-12666` | Ley Orgánica 6/1985, de 1 de julio, del Poder Judicial. |
| LOPDGDD | `BOE-A-2018-16673` | Ley Orgánica 3/2018, de 5 de diciembre, de Protección de Datos Personales y garantía de los derechos digitales. |
| LJCA | `BOE-A-1998-16718` | Ley 29/1998, de 13 de julio, reguladora de la Jurisdicción Contencioso-administrativa. |
| LEC | `BOE-A-2000-323` | Ley 1/2000, de 7 de enero, de Enjuiciamiento Civil. |
| LGS | `BOE-A-2003-20977` | Ley 38/2003, de 17 de noviembre, General de Subvenciones. |
| LGP | `BOE-A-2003-21614` | Ley 47/2003, de 26 de noviembre, General Presupuestaria. |
| Código Civil | `BOE-A-1889-4763` | Real Decreto de 24 de julio de 1889 por el que se publica el Código Civil. |

Son los identificadores y títulos que T017 copia a `data/normas.yaml`, **de aquí y no de `refs/`**. Restaurada la
grabación de `robots.txt`; lo nuevo son 21 ficheros: las siete búsquedas y los metadatos e índices de las siete normas.

### El `enum` de `rango` gana **tres** valores, no uno

`TestEsquemaDeNormas/rangos-grabados` pedía «Acuerdo», «Instrucción» y **«Ley Orgánica»**. D29 acertaba al decir «al
menos Ley Orgánica, y lo que traigan los demás resultados»: los otros dos salen de resultados de las búsquedas que no
son ninguna de las siete normas. El rango del Código Civil **no** añade valor: es uno de los que ya estaban.

### Lo que D29 no previó: un caso negativo que el `enum` deja sin error

Al entrar «Ley Orgánica» en el `enum`, `TestLeerNormas/rango-no-admitido` dejó de fallar: usaba justo ese valor como
ejemplo de rango **no** admitido (`internal/skills/normas_test.go:212-213`), así que esperaba un error y recibía
`nil`. Es la misma clase que D28 —un control que el cambio de esquema rompe y que por tanto es inseparable de él—,
pero ni el plan ni el juez la vieron, así que **esta tarea toca también `internal/skills/normas_test.go`**, fuera de
las rutas que declaraba.

El caso pasa a usar `rango: Bando`: un bando es un acto del alcalde y nunca aparece como rango en la legislación
consolidada del BOE, así que ninguna grabación futura puede meterlo en el `enum` y repetir el fallo. Escoger otro
valor del vocabulario estatal (por ejemplo «Decreto») lo dejaría expuesto a la siguiente norma autonómica que se
grabe.

**Para el plan**: D29 debería decir, como D28, que la tarea trae la expectativa de test que el `enum` arrastra. Queda
anotado para la revisión final del hito.
