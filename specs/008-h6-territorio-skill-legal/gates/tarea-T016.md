# T016 — en verde en el intento 1; una observación para la revisión final

**Estado**: marcada `[X]`. `make ci` en verde, en primer plano (`ci: todos los controles en verde`). El diff solo
toca rutas declaradas: `Makefile`, `internal/skills/jerarquia.go` e `internal/skills/jerarquia_test.go`, más
`tasks.md` y esta nota.

## Lo que entra

- `LeerJerarquia(contenido []byte) (Jerarquia, error)`: el lector común (`ValidarDocumentoYAML`) contra
  `schemas/jerarquia.yaml.json`, compilado una sola vez. Devuelve los defectos del lector tal cual, unidos y en el
  orden del documento, **sin el nombre del fichero delante**, igual que `LeerNormas`: quien lee el fichero lo pone,
  como ya hace `referenciasGeneradas` con el YAML de datos de cada referencia (`defectosDelError`).
- Tipos `Jerarquia`, `NivelNormativo` y `ReglaDeInterpretacion`. El valor del enumerado va en el campo `Codigo`
  (etiquetas `nivel` y `regla`), así no se repite `nivel.Nivel`.
- Tests: `TestLeerJerarquia` (valido, nivel-fuera-del-enumerado, niveles-en-otro-orden, regla-desconocida,
  clave-desconocida, clave-repetida), `TestEsquemaDeJerarquia` (compila; en cada lista, el enumerado es el de
  data-model §4.2 y `prefixItems` trae cada valor una vez, en su orden, con `minItems` igual y `items: false`) y
  `TestJerarquiaDelRepositorio` (el fichero real: los cinco niveles y las cuatro reglas en su orden, exigidos por
  igualdad con listas no vacías, así que no puede pasar en vacío).
- Rojo antes del código: el paquete de test no compilaba. Dos mutantes temporales, ya restaurados con
  `git checkout`, confirman que los controles fallan: un enumerado de niveles reordenado y ampliado hace fallar
  `TestEsquemaDeJerarquia/niveles`, y una clave `regla` repetida en `data/jerarquia.yaml` hace fallar
  `TestJerarquiaDelRepositorio` con `reglas/2: regla repetido en las líneas 54 y 55`.
- `make skills-check` gana `TestJerarquiaDelRepositorio`. Se comprobó ejecutando esa expresión `-run` con `-v`.

## Observación: los errores de lectura y compilación del esquema no tienen test

`compilarEsquemaDeJerarquia(ruta)` sigue la forma de `compilarEsquemaDeNormas` y `compilarEsquemaDelTerritorio`.
Pero sus dos ramas de error (carpeta en lugar del fichero, JSON que no compila) no tienen test, y las de los otros dos
sí: `TestCompilarEsquemaDeNormasDesdeUnaRuta` y `TestCompilarEsquemaDelTerritorioDesdeUnaRuta`, a través de
`internal/skills/export_test.go`. T016 no declara esa ruta (T007 sí la declaró para esto mismo), y la ruta constante del
paquete no produce nunca esos errores.

Para cerrarlo hacen falta dos cambios: una línea en `export_test.go`
(`CompilarEsquemaDeJerarquia = compilarEsquemaDeJerarquia`) y `TestCompilarEsquemaDeJerarquiaDesdeUnaRuta` en
`jerarquia_test.go`, con los mismos dos casos que sus hermanos. **Queda anotado para la revisión final del hito**.
No afecta a `make ci` ni al comportamiento.
