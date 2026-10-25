# Data model: H31.1 · `boe-legislacion` habla de la norma y lee lo que nombra

El hito no añade ningún tipo, campo ni formato: cambia texto de una skill, datos de tres evals y un manifiesto, con los
formatos de H26.1 y de H31. Lo que sigue son las entidades del spec («Key Entities») con lo que el hito les hace y las
reglas que las validan.

## 1. Cuerpo de `SKILL.md`

- **Qué es**: `skills/boe-legislacion/SKILL.md` desde el byte que sigue a la línea `---` que cierra el frontmatter
  (`cuerpoDeLaSkill`, `internal/skills/sincronia.go`). Lo carga entero el agente al activar la skill.
- **Partes**: el protocolo (pasos 1 a 5), «Cómo se cita», «Redacción en una fecha», «Redacción modificada», la tabla de
  comandos generada entre sus marcas y las reglas 1 a 7 (hasta hoy, 1 a 8).
- **Reglas**: como mucho 11 000 caracteres (runas) y menos de 500 líneas (gate de H31, `maximoDeCaracteresDelCuerpo`,
  `maximoDeLineasDelCuerpo`); cada referencia de `references/` enlazada desde él; las formas fijas `⚠` en bloques
  `text`; la tabla de comandos, la que regenera `make skills-sync`.
- **En este hito**: 10 992 → 10 993 caracteres; el texto de contracts/skill-boe-legislacion-v0-3-1.md §1.

## 2. Referencia a mano

- **Qué es**: `skills/boe-legislacion/references/<nombre>.md` sin la cabecera «generado desde» (`ReferenciasAMano`).
  El agente la lee si un paso del cuerpo la pide y la sesión necesita lo que trae.
- **Reglas**: enlazada desde `SKILL.md` y desde ninguna otra referencia; sin `⚠` al principio de una línea que no sea
  una línea de `SKILL.md`; los controles de contenido de `docs/SKILLS.md` §1 la leen como al cuerpo.
- **En este hito**: cambian `redaccion_modificada.md` (dos apartados y su párrafo de entrada) y `buscar_la_norma.md`
  («Si no aparece»); ninguna nueva (contracts/skill-boe-legislacion-v0-3-1.md §4).

## 3. Lectura de una eval

- **Qué es**: una entrada de `lecturas` de `evals/<skill>/<nn>-….yaml`, `{norma, bloque, motivo}`
  (`LecturaDeLaEval`, `internal/evals/formato.go`): un bloque que la caché de las sesiones de la skill sirve porque la
  eval lo provoca (ADR 0043, clase (b)).
- **Reglas** (H26.1, sin cambios): `motivo` no vacío ni solo de blancos; ningún bloque que sea comando de bloque o cita
  de su eval; no cuenta en ningún umbral. `ConsultasNecesarias` la anota sin fecha y con `--fecha 9999-12-31`, y la
  caché de cada sesión de la skill la sirve salvo que la eval de la sesión la tenga en `sin_servir`.
- **En este hito**: nueve nuevas (contracts/lecturas-y-grabaciones.md §1); la 17 gana la clave.

## 4. Entrada del manifiesto de grabaciones

- **Qué es**: un elemento de `normas` de `internal/evals/testdata/grabaciones.json`, `{busqueda, titulo_empieza_por,
  bloques, para}` (`EntradaDelManifiesto`), con la `fuente` del documento, `boe.legislacion-consolidada`.
- **Reglas**: lo lee `LeerManifiesto` (`TestManifiestoDeGrabaciones`); lo graba `TestGrabarEvals` en `grabar_datos`,
  solo de una fuente con fila revisada en `docs/SOURCES.md` de `main`; cada id de bloque, el de su índice grabado.
- **En este hito**: dos entradas ganan bloques y una frase de `para` (contracts/lecturas-y-grabaciones.md §2).

## 5. Grabación de un bloque

- **Qué es**: `internal/evals/testdata/boe.legislacion-consolidada/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_<norma>_texto_bloque_<id>.json`,
  `{formato, grabado_en, peticion, respuesta}`, con todas las versiones del bloque en la respuesta.
- **Reglas**: la escribe solo `grabar_datos`; una persona la revisa en el informe final; sirve la lectura sin fecha y
  la de cualquier `--fecha`.
- **En este hito**: ocho nuevas.

## 6. Umbral del informe del job

- **Qué es**: un elemento de `umbrales` de `informe.json` del trabajo `evals (<skill>)`, `{nombre, medida, umbral,
  cumple, decide, …}` (contrato del ADR 0029).
- **En este hito**: ninguno cambia ni se añade; los que deciden dan el veredicto del cierre (plan.md, «Controles de
  umbral»).
