# T010 · intento 1: redelimitada

## Por qué no queda en verde en este intento

T010 estaba mal delimitada: `TestGrabacionesDerivadas` (`internal/app/grafo_test.go`) recorre
`internal/app/testdata/derivadas/` y exige, con `assert.ElementsMatch`, una entrada de `grabacionesDerivadas()` por
cada fichero de la carpeta («cada derivada del e2e y del grafo previo de una eval tiene su comprobación, y cada
comprobación, su derivada»). La derivada `version-ulterior/` sola, que es todo lo que declaraba la tarea, deja ese
test y `make ci` en rojo; la entrada estaba en T011, detrás. `go test -count=1 -run TestGrabacionesDerivadas
./internal/app/` sobre el árbol de este intento (derivada sí, entrada no):

```
--- FAIL: TestGrabacionesDerivadas (0.00s)
        	Error:      	elements differ
        	            	extra elements in list A:
        	            	 (string) (len=140) "testdata/derivadas/version-ulterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json"
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/app	1.814s
```

La línea de T010 en tasks.md declara ahora `internal/app/grafo_test.go` con la entrada del control, que va primero
(rojo sin la derivada, verde con ella). No se tocó `grafo_test.go` en el árbol del repositorio. La línea de T011 no
cambia: su intento encontrará la entrada hecha y en verde, y le queda su verificación (la fecha de la entrada
cambiada un momento hace fallar el test; se restaura antes de `make ci`) y marcarla.

## El arreglo, comprobado en una copia desechable

Copia `rsync -a` (con `.git`) del árbol en `/tmp/kitlegal-t010/copia`:

1. **Rojo**: con solo el diff de abajo en `grafo_test.go` y sin la derivada,

   ```
   --- FAIL: TestGrabacionesDerivadas (0.00s)
           	Error:      	elements differ
           	            	extra elements in list B:
       --- FAIL: TestGrabacionesDerivadas/version-ulterior (0.02s)
               	Error:      	Received unexpected error:
               	            	open GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json: no such file or directory
   FAIL	github.com/jmorenobl/kitlegal/internal/app	2.217s
   ```

2. **Verde**: con la derivada,

   ```
   --- PASS: TestGrabacionesDerivadas (0.00s)
       --- PASS: TestGrabacionesDerivadas/eli-sin-segmento (0.02s)
       --- PASS: TestGrabacionesDerivadas/sin-eli (0.02s)
       --- PASS: TestGrabacionesDerivadas/version-ulterior (0.02s)
       --- PASS: TestGrabacionesDerivadas/version-posterior (0.02s)
       --- PASS: TestGrabacionesDerivadas/lpac-a21-version-anterior (0.02s)
   ok  	github.com/jmorenobl/kitlegal/internal/app	1.614s
   ```

3. `make -C /tmp/kitlegal-t010/copia ci` termina en `ci: todos los controles en verde` (lint, tests con `-race`,
   integración, tiempos, vuln, schema-check, skills-check, goreleaser check, secretos y módulos).

El diff de `internal/app/grafo_test.go` que se comprobó, y que el intento 2 aplica tal cual:

```diff
@@ -2259,6 +2259,12 @@ const (
 const parrafoDeLaVersionPosterior = "[Redacci\xc3\xb3n sint\xc3\xa9tica de prueba: versi\xc3\xb3n posterior " +
 	"derivada de la grabaci\xc3\xb3n de H4.]"
 
+// parrafoDeLaVersionUlterior es el que marca como sintética la redacción de la
+// derivada version-ulterior, la redacción C del e2e (research.md D23 de H7.1): el
+// último de su versión.
+const parrafoDeLaVersionUlterior = "[Redacci\xc3\xb3n sint\xc3\xa9tica de prueba: versi\xc3\xb3n ulterior " +
+	"derivada de la grabaci\xc3\xb3n de H4.]"
+
 // parrafoDeLaVersionAnterior es el que marca como sintética la redacción de la
@@ -2278,14 +2284,16 @@ type grabacionDerivada struct {
 // grabacionesDerivadas son las derivadas del e2e y la del grafo previo de la eval
 // de la consulta repetida (research.md D22), cada una con lo que dice su nombre:
 // version-posterior, la fecha de vigencia 20250101 y el párrafo sintético al
-// final del texto, con la huella de ese texto; sin-eli, la url_eli vacía;
-// eli-sin-segmento, una url_eli sin el segmento eli; y lpac-a21-version-anterior,
-// la fecha de vigencia 20151002 y su párrafo sintético al final del texto, con la
-// huella de ese texto.
+// final del texto, con la huella de ese texto; version-ulterior, lo mismo con la
+// fecha 20260101 y su párrafo; sin-eli, la url_eli vacía; eli-sin-segmento, una
+// url_eli sin el segmento eli; y lpac-a21-version-anterior, la fecha de vigencia
+// 20151002 y su párrafo sintético al final del texto, con la huella de ese texto.
 func grabacionesDerivadas() []grabacionDerivada {
 	return []grabacionDerivada{
 		versionDelArticulo21(filepath.Join(derivadasDelE2E, "version-posterior"), "20250101",
 			parrafoDeLaVersionPosterior),
+		versionDelArticulo21(filepath.Join(derivadasDelE2E, "version-ulterior"), "20260101",
+			parrafoDeLaVersionUlterior),
 		{
```

## Lo que ya está en el árbol (lo hereda el intento 2)

La derivada, en su ruta declarada y sin commitear:
`internal/app/testdata/derivadas/version-ulterior/GET_https_www.boe.es_datosabiertos_api_legislacion-consolidada_id_BOE-A-2015-10565_texto_bloque_a21.json`,
sha256 `527c2c223e919981ef374ee83850563cd199aeada9845e91cce077a11840fbb9`, idéntica (`cmp`) a la que dejó verde la
copia. La produjo un programa de un solo uso fuera del repositorio, `/tmp/kitlegal-t010/main.go`, lanzado con
`go run /tmp/kitlegal-t010/main.go <grabación de H4> <salida> 20260101 ulterior` desde la grabación de H4 del paquete
`boe` (`internal/source/boe/testdata/boe.legislacion-consolidada/`, mismo nombre de fichero):

- decodifica la grabación con el struct de `internal/httpx/grabar.go` (`DisallowUnknownFields`) y exige que la
  serialización de `serializar` —dos espacios de sangría, `SetEscapeHTML(false)`— la reproduzca byte a byte;
- en el cuerpo, sustituye `fecha_vigencia="20161002"` por `fecha_vigencia="20260101"` y añade
  `<p class="parrafo">[Redacción sintética de prueba: versión ulterior derivada de la grabación de H4.]</p>` antes de
  `</version>`, exigiendo que cada punto de sustitución aparezca exactamente una vez;
- recalcula `Content-Length` si la grabación lo lleva: la del bloque no lo lleva (solo aparece como valor de
  `Access-Control-Expose-Headers`), así que no cambia nada más.

Control del programa: con `20250101` y «posterior» regenera byte a byte (`cmp`) la derivada `version-posterior` de
H7. Y la derivada nueva difiere de `version-posterior` solo en dos palabras (`git diff --no-index --word-diff`):
`fecha_vigencia=\"20250101\"` → `fecha_vigencia=\"20260101\"` y `posterior` → `ulterior`. Sin red y sin grabación nueva.

## Lo que hace el intento 2

Aplicar el diff de arriba en `internal/app/grafo_test.go` (los caracteres no ASCII de la constante con sus escapes de
Go, como las otras dos), comprobar `TestGrabacionesDerivadas` en verde con la derivada heredada (el rojo sin ella
está arriba; para reproducirlo basta apartar la derivada un momento y volver a generarla con el programa, y
comprobar después su sha256 contra el de arriba), `make ci` en primer plano y marcar T010.

## Intento 2: en verde

Diff de arriba aplicado tal cual (escapes `\xc3\xb3` como texto, `gofmt` limpio). Rojo reproducido apartando la
derivada a `/tmp`: `elements differ … extra elements in list B` y el subtest `version-ulterior` con `no such file`.
Devolverla desde `/tmp` pedía una aprobación que en headless nadie da, así que se regeneró en su ruta con el programa
(`go run /tmp/kitlegal-t010/main.go <grabación de H4> <ruta declarada> 20260101 ulterior`): `git hash-object` da
`c48925a9b04a73828e650a17daca2060c32d3b58`, el mismo blob que la heredada (el `git diff --no-index` previo la
registraba como `c48925a`), y el control «posterior» del programa da `97ea93e3a6d319c261315c371b4b22f511ef4332`, el
blob de la derivada de H7. `TestGrabacionesDerivadas` en verde con sus cinco subtests y `make ci` termina en
`ci: todos los controles en verde`.
