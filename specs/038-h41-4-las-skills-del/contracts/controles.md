# Contrato: los controles de `make ci`

Todo en el paquete `evals`, en ficheros `_test.go`, sin nada exportado: ningún paquete, tipo, función de producción,
orden, bandera ni objetivo de `make` nuevo. Los corre `make test` (`go test -race -shuffle=on … ./...`), dentro de
`make ci`, sin red ni modelo. Las cifras y los rojos de §1.3, §2.3 y §3.3 se verificaron en una copia (research V6-V8).

## 1. `TestDescripcionesDelProducto` (FR-001, FR-002, FR-004; SC-001, SC-003)

Fichero nuevo: `internal/evals/descripciones_del_producto_test.go`.

### 1.1 La comprobación

`defectosDeLasDescripcionesDelProducto(raiz string) ([]string, error)` lee, de la raíz:

- los textos de la variante, con `descripcionesCortas(raiz)`; sus defectos no son de esta comprobación
  (`TestDescripcionesCortas` los da) y se descartan;
- cada subdirectorio de `skills/`, en orden de nombre (`fs.ReadDir` sobre `os.DirFS(raiz)`), y su `SKILL.md`
  (`fs.ReadFile`), con `skills.LeerFrontmatter`.

Por cada skill, en este orden, da como mucho una línea de cada clase:

1. **El texto** (FR-001): si la variante no tiene texto para la skill,
   `skills/<skill>: no tiene descripción en laboratorio/producto-descripciones-cortas`; si lo tiene y la `description`
   leída difiere en un byte, `skills/<skill>/SKILL.md: la description no es la de laboratorio/producto-descripciones-cortas/<skill>.txt`.
2. **La forma** (FR-002): con las líneas de la clave que da `descripcionYCuerpo`, si no empiezan por
   `description: >-` y su salto o no son al menos tres (la de la clave y dos sangradas),
   `skills/<skill>/SKILL.md: la description no está en un bloque >- de más de una línea`.
3. **La longitud** (FR-004): si la `description` leída pasa de `maximoDeUnaDescripcionCorta` (450) runas,
   `skills/<skill>/SKILL.md: la description tiene <n> caracteres (máximo 450)`.

Y al final, si la suma de las longitudes de todas las skills de `skills/` pasa de `maximoDeLasDescripcionesCortas`
(1700), `skills: las descripciones suman <n> caracteres (máximo 1700)` (research D7). La forma y la longitud se miden
en toda skill de `skills/`, tenga o no texto en la variante.

El error, sin defectos, es el de un directorio o un fichero que no se puede leer o el de un `SKILL.md` cuyo
frontmatter no lee `skills.LeerFrontmatter`, con su ruta delante: defecto `inesperado` (ADR 0023), sin casos propios.
Los 900 caracteres y las etiquetas XML no son de esta comprobación: los da `ValidarFrontmatter` en
`TestSkillsDelRepositorio`, que no cambia.

### 1.2 El test

Un subtest sobre la raíz del repositorio, que exige cero líneas, y uno por caso de §1.3, cada uno sobre una copia de
`skills/` y de `laboratorio/producto-descripciones-cortas/` en `t.TempDir()` (`os.CopyFS`), con su mutante aplicado y
sus líneas exactas, en su orden (`compararLasLineasDeLosDefectos`). Los mutantes de un `SKILL.md` lo leen y lo
escriben con `leerFichero` y `escribirFichero`, exigen que cambie algo y empiezan por `t.Helper()` (research V6).

### 1.3 Los casos

| Caso | Mutante | Líneas |
|---|---|---|
| la rama | ninguno: la raíz del repositorio | ninguna |
| un carácter añadido (US1.2) | `normativa`: la primera «`lo dice la skill.`» pasa a «`lo dice la skill..`» | `skills/normativa/SKILL.md: la description no es la de laboratorio/producto-descripciones-cortas/normativa.txt` |
| un carácter quitado (US1.2) | `consulta-juridica`: «`escrito presentar.`» → «`escrito presentar`» | `skills/consulta-juridica/SKILL.md: la description no es la de laboratorio/producto-descripciones-cortas/consulta-juridica.txt` |
| un carácter cambiado (US1.2) | `plazos-y-territorio`: «`tiene un municipio.`» → «`tiene un municipio;`» | `skills/plazos-y-territorio/SKILL.md: la description no es la de laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt` |
| en un escalar de una línea (FR-002) | `jurisprudencia`: el `SKILL.md` que da `skillConLaDescripcion` con su texto de la variante | `skills/jurisprudencia/SKILL.md: la description no está en un bloque >- de más de una línea` |
| en un bloque de una línea (FR-002) | `plazos-y-territorio`: las líneas de la clave pasan a `description: >-`, salto, dos espacios, el texto de la variante y un salto | `skills/plazos-y-territorio/SKILL.md: la description no está en un bloque >- de más de una línea` |
| una de 451 (FR-004) | `normativa`: «`lo dice la skill.`» seguido de 25 «é» | `skills/normativa/SKILL.md: la description no es la de laboratorio/producto-descripciones-cortas/normativa.txt` · `skills/normativa/SKILL.md: la description tiene 451 caracteres (máximo 450)` · `skills: las descripciones suman 1715 caracteres (máximo 1700)` |
| una suma de 1701 (FR-004) | `plazos-y-territorio`: «`tiene un municipio.`» seguido de 11 «é» (438 runas) | `skills/plazos-y-territorio/SKILL.md: la description no es la de laboratorio/producto-descripciones-cortas/plazos-y-territorio.txt` · `skills: las descripciones suman 1701 caracteres (máximo 1700)` |
| una quinta skill sin texto (caso límite) | `skills/quinta/SKILL.md` = `---\nname: quinta\ndescription: >-\n  Una skill\n  de ensayo.\n---\n\n# Quinta\n` | `skills/quinta: no tiene descripción en laboratorio/producto-descripciones-cortas` · `skills: las descripciones suman 1710 caracteres (máximo 1700)` |

Cada sustitución es la primera aparición en el `SKILL.md`, que está en la `description` porque el frontmatter va
delante del cuerpo.

## 2. `TestConsultaJuridicaDesdeLaBase` (FR-003; SC-002)

Mismo fichero que §1.

### 2.1 Lo fijado

`huellasDeConsultaJuridicaEnLaBase`: la SHA-256 en hexadecimal de cada fichero de `skills/consulta-juridica/` en la
base de la rama, por su ruta relativa a la skill; la del `SKILL.md`, de su contenido sin las líneas de la clave
`description` (`descripcionYCuerpo`). En `ab951c22` es un solo fichero (research V3):

| Ruta | SHA-256 |
|---|---|
| `SKILL.md` | `bd089909060d0fd82c1c6d2ee4ae148f6f0a1b2b7d7eefd476d164dc610a205a` |

### 2.2 La comprobación

`defectoDeConsultaJuridica(raiz string) (string, error)` lee los ficheros de `skills/consulta-juridica/` de la raíz
con `leerLosFicheros`, quita al `SKILL.md` las líneas de la clave y, en orden de ruta sobre la unión de las dos
listas, da el primero que falta, sobra o difiere, o la cadena vacía:

- `skills/consulta-juridica/<ruta> falta: está en la base de la rama (ab951c22)`;
- `skills/consulta-juridica/<ruta> sobra: no está en la base de la rama (ab951c22)`;
- `skills/consulta-juridica/<ruta> difiere de la base de la rama (ab951c22)`.

El error, sin defecto, es el de `leerLosFicheros` (un directorio que no se puede recorrer): defecto `inesperado`.

### 2.3 Los casos

| Caso | Mutante | Defecto |
|---|---|---|
| la rama | ninguno: la raíz del repositorio | ninguno |
| el cuerpo (US2.3) | «`## Método`» → «`## Métodos`» | `skills/consulta-juridica/SKILL.md difiere de la base de la rama (ab951c22)` |
| `name` (US2.3) | «`name: consulta-juridica`» → «`name: consulta-juridico`» | `skills/consulta-juridica/SKILL.md difiere de la base de la rama (ab951c22)` |
| un fichero de más (US2.3) | `references/otra.md` con `# Otra\n` | `skills/consulta-juridica/references/otra.md sobra: no está en la base de la rama (ab951c22)` |
| un fichero de menos (US2.3) | sin `SKILL.md` | `skills/consulta-juridica/SKILL.md falta: está en la base de la rama (ab951c22)` |
| solo la clave (FR-003) | «`Punto de entrada de toda`» → «`Punto de entrada de cualquier`» | ninguno |

Cada caso, sobre su copia de §1.2.

## 3. `TestSkillsDelProductoDesdeControl` (FR-005; SC-002)

Fichero existente: `internal/evals/laboratorio_familias_test.go`. El único cambio en las suites congeladas de H39 a
H41.3 (FR-005).

### 3.1 Qué cambia

- `TestSkillsDelProductoDesdeControl` lee una vez los textos de la variante (`descripcionesCortas(raizDelRepositorio)`,
  `require.NoError`) y pasa a `comprobarElProductoDesdeControl` el de la skill, en un parámetro más.
- **«desde control»**: lo esperado (su base, sin lo generado, con su declaración) y lo obtenido (`skills/<skill>` sin
  lo generado) se comparan con `compararLosFicheros` como hoy, pero con el `SKILL.md` de los dos sin las líneas de la
  clave `description` (`descripcionYCuerpo`).
- **«frontmatter»**: el de su control y el de `skills/<skill>`, los dos sin las líneas de la clave, con `assert.Equal`
  como hoy.
- En los dos subtests, además, la `description` que lee `skills.LeerFrontmatter` de `skills/<skill>/SKILL.md` es el
  texto de la variante: `la description de skills/<skill>/SKILL.md es la de laboratorio/producto-descripciones-cortas/<skill>.txt`.
- Dos auxiliares del fichero: `sinLasLineasDeLaDescripcion` (los ficheros con el `SKILL.md` sin las líneas de la clave)
  y `exigirLaDescripcionDeLaVariante` (la aserción del punto anterior).
- Los comentarios de `TestSkillsDelProductoDesdeControl`, `comprobarElProductoDesdeControl` y `declaracionDelProducto`
  dicen que el frontmatter es el de su control con la `description` de la variante (FR-005 de H41.4).

### 3.2 Qué no cambia

`declaracionesDelProducto` (bases, cambios del cuerpo y movimientos), `sinLoGenerado`, `conLaDeclaracion`,
`compararLosFicheros` y sus mensajes, el subtest «movimientos», `huellasDeLosControles` y `TestVariantesDeControl`.

### 3.3 Sus rojos (research V7)

Con el cambio, sale en rojo con la ruta del primer fichero que difiere, falta o sobra con un carácter cambiado en el
cuerpo, en `name` o en `metadata` de un `SKILL.md`, en una referencia, o con un fichero de más; y con la `description`
distinta de la variante. Una `description` con el mismo texto en otra forma no es suya: la da §1 (FR-002).

## 4. Lo que no cambia

Sin cambiar ningún valor esperado (FR-006, FR-009; research V4): `TestDescripcionesCortas`,
`TestDescripcionesCortasPorLaEstructura`, `TestConLaDescripcion`, `TestVariantesDeControl`,
`TestEleccionDeLosMutantes` (su mutante de la descripción de `plazos-y-territorio` busca `description: >-\n  `),
`TestSkillsDelRepositorio`, `TestArnesDeLasFamilias` y los guiones de `internal/app/testdata/script/`. El arnés
(`internal/evals/laboratorio*.go` que no son `_test.go`), `cmd/laboratorio`, `scripts/`, el `Makefile`,
`.golangci.yml` y `.github/` no cambian.

## 5. Uso, de fuera adentro

Las líneas de §1 y §2 las lee el mantenedor cuando `make ci` sale en rojo, una vez por ejecución: una línea por
defecto, de unos 60 bytes (la de la suma) a unos 130 (la del texto de `plazos-y-territorio`), con la skill o la ruta. Como mucho tres por skill y una de la suma, 13 con cuatro
skills, y una sola en §2; no dependen del uso del kit (normas, bloques, asuntos) y dejan de darse cuando `skills/`
vuelve a ser lo declarado.
