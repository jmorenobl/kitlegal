# T006: por qué no quedó en verde

## Intento 1: tarea mal delimitada, redelimitada en tasks.md (2026-10-24)

El trabajo de la tarea está hecho y queda en el árbol, que hereda el intento siguiente:

- `skills/legal-core/SKILL.md` reescrito según contracts/skills-reescritas.md, «legal-core»: título y frase,
  `## Referencias` (las seis, una línea y el único enlace de cada una), `## Protocolo` (pasos 1 a 11 con lo que
  enumera «Riesgos»), `## Cómo se presenta el territorio` (las tres líneas de cobertura), `## Cómo se presenta un
  plazo` (las nueve formas `⚠` sin cambio), `## Comandos` (párrafo y región generada sin cambio) y `## Reglas` (las
  nueve; la 7 con la línea sin consulta sola en su bloque `text`). Cuerpo: 167 líneas y 10 860 runas, medidas desde la
  línea que sigue al `---` que cierra el frontmatter (v1.2.3: 283 y 19 225). Descripción: 888 caracteres, la del
  borrador del contrato.
- `skills/legal-core/references/municipios.md` (63 líneas) y `skills/legal-core/references/calculo_de_plazos.md`
  (68 líneas), con los apartados de la tarea, sin líneas `⚠`, sin enlaces, sin órdenes y sin normas «Ley N/AAAA».
- `CHANGELOG.md`: entrada de `legal-core` v1.3 en «Cambiado».
- Tres líneas en gates/supuestos.md.

### Lo que sale en rojo

`make fmt-check lint test`: fmt-check y lint en verde; en `test`, un solo test en rojo en todo el repositorio
(`go test -count=1 ./internal/evals/` lista solo `--- FAIL: TestMatrizDeLasPropuestas`):

```
        	            	-  01-territorio-municipio-cubierto.yaml: medida: rutas de su conjunto (skills/legal-core/references/cobertura_por_materia.md, skills/legal-core/references/cobertura_por_materia_renombrada.md)
        	            	+  01-territorio-municipio-cubierto.yaml: medida: rutas de su conjunto (skills/legal-core/references/calculo_de_plazos.md, skills/legal-core/references/calculo_de_plazos_renombrada.md)
        	            	 ## etiqueta-evals-sin-nada-elegido
        	Test:       	TestMatrizDeLasPropuestas
        	Messages:   	la matriz de cada ejecución, byte a byte
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/evals	138.044s
make: *** [test] Error 1
```

`make test-integration`: el mismo rojo, el mismo test. `make test-tiempos vuln schema-check skills-check
goreleaser-check secrets mod-verify mod-tidy-check`: en verde. `make skills-sync` no cambia nada (la región de la tabla
de comandos es la de v1.2.3) y `go test -count=1 -run '^TestSkillsEmpotradas$' .` sale `ok`.

### Causa

`mutantesDeLaEleccion` (`internal/evals/eleccion_test.go`) toma para los mutantes `referencia-de-legal-core` y
`referencia-renombrada` el primer `.md` de `skills/legal-core/references` por orden de nombre
(`ficheroDelRepositorio(..., esMarkdown)`). Con la referencia a mano `calculo_de_plazos.md`, que el contrato nombra
así, el primero deja de ser `cobertura_por_materia.md` (generada) y pasa a ser una a mano: las elecciones son las
mismas (12 de legal-core medidas, 0 reutilizadas), pero el golden de la matriz, byte a byte, lleva la ruta. El test
dice que el mutante es el cambio de una referencia **generada** («que elige lo mismo que un cambio de una referencia
generada de su skill»); el de la referencia a mano es otro mutante, `referencia-a-mano-de-legal-core` (T002).

No se puede arreglar dentro de las rutas congeladas del intento: `internal/evals/eleccion_test.go` no es el
`x_test.go` de un `x.go` declarado. La línea de T006 en tasks.md queda redelimitada con esa ruta.

### Arreglo, comprobado en una copia desechable

Copia del árbol con los cambios sin commitear (`rsync -a --exclude .git … $TMPDIR/copia-t006/`) y la herramienta Edit
sobre la copia. El mutante elige la primera referencia **generada** por orden de nombre, con
`skills.DatosDeLaReferencia`, y el golden no cambia:

```diff
@@ -13,6 +13,8 @@
 
 	"github.com/stretchr/testify/assert"
 	"github.com/stretchr/testify/require"
+
+	"github.com/jmorenobl/kitlegal/internal/skills"
 )
 
@@ -136,7 +138,7 @@
 func mutantesDeLaEleccion(t *testing.T, repositorio fs.FS, base arbolDelRepositorio) []mutanteConNombre {
 	t.Helper()
 
-	referencia := ficheroDelRepositorio(t, repositorio, "skills/legal-core/references", esMarkdown)
+	referencia := referenciaGenerada(t, repositorio, "legal-core")
 	cli := ficheroDelRepositorio(t, repositorio, "internal/cli", esCodigoSinTests)
 	docs := ficheroDelRepositorio(t, repositorio, "docs", esMarkdown)
 
@@ -312,7 +314,32 @@
 func esMarkdown(nombre string) bool {
 	return strings.HasSuffix(nombre, ".md")
 }
+
+// referenciaGenerada es la ruta de la primera referencia generada de la skill,
+// por orden de nombre: la que abre con la cabecera «generado desde» que lee
+// skills.DatosDeLaReferencia. El mutante de la referencia de legal-core y el de
+// la renombrada cambian una generada, como en H30; el de la referencia a mano,
+// una a mano (FR-012 y FR-014 de H31). Desde H31, la primera referencia de
+// legal-core por orden de nombre es a mano.
+func referenciaGenerada(t *testing.T, repositorio fs.FS, skill string) string {
+	t.Helper()
 
+	directorio := path.Join("skills", skill, "references")
+
+	return ficheroDelRepositorio(t, repositorio, directorio, func(nombre string) bool {
+		if !esMarkdown(nombre) {
+			return false
+		}
+
+		contenido, err := fs.ReadFile(repositorio, path.Join(directorio, nombre))
+		require.NoError(t, err)
+
+		_, generada := skills.DatosDeLaReferencia(contenido)
+
+		return generada
+	})
+}
+
 // evalsDeLasSkills son las evals del listado que son de alguna de las skills,
```

Sobre la copia, con el arreglo:

- `go -C <copia> test -race -tags=integration -count=1 -run '^(TestMatrizDeLasPropuestas|TestEleccionDeLosMutantes|TestConjuntoDeCadaEval|TestEvalsDelRepositorio)$' ./internal/evals/`: `ok`.
- `make -C <copia> fmt-check lint`: `0 issues.`
- `make -C <copia> test`: en rojo solo `TestCalculoDeLaDecisionDelCheckout`, `TestHuellaDeLasEvals` y
  `TestFilaConReutilizadas`, porque la copia no es un repositorio git (`git rev-parse HEAD` y `git ls-files` salen con
  `fatal: not a git repository`); en el repositorio los tres pasan en este intento.

### Lo que queda para el intento siguiente

Aplicar el diff de arriba a `internal/evals/eleccion_test.go` (rojo hoy en el repositorio, verde con él), correr
`make ci` en primer plano (por objetivos: `make fmt-check lint test` · `make test-integration` · `make test-tiempos vuln
schema-check skills-check goreleaser-check secrets mod-verify mod-tidy-check`) y marcar T006 `[X]` si sale en verde. El
resto de la tarea no necesita cambios.

## Intento 2: en verde (2026-10-24)

Se aplicó a `internal/evals/eleccion_test.go` el diff de arriba, tal cual (`referenciaGenerada`, con
`skills.DatosDeLaReferencia`, para los mutantes `referencia-de-legal-core` y `referencia-renombrada`); el resto del
árbol del intento 1 no cambió. Comprobado antes, sobre ese árbol: las nueve formas `⚠` del plazo y la región generada
de «Comandos» son byte a byte las de v1.2.3 (`git diff HEAD` no toca ninguna línea que empiece por `⚠` ni ninguna de
la región), cada referencia tiene un solo enlace en `SKILL.md`, la descripción mide 888 caracteres y el cuerpo 167
líneas y 10 860, y las dos referencias a mano llevan los apartados de la tarea sin `⚠`, sin enlaces, sin «Ley N/AAAA»
ni `BOE-A-`.

En verde, en primer plano y en este orden: `make skills-check`; `make skills-sync` sin cambios en el árbol;
`go test -count=1 -run '^TestSkillsEmpotradas$' . ./internal/app/`;
`go test -race -count=1 -run '^(TestMatrizDeLasPropuestas|TestEleccionDeLosMutantes|TestConjuntoDeCadaEval|TestEvalsDelRepositorio)$' ./internal/evals/`;
y `make ci` por objetivos: `make fmt-check lint test` (`0 issues.`, `internal/evals` ok en 139,5 s),
`make test-integration` (ok) y `make test-tiempos vuln schema-check skills-check goreleaser-check secrets mod-verify
mod-tidy-check` (sin vulnerabilidades, sin fugas, `go mod tidy -diff` vacío). T006 queda `[X]`.
