# Contrato: la tabla de comandos, sin lo que da la llamada (FR-001 a FR-004)

La genera `RenderizarTabla(applets, descripciones)` (`internal/skills/comandos.go`) desde `--describe`, y `make
skills-sync` la escribe entre sus marcas en el cuerpo de cada skill con `kitlegal-applets`. La lee el agente, en el
cuerpo, antes de su primera llamada.

## 1. Los bytes

Para cada applet, en el orden de `kitlegal-applets`:

```text
\n
### `kitlegal <applet>`\n
\n
| Orden | Herramienta | Qué hace |\n
|---|---|---|\n
| `<sintaxis de la orden>` | `<applet>_<verbo>` | <lo que hace> |\n      (una por verbo con fila, en su orden)
```

y, detrás de la última sección, una línea en blanco (`\n`). Cada celda, con la barra como `\|` y un salto como un
espacio, como hoy. Las mismas entradas dan siempre los mismos bytes.

Lo que ya no escribe: la cuarta columna («Qué devuelve en `data`», con `objeto`, `lista de objetos`, `sin forma
declarada` y sus claves), la línea «La orden y la herramienta de cada fila devuelven el mismo sobre: …» y la línea
«Banderas comunes: …».

**Ejemplo**, la región de `jurisprudencia` (`skills/jurisprudencia/SKILL.md` de `f89093dd`, 796 caracteres con sus dos
marcas):

```text
<!-- inicio de la tabla de comandos: generada desde --describe con make skills-sync, no editar -->

### `kitlegal cita`

| Orden | Herramienta | Qué hace |
|---|---|---|
| `kitlegal cita preparar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--texto <texto>]` | `cita_preparar` | Dice qué tiene que hacer la persona para encontrar una sentencia en el buscador del CENDOJ: la dirección y cada casilla con su valor, o la dirección de una búsqueda por texto. No consulta nada. |
| `kitlegal cita cotejar [<ecli>] [--roj <roj>] [--resolucion <resolucion>] [--fecha <fecha>] [--documento <documento>]` | `cita_cotejar` | Lee la ficha del documento del CENDOJ que trae la persona y dice si es la sentencia que se pidió. No consulta nada. |

<!-- fin de la tabla de comandos -->
```

## 2. Lo que no cambia (FR-002)

- `LeerDescripcionDeVerbo` lee del documento lo mismo y da los mismos defectos: `DescripcionDeVerbo.Devuelve`, `Sobre` y
  `Banderas` se siguen leyendo, aunque la tabla no los escriba (`TestDescripcionDeVerbo` y
  `probarDefectosDeLaDescripcion`, sin cambiar lo que esperan).
- `RenderizarTabla` sigue dando, sin tabla, «`<applet> <verbo>` no declara las mismas banderas globales que `<…>`» y
  «`<applet> <verbo>` no declara el mismo sobre que `<…>`» cuando dos verbos de la tabla difieren
  (`TestRenderizarTabla`, casos `banderas-distintas` y `sobre-distinto`, `comandos_test.go:975-990`, sin cambios), y los
  de hoy sin ningún applet o con uno sin verbos.

## 3. Lo que se retira de `comandos.go`

Sin uso tras §1, y `unused` lo daría (research D1, V3): `Devuelve.texto`, `lineaDelSobre`, `lineaDeLasBanderas`,
`comoCodigo`, `textoDeObjeto`, `textoDeListaDeObjetos`, `textoSinFormaDeclarada` y `valorDeUnaBandera`. Se quedan
`enumerar` y `Bandera.escrita`.

## 4. Los tests que fijan los bytes (FR-003)

| Test | Qué cambia | En rojo con la tabla de cuatro columnas |
|---|---|---|
| `TestRenderizarTabla` (`internal/skills/comandos_test.go`) | `columnasDeLaTabla`, `filaDeBuscar`, `filaDeArticulo`, `filaDeVacio` y las filas de cada caso, con tres celdas; `pieDeLaTabla` = `"\n"`; fuera `lineaDeLasBanderasDePrueba` | sí: la tabla esperada y la renderizada difieren |
| `TestRegenerarYComparar` y `TestCuerpoDeLaSkill` (`internal/skills/sincronia_test.go`) | `regionDeAlfa`, con las dos filas de tres celdas y sin las dos líneas | sí: la región de `alfa` regenerada y la esperada difieren |
| `TestOrdenesDeLasSkillsEmpotradas/control` (`internal/app/skills_test.go`) | la cabecera del fixture y las filas sin su cuarta celda; lo esperado, igual | no: el subtest lee la orden y la herramienta de cada fila y no la cabecera (research D11) |

`make skills-sync` rehace las tablas de `normativa`, `plazos-y-territorio` y `jurisprudencia` en la misma tarea, y
`make skills-check` sale con 0, sin defectos ni deriva (FR-004).

## 5. Tamaño y uso

Una tabla por pregunta en que se activa la skill, dentro del cuerpo, sin ninguna invocación: la `SKILL.md` no pide
`--describe`. Con los applets de hoy: `normativa` 1436 caracteres (2673 en `main`), `plazos-y-territorio` 929 (1498),
`jurisprudencia` 796 (1302) (research P1). Crece con los verbos de sus applets, de 150 a 400 caracteres por fila (las de hoy), no
con el uso: con meses de uso diario, cientos de normas y miles de bloques consultados, es la misma. Lo que deja de dar —las
claves de `data`, el sobre y las banderas— lo trae cada llamada en su resultado y, en el modo herramienta, el esquema de
salida que anuncia el servidor con cada herramienta (research V18); `--describe` de cada orden, para quien lo pida.
