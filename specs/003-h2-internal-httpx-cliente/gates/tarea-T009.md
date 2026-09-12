# T009 — nota de delimitación (intento 1, cerrada en verde)

La tarea quedó **en verde y marcada `[X]`**. Esta nota registra lo único que no
estaba previsto en su enunciado: sus rutas declaradas eran incompletas, y la
tarea las ha ampliado con dos ficheros de test, con la razón que sigue.

## Qué faltaba

`T009` inserta el decorador `robots` en la **cadena compartida** por todo el
paquete, por encima de los reintentos y del ritmo. Eso cambia dos cosas que los
tests de los dos escalones vecinos —`ritmo_test.go` (T007) y `reintentos_test.go`
(T008)— medían dando por supuesto que la primera petición que el cliente hace a
un sitio es la del recurso:

1. **El turno recién nacido del cupo de cada sitio ya no lo ocupa el recurso,
   sino el `robots.txt`** (FR-017: la obtención respeta el ritmo del sitio como
   cualquier otra petición). `TestRitmoNoRetrasaOtroSitio` medía que la operación
   contra el segundo sitio no espera nada; ahora espera **un** intervalo, el
   suyo, y lo que distingue «cupo por sitio» de «cupo compartido» es que no
   espere **dos**. La medida se ha reescrito así, con el mismo intervalo y el
   mismo margen.
2. **Contra un sitio que ya no escucha, lo primero que falla es su `robots.txt`,
   no el recurso**, y un fallo de transporte que sobrevive a los reintentos
   deniega con clase 5 en vez de declarar la fuente caída con clase 4 (FR-015
   prevalece sobre FR-029, y así lo fija el propio enunciado de T009). La
   subprueba de `TestReintentosAgotados` que cierra el servidor perdía por eso su
   sujeto. Se ha conservado **entero** lo que medía —el fallo de transporte del
   recurso se reintenta tantas veces como un 5xx y termina en clase 4— añadiendo
   antes del cierre una petición que deja el `robots.txt` del sitio ya cacheado.

Ninguno de los dos ficheros de producto (`ritmo.go`, `reintentos.go`) se ha
tocado, ni se ha rebajado ninguna comprobación: las dos tablas siguen midiendo lo
mismo, con la aritmética que la cadena nueva impone.

No había forma de evitarlo desde las rutas declaradas: la alternativa —que la
obtención del `robots.txt` no ocupe turno en el sitio, o que un transporte
agotado no deniegue— contradice FR-015 y FR-017.

## Qué se ha hecho con ello

- La línea de **T009 en `tasks.md`** declara ahora los dos ficheros, con el
  porqué, y deja explícito que `ritmo.go` y `reintentos.go` no se tocan.
- `gates/tarea-actual.json` lleva las dos rutas nuevas en `.rutas`, de modo que
  el guardián de diff las juzgue declaradas y no coladas.
- El resto de la delimitación de la tarea se ha cumplido tal cual: ningún otro
  fichero fuera de sus rutas, nada bajo `testdata/` ni `schemas/`, ninguna
  supresión de lint (los cinco falsos positivos de `misspell` sobre términos
  españoles se han resuelto reescribiendo el término, como manda la nota de
  `tasks.md`), y sin plataforma remota.

## Verificación

`make ci` en verde, con `internal/httpx` al 97,5 % de cobertura y
`TestDependenciasDelBinario` intacto (supuesto S1: `go mod tidy -diff` limpio y
ningún módulo nuevo en el grafo más allá de `github.com/temoto/robotstxt`
v1.1.2).
