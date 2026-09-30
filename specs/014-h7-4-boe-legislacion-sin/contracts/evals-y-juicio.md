# Contrato: el formato de eval, la eval 20 y el juicio de la respuesta a la pregunta

Research D9-D12. Las evals de hoy conservan su juicio (FR-053, FR-062): las claves nuevas son opcionales y, sin ellas,
`Juzgar` hace lo de hoy.

## 1. `schemas/eval.yaml.json` (tarea `[datos]`)

Dos propiedades nuevas:

```json
"no_se_activan": {
  "type": "array", "minItems": 1, "uniqueItems": true,
  "items": { "type": "string", "pattern": "^[a-z0-9]+(-[a-z0-9]+)*$" }
},
"redacciones_modificadas": {
  "type": "array", "minItems": 1,
  "items": { "$ref": "#/$defs/redaccion-modificada" }
}
```

```json
"redaccion-modificada": {
  "type": "object",
  "additionalProperties": false,
  "required": ["norma", "bloque", "fecha_vigencia", "fecha_vigencia_reciente"],
  "properties": {
    "norma": { "$ref": "#/$defs/norma" },
    "bloque": { "$ref": "#/$defs/bloque" },
    "fecha_vigencia": { "$ref": "#/$defs/fecha" },
    "fecha_vigencia_reciente": { "$ref": "#/$defs/fecha" }
  }
},
"fecha": { "type": "string", "pattern": "^[0-9]{4}(0[1-9]|1[0-2])(0[1-9]|[12][0-9]|3[01])$" }
```

`no_se_activan` vale en toda eval; `redacciones_modificadas` entra en la lista del `else` (una eval de no activación no
la admite), como `hallazgos`. Los casos de `formato_test.go` que lo fijan: una eval con cada clave bien formada; con
`no_se_activan: []`, con un nombre que no es de skill y con uno repetido; `redacciones_modificadas` en una eval de no
activación, sin `fecha_vigencia_reciente`, con una fecha de siete cifras o con un mes 13.

## 2. El juicio (`Juzgar`)

Orden de los motivos de una sesión, con lo nuevo en negrita: sin terminar; la activación; **cada skill de
`no_se_activan` que la sesión activó** («se activó la skill <nombre>, que la eval dice que no se activa»); cada comando
ausente; cada comando prohibido; cada cita ausente; cada aviso ausente; cada hallazgo ausente; **cada redacción
modificada ausente** («redacción modificada ausente: <norma> <bloque> <fecha_vigencia> <fecha_vigencia_reciente>»); cada
elemento del territorio ausente; cada expresión prohibida; el modelo (lo pone `EscribirInforme`). `Pasa` exige además
que no se active ninguna de `no_se_activan` y que no falte ninguna redacción esperada.

**Líneas esperadas.** `ExtraerRedaccionesModificadas(respuesta)`, por cada línea de la respuesta con la forma de la
etiqueta de `version-obsoleta` (`formasDeHallazgo`), da la primera cita de la línea (`ExtraerCitas`) y sus dos primeras
fechas de ocho cifras con un mes y un día posibles, sin otra cifra a los lados, en su orden; nada si la línea no tiene
cita o tiene menos de dos fechas. Una esperada está si alguna extraída es igual (norma, bloque y las dos fechas en su
orden). `ResultadoDeEval` publica `redacciones_modificadas_encontradas` y `redacciones_modificadas_ausentes`, con el texto
de cada una, en el orden de la eval; vacías, nunca `null`, si la eval no las espera. La serie publica, en `formas`, la
de cada clase de sus hallazgos (como hoy) y, detrás, `⚠ REDACCIÓN MODIFICADA: <texto>` por cada una esperada.

**Casos de `TestJuzgar…` que lo fijan** (FR-052, FR-094): con las dos líneas de la eval 20 (una por bloque, con su cita
y sus fechas), pasa; con una sola, con las dos sin cita, con la cita de un bloque y las fechas del otro, con las fechas
en otro orden o con las dos en una sola línea, no pasa y nombra cada ausente; una eval sin la clave, juicio de hoy. Y la
sesión de una eval de `legal-core` (`activa: true`, `no_se_activan: [boe-legislacion]`) que activa `legal-core` y
`boe-legislacion`: no pasa, con el motivo que nombra `boe-legislacion`; la misma sin activar `boe-legislacion`, pasa.

## 3. La eval 20 (`evals/boe-legislacion/20-lcsp-dos-bloques-redaccion-cambiada.yaml`, tarea `[datos]`)

```yaml
# Eval informativa (ADR 0016): la consulta repetida sobre dos preceptos de una misma norma que el BOE modificó de verdad.
# El grafo de la sesión ya registró una lectura del art. 118 (bloque a1-30) y otra de la disposición adicional tercera
# (bloque da-3) de la LCSP que vieron su redacción original (vigencia 20180309), derivadas de las respuestas grabadas
# quitándoles solo la redacción posterior; la caché sirve las grabadas, con las vigentes (20200206, BOE-A-2020-1651;
# 20230101, BOE-A-2022-22128). La sesión tiene que leer los dos bloques, en una orden o en dos, comprobar con kitlegal
# graph check y la norma, no pedir graph show, citar los dos bloques y llevar una línea ⚠ REDACCIÓN MODIFICADA: por
# bloque, cada una con su cita y sus dos fechas.
pregunta: "Hace tiempo te pregunté qué exige la LCSP para el expediente de un contrato menor, en su artículo 118, y qué añade su disposición adicional tercera para los ayuntamientos. ¿Qué dicen ahora?"
activa: true
informativa: true
grafo_previo:
  grabaciones: lcsp-a1-30-y-da-3-redaccion-original
  comandos:
    - applet: boe
      norma: BOE-A-2017-12902
      bloque: a1-30
    - applet: boe
      norma: BOE-A-2017-12902
      bloque: da-3
comandos:
  - applet: boe
    norma: BOE-A-2017-12902
    bloque: a1-30
  - applet: boe
    norma: BOE-A-2017-12902
    bloque: da-3
  - applet: graph
    verbo: check
    norma: BOE-A-2017-12902
prohibidos:
  - applet: graph
    verbo: show
citas:
  - norma: BOE-A-2017-12902
    bloque: a1-30
  - norma: BOE-A-2017-12902
    bloque: da-3
hallazgos:
  - version-obsoleta
redacciones_modificadas:
  - norma: BOE-A-2017-12902
    bloque: a1-30
    fecha_vigencia: "20180309"
    fecha_vigencia_reciente: "20200206"
  - norma: BOE-A-2017-12902
    bloque: da-3
    fecha_vigencia: "20180309"
    fecha_vigencia_reciente: "20230101"
```

Los dos bloques se leen con `articulo` (una orden por bloque) o con `articulos` (una orden con los dos): la forma
bloque del comando esperado ya admite los dos verbos (`leeElBloque`). El conjunto queda con 20 evals: 10 positivas que
deciden, 2 de no activación y 8 informativas (FR-054; límite de `conjunto.go:206`).

## 4. El grafo previo de la eval 20 (tarea `[datos]`)

`testdata/evals/grafo-previo/lcsp-a1-30-y-da-3-redaccion-original/` con dos derivadas, cada una con el nombre de su
grabación de H4:

- `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_a1-30.json`, la
  derivación de la grabación de `a1-30` con 20180309 (byte a byte la misma que la de la 19);
- `GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2017-12902_texto_bloque_da-3.json`, la de
  `da-3` con 20180309 (FR-051).

`internal/app/grafo_test.go`: `derivadasDelGrafoPrevio()` gana dos entradas con esa subcarpeta (argumentos `articulo
BOE-A-2017-12902 a1-30` y `… da-3`, fecha 20180309). La tarea las escribe con
`go test ./internal/app/ -run '^TestGrabacionesDerivadas$' -actualizar-derivadas` y `make ci` las comprueba sin la
bandera: byte a byte su derivación, la redacción de su fecha que trae la grabada y ningún fichero sin entrada en la
carpeta (FR-096; SC-012).

`probarGrafosPrevios` (`TestEvalsDelRepositorio/grafo-previo`) prepara el grafo previo de la eval 20 como el de la 19 y
comprueba lo que verá la sesión; `compruebaLaSesion` gana una comprobación: cada redacción de `redacciones_modificadas`
es un `version-obsoleta` de su norma y su bloque, con su `fecha_vigencia` como la superada y su
`fecha_vigencia_reciente` como la leída (la eval no puede esperar unas fechas que las grabaciones no dan).

## 5. Las evals de `legal-core` (FR-004)

Las tres (`01-territorio-municipio-cubierto.yaml`, `02-territorio-municipio-no-cubierto.yaml`,
`03-no-activa-receta-de-cocina.yaml`) ganan, detrás de `activa`, `no_se_activan: [boe-legislacion]`. Nada más cambia en
ellas (fuera de alcance).

## 6. La respuesta a la pregunta (`LeerSesion`, FR-060)

`Sesion.Respuesta` es el `result` del **primer** mensaje `result` del transcript, si tiene subtype `success` e
`is_error` falso; vacía en otro caso. Los demás campos, como hoy (`Fin`, `Terminada`, `MotivoSinTerminar` y
`ErrorDelResultado`, del último mensaje; las activaciones, de todo el transcript). Lo comparten el job y el sondeo.

Caso nuevo de `TestLeerSesion` (tarea `[datos]`), `internal/evals/testdata/sesiones/leer-sesion/respuesta-antes-de-una-tarea-en-segundo-plano/`:
`sesion.jsonl` con `system/init`; un `assistant` con el `tool_use` de `Skill` (`boe-legislacion`); un `assistant` con un
`tool_use` de `Bash` en segundo plano; su `user` con el `tool_result`; un `assistant` con el texto de la respuesta; un
`result` `success` con la respuesta con cita (`respuestaConCita`); un `system` con `subtype` `task_notification`
(`task_id`, `tool_use_id`, research V4); un `assistant` con la réplica; y un `result` `success` con la réplica («Esa tarea
en segundo plano era solo una búsqueda auxiliar…»). `codigo-de-la-sesion` con `0` y `sesion.err` vacío. Se espera la
sesión activada y terminada, con `Respuesta` la de la pregunta y `Fin` `result success`.
