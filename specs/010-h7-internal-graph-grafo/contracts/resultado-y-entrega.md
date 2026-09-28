# Contrato: operaciones en el `Resultado` y entrega por el kernel

FR-020 a FR-035. Decisiones: [../research.md](../research.md) D1-D7. Tipos: [../data-model.md](../data-model.md) §1-§2.

## 1. Lo que declara un applet

- `schema.Resultado` gana el campo `Grafo schema.Observado`. Su valor cero es válido y es lo que devuelve todo applet
  que no emite (`skills`, los de ejemplo, los verbos de `boe` que no emiten, los de `graph`): **ningún applet
  implementa nada nuevo** y el contrato `Applet` del ADR 0005 —`Nombre`, `Descripcion`, `Verbos`, y `Ejecutar` en los
  `Argumentos` de cada verbo— no cambia (FR-020, FR-047).
- `Observado.Vigencia` es la vigencia de la consulta, o cero si el applet no la declara (FR-065).
- Una operación es un `schema.Nodo`, una `schema.Arista` o un `schema.Texto`; ninguna lleva fuente, url ni fecha
  (FR-021). `TestSobre` fija los cinco campos de `Resultado`: `Procedencia`, `Datos`, `Legible`, `Ensayo`, `Grafo`.

## 2. El puerto

```go
// internal/core/graphstore.go
type GraphStore interface {
	Apply(ctx context.Context, lote Lote) error
}
```

`Apply` es transaccional e idempotente (FR-022); lo implementan `*graph.Almacen` (SQLite) y `graph.Nulo` (descarta).
El puerto no tiene `Close`: cada entrega abre, aplica y cierra.

## 3. Cuándo se entrega

| Invocación | ¿Entrega? | Por qué es así por construcción |
|---|---|---|
| Termina con 0, sin `--no-graph`, con operaciones | sí, al almacén del registro | `Montador.Emitir` entrega tras presentar (§4) |
| Termina con 0 y `--no-graph` | se entrega a `graph.Nulo`, que no resuelve ninguna ruta ni abre nada (FR-031) | `Main` elige el almacén con la bandera |
| Termina con 0 sin operaciones | no; no se resuelve la ruta ni se abre `world.db` (FR-030) | `Emitir` mira `len(Operaciones)` antes de nada |
| Termina con un código distinto de 0 | no (FR-032) | un fallo sale por `emitirFallo`, que no entrega |
| `--dry-run` | no (FR-034) | sin fallo, `Main` vuelve antes de `Emitir`; con fallo, `Emitir` recibe el resultado vacío y sale por `emitirFallo` (`internal/app/main.go`, research V17) |
| `--offline` | sí (FR-035) | la bandera no llega a la entrega |
| La ayuda, `--describe`, `version` | no | no llegan a `Emitir` |
| Montar el sobre o presentarlo falla | no | `Emitir` sale por `emitirFallo` antes de entregar |
| Registro sin almacén (valor cero, tests, preparación de la caché de las evals) | no | `Montador.Grafo` nulo |
| El registro no se construye (`fallarAlArrancar`) | no | montador sin almacén |

## 4. Cómo se entrega (`internal/cli`)

```go
type Montador struct {
	Ahora Reloj           // nulo: el reloj del sistema
	Grafo core.GraphStore // nulo: no entrega
}

func (m Montador) Emitir(ctx context.Context, p Presentador, enJSON bool, res schema.Resultado, err error) int
func LoteDe(sobre schema.Sobre, observado schema.Observado) core.Lote // internal/cli/entrega.go
```

1. `Emitir` monta el sobre y lo presenta exactamente como hoy (sobre, tabla o `Legible`). Si la presentación falla,
   el desenlace es el de hoy y no se entrega nada.
2. Presentado con éxito, si `m.Grafo` no es nulo y `res.Grafo.Operaciones` no está vacío, `LoteDe` toma **del sobre
   presentado** `Fuente`, `URL` y `FechaConsulta.Format(time.RFC3339Nano)` —el mismo texto que el sobre escribe, también
   cuando la fecha la puso el reloj del montador (research V2, V18)— y del `Observado` la vigencia y las operaciones.
3. `m.Grafo.Apply(ctx, lote)` con el contexto que `Main` crea con el **plazo de `--timeout` de la invocación**
   (FR-014): el mismo instante límite con el que se ejecutó el applet.
4. Si `Apply` falla, `Emitir` escribe por `p.Aviso` **una línea**:

   ```text
   kitlegal: lo observado no ha llegado al grafo del mundo: <causa>
   ```

   donde `<causa>` es el mensaje del error (el de contracts/almacen-world-db.md §6, que empieza por `grafo: `) con
   cada salto de línea `\n` sustituido por un espacio (un `\r` se deja tal cual). Por ejemplo, con un `world.db` que no
   es una base: `kitlegal: lo observado no ha llegado al grafo del mundo: grafo: "<ruta>" no es una base de datos
   utilizable; no se modifica` (en una sola línea). El código de salida sigue
   siendo 0 y la salida estándar no cambia (FR-033). El error de escribir esa línea no se propaga (research D7).
5. `Emitir` devuelve el código de siempre.

`internal/app/main.go`: `desenlace` gana `limite` —el instante en que vence el plazo de `--timeout` con el que se
ejecutó el applet, `ctx.Deadline()`— y `sinGrafo`; `Main` crea el contexto de la entrega con
`context.WithDeadline(context.Background(), fin.limite)` (un desenlace que no ejecutó el applet no trae límite y su
contexto nace vencido, sin consecuencia: no hay resultado correcto que entregar) y el montador con `Grafo` =
`graph.Nulo{}` con la bandera, o el almacén del registro sin ella. `fallarAlArrancar` usa un montador sin almacén y
`context.Background()`.

## 5. El registro

```go
func (r *Registro) EntregarAlGrafo(almacen core.GraphStore)
```

Uno nuevo sustituye al anterior; nulo deja el registro sin entrega, como el valor cero. `RegistroDeProduccion`
(`internal/app/registro.go`) y `registroDeE2E` (`internal/app/ejemplo/kitlegal-e2e/main.go`) llaman
`EntregarAlGrafo(graph.Nuevo())`; la preparación del grafo previo de una eval, `EntregarAlGrafo(graph.Nuevo(
graph.ConDirectorio(<caché de la sesión>)))` (contracts/evals-y-skill.md §3). Construirlo no resuelve ninguna ruta ni
abre nada.

## 6. `--no-graph`

La ayuda de la bandera pasa a ser exactamente `No entrega al grafo del mundo nada de lo que observa la invocación.`
(FR-031), y el comentario de `cli.Globales.SinGrafo` deja de decir que el grafo llega en H17. La ayuda de un verbo
(Kong) parte esa frase en dos líneas porque no cabe en la columna: la cadena exacta la fija `TestGlobales`, y el
guion la busca con `\s+` entre sus palabras. Los verbos de `graph` no
entregan nada: la bandera no cambia lo que leen (FR-031, contracts/applet-graph.md §2).

## 7. Qué lo vigila

| Control | Test |
|---|---|
| Cinco campos del `Resultado`; `schema` sin importaciones nuevas | `TestSobre`, `TestContexto` (`internal/core/schema`) |
| Lote con la procedencia del sobre, también fechada por el reloj; nada sin operaciones; nada en fallo; una línea de aviso y el mismo código; error de escritura del aviso sin efecto | `TestEntregaDelMontador` (`internal/cli/entrega_test.go`) con un `GraphStore` y un `Presentador` de prueba |
| `--no-graph` elige `graph.Nulo`; el plazo es el de `--timeout`; `--dry-run` no entrega | `TestEntregaDelKernel` (`internal/app/main_test.go`) |
| La ayuda de `--no-graph` | `TestGlobales` (`internal/cli/globales_test.go`) y el guion `h7-grafo-applet` |
| Procedencia de cada nodo, arista y texto igual a la del sobre (FR-089, SC-002) | `TestLaEntregaLlevaLaProcedenciaDelSobre` (`internal/app/grafo_test.go`) y el guion `h7-grafo-memoria` |
