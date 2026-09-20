# Contrato: `internal/core/ids` — código INE y DIR3

El paquete de identificadores que H6 estrena. Dominio puro: no importa `os`, `io`, `io/fs`, `log`, `net/http` ni
`database/sql`, ni ningún paquete interno fuera de `internal/core/schema` (research.md V5, V6, V7).

## 1. Superficie

```go
// Código INE de municipio.
type CodigoINE struct{ /* campos privados */ }

func AnalizarCodigoINE(entrada string) (CodigoINE, error)
func AnalizarCodigoINEConDigito(entrada string) (codigo CodigoINE, digito byte, err error)
func (c CodigoINE) String() string      // cinco cifras con sus ceros
func (c CodigoINE) Provincia() string   // dos cifras
func (c CodigoINE) ComprobarDigito(declarado, oficial byte) error

// Código DIR3 de un ayuntamiento.
type DIR3 struct{ /* campos privados */ }

func AnalizarDIR3(entrada string) (DIR3, error)
func DIR3DeAyuntamiento(c CodigoINE, digito byte) DIR3
func (d DIR3) String() string           // L01PPMMMD
func (d DIR3) CodigoINE() CodigoINE
func (d DIR3) Digito() byte
```

Nada más: ELI, ECLI, CELEX y NIF entran con sus hitos (FR-034). Un test de superficie (`TestSuperficieDeIds`, al
estilo de `internal/httpx/superficie_test.go`) exige que lo exportado sea exactamente esto. **`ComprobarDigito` es
método de `CodigoINE`**, no función de paquete: la misma forma que fijan [data-model.md](../data-model.md) §1.1 y
research.md D8.

## 2. Gramáticas

| Identificador | Forma | Rechaza |
|---|---|---|
| Código INE | `^[0-9]{5}$` con provincia `01`-`52` y municipio `001`-`999` | vacío, no cifras ASCII, otra longitud, provincia `00` o `>52`, municipio `000` |
| Código INE con dígito | `^[0-9]{6}$`, las cinco primeras como arriba | ídem |
| DIR3 de ayuntamiento | `^[Ll]01[0-9]{6}$`, normalizado a `L` mayúscula | cualquier otra letra, otra longitud, cifras no ASCII |

Las cifras se comprueban byte a byte, sin `regexp`, como `internal/source/boe/ids.go`. **Los patrones de los esquemas
de `data/territorio/` y estas gramáticas se atan con un test**, el subtest
`TestTerritorioDelRepositorio/gramaticas` de `internal/skills` —donde ya vive el lector de esquemas—, al estilo de
`TestGramaticasCoincidenConBoe` (research.md V38): aceptan y rechazan exactamente lo mismo.

## 3. Dígito de control

`ids` **no calcula** el dígito: lo compara. El oficial viene de `data/territorio/municipios.yaml` (research.md D9).

```go
func (c CodigoINE) ComprobarDigito(declarado, oficial byte) error   // nil si coinciden; error de argumentos si no
```

El mensaje dice el dígito recibido y el oficial. Quien resuelve decide el orden: primero «¿está el municipio en la
relación?» (si no, código 3), después el dígito (si no coincide, código 2), porque sin municipio no hay dígito oficial
con el que comparar.

## 4. Propiedades

1. **Idempotencia**: `AnalizarCodigoINE(c.String())` devuelve el mismo valor; ídem para `DIR3` (FR-032).
2. **Ida y vuelta estable**: para toda entrada aceptada, `String()` de lo analizado vuelve a analizarse al mismo
   valor, y `String()` de eso es la misma cadena.
3. **Relación INE ↔ DIR3**: `DIR3DeAyuntamiento(c, d).CodigoINE() == c` y `.Digito() == d`; y para todo DIR3
   analizado, `DIR3DeAyuntamiento(d.CodigoINE(), d.Digito()).String() == d.String()` (FR-031).
4. **Sin `panic` y sin valor por omisión**: toda entrada inválida devuelve error (FR-033).

## 5. Errores

Un tipo propio con `Clase() schema.Clase` → `schema.ClaseArgumentos`, que el kernel traduce a **código 2**
(research.md V18). El error nombra la entrada con `%q` y dice qué tiene de malo.

**El fuzz no puede comprobar el código numérico**: `depguard` deniega `internal/cli` también a los `_test.go` del
dominio (research.md V7). La comprobación de clase va en el fuzz (`schema.ClaseArgumentos`); la del código 2, en
`internal/app` (`TestCodigosDeTerritorio`) y en el e2e.

## 6. Fuzz y corpus

Un fichero de test por fichero de código, **también para el fuzz**: cada objetivo vive con el analizador que ejercita
y no hay ningún `ids.go` ni `ids_test.go` en el paquete (research.md D8; inventario de tests de
[plan.md](../plan.md)). El corpus se indexa por el **nombre del objetivo**, no por el del fichero que lo declara.

| Objetivo | Fichero | Corpus versionado |
|---|---|---|
| `FuzzCodigoINE` | `internal/core/ids/ine_test.go` | `internal/core/ids/testdata/fuzz/FuzzCodigoINE/` |
| `FuzzCodigoDIR3` | `internal/core/ids/dir3_test.go` | `internal/core/ids/testdata/fuzz/FuzzCodigoDIR3/` |

- Semillas con `F.Add`: `"28074"`, `"280748"`, `"01001"`, `"52001"`, `""`, `"2807"`, `"2807a"`, `"00074"`, `"28000"`,
  `"L01280748"`, `"l01280748"`, `"L0128074"`, `"X01280748"`.
- El corpus versionado **se ejecuta en `make ci`** sin `-fuzz`, como corpus semilla (research.md V3). El fuzz largo se
  lanza a mano, como el de H4: `go test -run '^$' -fuzz '^FuzzCodigoINE$' -fuzztime 30s ./internal/core/ids/`.
- Propiedades que comprueba cada objetivo: ningún `panic`; lo aceptado cumple la gramática de referencia y lo
  rechazado no; toda entrada aceptada tiene ida y vuelta estable (§4); todo error es de clase `argumentos` y nombra la
  entrada.
- El corpus versionado entra en una tarea `[datos]` por vivir bajo `testdata/` (research.md V10); no provoca pausa,
  porque no está bajo el `testdata/` de la raíz (V9, FR-086).

## 7. Qué lo vigila

| Control | Test | En `make ci` |
|---|---|---|
| Gramáticas, normalización e idempotencia | `TestAnalizarCodigoINE`, `TestAnalizarDIR3`, `TestIdaYVuelta` | sí |
| Relación INE ↔ DIR3 | `TestDIR3DeAyuntamiento` | sí |
| Dígito de control | `TestComprobarDigito` | sí |
| Superficie exportada | `TestSuperficieDeIds` | sí |
| Fuzz con corpus versionado | `FuzzCodigoINE`, `FuzzCodigoDIR3` | sí (semillas) |
| Gramáticas atadas a los esquemas de datos | `TestTerritorioDelRepositorio/gramaticas` | sí |
| Dominio puro | `TestArquitectura` R1 y `depguard` (automáticos para todo paquete nuevo bajo `internal/core`) | sí |
| Cobertura ≥ 85 % en `internal/core/**` | `codecov.yml` | en la propuesta de cambio |
