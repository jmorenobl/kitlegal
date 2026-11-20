# Contrato: las evals de la frontera, su juicio y sus umbrales

FR-012 a FR-019, FR-024, FR-026; SC-004, SC-006 a SC-008. Todo en `internal/evals`, `schemas/eval.yaml.json`,
`evals/` y la grabación del manifiesto de H9.

## 1. El formato (`schemas/eval.yaml.json`, `internal/evals/formato.go`)

`no_cubierto`, con `additionalProperties: false`:

```json
"no_cubierto": {
  "type": "object",
  "additionalProperties": false,
  "if": { "required": ["fuera_de_lo_declarado"] },
  "then": { "not": { "anyOf": [{ "required": ["norma"] }, { "required": ["precepto"] }] } },
  "else": { "required": ["norma", "precepto"] },
  "properties": {
    "fuera_de_lo_declarado": { "const": true },
    "norma": "… (sin cambios)",
    "precepto": "… (sin cambios)"
  }
}
```

y, en la raíz, `tambien_la_resuelven` con la forma de `se_activan` (lista no vacía, sin repetir, de nombres
`^[a-z0-9]+(-[a-z0-9]+)*$`). Las demás reglas de `no_cubierto` (solo con `activa: true`, nunca con `plazo`…) no cambian.
`NoCubiertoEsperado` gana `FueraDeLoDeclarado bool` (`yaml:"fuera_de_lo_declarado"`) y `Eval`, `TambienLaResuelven
[]string` (`yaml:"tambien_la_resuelven"`). `esDeCobertura` es `Norma != "" || FueraDeLoDeclarado`.

| Caso de `TestFormatoDeLaCobertura` (nuevo, `formato_test.go`) | Resultado |
|---|---|
| `no_cubierto: {norma, precepto}` | se lee; es de cobertura |
| `no_cubierto: {fuera_de_lo_declarado: true}` | se lee; es de cobertura |
| `no_cubierto: {fuera_de_lo_declarado: true, norma: …, precepto: …}` (el mutante de FR-012) | rechazada, con un defecto en `no_cubierto` |

`TestFormatoDePlazos` no cambia ni sus casos ni sus mensajes (FR-012; research V4).

## 2. El juicio (`internal/evals/plazos.go`)

- `NoCubiertoEsperado.declarada`: con `FueraDeLoDeclarado`, verdadero si `ExtraerNoCubierto` da alguna declaración;
  si no, como hoy, si alguna lleva en su misma línea la norma y la cita del precepto.
- `juzgarElPlazo`, en una eval de cobertura sin declarar: el motivo `declaración de no cubierto ausente: ` seguido de la
  norma esperada o, en una de `fuera_de_lo_declarado`, de la constante `fuera de lo declarado`
  (`motivoFueraDeLoDeclarado`). Cada línea del plazo sigue dando `línea del plazo en un asunto no cubierto: <fecha>`.
- Ninguna constante de `internal/evals` lleva texto de la plantilla, del motivo ni de la cola.

`TestJuzgarLaDeclaracionDeNoCubierto` (nuevo, `plazos_test.go`): `juzgarElPlazo` sobre once respuestas constantes, las
de FR-013 y la historia 3, escenario 4; cada caso fija `NoCubiertoDeclarado` y los motivos, en rojo si uno se clasifica
al revés o da otro motivo. Que la línea con la norma y la cita cuente como declarada en lo regulado aparte ya lo fijan
`TestJuzgarPlazos` y los demás tests de hoy:

| Eval | Respuesta | Declarada | Motivos |
|---|---|---|---|
| fuera de lo declarado | sin ninguna línea `⚠ NO CUBIERTO:` | no | `declaración de no cubierto ausente: fuera de lo declarado` |
| fuera de lo declarado | solo una línea `⚠ PLAZO: 2026-12-02 …` con su cita | no | el de arriba y `línea del plazo en un asunto no cubierto: 2026-12-02` |
| fuera de lo declarado | `**⚠ NO CUBIERTO:** …` (negritas) | sí | ninguno |
| fuera de lo declarado | `⚠️ NO CUBIERTO: …` | sí | ninguno |
| fuera de lo declarado | `⚠ NO CUBIERTO: a quién reclamar y qué escrito presentar. Está fuera de lo que kitlegal cubre, que no da plazos, órganos, vías ni escritos para esto.` (despacho 44568601898) | sí | ninguno |
| fuera de lo declarado | `⚠ NO CUBIERTO: el día en que vence el plazo: es un plazo laboral. kitlegal no da sus plazos, órganos, vías ni escritos.` (el mismo despacho) | sí | ninguno |
| lo regulado aparte (`BOE-A-2003-23186`, `daprimera`) | sin la línea | no | `declaración de no cubierto ausente: BOE-A-2003-23186` |
| lo regulado aparte | la línea sin `BOE-A-2003-23186` | no | el mismo |
| lo regulado aparte | la línea sin `[BOE-A-2015-10565, bloque daprimera]` | no | el mismo |
| lo regulado aparte | la etiqueta en una línea y la cita en la siguiente | no | el mismo |
| lo regulado aparte | con el motivo «fuera de lo que kitlegal cubre» y sin la norma ni la cita | no | el mismo |

Los tests de hoy del juicio de lo regulado aparte (`TestJuzgarPlazos`, `TestJuzgarLasLecturasSinServir`,
`TestUmbralesDePlazosYTerritorio` y los de `informe_test.go`) cambian solo sus constantes con la línea vieja por la
plantilla con sus datos; lo que comprueban no cambia (FR-021).

## 3. `tambien_la_resuelven` y los umbrales

- `Eval.activadaPor(sesion Sesion, skill string) bool`: `sesion.Activada(skill)` o alguna de `TambienLaResuelven`. La
  usan `juzgarEnModo` y `juzgarSesion` para `ResultadoDeEval.Activada`, de la que salen el veredicto de la eval y
  `sin_activar`. `ComprobarLasSkillsNombradas` mira también `tambien_la_resuelven`, con su mensaje de hoy
  («… tambien_la_resuelven nombra <nombre> …»).
- `descripcionDelUmbralDeCoberturaSinDeclarar` pasa a ser, palabra por palabra, la de FR-014. Ningún otro código de
  `umbrales.go` cambia (research V12).
- Con las evals de §5, `consulta-juridica` lleva, con el plan del job, en este orden:
  `sin_activar:claude-sonnet-5-5:orden` (total 12), `cobertura_sin_declarar:claude-sonnet-5-5:orden` (total 6),
  `sin_activar:claude-sonnet-5-5:herramienta` (12), `cobertura_sin_declarar:claude-sonnet-5-5:herramienta` (6), los
  cuatro con `Decide: true` y umbral 0; detrás, los cuatro `fuera_de_lo_grabado` de hoy, que no deciden, y el de su
  descripción si el informe la lleva, como hoy. Lo fija `TestUmbralesDeLasSkills` (subtest `consulta-juridica`).

Ejemplo, el segundo, en `umbrales` de `informe.json` (446 bytes con `wc -c`; cuatro más por informe de `consulta-juridica`,
uno por medición):

```json
{"nombre": "cobertura_sin_declarar:claude-sonnet-5-5:orden", "descripcion": "Respuestas de claude-sonnet-5-5 en el modo orden sin una línea ⚠ NO CUBIERTO: —en lo regulado aparte, con la norma que lo rige y la cita de su precepto—, o con una línea ⚠ PLAZO:, sobre sus respuestas medidas en las evals que esperan una declaración de no cubierto", "medida": 0, "total": 6, "comparacion": "<=", "umbral": 0, "cumple": true, "decide": true}
```

## 4. La regla «frontera en dos materias» (`internal/evals/conjunto.go`)

La cuarta de `ReglasDeConsultaJuridica`, detrás de «una sola capacidad», con nombre `frontera en dos materias`: cuenta
las normas de `data/normas.yaml` sin `vertical` que cita (`normasCitadas`) una eval activa que no es de cobertura y otra
activa de cobertura con `fuera_de_lo_declarado` que prohíbe `plazos calcular`. Con menos de 2, el defecto:

`hay <n> normas de data/normas.yaml sin vertical con una eval activa a cada lado de la frontera y el conjunto lleva al menos 2: <normas, o «ninguna»>; evals activas: <ficheros>`

Con las evals de §5: 2 (`BOE-A-1980-22501`, `BOE-A-2015-11430`). En `TestEvalsDelRepositorio`: el conjunto de la rama la
cumple; el mutante sin la 03 da el defecto con 1 (`BOE-A-2015-11430`). El mutante de «no activación de cada capacidad»
quita `plazos-y-territorio` de `no_se_activan` en la 02 y en la 03; el subtest «formato» lleva las cuatro evals.

## 5. Las evals

`evals/consulta-juridica/01-poliza-tras-el-fallecimiento.yaml`: gana, al final,

```yaml
no_cubierto:
  fuera_de_lo_declarado: true
```

y su comentario cambia «No exige ninguna línea fija: lo que no cubre lo dice en prosa» por que la respuesta declara con
la plantilla lo que no cubre (a quién reclamar, qué vía), con el motivo fuera de lo declarado. Lo demás no cambia.

`evals/consulta-juridica/03-seguro-prorroga.yaml` (nueva, dentro de lo cubierto, seguros):

```yaml
# Eval positiva de la skill de entrada dentro de lo cubierto, de una materia sin paquete (seguros): leer y citar una
# norma lo cubre la capacidad de normativa. El otro lado de la frontera, en la misma norma, es la 01 (regla «frontera
# en dos materias»; ADR 0058). Sin lecturas: las declara la persona tras ver una sesión (ADR 0043, punto 2).
pregunta: "¿Qué dice la Ley de Contrato de Seguro sobre la prórroga de la póliza?"
activa: true
se_activan: [normativa]
no_se_activan: [jurisprudencia, plazos-y-territorio]
comandos:
  - applet: boe
    norma: BOE-A-1980-22501
    bloque: aveintidos
citas:
  - norma: BOE-A-1980-22501
    bloque: aveintidos
```

`evals/consulta-juridica/04-despido-plazo.yaml` (nueva, fuera de lo cubierto, laboral):

```yaml
# Eval de la skill de entrada fuera de lo cubierto, de una materia sin paquete (laboral): la sesión lee el art. 59 del
# Estatuto de los Trabajadores en la fecha del despido y lo cita, no pide plazos calcular, y declara con la plantilla y
# el motivo de lo no declarado el día en que vence el plazo. El otro lado de la frontera, en la misma norma, es la 02
# (regla «frontera en dos materias»; ADR 0058). Sin lecturas (ADR 0043, punto 2).
pregunta: "Me despidieron el 2 de noviembre de 2026. ¿Qué dice el Estatuto de los Trabajadores del plazo para reclamar y hasta qué día puedo hacerlo?"
activa: true
no_se_activan: [jurisprudencia]
comandos:
  - applet: boe
    norma: BOE-A-2015-11430
    bloque: a59
    fecha: "2026-11-02"
citas:
  - norma: BOE-A-2015-11430
    bloque: a59
prohibidos:
  - applet: plazos
    verbo: calcular
no_cubierto:
  fuera_de_lo_declarado: true
```

`evals/normativa/17-ltaibg-plazo-por-materia.yaml`: gana, tras su comentario de hoy,

```yaml
# El plazo del acceso a la información lo cubre también plazos-y-territorio (acceso-resolver, data/plazos.yaml), que
# lee su precepto con su paso 6: una respuesta que solo la activa a ella es una ruta que el producto permite, no un
# defecto, y su activación cuenta como la de normativa (ADR 0058, punto 7; «Antes de lanzar» del hito H33).
tambien_la_resuelven:
  - plazos-y-territorio
```

Sus comandos y sus citas no cambian. Las 07, 08 y 12 de `plazos-y-territorio` no cambian.

## 6. Lo grabado (FR-018)

`internal/evals/testdata/grabaciones.json` (manifiesto de H9, `"fuente": "boe.legislacion-consolidada"`), al final de
`normas`:

```json
{"busqueda": "estatuto de los trabajadores", "titulo_empieza_por": "Real Decreto Legislativo 2/2015,", "bloques": ["a59"], "para": "eval 04-despido-plazo de consulta-juridica (H33, FR-016, FR-018): el art. 59 del Estatuto de los Trabajadores, que la sesión lee en la fecha del despido; la búsqueda, los metadatos y el índice ya están en las grabaciones de H5, y solo el bloque es nuevo"}
```

Lo graba el paso `grabar_datos` con `TestGrabarEvals` (research D11). `TestCacheDeLasSesiones`, subtest
`consulta-juridica`: desde la caché de una sesión de la 04, `boe articulo BOE-A-2015-11430 a59 --fecha 2026-11-02
--offline` sale con 0; desde la de la 01, los arts. 22 y 4 de la Ley 50/1980, como hoy (la 03 usa el 22).

## 7. El trabajo de evals (FR-019)

`TestDefinicionDelJob`: el peor caso de `consulta-juridica` pasa a `SesionesPorGrupo: []int{25, 24}`, concurrencia 1,
15 013 s, con su texto (`15013 s` … `(⌈25 / 1⌉ + ⌈24 / 1⌉)` …), bajo el tope de 21 120 s. `TestMatrizDeLasPropuestas`:
su golden con las huellas y las evals nuevas de cada ejecución. `.github/workflows/evals.yml`: solo el comentario
(contracts/skills.md §4); ni `VERSION_DE_CLAUDE_CODE`, ni `MODELO_DE_EVALS`, ni la concurrencia, ni el tope.

## 8. Uso, de fuera adentro

- **La línea `⚠ NO CUBIERTO:` de una respuesta** la escribe el agente y la lee la persona: como mucho una por asunto no
  cubierto que la pregunta pide, ninguna por lo que no pide; una línea de unos 100 a 350 bytes (con su salto: la de los
  tributos, 324; las dos del despacho 44568601898, 125 y 156; medido con `wc -c`). Las órdenes por pregunta no crecen: §4 de
  `consulta-juridica` no pide ninguna, y la 04 hace una, `boe articulo` en la fecha; `plazos calcular` queda prohibido.
  Deja de darse cuando lo que pide entra en lo cubierto (un paquete con su fila, o una capacidad de otro hito).
- **La marca `⚠ NO VERIFICADO:`**, una por lectura hecha con la skill de normativa para completar, solo tras la línea
  fuera de lo declarado (140 bytes la forma); ninguna en lo regulado aparte. No entra en ningún umbral.
- **Los umbrales** (§3) los lee la persona en el informe del job, una vez por medición: cuatro más en el informe de
  `consulta-juridica`, unos 1800 bytes, fijos con el uso.
