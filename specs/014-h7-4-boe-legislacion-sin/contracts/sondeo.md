# Contrato: los errores de uso del sondeo

`scripts/evals-sondeo.sh`, `internal/evals/sondeo.go` y el punto de entrada `TestSondeo` (research D18). Sustituye al
§2, punto 3, del contrato del sondeo de H7.3 en lo que dice del registro de `go test` en los errores de uso; sus
códigos de salida (§6) se quedan: 1 el guion, 2 `make`.

## 1. Qué es un error de uso (FR-080)

Lo que hoy devuelve `comprobarElSondeo` por los argumentos (los de `SKILL`, `EVALS`, `MODELO`, `REPETICIONES` y
`CONCURRENCIA`, unidos, uno por línea, en ese orden) o por la credencial (`errSinSuscripcion`), con el mismo texto:
ahora envuelto en `errorDeUso`. No lo son: que la definición del job no se pueda leer, que el binario no se construya,
que las skills no se instalen, que el repartidor falle o que el directorio de sesiones no se pueda crear (FR-081).

La ruta de la definición del job llega en `SondeoAEjecutar.RutaDeLaDefinicionDelJob` (nuevo; `TestSondeo` le da
`rutaDeLaDefinicionDelJob`, la de hoy), para que un test pueda darle una que no se puede leer.

## 2. `TestSondeo`

Con un `errorDeUso` (`errors.As`), escribe su mensaje, con un salto de línea final, en `uso.txt` del temporal y termina
sin fallar; no escribe `salida.txt`. Con cualquier otro error, falla como hoy (`require.NoError`). Sin error, escribe
`salida.txt` como hoy. Ninguna sesión se abre antes de la comprobación (hoy).

## 3. `scripts/evals-sondeo.sh`

Tras `go test`, en este orden:

1. si `go test` falló: imprime `go-test.log` en la salida de error y sale con 1 (hoy);
2. si `uso.txt` existe y no está vacío: lo imprime en la salida de error, sin nada más, y sale con 1;
3. si no: imprime `salida.txt` (hoy).

El temporal se borra al salir, con el código que sea (hoy).

Ejemplo, con `REPETICIONES=0` y `MODELO` vacío (dos líneas, 77 bytes; como mucho una línea por argumento y la de la
credencial, < 1 KB):

```text
MODELO: está vacío
REPETICIONES: «0» no es un entero mayor o igual que 1
```

Sin `CLAUDE_CODE_OAUTH_TOKEN` (una línea):

```text
falta la credencial: CLAUDE_CODE_OAUTH_TOKEN, el token de la suscripción que da claude setup-token, no está en el entorno o está vacía
```

## 4. Tests (FR-099; SC-009)

- `TestGuionDelSondeo` (con el `go` sustituto de hoy, envuelto en uno que deja `uso.txt`) gana tres casos: el
  sustituto deja `uso.txt` con dos líneas y sale con 0 (`uso-con-dos-argumentos-que-no-valen`) → el guion sale con 1,
  su salida de error es exactamente esas dos líneas y su salida estándar, vacía, aunque haya `salida.txt`; lo mismo con
  la línea de la credencial (`uso-sin-la-credencial`); y el sustituto deja `uso.txt` y sale con 1
  (`go-test-sale-con-1-y-deja-uso`) → el guion sale con 1 con el registro y sin `uso.txt`, que fija el orden de §3. El
  de hoy en que el sustituto falla sigue dando el registro entero con 1 (FR-081).
- `TestComprobarElSondeo` (hoy) comprueba además que sus errores de argumentos y de credencial son `errorDeUso` y, en
  la subprueba `job-ilegible`, que el de una definición del job que no se puede leer, que llega antes que un argumento
  que no vale y que la credencial que falta, no lo es.
- `TestSondear` comprueba que su error, en los casos de un argumento que no vale y de la credencial ausente o vacía, es
  un `errorDeUso` y que con él no se llama a `PrepararElArbol` ni al repartidor (0 sesiones) ni se escribe en el
  temporal.
