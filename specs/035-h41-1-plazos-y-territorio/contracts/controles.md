# Contrato: los controles de `make ci` que cambia H41.1

FR-004 a FR-008, FR-012; SC-001, SC-002, SC-005. Research V4, V5, V7, V13, D1, D4, D5, D11, D12; data-model §3 y §4.

## 1. `TestSkillsDelProductoDesdeControl` (`internal/evals/laboratorio_familias_test.go`; FR-005)

### 1.1 La declaración

`declaracionDelProducto` gana el campo `desde`: la variante versionada de la familia sobre la que se aplican los
cambios y los movimientos; vacío, la de control (las de `normativa` y `jurisprudencia` no lo llevan y no cambian). La
de `plazos-y-territorio`, en `declaracionesDelProducto`:

```go
// La forma sin condición de H41 con M1 y nada más (FR-001 a FR-005 de
// H41.1; ADR 0053, punto 1).
skillDeLaFamiliaDePlazos: {
	desde: "plazos-y-territorio-sin-condicion",
	cambios: []cambioDelCuerpo{
		{
			antes: "    `⚠ REDACCIÓN A FECHA:`. Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has " +
				"podido leerlo, sin\n" +
				"    afirmar qué dice.\n",
			despues: "    `⚠ REDACCIÓN A FECHA:`.\n",
		},
	},
	movimientos: []movimientoALaReferencia{
		{
			texto:          siSaleConCuatroOCinco,
			referencia:     "calculo_de_plazos.md",
			enLaReferencia: "\n" + siSaleConCuatroOCinco + "\n",
		},
	},
},
```

con `siSaleConCuatroOCinco = "Si sale con 4 o 5, **`⚠ PLAZO:` se mantiene**: di aparte que no has podido leerlo, sin
afirmar qué dice."` (104 caracteres; research V3).

### 1.2 Lo que comprueba

Por cada skill de `declaracionesDelProducto` —`jurisprudencia`, `normativa` y `plazos-y-territorio`—, en subtests con su
nombre:

1. **«desde control»**: la base —`desde`, o el control si está vacío— sin lo generado, con los cambios y los movimientos,
   es `skills/<skill>` sin lo generado, ruta a ruta y byte a byte. Falla con la ruta del primer fichero que difiere,
   falta o sobra: `skills/plazos-y-territorio/<ruta> difiere de lo que dan plazos-y-territorio-sin-condicion y su
   declaración` (`falta: está en …`, `sobra: no está en …`); en `normativa` y `jurisprudencia`, «su control» en lugar del
   nombre, como hoy (research D5).
2. **«movimientos»**: cada texto movido está, con los blancos juntados, en el cuerpo de la base («lo que llega a <ruta>
   sale del SKILL.md de <base>»), no en el de `skills/<skill>`, y sí al final de su referencia.
3. **«frontmatter»**: el de `skills/<skill>` es el de su control, byte a byte.

Que lo generado sea lo que da `make skills-sync` lo sigue comprobando `TestSkillsDelRepositorio`. El subtest de H41 que
comparaba `skills/plazos-y-territorio` con su control, con lo generado, sale, y los comentarios de
`declaracionesDelProducto` y del test dejan de decir que esa skill es su control.

## 2. `TestRemisionesDelRepositorio` (`internal/skills/remisiones_test.go`; FR-007)

### 2.1 Sin exclusiones, con los fallos que admite cada skill

Sale `fueraDeLasRemisiones` y su `continue`. Entra:

```go
// fallosAdmitidosDelProducto son, por skill de skills/, los que la regla da con
// su SKILL.md y el test admite; una skill que no está aquí no admite ninguno.
var fallosAdmitidosDelProducto = map[string][]string{
	// La viñeta de jerarquia_normativa.md queda como en v0.9.0 (FR-002 de H41.1;
	// ADR 0053, punto 1 y «Lo que cuesta»): ninguna eval llega a su condición.
	"plazos-y-territorio": {
		"plazos-y-territorio: references/jerarquia_normativa.md: " + remisionSinElVerbo,
	},
}
```

Por cada skill de `skills/` con `references/`, en orden de nombre, `assert.Equal(t, fallosAdmitidosDelProducto[skill],
fallosDeLaRegla(skill, segmentos, referencias), "%s: la regla da exactamente los fallos que admite", skill)`: en rojo si
la regla da otro fallo —el mensaje de `testify` lista los obtenidos, cada uno con la skill, la referencia y la regla— y
si deja de dar el admitido. Después, la lista del plan, como hoy. `fallosDeControl` y el subtest «control» no cambian.

### 2.2 La lista del plan de `plazos-y-territorio`

En `remisionesDelPlan`, en este orden (research V4):

```go
"plazos-y-territorio": {
	"- [jerarquia_normativa.md](references/jerarquia_normativa.md): para el nivel que regula algo, con el " +
		"autonómico de la `comunidad` de `data`, y la norma que prevalece.",
	"Antes de la primera orden o llamada, lee [municipios.md](references/municipios.md): qué hacer si falta el " +
		"municipio o no sale con 0, qué trae `data`, la plantilla del territorio y qué dice su cobertura.",
	"Antes del 8, lee [calculo_de_plazos.md](references/calculo_de_plazos.md): los municipios que cuentan, el " +
		"plazo sin calcular y la lectura en la fecha de inicio.",
},
```

Una frase distinta da `plazos-y-territorio: la remisión <n> no es la del plan`, con las dos listas. La de
`administrativo.md` está en la región de los paquetes: la regla la mira (cumple), la lista no.

## 3. `TestArnesDeLasFamilias`: `origenesDeLaVariante` (`internal/evals/laboratorio_sondeo_test.go`; research D12)

Hoy, la skill de cada sesión de `plazos-y-territorio-producto` tiene que ser, byte a byte, la de `skills/` y también la
de `laboratorio/plazos-y-territorio/plazos-y-territorio-control/skills`, porque en H41 eran la misma (research V13). Tras
el hito, `origenesDeLaVariante` devuelve, en toda `<skill>-producto`, solo `skills/`:

```go
// origenesDeLaVariante son los directorios de skills cuya skill de la familia
// tiene que ser, ruta a ruta y byte a byte, la de las sesiones de la variante
// (contracts/arnes.md §5 de H41): laboratorio/<skill>/<variante>/skills en una
// variante versionada, y skills/ del árbol en <skill>-producto (FR-022 de H41).
func origenesDeLaVariante(skill, variante string) []string {
	if variante != skill+sufijoDelProductoDelTest {
		return []string{filepath.Join(raizDelRepositorio, "laboratorio", skill, variante, "skills")}
	}

	return []string{filepath.Join(raizDelRepositorio, "skills")}
}
```

Nada más del test cambia: los casos, las sesiones que abre cada invocación, que cada sesión vea solo su skill enlazada a
la copia de su variante, el servidor del modo herramienta y `sondeo.json`. El arnés, que copia `skills/<skill>` para
`<skill>-producto`, no cambia (FR-008).

## 4. Lo que no cambia y sigue en verde (FR-006, FR-008, FR-012; SC-005)

Sin cambiar ningún valor esperado: `TestVariantesDeControl` (las huellas SHA-256 de los cinco ficheros de
`plazos-y-territorio-control`), `TestVariantesDePlazosYTerritorio` (cada forma, su declaración sobre control),
`TestSkillsDelRepositorio` (tope de 11 000 y derivas), `TestVerticalDeEnsayo` (con la línea de la vertical),
`TestRecetasDelSondeo`, `TestLaboratorioFueraDelProducto`, `TestFrases`, `TestArnesDeNormativa`,
`TestArnesDelLaboratorio`, `TestVariantesDeNormativa`, los de la tabla (`TestTablaDeLasFamilias`, `TestLecturaConCd`,
`TestTablaDeNormativa`, `TestTablaDelLaboratorio`), y las declaraciones de `normativa` y `jurisprudencia` en
`TestSkillsDelProductoDesdeControl`. Verificado sobre una copia con el estado de destino (research V5, V13, V14).

## 5. Mutaciones de un solo cambio (SC-001)

Cada una, sobre el repositorio tras el hito, pone en rojo el test de su fila con ese mensaje (research V7; quickstart.md,
escenario 5):

| Mutación | Test en rojo | Lo que dice |
|---|---|---|
| un byte añadido a `skills/plazos-y-territorio/references/calculo_de_plazos.md` | `TestSkillsDelProductoDesdeControl` | `skills/plazos-y-territorio/references/calculo_de_plazos.md difiere de lo que dan plazos-y-territorio-sin-condicion y su declaración` |
| un byte al principio de `skills/normativa/references/con_herramientas.md` (la rama con `desde` vacío) | `TestSkillsDelProductoDesdeControl/normativa/desde_control` | `skills/normativa/references/con_herramientas.md difiere de lo que dan su control y su declaración` |
| un byte al principio de `skills/jurisprudencia/references/documento_aportado.md` (la rama con `desde` vacío) | `TestSkillsDelProductoDesdeControl/jurisprudencia/desde_control` | `skills/jurisprudencia/references/documento_aportado.md difiere de lo que dan su control y su declaración` |
| un byte añadido a un fichero de `plazos-y-territorio-control` | `TestVariantesDeControl` | `<ruta> difiere: su huella es …` |
| un byte añadido al `SKILL.md` de `plazos-y-territorio-lista-con-verbo` | `TestVariantesDePlazosYTerritorio` | `<ruta> difiere: la variante es la que da su declaración sobre plazos-y-territorio-control…` |
| «sin calcular» → «sin contar» en la remisión de `calculo_de_plazos.md` | `TestRemisionesDelRepositorio` | `plazos-y-territorio: la remisión 3 no es la del plan` |
| «Lee » delante del enlace de `jerarquia_normativa.md` | `TestRemisionesDelRepositorio` | `plazos-y-territorio: la regla da exactamente los fallos que admite` |
| 49 caracteres más en el cuerpo de `plazos-y-territorio` (10 931; 11 001 con la vertical) | `TestVerticalDeEnsayo` | `plazos-y-territorio: SKILL.md: cuerpo de 11001 caracteres (máximo 11000)` |

Un cuerpo de más de 11 000 caracteres sin la vertical lo da `TestSkillsDelRepositorio` con el mismo defecto, como hoy.
Las mutaciones del bloque de FR-013 de `h41-skills-instaladas.txtar`, una por cada `grep`, están en
contracts/aceptacion.md §2.
