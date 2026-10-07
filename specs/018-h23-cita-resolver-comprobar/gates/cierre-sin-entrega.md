# H23, run 94cbbfec: cerrado sin entrega (2026-10-07)

El run se detuvo en `grabar_datos`, tras T010 de 23, con las diez primeras tareas en verde. No se reanuda: lo
decidió Jorge el 2026-10-07.

## Qué pasó

- **2026-10-07, 01:43 y 01:44.** El test de grabación pidió la página del buscador (200, con su cookie) y envió
  el formulario con `ECLI:ES:TS:2023:3144`. El envío recibió un 302, las dos veces. El test no guardó nada.
- **08:28.** La misma prueba a mano del 2026-10-02 (`docs/JURISPRUDENCIA.md` §3), con `curl` y el agente
  `kitlegal/0.3 (+https://kitlegal.es/bot)`: la página, 200; el envío, `302` con
  `Location: captchalogin.jsp?prevaction=query&…&ECLI=ECLI%3AES%3ATS%3A2023%3A3144&`. El buscador pide un CAPTCHA.
- **La petición de kitlegal era correcta.** Volcada contra un servidor local: la cookie de la página, los seis
  campos y `X-Requested-With`. No es un defecto de `internal/httpx`.
- **Una persona con un navegador no lo recibe.** Jorge buscó ese ECLI en su navegador y en una ventana privada,
  sin cookies: la sentencia, sin CAPTCHA. La página no ha cambiado desde julio (`guid=202607230816`) y su
  formulario no lleva ninguna clave. El buscador distingue al cliente que se identifica como programa.

## Qué se decidió

No se sortea: ni otro agente, ni un navegador, ni resolver el CAPTCHA (ADR 0036). `cita resolver` contra el
CENDOJ no se entrega. El objetivo del hito —que ninguna sentencia citada sea inventada— pasa a un hito más
pequeño en el que la comprobación la hace la persona con su navegador, con la consulta que le prepara la skill, y
trae el documento.

## Qué queda en esta rama

Material de lectura para ese hito, no código del producto: el reconocimiento de ECLI y de ROJ
(`internal/source/cendoj/referencia.go`, con sus tests y su fuzz), el spec, el plan y las tareas. El formulario de
`internal/httpx` y la grabación no se usan.
