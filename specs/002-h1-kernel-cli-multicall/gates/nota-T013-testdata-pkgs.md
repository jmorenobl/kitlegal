# T013 · Cómo se demostró que `TESTDATA_PKGS` está vivo

T013 solo toca el `Makefile`, así que su rojo → verde no puede ser un `_test.go`: el control se ejerce
como manda la obligación 3 del plan y el escenario 9 de `quickstart.md`, **sobre una copia desechable del
árbol fuera del repositorio** (`$(mktemp -d)`), introduciendo ahí el defecto deliberado que hoy no se
puede introducir en el árbol de trabajo —`internal/app/testdata/` es material protegido y T013 no es
`[datos]`—. Nada de este ejercicio entra en el repositorio: los defectos solo existieron en esa copia
bajo `/tmp`, que es desechable y queda fuera de todo control del proyecto.

## Lo que se ejerció, y en qué orden

Defectos plantados en la copia:

| Ruta | Defecto | Para qué |
|---|---|---|
| `internal/app/testdata/roto/roto.go` | `fmt.Println` | R5: la prohibición de escribir en la salida estándar alcanza al paquete hermano de la raíz de composición (D18, D19) |
| `internal/app/testdata/roto/otro.go` | ninguno (segundo fichero Go del paquete) | que la lista **no repita** el directorio |
| `internal/app/testdata/malformato/malformato.go` | importaciones desordenadas, espacios sobrantes | que `fmt` y `fmt-check` también desciendan ahí |
| `internal/app/testdata/script/guion.txtar` | — | que un subdirectorio **sin** ficheros Go quede fuera de la lista |

1. **Rojo** (con el `Makefile` de H0): `make lint` → `0 issues`, código 0. El defecto es invisible; es
   exactamente el agujero que describe D19 —los comodines de Go no descienden a `testdata`—.
2. **Verde** (con el `Makefile` de T013): la receta se expande a
   `golangci-lint run ./... ./internal/app/testdata/malformato ./internal/app/testdata/roto` —`roto` una
   sola vez pese a sus dos ficheros, `script/` ausente, orden estable por `sort`— y `make lint` falla
   nombrando la regla: ``use of `fmt.Println` forbidden because "…solo internal/render escribe en
   stdout"`` (forbidigo), más los avisos de `gofumpt`, `goimports` y `revive` del otro paquete.
   `make fmt-check` falla con el diff del fichero mal formateado y `make fmt` lo corrige de verdad.
3. **El caso vacío de hoy**: en el árbol real, donde `internal/app/testdata/` todavía no existe,
   `TESTDATA_PKGS` queda vacía y las tres recetas son literalmente las de H0. `make ci` en verde.

## Lo que este ejercicio descartó

El riesgo real era que `golangci-lint` se negara a lintar bajo un directorio `testdata` aunque se le
nombrara —tiene exclusiones de directorio por omisión en su historia—, lo que habría obligado a tocar
`.golangci.yml`, que no es ruta de esta tarea. Con la versión fijada en `tools/golangci-lint/go.mod`
(v2.13.2) **no ocurre**: enumerado explícitamente, el paquete se linta y se formatea como cualquier otro.
Por eso T013 se cierra sin ninguna exclusión nueva ni excepción de lint.

La segunda mitad de D19 —añadir estos mismos paquetes al **test de arquitectura**— es T016, no T013.
