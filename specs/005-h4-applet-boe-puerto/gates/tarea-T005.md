# T005 · Errores con clase e identificadores de entrada

## Intento 1 (2026-09-13, 12:23–12:48)

El trabajo quedó **completo y en verde**, pero la tarea no se marcó `[X]`.

- El implementador escribió `errores.go`, `errores_test.go`, `ids.go`, `ids_test.go` y las dos entradas de
  `misspell.ignore-rules` (`capitulo`, `disposicion`), lanzó `make ci` **en segundo plano** y terminó su turno con
  «esperaré a que termine». En una sesión headless (`claude -p`) terminar el turno cierra la sesión: la notificación
  del proceso en segundo plano nunca llegó y la marca `[X]` nunca se escribió.
- El workflow ejecutó su propia verificación (`gates/ci.log`, `kitlegal-verificacion exit=0`) y commiteó el trabajo
  como `c1cbd3d feat(H4): T005`. Al seguir `[ ]` en `tasks.md`, `siguiente_tarea` la volvió a elegir como intento 2.

## Intento 2 (2026-09-13)

- Revisión de lo commiteado contra la tarea: `Error{URL, Instante, Causa}` con `Error()`, `Unwrap()` y `Clase()`
  (`schema.ConClase`) y los cinco constructores sin exportar, uno por clase, ninguno de identidad humana (FR-100,
  D10); `TestErrorDeBoeMensajes` con las diez formas de mensaje del contrato errores-y-codigos §2, los textos de
  `httpx` y `cache` conservados detrás del contexto del verbo, y clase y código de cada constructor bajo cero, una y
  dos capas de `%w`; `ValidarNorma` y `ValidarBloque` con las gramáticas de data-model §5 y «argumentos» nombrando el
  valor y la forma esperada (FR-080, FR-081, D9); `TipoDesdeID` con las diez reglas en el orden de `_tipo_from_id`
  (`refs/boe.py` 155-176; comprobado que `c` a secas no casa la regla 3 tampoco en Python, porque
  `re.match(r"^c[ivxlcdm]")` exige el segundo carácter); `TestValidarNorma`, `TestValidarBloque` (vacío, `/`, `\`,
  `..`, `?`, `#`, `%`, espacios, controles, no ASCII, 65 caracteres), `TestTipoDesdeID` (`a21`, `da3`, `dt1`, `ti`,
  `cv3`, `preambulo` y varios sin regla) y `FuzzIDDeBloque` con las tres semillas (sin pánico, id aceptado seguro como
  segmento de la petición, tipo determinista) (FR-082, SC-007); `.golangci.yml` tocado solo en
  `misspell.ignore-rules` con `capitulo` y `disposicion` y su comentario (D18). Nada que añadir ni corregir.
- `make ci` en primer plano sobre el árbol actual: en verde (`internal/source/boe` con cobertura 100 %). Se marca
  `[X]`.

## Lección para el proceso

En modo headless la verificación determinista se ejecuta **en primer plano** (`rtk proxy make ci` con tiempo de
espera amplio) y la marca `[X]` se escribe en el mismo turno, antes de terminar. Un proceso en segundo plano seguido
de fin de turno gasta un intento sin dejar rastro en `gates/`.
