# Contrato: la sincronía reparte cada tipo y cada fila a su paquete

Lo consume `make skills-sync` (`scripts/skills-sync.sh` → `TestSkillsDelRepositorio -args -regenerar-skills`) y lo
comprueba `make skills-check`. Nada exportado cambia de firma; todo vive en `internal/skills/regiones.go`.

## 1. La región `recursos_y_plazos` de un paquete (FR-001, FR-002)

`filasDeLosRecursosYPlazos(datos, region)`:

1. Lee `data/plazos.yaml` y `data/normas.yaml` como hoy (`usar`).
2. Para cada tipo, en el orden del fichero, `paquetesDelTipo`. Junta un defecto por tipo con 0 o con 2 o más paquetes.
3. Con algún defecto: `&defectoDeLosDatos{fichero: "plazos.yaml", causa: errors.Join(defectos...)}`. La región no se
   regenera, la referencia no se escribe y `make skills-sync` sale con error, como con cualquier defecto de
   `data/plazos.yaml` de hoy.
4. Sin defectos: `filasDeLosTipos` (sin cambiar) sobre los tipos cuyo paquete es `region.paquete`, en su orden. Un tipo
   con unidades sigue dando una fila por unidad.

La forma de cada fila no cambia. Ejemplo, una fila de hoy de `administrativo.md` (100 bytes con su salto, medido con
`wc -c`; la de `contencioso`, con dos preceptos, es la más larga):

```text
| recurso de alzada | `alzada` | 1 mes | `lpac` | art. 122.1 LPAC [BOE-A-2015-10565, bloque a122] |
```

## 2. Los dos defectos (FR-003, FR-004, §C.1)

Por cada skill con una referencia que lleva la región, con el prefijo de los defectos de los datos (`<skill>: `), una
línea por tipo:

```text
plazos-y-territorio: data/plazos.yaml: el tipo prueba-fiscal no es de ningún paquete: ninguna norma de sus preceptos lleva vertical en data/normas.yaml
plazos-y-territorio: data/plazos.yaml: el tipo mixto es de más de un paquete: administrativo, ensayo
```

La primera, 153 bytes con su salto (medido con `wc -c`); la segunda, unos 75 más el código y los paquetes. Un defecto de los datos sale una vez por
skill (`sinRepetir`, `regiones.go:782`), aunque la skill tenga varias referencias con la región. Deja de darse en cuanto
las normas de los preceptos propios del tipo comparten un `vertical`. Nadie la lee más de una vez por `make skills-sync`.

Lo que no es defecto (Fuera de alcance): un paquete con tipos y sin ninguna referencia con la región; sus tipos no
salen en ninguna.

## 3. La región `cobertura_por_materia` y `filasSinRegion` (FR-005, FR-006, FR-007)

- `paqueteDeLaFila(entrada, normas)`: vacío si la norma del precepto no está en `data/normas.yaml`, como hoy (la línea
  no podría nombrarla); si no, el `vertical` de la norma que la rige, si la fila la lleva y tiene `vertical`, y si no el
  de la norma de su precepto.
- `filasDeLaCobertura`: la fila sale en la referencia de `paqueteDeLaFila(entrada, normas)`. La línea, con
  `region.linea.linea(entrada, norma)` y la norma del precepto, como hoy.
- `filasSinRegion(cubiertos)`: defecto, con el texto de hoy, para cada fila cuyo `paqueteDeLaFila` es vacío o no está
  entre los `cubiertos`:

```text
data/cobertura.yaml: la fila de «procedimientos de ensayo que la ley común deja aparte» no la genera ninguna región cobertura_por_materia
```

## 4. Lo de hoy, igual (FR-008)

Con `data/` de `main`: los siete tipos y las nueve filas son de `administrativo` (data-model §2 y §3) y
`administrativo.md` sale byte a byte la de hoy (visto en P1: `diff` vacío). Lo comprueban `TestRegiones/sincronia`,
`make skills-check` y `TestSkillsDelProductoDesdeControl`.

## 5. `TestVerticalDeEnsayo` crecida (FR-019, FR-020, SC-002)

Ficheros de la vertical: data-model §7. `instalarLaVertical` añade también `tipos:` de su `data/plazos.yaml`. El test,
sobre la copia sincronizada, exige:

| Comprobación | Rojo con el mutante (US2.4) |
|---|---|
| la fila del tipo `ensayo` está en `recursos_y_plazos` de `plazos-y-territorio/references/ensayo.md` y en ninguna otra referencia de la copia | «un tipo asignado a todas las referencias» (y la regla de hoy, US1.4): sale también en `administrativo.md` |
| la fila «procedimientos de ensayo que la ley común deja aparte» está en `ensayo.md` y en ninguna línea de `administrativo.md` | «una fila asignada por su precepto»: va a `administrativo.md` |
| subtest `sin-la-region`: con la `ensayo.md` de `plazos-y-territorio` sin la región de la cobertura, `skills.Regenerar` da el defecto de §3 para esa fila | «`filasSinRegion` con la regla vieja»: la fila es de lo administrativo por su precepto y no da defecto |
| tras sincronizar, con una línea de 69 caracteres y su salto al final del cuerpo de `jurisprudencia`, `normativa` y `plazos-y-territorio`, `skills.Regenerar` da 0 defectos | «la línea de `normativa` sin el movimiento»: 10 926 + 70 + 70 > 11 000 (P2: 11 066 con dos líneas de paquete) |
| el diff de la copia, 0 `.go`, y ningún `SKILL.md` cambiado fuera de las marcas | lo de hoy (`TestVerticalDeEnsayo/mutante`) |
| la fila de lo cubierto de la vertical es la última de la región y va detrás de la que era la última del repositorio | la fila puesta en otro sitio; ya no depende de que esa sea la del territorio |

Cuerpos con la línea de ensayo y 70 más (spec, Assumptions): `normativa` unos 10 891, `jurisprudencia` 10 983,
`plazos-y-territorio` 10 827.

## 6. Los subtests de los defectos (US3, SC-003)

En `TestRegiones/defectos`, sobre `internal/skills/testdata/regiones/plazos.yaml` (copia de los tipos y cómputos de hoy,
research D10) con las normas sintéticas de `normasDeLasRegiones`, que ganan una norma de prueba con `vertical: ensayo`:

| Subtest | Cambio en el `plazos.yaml` de prueba | Esperado |
|---|---|---|
| `tipo-sin-paquete` | un tipo con su precepto en la norma sintética sin `vertical` (la LGT de `normasDeLasRegiones`) | el defecto «no es de ningún paquete» con su código, y ninguna referencia regenerada |
| `tipo-de-dos-paquetes` | un tipo con un precepto de la norma de prueba delante de dos de la LPAC, para que el orden alfabético y sin repetir solo salga si se ordenan | «es de más de un paquete: administrativo, ensayo» |
| `computo-de-otro-paquete` | un cómputo nuevo que cita la norma de prueba, y un tipo de la LPAC con él | 0 defectos; el tipo sale en `administrativo.md` |
| `fila-con-su-precepto-en-una-norma-que-no-esta` (en `cobertura.yaml`, con la `ensayo.md` con su región) | una fila con la norma de prueba por norma que la rige y su precepto en una norma que no está en `normasDeLasRegiones` | el defecto de §3 para esa fila, como en `main` |

Sin la regla (mutante de US3.4), los dos primeros dan 0 defectos y el tercero, si se cuentan los preceptos del cómputo,
el de dos paquetes.
