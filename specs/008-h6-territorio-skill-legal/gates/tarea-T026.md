# T026: por qué no quedó en verde

## Intento 1 (2026-09-22)

**Estado: sin marcar.** La rama está publicada y la propuesta de cambio, abierta
([#39](https://github.com/jmorenobl/kitlegal/pull/39)). La primera ejecución de evals, 35714659803, salió `aprobado` en
sus dos trabajos, y sus datos y salidas están en `gates/evals-cierre.md`. Pero `ci` salió en rojo sobre la misma
cabeza, `47f3090`. Su arreglo toca código fuera de `specs/008-h6-territorio-skill-legal/`, así que esta ejecución no
podrá cubrir la cabeza que se fusione (SC-015). Además, una respuesta incumple SC-013, que queda por decidir (§ 2). No
se fusionó nada, no se empujó a `main`, no se forzó nada ni se borró ninguna rama, etiqueta o release. La ejecución de
evals no se relanzó.

Hecho en este intento: prerrequisitos en verde (sesión de `gh`, secreto `CLAUDE_CODE_OAUTH_TOKEN`, etiquetas `evals` y
`evals-prueba-de-red`, cuerpo `gates/pr-h6.md`); `git push -u origin 008-h6-territorio-skill-legal`;
`gh pr create --base main --head 008-h6-territorio-skill-legal --title 'feat(H6): `territorio` + skill `legal-core` v0'
--body-file specs/008-h6-territorio-skill-legal/gates/pr-h6.md`, que abrió #39; espera de las dos ejecuciones
disparadas por la apertura; lectura de los informes; y el registro.

### 1. `ci` en rojo: la carga del territorio, en el borde de su presupuesto de tiempo

`gh pr checks` terminó en 1 con `ci fail`. En la ejecución 35714659927, el paso «Ejecutar los controles» (`make ci`)
pasó `make test`, con `internal/app` en `ok` en 11,497 s, y cayó en `make test-integration` con el mismo test:

```text
FAIL: testdata/script/territorio-matriz.txtar:192: cronometra: el programa tardó el máximo o más: /tmp/kitlegal-e2e-2363664719/kitlegal ["territorio" "resolver" "Leganés" "--json"] tardó 235.378284ms y el máximo es 200ms
```

La salida entera del paso fallido va al final de esta nota. Como `make ci` falló, el paso «Publicar el perfil de
cobertura» salió `skipped` y Codecov no emitió ningún estado. No es un problema de configuración.

**No es ruido que se pueda esperar a que pase.** Las demás invocaciones del mismo guion en el runner tardaron entre 181
y 303 ms por bloque (tiempos de testscript en la salida), así que una sola invocación ronda el máximo. El mismo guion
pasó en `make test` y falló minutos después en `make test-integration` del mismo trabajo. La cota del plan
(«Performance Goals»: por debajo de 200 ms con el análisis de los ficheros embebidos incluido) descansaba en
research.md S3, una estimación que el plan declara «no medida todavía». El runner la desmiente: sin margen.

**Medida en local, sobre `47f3090`** (Apple M3, go1.27.1):

- El binario (`make build`) tarda unos 0,05 s en `territorio resolver Leganés --json`, y 0,00 s tanto en `--help`
  como en `territorio resolver --describe`. Todo el tiempo es la carga. `GODEBUG=inittrace=1` no muestra ningún `init`
  apreciable.
- Un benchmark temporal de `territorio.Cargar` sobre los datos congelados reales (fichero `zz_perfil_tmp_test.go`,
  borrado antes de terminar y nunca commiteado) da **42 ms/op, 36 444 020 B/op y 756 111 allocs/op** para 830 KB de
  YAML (632 KB de relación y 183 KB de correspondencia).
- Perfil de memoria (`-memprofile`, `alloc_space`): `decodificar` hace el 81,8 % de lo asignado. De eso, el árbol de
  nodos que el lector de YAML construye para toda la relación es el 61 % del total (`(*parser).node` y `scalar`), y la
  decodificación fila a fila (`(*Node).Decode`, con un `newDecoder` por fila) es el 10,7 %. `construir` hace el 18,2 %
  (`indexar`, el 10,9 %).
- Perfil de CPU: la función `Cargar` es el 16 % de las muestras y la recogida de basura y la devolución de memoria
  (`madvise`, `gcBgMarkWorker`) son la mayor parte del resto. En el runner, con 4 vCPU y los paquetes de
  `go test ./...` en paralelo (`internal/skills` 103 s, `internal/httpx` 34 s, a la vez que `internal/app`), eso se va a
  unos 235 ms.

**Por qué no se toca la cota ni la medida.** Subir los 200 ms es un umbral rebajado. Medir el mínimo de varias
tiradas o sacar el cronometraje de `make ci` esconde la causa, y por el ritmo de las demás invocaciones ni siquiera
dejaría margen. La causa es que cada invocación asigna 36 MB para leer 830 KB, y ahí va el arreglo.

**Arreglo: tarea nueva T029, colocada en `tasks.md` antes de esta.** Delimitada a `internal/core/territorio/`:

- test primero, con un control determinista del coste de la carga sobre una entrada sintética del tamaño real, que
  hoy falla y exige como mucho un tercio de los bytes y de las asignaciones actuales;
- la misma salida y los mismos defectos que hoy;
- la cobertura no baja del 100 %;
- sin tocar la cota, la orden de cronometraje, el guion, los datos congelados, sus esquemas, el kernel ni el applet, y
  sin dependencias nuevas.

Por el perfil, con un tercio basta: sin el árbol de nodos, lo que queda es `construir` (unos 6,6 MB) y las cadenas y
los mapas de las filas. Si el presupuesto solo se sostiene cambiando alguna de las cosas excluidas, T029 se detiene y
lo anota: ese cambio ya es de alcance.

La tarea no cabe en T026: sus rutas son `specs/`, y el commit de la tarea se hace después del intento. Su push
publicaría la cabeza sin el arreglo.

**Qué exige esto al intento 2.**

- T029 cambia `internal/core/territorio/`, que queda fuera de `specs/` y dentro del filtro del job de evals. La
  ejecución 35714659803 deja de cubrir la cabeza (SC-015; `gates/pr-h6.md`, «Pendientes»).
- El intento 2 empuja la rama con T029 y espera `ci` en verde, y con él el estado de Codecov.
- El flujo no escucha `synchronize`, así que pone la etiqueta `evals` en #39 (`gh pr edit 39 --add-label evals`)
  para obtener **una** ejecución nueva sobre ese commit. La registra en `gates/evals-cierre.md` como vigente, sobre la
  del intento 1, que se conserva.
- Si el cuerpo `gates/pr-h6.md` cambió, lo sincroniza con `gh pr edit 39 --body-file …`.
- Es medir otro commit, no relanzar para buscar otro resultado.

### 2. SC-013: una respuesta de seis nombra boletines que el applet no devolvió

Las seis sesiones de Tordesillas pasan por `territorio resolver`, dan solo el BOE y declaran los dos aspectos
`no-configurado`, diciendo que eso no significa que no existan. Cinco no nombran ningún boletín que falte. La sesión
`02-territorio-municipio-no-cubierto-claude-sonnet-5-02`, del modelo que decide, dice: «Si necesitas el nombre exacto
del boletín autonómico (normalmente el BOCyL) o el provincial (BOP de Valladolid), tendría que confirmártelo por otra
vía». Contradice la segunda cláusula de SC-013 y la regla de `skills/legal-core/SKILL.md:71`, que ya lo prohíbe
expresamente («ni su nombre, ni su sigla…»). El juez la da por buena, porque el esperado de territorio de FR-084
declara lo que tiene que aparecer y no lo que no puede aparecer, así que el veredicto `aprobado` no dice nada de esta
cláusula.

**Decisión pendiente, y humana, porque es de alcance.** Respuesta a respuesta, SC-013 no se cumple (1 de 3 en el
modelo que decide). Por serie y umbral, como SC-015 cuenta el verde, se cumple (2 de 3). El spec no dice cuál de las
dos lecturas aplica («la forma exacta de contar el verde está en SC-015, y SC-013 es la comprobación de la respuesta
en esa misma ejecución»).

Recomendación: que la cláusula la juzgue la máquina, para que el umbral de ADR 0016 se aplique igual que al resto y
SC-013 deje de ser una lectura a mano. Por ejemplo, un esperado de territorio que liste los boletines que la respuesta
no puede nombrar, o que el juez compare los boletines nombrados con los que devolvió el applet. Eso cambia FR-084, el
esquema de eval y el juez, así que es spec y material de esquema, fuera de T029 y de esta tarea.

Hasta que se decida, el intento 2 registra cada respuesta que nombre un boletín no devuelto, igual que aquí, y aplica
la lectura que se haya decidido. Si no se ha decidido, se detiene en este punto.

### 3. La orden 2 del §14 no imprimía nada (corregido en este intento)

Tal como estaba escrita, `sed -n "/^# Informe de evals/,/^## Sesiones/p"` no casaba con ninguna línea, porque
`gh run view --log` pone delante de cada una `trabajo<TAB>paso<TAB>hora `. Terminaba con `código: 0` sin imprimir
nada, así que una lectura desatendida no habría distinguido «sin informe» de «informe vacío».

`quickstart.md` §14 está en las rutas de esta tarea y se corrige aquí. La orden nueva quita el prefijo como la de H5.1
(`cut -f3- | sed -E "s/^[^ ]+ //"`) y termina en 1 si no encuentra los dos informes. Probada sobre la ejecución
35714659803: imprime los dos informes y `código: 0`. Las dos salidas, la vieja y la corregida, están en
`gates/evals-cierre.md`.

### Salida entera del paso fallido de `ci` (`gh run view 35714659927 --log-failed`)

~~~~text
ci	Ejecutar los controles	﻿2026-09-22T10:12:55.3627011Z ##[group]Run make ci
ci	Ejecutar los controles	2026-09-22T10:12:55.3627468Z ^[[36;1mmake ci^[[0m
ci	Ejecutar los controles	2026-09-22T10:12:55.3657550Z shell: /usr/bin/bash -e {0}
ci	Ejecutar los controles	2026-09-22T10:12:55.3657934Z env:
ci	Ejecutar los controles	2026-09-22T10:12:55.3658369Z   GOTOOLCHAIN: local
ci	Ejecutar los controles	2026-09-22T10:12:55.3658692Z ##[endgroup]
ci	Ejecutar los controles	2026-09-22T10:12:55.7825726Z go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
ci	Ejecutar los controles	2026-09-22T10:12:59.2684730Z go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
ci	Ejecutar los controles	2026-09-22T10:13:55.1675869Z 0 issues.
ci	Ejecutar los controles	2026-09-22T10:13:55.1993118Z go test -race -shuffle=on -coverprofile=coverage.out ./...
ci	Ejecutar los controles	2026-09-22T10:14:56.2081383Z ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.026s	coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-22T10:14:56.3869007Z 	github.com/jmorenobl/kitlegal/data		coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-22T10:14:59.1222726Z ok  	github.com/jmorenobl/kitlegal/internal	2.160s	coverage: [no statements]
ci	Ejecutar los controles	2026-09-22T10:15:14.7076891Z ok  	github.com/jmorenobl/kitlegal/internal/app	11.497s	coverage: 93.6% of statements
ci	Ejecutar los controles	2026-09-22T10:15:14.7125822Z 	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-22T10:15:14.7142159Z 	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e		coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-22T10:15:14.7143864Z ok  	github.com/jmorenobl/kitlegal/internal/cache	7.325s	coverage: 90.5% of statements
ci	Ejecutar los controles	2026-09-22T10:15:15.8608851Z ok  	github.com/jmorenobl/kitlegal/internal/cli	1.360s	coverage: 98.6% of statements
ci	Ejecutar los controles	2026-09-22T10:15:15.8610133Z ?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ci	Ejecutar los controles	2026-09-22T10:15:16.7214713Z ok  	github.com/jmorenobl/kitlegal/internal/core/ids	1.036s	coverage: 100.0% of statements
ci	Ejecutar los controles	2026-09-22T10:15:17.3982871Z ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.021s	coverage: 90.1% of statements
ci	Ejecutar los controles	2026-09-22T10:15:19.7325964Z ok  	github.com/jmorenobl/kitlegal/internal/core/territorio	2.102s	coverage: 100.0% of statements
ci	Ejecutar los controles	2026-09-22T10:15:27.9587356Z ok  	github.com/jmorenobl/kitlegal/internal/evals	7.422s	coverage: 99.1% of statements
ci	Ejecutar los controles	2026-09-22T10:15:53.4686848Z ok  	github.com/jmorenobl/kitlegal/internal/httpx	29.417s	coverage: 97.2% of statements
ci	Ejecutar los controles	2026-09-22T10:15:53.4706481Z ok  	github.com/jmorenobl/kitlegal/internal/render	1.016s	coverage: 95.8% of statements
ci	Ejecutar los controles	2026-09-22T10:15:53.4716320Z ok  	github.com/jmorenobl/kitlegal/internal/skills	19.445s	coverage: 98.2% of statements
ci	Ejecutar los controles	2026-09-22T10:15:59.6355158Z ok  	github.com/jmorenobl/kitlegal/internal/source/boe	5.093s	coverage: 99.4% of statements
ci	Ejecutar los controles	2026-09-22T10:15:59.7010905Z go test -race -tags=integration -coverprofile=coverage-integration.out ./...
ci	Ejecutar los controles	2026-09-22T10:16:01.8946850Z ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.020s	coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-22T10:16:01.8964264Z 	github.com/jmorenobl/kitlegal/data		coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-22T10:16:03.2707453Z ok  	github.com/jmorenobl/kitlegal/internal	1.836s	coverage: [no statements]
ci	Ejecutar los controles	2026-09-22T10:16:12.8217466Z --- FAIL: TestEntregaDelHito (0.00s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8219861Z     --- FAIL: TestEntregaDelHito/territorio-matriz (4.63s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8224579Z         testscript.go:609: # La matriz territorial de la entrega de H6, de extremo a extremo sobre el
ci	Ejecutar los controles	2026-09-22T10:16:12.8236530Z             # binario compilado: el municipio del territorio configurado (Leganés), uno de
ci	Ejecutar los controles	2026-09-22T10:16:12.8238118Z             # una comunidad sin configurar (Tordesillas), dos de régimen foral (Navarra y
ci	Ejecutar los controles	2026-09-22T10:16:12.8239476Z             # País Vasco), un nombre ambiguo, un nombre y un código bien formado que no están
ci	Ejecutar los controles	2026-09-22T10:16:12.8240877Z             # en la relación y un código que no llega a serlo, cada uno con su código de
ci	Ejecutar los controles	2026-09-22T10:16:12.8242181Z             # salida exacto y el contenido de su sobre; la igualdad byte a byte por nombre,
ci	Ejecutar los controles	2026-09-22T10:16:12.8243653Z             # por código y con --offline; la misma salida desde otro directorio de trabajo;
ci	Ejecutar los controles	2026-09-22T10:16:12.8244870Z             # --describe con su argumento; y el tiempo de respuesta (FR-009, FR-056, FR-090,
ci	Ejecutar los controles	2026-09-22T10:16:12.8260068Z             # FR-091, US1 a US4, SC-001 a SC-004, contrato del applet §8).
ci	Ejecutar los controles	2026-09-22T10:16:12.8260783Z             #
ci	Ejecutar los controles	2026-09-22T10:16:12.8262220Z             # Nada de lo que se espera está escrito de memoria: cada nombre, código, dígito
ci	Ejecutar los controles	2026-09-22T10:16:12.8263558Z             # de control, DIR3, provincia, comunidad, régimen y boletín se ha leído de los
ci	Ejecutar los controles	2026-09-22T10:16:12.8264549Z             # ficheros congelados de data/territorio/, que el binario lleva dentro. La
ci	Ejecutar los controles	2026-09-22T10:16:12.8265745Z             # fecha_consulta de toda respuesta del applet es la «fecha» de municipios.yaml,
ci	Ejecutar los controles	2026-09-22T10:16:12.8267061Z             # 2026-02-04, la más antigua de los cuatro ficheros, y no la del reloj
ci	Ejecutar los controles	2026-09-22T10:16:12.8267893Z             # (contrato del applet §2).
ci	Ejecutar los controles	2026-09-22T10:16:12.8268377Z             #
ci	Ejecutar los controles	2026-09-22T10:16:12.8269317Z             # Como en boe-codigos.txtar, testscript solo distingue el éxito del fallo, así
ci	Ejecutar los controles	2026-09-22T10:16:12.8270699Z             # que los códigos distintos de cero los comprueba el intérprete de órdenes;
ci	Ejecutar los controles	2026-09-22T10:16:12.8272083Z             # dentro de las comillas simples no hay expansión de variables y el binario se
ci	Ejecutar los controles	2026-09-22T10:16:12.8273310Z             # nombra por el PATH, donde el e2e antepone el que acaba de construir.
ci	Ejecutar los controles	2026-09-22T10:16:12.8274679Z             # territorio no pide nada a ninguna fuente: sin reproducción de la que servir,
ci	Ejecutar los controles	2026-09-22T10:16:12.8280016Z             # cualquier petición fallaría, y todo lo que sigue termina bien (FR-043). (0.001s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8282234Z             # --- Leganés, el municipio del territorio configurado (US1, SC-001) ---
ci	Ejecutar los controles	2026-09-22T10:16:12.8283140Z             #
ci	Ejecutar los controles	2026-09-22T10:16:12.8284099Z             # Código 0 y el sobre entero, anclado a los dos extremos: las ocho claves de
ci	Ejecutar los controles	2026-09-22T10:16:12.8286704Z             # `data` en su orden, ni una más ni una menos. (0.194s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8287629Z             # Cada dato con su source (FR-005). (0.001s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8288762Z             # El estatal y los dos niveles configurados, con el mismo boletín en los dos y
ci	Ejecutar los controles	2026-09-22T10:16:12.8289997Z             # el motivo de la equivalencia en el provincial; y la cobertura configurada. (0.001s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8291362Z             # Por nombre, por código, por código con su dígito de control, con --offline y
ci	Ejecutar los controles	2026-09-22T10:16:12.8292535Z             # con el nombre sin tilde ni mayúsculas: byte a byte la misma salida, huella y
ci	Ejecutar los controles	2026-09-22T10:16:12.8293744Z             # fecha_consulta incluidas (FR-009, FR-015, US1 escenarios 3 y 4, SC-001). (0.757s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8294770Z             # Y lo mismo se ejecute desde donde se ejecute: en un directorio de trabajo sin
ci	Ejecutar los controles	2026-09-22T10:16:12.8295904Z             # nada dentro la respuesta no cambia, porque los datos viajan en el binario
ci	Ejecutar los controles	2026-09-22T10:16:12.8296558Z             # (FR-056). (0.189s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8297334Z             # --- Tordesillas, fuera del territorio configurado (US2, SC-002) ---
ci	Ejecutar los controles	2026-09-22T10:16:12.8297893Z             #
ci	Ejecutar los controles	2026-09-22T10:16:12.8298684Z             # Código 0, las ocho claves y los datos nacionales completos. (0.181s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8299649Z             # Los dos boletines declarados no cubiertos, con un vocabulario que no permite
ci	Ejecutar los controles	2026-09-22T10:16:12.8300426Z             # concluir que no existan (FR-020, FR-022)... (0.000s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8301330Z             # ... y ningún boletín inventado: solo el estatal, ningún nivel autonómico ni
ci	Ejecutar los controles	2026-09-22T10:16:12.8302307Z             # provincial, nada del único otro boletín que los datos conocen —el de la
ci	Ejecutar los controles	2026-09-22T10:16:12.8304347Z             # Comunidad de Madrid— y ninguna forma del código ni del nombre de los de
ci	Ejecutar los controles	2026-09-22T10:16:12.8316898Z             # Castilla y León ni de la provincia (FR-008, FR-021, SC-002). (0.003s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8318124Z             # --- Régimen foral, con la comunidad sin configurar (US3, SC-003) ---
ci	Ejecutar los controles	2026-09-22T10:16:12.8318843Z             #
ci	Ejecutar los controles	2026-09-22T10:16:12.8319876Z             # Navarra: el régimen consta aunque su comunidad no tenga boletines. (0.303s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8321102Z             # País Vasco, igual. (0.208s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8322368Z             # --- Nombre ambiguo, inexistente y código mal formado (US4, SC-004) ---
ci	Ejecutar los controles	2026-09-22T10:16:12.8323071Z             #
ci	Ejecutar los controles	2026-09-22T10:16:12.8324109Z             # Un nombre que es el de dos municipios: código 2, clase argumentos y los dos
ci	Ejecutar los controles	2026-09-22T10:16:12.8325599Z             # candidatos —todos los que tiene— con su código INE y su provincia, ordenados
ci	Ejecutar los controles	2026-09-22T10:16:12.8326967Z             # por código, en el sobre y en la salida de error (FR-013, FR-014). (0.233s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8328387Z             # Un nombre que no es el de ningún municipio: código 3 (FR-011). (0.219s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8329652Z             # Un código bien formado que no está en la relación: código 3. La provincia 01
ci	Ejecutar los controles	2026-09-22T10:16:12.8331045Z             # es de las que trae la relación —Amurrio, 01002, se acaba de resolver por su
ci	Ejecutar los controles	2026-09-22T10:16:12.8332240Z             # nombre, y por su código da lo mismo— y el municipio 999 no lo tiene: la
ci	Ejecutar los controles	2026-09-22T10:16:12.8333593Z             # relación congelada acaba la provincia 01 en 01902 (FR-011). (0.750s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8334789Z             # Una provincia fuera de 01-52 no forma un código: la entrada no llega a
ci	Ejecutar los controles	2026-09-22T10:16:12.8336415Z             # buscarse, así que es código 2 con clase argumentos, y nunca 3, en los dos
ci	Ejecutar los controles	2026-09-22T10:16:12.8337800Z             # bordes del intervalo (FR-011, FR-012, contrato de identificadores §2). (0.977s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8339067Z             # Y el código de Leganés con un dígito de control que no es el oficial: código
ci	Ejecutar los controles	2026-09-22T10:16:12.8340333Z             # 2, y el mensaje dice cuál se esperaba (FR-012). (0.347s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8341337Z             # --- --describe (FR-007, contrato del applet §1) ---
ci	Ejecutar los controles	2026-09-22T10:16:12.8341959Z             #
ci	Ejecutar los controles	2026-09-22T10:16:12.8342912Z             # Con su argumento, el esquema de entrada y salida del verbo, con la consulta
ci	Ejecutar los controles	2026-09-22T10:16:12.8344353Z             # entre lo exigido, y sin ejecutar nada: ni un sobre ni ningún dato del
ci	Ejecutar los controles	2026-09-22T10:16:12.8345619Z             # municipio. (0.011s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8346745Z             # Sin él, el análisis de la invocación, que corre antes de describir, termina
ci	Ejecutar los controles	2026-09-22T10:16:12.8347719Z             # en 2 y no describe nada. (0.007s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8348717Z             # --- Tiempo de respuesta (plan, Performance Goals; research.md S3) ---
ci	Ejecutar los controles	2026-09-22T10:16:12.8349340Z             #
ci	Ejecutar los controles	2026-09-22T10:16:12.8350425Z             # cronometra mide la invocación entera, arranque del proceso y análisis de los
ci	Ejecutar los controles	2026-09-22T10:16:12.8367985Z             # ficheros embebidos incluidos, y cada respuesta sigue siendo la misma. (0.236s)
ci	Ejecutar los controles	2026-09-22T10:16:12.8369625Z             > cronometra 200ms $KITLEGAL_BIN territorio resolver Leganés --json
ci	Ejecutar los controles	2026-09-22T10:16:12.8370450Z             [stdout]
ci	Ejecutar los controles	2026-09-22T10:16:12.8401515Z             {"ok":true,"fuente":"kitlegal.territorio","url":"kitlegal:applet/territorio","fecha_consulta":"2026-02-04T00:00:00Z","hash":"sha256:b730a9076336bd84fa84c91f3efe59c462ee589c68432c06971b59dcd7a8d570","data":{"municipio":{"nombre":"Leganés","source":"ine.municipios"},"codigo_ine":{"codigo":"28074","digito_de_control":"5","source":"ine.municipios"},"provincia":{"codigo":"28","nombre":"Madrid","source":"ine.municipios"},"comunidad":{"codigo":"13","nombre":"Comunidad de Madrid","source":"ine.municipios"},"dir3":{"codigo":"L01280745","source":"mpt.rel"},"regimen":{"valor":"comun","source":"data/territorio/comunidades/13.yaml"},"boletines":[{"nivel":"estatal","codigo":"BOE","nombre":"Boletín Oficial del Estado","url":"https://www.boe.es/","motivo":"","source":"data/territorio/estado.yaml"},{"nivel":"autonomico","codigo":"BOCM","nombre":"Boletín Oficial de la Comunidad de Madrid","url":"https://www.bocm.es/","motivo":"","source":"data/territorio/comunidades/13.yaml"},{"nivel":"provincial","codigo":"BOCM","nombre":"Boletín Oficial de la Comunidad de Madrid","url":"https://www.bocm.es/","motivo":"Comunidad uniprovincial: el BOCM hace también de boletín provincial.","source":"data/territorio/comunidades/13.yaml"}],"cobertura":{"boletin_autonomico":"configurado","boletin_provincial":"configurado","dir3":"verificado"}}}
ci	Ejecutar los controles	2026-09-22T10:16:12.8408842Z             
ci	Ejecutar los controles	2026-09-22T10:16:12.8411479Z             FAIL: testdata/script/territorio-matriz.txtar:192: cronometra: el programa tardó el máximo o más: /tmp/kitlegal-e2e-2363664719/kitlegal ["territorio" "resolver" "Leganés" "--json"] tardó 235.378284ms y el máximo es 200ms
ci	Ejecutar los controles	2026-09-22T10:16:12.8412826Z             
ci	Ejecutar los controles	2026-09-22T10:16:12.8413491Z FAIL
ci	Ejecutar los controles	2026-09-22T10:16:12.8413900Z coverage: 93.6% of statements
ci	Ejecutar los controles	2026-09-22T10:16:12.8414596Z FAIL	github.com/jmorenobl/kitlegal/internal/app	10.920s
ci	Ejecutar los controles	2026-09-22T10:16:12.8435875Z 	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-22T10:16:12.8437383Z 	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e		coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-22T10:16:15.0257376Z ok  	github.com/jmorenobl/kitlegal/internal/cache	9.280s	coverage: 96.2% of statements
ci	Ejecutar los controles	2026-09-22T10:16:15.0266702Z ok  	github.com/jmorenobl/kitlegal/internal/cli	1.511s	coverage: 98.6% of statements
ci	Ejecutar los controles	2026-09-22T10:16:15.0287529Z ?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ci	Ejecutar los controles	2026-09-22T10:16:16.5683563Z ok  	github.com/jmorenobl/kitlegal/internal/core/ids	1.045s	coverage: 100.0% of statements
ci	Ejecutar los controles	2026-09-22T10:16:16.5736640Z ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.031s	coverage: 90.1% of statements
ci	Ejecutar los controles	2026-09-22T10:16:19.1814268Z ok  	github.com/jmorenobl/kitlegal/internal/core/territorio	2.100s	coverage: 100.0% of statements
ci	Ejecutar los controles	2026-09-22T10:16:23.7567480Z ok  	github.com/jmorenobl/kitlegal/internal/evals	6.131s	coverage: 99.1% of statements
ci	Ejecutar los controles	2026-09-22T10:16:54.9252824Z ok  	github.com/jmorenobl/kitlegal/internal/httpx	34.586s	coverage: 97.2% of statements
ci	Ejecutar los controles	2026-09-22T10:16:54.9306932Z ok  	github.com/jmorenobl/kitlegal/internal/render	1.016s	coverage: 95.8% of statements
ci	Ejecutar los controles	2026-09-22T10:18:09.9560409Z ok  	github.com/jmorenobl/kitlegal/internal/skills	103.470s	coverage: 98.2% of statements
ci	Ejecutar los controles	2026-09-22T10:18:09.9562451Z ok  	github.com/jmorenobl/kitlegal/internal/source/boe	16.296s	coverage: 99.4% of statements
ci	Ejecutar los controles	2026-09-22T10:18:09.9563558Z FAIL
ci	Ejecutar los controles	2026-09-22T10:18:09.9660878Z make: *** [Makefile:73: test-integration] Error 1
ci	Ejecutar los controles	2026-09-22T10:18:09.9680420Z ##[error]Process completed with exit code 2.
~~~~
