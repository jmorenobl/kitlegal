# Varios bloques a la vez

Cuándo se leen varios bloques de una norma con una sola orden, y cómo se escribe esa orden, con su comprobación
detrás, en Bash, en PowerShell y con herramientas.

## Cuándo

Solo cuando hacen falta varios bloques a la vez y todos salen del índice. `kitlegal boe articulos` los devuelve en el
orden pedido, y detrás va `kitlegal graph check` de esa misma norma y esos mismos bloques, en la misma orden.

## La orden

```bash
kitlegal boe articulos <norma> <bloques>... --json && kitlegal graph check <norma> <bloques>... --json
```

En PowerShell 5.1, que no tiene `&&`, la misma orden es:

```powershell
kitlegal boe articulos <norma> <bloques>... --json; if ($LASTEXITCODE -eq 0) { kitlegal graph check <norma> <bloques>... --json }
```

Con herramientas, `boe_articulos` con `norma` y `bloques`, y detrás `graph_check` con esa `norma` y esos mismos
bloques en `bloques`.
