# T007: por qué no quedó en verde

## Intento 1: tarea mal delimitada, redelimitada en tasks.md (2026-10-24)

El trabajo de la tarea está hecho y queda en el árbol, que hereda el intento siguiente:

- `skills/boe-legislacion/SKILL.md` reescrito según contracts/skills-reescritas.md, «boe-legislacion»: título,
  `## Protocolo` con su párrafo y los pasos 1 a 5 (Identificar la norma, Leer, Completar, Responder y De un precepto no
  leído, nada), `## Cómo se cita`, `## Redacción en una fecha` (las tres formas fechadas juntas en un bloque `text`),
  `## Redacción modificada` (la línea `⚠ REDACCIÓN MODIFICADA:` en su bloque `text`), `## Comandos` con la región
  generada sin cambio y `## Reglas`, las ocho con su número (la 3 con las tres etiquetas de aviso, la 7 sin su frase y
  enlazando `lecturas_que_fallan`, la 8 con la línea sin consulta sola en su bloque `text`). En el cuerpo siguen
  `kitlegal boe indice BOE-A-2015-10565 --json` y la orden de `a21` en Bash y en PowerShell. Cuerpo: 151 líneas y
  10 976 runas, medidas desde el byte que sigue al `---` que cierra el frontmatter (v0.2.2: 283 y 24 439).
  Descripción: la «Descripción propuesta» del contrato, 878 caracteres.
- Las siete referencias a mano nuevas, entre 24 y 51 líneas, sin `⚠`, sin enlaces y con un título y un párrafo que
  dice lo que lleva cada una: `con_herramientas.md`, `buscar_la_norma.md` (con `kitlegal boe buscar <términos> --json`),
  `varios_bloques.md` (las órdenes de `articulos` en Bash y en PowerShell), `lecturas_que_fallan.md` (la frase de la
  regla 7 sola en un bloque `text`), `como_se_cita.md`, `redaccion_en_una_fecha.md` (con
  `kitlegal boe articulo BOE-A-1985-5392 a22 --fecha 2005-06-15 --json`) y `redaccion_modificada.md`. Sale el ejemplo
  de `analisis`. `normas.md` no cambia.
- `CHANGELOG.md`: entrada de `boe-legislacion` v0.3 en «Cambiado».
- Dos líneas en gates/supuestos.md.

Cada regla de la tabla (research.md, «Tabla de reglas», B01–B132) está en el sitio que da la tabla; repasado a mano
contra el texto final.

### Lo que sale en rojo

`make fmt-check lint test`: fmt-check y lint en verde (`0 issues.`); en `test`, en rojo solo dos subtests de todo el
repositorio, en `internal/evals` (141,7 s):

```
--- FAIL: TestControlesSobreLasReferenciasAMano (0.04s)
    --- FAIL: TestControlesSobreLasReferenciasAMano/sin-las-ordenes-de-articulos (0.01s)
        conjunto_test.go:3760:
            	Error Trace:	/Users/jorge/Projects/kitlegal/internal/evals/conjunto_test.go:3835
            	            				/Users/jorge/Projects/kitlegal/internal/evals/conjunto_test.go:3846
            	            				/Users/jorge/Projects/kitlegal/internal/evals/conjunto_test.go:3760
            	Error:      	Should not be zero, but was 0
            	Messages:   	/tmp/claude-501/TestControlesSobreLasReferenciasAManosin-las-ordenes-de-articul2702305941/001/skills/boe-legislacion/SKILL.md lleva lo que el mutante quita
    --- FAIL: TestControlesSobreLasReferenciasAMano/ordenes-de-articulos-en-una-referencia (0.01s)
        conjunto_test.go:3760:
            	Error Trace:	/Users/jorge/Projects/kitlegal/internal/evals/conjunto_test.go:3835
            	            				/Users/jorge/Projects/kitlegal/internal/evals/conjunto_test.go:3857
            	            				/Users/jorge/Projects/kitlegal/internal/evals/conjunto_test.go:3760
            	Error:      	Should not be zero, but was 0
            	Messages:   	/tmp/claude-501/TestControlesSobreLasReferenciasAManoordenes-de-articulos-en-un4048900809/001/skills/boe-legislacion/SKILL.md lleva lo que el mutante quita
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	141.697s
make: *** [test] Error 1
```

`make test-integration`: el mismo rojo, los mismos dos subtests. `make test-tiempos vuln schema-check skills-check
goreleaser-check secrets mod-verify mod-tidy-check`: en verde. `make skills-sync` no cambia nada (ninguna línea de la
región generada difiere de `HEAD`), `make skills-check` sale `ok`, y `go test -count=1 -run '^TestSkillsEmpotradas$' .`
y `go test -count=1 -run '^TestOrdenesParaPowerShell$' ./internal/evals/` salen `ok`.

### Causa

Los dos mutantes de `mutantesDeLasOrdenes` (`internal/evals/conjunto_test.go`, de T003) quitan de `SKILL.md` las líneas
que llevan `kitlegal boe articulos <norma> <bloques>... --json` (`quitarDelSkillMd`, que exige al menos una). El
contrato y la propia tarea llevan esas órdenes, con su forma para PowerShell, a `references/varios_bloques.md`: el
`SKILL.md` ya no las lleva, y los dos mutantes caen en el `require.NotZero` antes de mutar nada. El control sí funciona
sobre la skill entera: `ordenes-para-powershell` de `TestEvalsDelRepositorio` sale en verde porque la orden de Bash de
`articulos` está en `varios_bloques.md`.

No se puede arreglar dentro de las rutas congeladas del intento: `internal/evals/conjunto_test.go` no es el `x_test.go`
de un `x.go` declarado. La línea de T007 en tasks.md queda redelimitada con esa ruta.

### Arreglo, comprobado en una copia desechable

Copia del árbol con los cambios sin commitear (`rsync -a --exclude .git --exclude .codegraph --exclude dist` a
`$TMPDIR/copia-t007/`) y la herramienta Edit sobre la copia. Los mutantes de las órdenes de `articulos` las toman de
donde están ahora: el primero las lleva de `varios_bloques.md` al final de `SKILL.md`, en un bloque `text`, y no da
ningún defecto; el segundo las quita de `varios_bloques.md` y da el defecto de hoy. `conLasLineasMovidas` pierde el
parámetro `enBloqueDeTexto`, que tras el cambio solo recibe `true`:

```diff
@@ -3820,7 +3820,13 @@
 func quitarDelSkillMd(t *testing.T, dir string, casa func(linea string) bool) string {
 	t.Helper()
 
-	ruta := filepath.Join(dir, ficheroDeLaSkill)
+	return quitarDelFichero(t, filepath.Join(dir, ficheroDeLaSkill), casa)
+}
+
+// quitarDelFichero quita del fichero de la ruta cada línea con la que casa, que
+// tiene que haber al menos una, y devuelve las quitadas, juntas.
+func quitarDelFichero(t *testing.T, ruta string, casa func(linea string) bool) string {
+	t.Helper()
 
 	var quedan, quitadas strings.Builder
 
@@ -3838,28 +3844,45 @@
 	return quitadas.String()
 }
 
-// sinLasLineas es el mutante que quita del SKILL.md cada línea con la que casa.
-func sinLasLineas(casa func(linea string) bool) func(t *testing.T, dir string) {
+// enBloqueDeTexto son las líneas, que terminan en un salto de línea, en un
+// bloque de código text, como se enseña una forma fija.
+func enBloqueDeTexto(lineas string) string {
+	return aperturaDeBloqueDeTexto + "\n" + lineas + cierreDeBloqueDeCodigo + "\n"
+}
+
+// sinLasLineas es el mutante que quita del fichero de la skill, su ruta dentro
+// de ella, cada línea con la que casa.
+func sinLasLineas(fichero string, casa func(linea string) bool) func(t *testing.T, dir string) {
 	return func(t *testing.T, dir string) {
 		t.Helper()
 
-		quitarDelSkillMd(t, dir, casa)
+		quitarDelFichero(t, filepath.Join(dir, fichero), casa)
 	}
 }
 
 // conLasLineasMovidas es el mutante que lleva del SKILL.md a la referencia a
-// mano cada línea con la que casa: tal cual o, con enBloqueDeTexto, en un bloque
-// de código text, como se enseña una forma fija.
-func conLasLineasMovidas(casa func(linea string) bool, enBloqueDeTexto bool) func(t *testing.T, dir string) {
+// mano cada línea con la que casa, en un bloque de código text.
+func conLasLineasMovidas(casa func(linea string) bool) func(t *testing.T, dir string) {
 	return func(t *testing.T, dir string) {
 		t.Helper()
 
-		movidas := quitarDelSkillMd(t, dir, casa)
-		if enBloqueDeTexto {
-			movidas = aperturaDeBloqueDeTexto + "\n" + movidas + cierreDeBloqueDeCodigo + "\n"
-		}
+		escribirLaReferencia(t, dir, referenciaAMano, "# Lo que decía SKILL.md\n\n"+
+			enBloqueDeTexto(quitarDelSkillMd(t, dir, casa)))
+	}
+}
 
-		escribirLaReferencia(t, dir, referenciaAMano, "# Lo que decía SKILL.md\n\n"+movidas)
+// conLasLineasLlevadasAlSkillMd es el mutante que lleva cada línea con la que
+// casa del fichero de la skill, su ruta dentro de ella, al final de su SKILL.md,
+// en un bloque de código text.
+func conLasLineasLlevadasAlSkillMd(fichero string, casa func(linea string) bool) func(t *testing.T, dir string) {
+	return func(t *testing.T, dir string) {
+		t.Helper()
+
+		llevadas := quitarDelFichero(t, filepath.Join(dir, fichero), casa)
+		ruta := filepath.Join(dir, ficheroDeLaSkill)
+
+		require.NoError(t, os.WriteFile(ruta, []byte(string(contenidoDelFichero(t, ruta))+"\n"+
+			enBloqueDeTexto(llevadas)), 0o600))
 	}
 }
 
@@ -3892,8 +3915,9 @@
 
 // mutantesDeLasOrdenes son los del control 3: en una referencia a mano, la
 // orden de Bash de articulo sin su forma para PowerShell, con ella en otra
-// referencia y con ella detrás; y las órdenes de articulos de SKILL.md llevadas
-// a una referencia, tal cual, o quitadas.
+// referencia y con ella detrás; y las órdenes de articulos, que desde H31 están
+// en references/varios_bloques.md (contracts/skills-reescritas.md de H31,
+// «boe-legislacion»), llevadas de ahí a SKILL.md, o quitadas.
 func mutantesDeLasOrdenes() []mutanteDeUnControl {
 	const (
 		enBash       = "# Lecturas\n\n```bash\n" + bashDeArticulo + "\n```\n"
@@ -3905,6 +3929,7 @@
 	conLasDeArticulos := func(linea string) bool {
 		return strings.Contains(linea, "kitlegal boe articulos <norma> <bloques>... --json")
 	}
+	referenciaDeArticulos := path.Join(directorioDeLasReferencias, "varios_bloques.md")
 
 	return []mutanteDeUnControl{
 		{
@@ -3924,12 +3949,12 @@
 			}),
 		},
 		{
-			nombre: "ordenes-de-articulos-en-una-referencia", control: controlDeLasOrdenes, skill: skillBoeLegislacion,
-			mutar: conLasLineasMovidas(conLasDeArticulos, false),
+			nombre: "ordenes-de-articulos-en-skill-md", control: controlDeLasOrdenes, skill: skillBoeLegislacion,
+			mutar: conLasLineasLlevadasAlSkillMd(referenciaDeArticulos, conLasDeArticulos),
 		},
 		{
 			nombre: "sin-las-ordenes-de-articulos", control: controlDeLasOrdenes, skill: skillBoeLegislacion,
-			mutar: sinLasLineas(conLasDeArticulos),
+			mutar: sinLasLineas(referenciaDeArticulos, conLasDeArticulos),
 			defectos: []string{"boe-legislacion: ninguna orden de Bash lee con «kitlegal boe articulos» y comprueba " +
 				"detrás con kitlegal graph check"},
 			exactos: true,
@@ -3948,7 +3973,7 @@
 			nombre: "linea-sin-consulta-movida-en-" + skill, control: controlDeLaLineaSinConsulta, skill: skill,
 			mutar: conLasLineasMovidas(func(linea string) bool {
 				return strings.TrimSpace(linea) == lineaSinConsultaDeLasSkills
-			}, true),
+			}),
 			defectos: []string{skill + ": SKILL.md no dice en un bloque text la línea «" + lineaSinConsultaDeLasSkills +
 				"»"},
 			exactos: true,
@@ -3969,7 +3994,7 @@
 	movida := func(nombre string, control controlDeContenido, forma *regexp.Regexp, defecto string) {
 		mutantes = append(mutantes, mutanteDeUnControl{
 			nombre: nombre, control: control, skill: skillBoeLegislacion,
-			mutar: conLasLineasMovidas(forma.MatchString, true), defectos: []string{defecto},
+			mutar: conLasLineasMovidas(forma.MatchString), defectos: []string{defecto},
 		})
 	}
```

Sobre la copia, con el arreglo y el texto final de la skill:

- `go -C <copia> test -race -count=1 -run '^(TestControlesSobreLasReferenciasAMano|TestEvalsDelRepositorio|TestOrdenesParaPowerShell|TestProsaDeLaSkill|TestLineaSinConsultaDeUnaSkill)$' ./internal/evals/`:
  `ok`.
- `go -C <copia> test -count=1 -v -run '^TestControlesSobreLasReferenciasAMano$/articulos' ./internal/evals/`: `PASS`
  de `ordenes-de-articulos-en-skill-md` y de `sin-las-ordenes-de-articulos`.
- `make -C <copia> fmt-check lint`: `0 issues.`

### Lo que queda para el intento siguiente

Aplicar el diff de arriba a `internal/evals/conjunto_test.go` (rojo hoy en el repositorio, verde con él), correr
`make ci` en primer plano (por objetivos: `make fmt-check lint test` · `make test-integration` · `make test-tiempos vuln
schema-check skills-check goreleaser-check secrets mod-verify mod-tidy-check`) y marcar T007 `[X]` si sale en verde. El
resto de la tarea no necesita cambios.

## Intento 2: en verde (2026-10-24)

Hereda el árbol del intento 1 sin tocar la skill, sus referencias ni `CHANGELOG.md`, y aplica a
`internal/evals/conjunto_test.go` el diff de arriba, tal cual. Antes del diff, `go test -count=1 -run
'^TestControlesSobreLasReferenciasAMano$/articulos' ./internal/evals/` da los dos subtests en rojo con el `require.NotZero`
de la nota; con el diff, `ordenes-de-articulos-en-skill-md` y `sin-las-ordenes-de-articulos` en `PASS`.

Comprobado en esta sesión, todo en primer plano:

- `make skills-check`: `ok` en `internal/app`, `internal/skills` e `internal/evals`; `make skills-sync` deja el árbol
  como estaba (el mismo diff de `SKILL.md`, `normas.md` sin cambio).
- `TestSkillsEmpotradas` (1), `TestOrdenesParaPowerShell` (15) y `TestEvalsDelRepositorio` (24): en verde.
- `make ci` por objetivos: `make fmt-check lint test` (28 paquetes `ok`, lint `0 issues.`), `make test-integration` y
  `make test-tiempos vuln schema-check skills-check goreleaser-check secrets mod-verify mod-tidy-check`: los tres con
  salida 0.
- Cuerpo de `SKILL.md`: 151 líneas y 10 976 runas tras el frontmatter; descripción por debajo de 900 caracteres.

T007 queda marcada `[X]`.
