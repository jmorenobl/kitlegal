# Lo ensayo

El paquete de ensayo en `plazos-y-territorio`: el de `TestVerticalDeEnsayo`, que prueba que una materia entra sin Go y
sin editar ningún `SKILL.md`. No es una materia: sus normas, la Ley 1/1900 (`BOE-A-1900-1`) y la Ley 2/1900
(`BOE-A-1900-2`), son datos de un test. Las tablas las genera `make skills-sync` desde `data/`, y en la vertical van
vacías.

## Lo que la ley regula aparte

Los motivos de la línea; el primero es el de la fila con norma:

```text
se rige por <norma> (<identificador de la norma>), no por el procedimiento común (<precepto> [<identificador>, bloque <id>])
lo regula aparte la ley, no el procedimiento común (<precepto> [<identificador>, bloque <id>])
```

<!-- inicio de cobertura_por_materia: generada desde data/cobertura.yaml y data/normas.yaml con make skills-sync, no editar -->
<!-- fin de cobertura_por_materia -->

## Recursos y plazos

El tipo que se pasa a `plazos calcular` es el de su código, y la línea del plazo lo nombra como su primera columna.

<!-- inicio de recursos_y_plazos: generada desde data/plazos.yaml y data/normas.yaml con make skills-sync, no editar -->
<!-- fin de recursos_y_plazos -->
