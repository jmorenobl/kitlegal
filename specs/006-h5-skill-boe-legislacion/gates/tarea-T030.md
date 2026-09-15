# T030 · intento 1 de 3 (2026-09-15): sin marcar

Cabeza `6d68c21be163d2f860eefbe1c7e1871f33ba701c`, propuesta de cambio [#27](https://github.com/jmorenobl/kitlegal/pull/27).
Dos motivos, cada uno con su arreglo en una tarea nueva colocada **antes** de T030 en `tasks.md` (T033, T034 y T035,
tras T029). T030 no puede llevar ninguno: sus rutas son de `gates/`, y el push del intento siguiente tiene que publicar
la cabeza con los arreglos.

## Lo que sí quedó hecho

1. `git push -u origin h5-skill-boe-legislacion`: rama nueva en el remoto (el gancho `pre-push` la deja pasar).
2. `gh pr view h5-skill-boe-legislacion` → `no pull requests found for branch "h5-skill-boe-legislacion"`; `gh pr create`
   con `--body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md` → #27. Tras anotar S8 y este intento en
   *Pendientes* de `gates/pr-h5.md`, el cuerpo se sincronizó con `gh pr edit 27 --body-file` (el mismo fichero).
3. Checks, `gh run list` y check-runs y estados por la API: `gates/evidencia-plataforma.md`. `ci` en verde; los cuatro
   estados de Codecov presentes, con cifra y objetivo; **`codecov/patch` en rojo** (motivo 1).
4. Quickstart §12.1: el secreto `CLAUDE_CODE_OAUTH_TOKEN` y las etiquetas `evals` y `evals-prueba-de-red` existen.
5. Quickstart §12.2, tal cual: ejecución 34922606273 identificada por el único evento `labeled` y `workflowName`;
   **se detuvo en «Retirar Python del runner» con código 100** (motivo 2). Quinta y sexta órdenes con 1 sin salida,
   orden de `--log-failed` registrada entera y etiqueta retirada con la séptima: `gates/prueba-de-red.md`, con la evidencia
   de S2, S4, S7, S9 y S12.
6. S8: lo comprobable con la propuesta abierta, en `gates/pr-h5.md` (*Pendientes*).
7. `make ci` local sobre la cabeza, en primer plano de la sesión (log en `gates/ci.log`, ignorado): código 0,
   `ci: todos los controles en verde`, ningún `FAIL`.

## Motivo 1 · `codecov/patch` en rojo: 93,54 % del diff frente al objetivo de 94,70 %

**No es configuración ni umbral.** El estado mide 20 ficheros y 2 090 líneas del diff (no es un verde vacío), es
bloqueante (`informational: false`) y su objetivo `auto` es la cobertura de la base (`90c3637`, 94,70 %): la regla «no
retroceder» aplicada al código nuevo. Los otros tres estados están en verde con cifra y objetivo. Faltan 135 líneas; para
llegar al objetivo harían falta 1 980 de 2 090 (25 más), pero el arreglo no persigue la cifra: cubre cada rama alcanzable.

**No es cobertura que solo den los tests de integración** (lo primero que hay que descartar, H3): la lista de abajo sale
de la **unión** de los dos perfiles que publica CI, los mismos que dejó el `make ci` de T029 sobre `536359c`, y ningún
`.go` cambió desde entonces (`git diff --name-only 536359c HEAD -- '*.go'` vacío). Orden que lista los bloques con cuenta
cero en los dos perfiles (Codecov cuenta líneas; un bloque de una sentencia es una línea):

```bash
rtk proxy perl -ne 'next if /^mode:/; if (/^github\.com\/jmorenobl\/kitlegal\/(internal\/(?:evals|skills)\/[a-z]+\.go):(\d+)\.\d+,(\d+)\.\d+ (\d+) (\d+)$/) { $k="$1:$2-$3"; $f{$k}=$1; $a{$k}=$2; $n{$k}=$4; $c{$k}+=$5 } END { for $k (sort { $f{$a} cmp $f{$b} or $a{$a} <=> $a{$b} } keys %c) { print "$k ($n{$k} sent.)\n" if $c{$k}==0 } }' coverage.out coverage-integration.out
```

Coincide fichero a fichero con la tabla de `codecov[bot]` (las «2 more» son `comandos.go` y `frontmatter.go`). Casi todo
son ramas de error de una sentencia en los dos paquetes nuevos:

| Fichero | Codecov | Bloques sin cubrir (líneas) | Qué son |
|---|---|---|---|
| `internal/evals/trazas.go` | 45 | 227, 257, 283, 291, 294, 346, 452, 457, 480, 519, 521, 528, 582, 608, 627, 637, 641, 646, 654, 659, 671, 691, 707, 723, 728, 743, 745, 770, 788, 799, 804, 814, 852, 882, 887-889, 908-910, 964, 981, 1001, 1022, 1061-1063, 1085 | defectos de línea de la traza: argv de `execve` sin la forma de lista de cadenas, escapes inválidos (`\` final, `\x` sin dos hexadecimales, desconocido, octal fuera de octeto), argumentos de `connect` y direcciones sin forma, familia distinta de `AF_INET`, `AF_INET6` y `AF_UNIX`, código final o resultado no entero, línea cortada o posterior a la final, fichero `t.<n>` mal nombrado o con número no entero, hilos creados dos veces o sueltos; fallos al leer el directorio o un fichero; el registro de applets y las banderas globales (constantes); `argv` vacío y ramas de control |
| `internal/skills/sincronia.go` | 30 | 172, 181, 217, 222, 234, 262, 293, 320, 404, 439, 444, 475, 484, 488, 492, 496, 506, 524, 539, 544, 549, 558, 573, 599, 601, 606, 633, 642, 671, 680 | al aplicar arreglos, fallos al retirar, crear el directorio, enlazar y escribir; carpetas y ficheros de la skill que no se pueden consultar, listar o leer, o que no son directorio; propagación de esos errores |
| `internal/evals/preparar.go` | 19 | 76, 98, 132-134, 140, 146, 166-168, 174, 218, 247, 251, 268, 311, 316-318, 322, 326, 351 | fallos al crear, copiar y retirar los temporales de las grabaciones y al escribir `eval.txt` y la pregunta; entrada del conjunto que no es fichero regular; registro de applets (constante); dos textos de causa |
| `internal/skills/esquemas.go` | 14 | 33, 40, 45, 198, 266, 275, 280, 324, 388-391, 461, 500, 504 | esquema que no es JSON, que no se registra o no compila; YAML inválido o que no se decodifica, codificación JSON; error que no es de validación; recorrido de nodos secuencia |
| `internal/evals/sesion.go` | 12 | 228, 245, 263, 267, 271, 291, 295, 308, 312, 333, 337, 344 | mensajes de `stream-json` sin `type`, `system` y `assistant` y `result` mal formados, `system/init` sin `model` o `claude_code_version`, `assistant` sin `message.content`, `tool_use` de `Skill` sin entrada o sin `input.skill`, `result` sin `subtype`/`is_error` o sin `result` |
| `internal/skills/normas.go` | 5 | 91, 96, 123, 176, 186 | esquema incrustado de las normas (constante); identificador o norma que no se decodifican |
| `internal/evals/formato.go` | 3 | 84, 89, 103 | esquema incrustado del formato de eval (constante) y su propagación |
| `internal/evals/informe.go` | 2 | 230, 496 | codificación del informe (tipo fijo); fallo al retirar un informe ya escrito |
| `internal/skills/skill.go` | 2 | 92, 102 | `SKILL.md` que no se puede consultar o leer |
| `internal/evals/conjunto.go` | 1 | 103 | eval que no se puede leer |
| `internal/skills/comandos.go` | 1 | 434 | error del decodificador JSON al leer el primer token |
| `internal/skills/frontmatter.go` | 1 | 244 | bloque sin sentencias en la unión de perfiles |

**Arreglo**: tests que provocan la condición real, con entradas mal formadas como literales y fallos de disco por la
estructura del directorio temporal (no por permisos, que como root no fallan), sin tocar datos de prueba, esquemas,
umbrales ni la configuración de Codecov; las ramas que solo pueden fallar por una constante del paquete se
reestructuran para comprobarse desde una entrada que un test pueda dar, o se justifican. Por paquete, para que cada
intento tenga un alcance acotado y su propia evidencia: **T034** (`internal/evals`, 82 líneas) y **T035**
(`internal/skills`, 53 líneas). Cada una actualiza la tabla de cobertura de `gates/pr-h5.md`, fechada por commit.

## Motivo 2 · la prueba de red descubre un defecto del job

**Hecho** (`gates/prueba-de-red.md`): en `ubuntu-24.04`, el filtro del paso elige 110 paquetes de Python y
`sudo apt-get purge -y --auto-remove` termina con 100 antes de cambiar nada:

```text
The following packages have unmet dependencies:
 shim-signed : Depends: grub-efi-amd64-signed (>= 1.191~) but it is not going to be installed or
E: Error, pkgProblemResolver::Resolve generated breaks, this may be caused by held packages.
               Depends: grub2-common (>= 2.04-1ubuntu24) but it is not going to be installed
##[error]Process completed with exit code 100.
```

**Por qué D17 no lo previó**: V57 comprobó la purga en un contenedor de `ubuntu:24.04` donde ningún paquete del sistema
depende de Python. En la imagen del runner, `python3`, `python3-apt`, `python3-debconf`, `python3-netplan`,
`python3-commandnotfound`, `python3-software-properties`… son dependencias de paquetes del sistema; purgarlos con
`--auto-remove` arrastra a sus dependientes hasta la cadena de arranque (`shim-signed` → `grub-efi-amd64-signed` →
`grub2-common`) y el resolvedor lo rechaza. No hay ningún paquete retenido en el mensaje: S7 (3) y (4) y la purga de S2
difieren por la forma del grafo de paquetes de la imagen, que además cambia con cada versión de la imagen.

**Alternativas para el arreglo**:

- **Quitar la parte de paquetes del paso** (elegida, **T033**): la búsqueda como root en todo el sistema de ficheros, la
  retirada con la regla del prefijo y de lo usado y la búsqueda final son exactamente la garantía que piden FR-073 y
  FR-081 («ningún intérprete de Python accesible») y lo que registra `sin-python.txt`. La purga no retiraba ningún
  intérprete que la búsqueda no encuentre: el del sistema (`python3.12`, sus enlaces `python3` y `python`) y
  `libpython3.12.so*` se encuentran por nombre; lo que queda en disco (biblioteca estándar, `dist-packages`, guiones)
  no se ejecuta sin intérprete. Y deja de depender del grafo de paquetes de cada versión de la imagen, el mismo
  argumento con el que D17 rechaza las listas de sitios.
- *Forzar `apt-get`* (`--allow-remove-essential`, o aceptar lo que arrastre): se llevaría la cadena de arranque y las
  herramientas del sistema que dependen de Python, distinto en cada imagen; rechazada.
- *`dpkg --purge --force-depends`*: ejecuta los guiones `prerm` de los paquetes de Python (`py3clean`), que necesitan el
  intérprete que se está retirando; su éxito depende del orden; rechazada.
- *Retirar los ficheros de las listas de dpkg* (`dpkg-query -L`, sin `apt` ni guiones): limpia más, pero no retira
  ningún intérprete que la búsqueda no vea y añade supuestos nuevos (formato de las listas, desvíos, directorios
  compartidos); rechazada.
- *Contenedor*: sigue rechazado por D13.

T033 actualiza con esa decisión el paso de `.github/workflows/evals.yml`, el contrato del job §1, research D17, V56, V57,
S2 y S7, el plan y quickstart §12.2, y lo comprueba en contenedores desechables reproduciendo el caso del runner (un
paquete que `apt` no retira depende de uno de Python): el paso anterior termina con 100 y el nuevo con 0 y
`resultado: ninguno`.

**Lo que solo dirá la próxima prueba de red** (S7 (1)-(3) y lo posterior): si la imagen tiene Python en un sistema de
ficheros de solo lectura (p. ej. un snap), `rm` fallará y el paso se detendrá con su error, como dice el contrato; cuánto
tarda la búsqueda en toda la imagen; y todo lo que hay tras el paso (S1 con el secreto, S4, S5, S6, S9, S10, S11).

## Tareas nuevas

T033, T034 y T035 van tras T029 y antes de la fase 12, en ese orden (son independientes; T033 primero porque sin ella no
hay prueba de red posible). T029 sigue marcada: cada tarea nueva actualiza lo que cambia de `gates/pr-h5.md` (la tabla de
cobertura o la decisión del paso). En la trazabilidad de `tasks.md`, FR-073 nombra T033 y el punto 9 de la Definition
of Done, T034 y T035. Ninguna línea nueva lleva etiqueta: la extracción del workflow da, de T033,
`.github/workflows/evals.yml` y dos ficheros de `gates/`; de T034, los seis `.go` de `internal/evals` y ficheros de
`gates/`; de T035, los seis de `internal/skills`, `internal/skills/export_test.go` y ficheros de `gates/`.

## Qué hace el intento siguiente de T030

Con T033-T035 en verde: `git push -u origin h5-skill-boe-legislacion` (avance rápido); `gh pr view` encuentra #27 y no
se crea nada; checks y check-runs de la cabeza nueva (se espera `codecov/patch` en verde; si no, se vuelven a listar los
bloques con la orden de arriba); quickstart §12.1 y §12.2 enteras (la etiqueta está quitada, así que la primera orden no
quita nada y la segunda crea un evento `labeled` nuevo); y reescribe `gates/evidencia-plataforma.md` y
`gates/prueba-de-red.md`. Como `gates/pr-h5.md` cambiará con las tareas nuevas, conviene volver a sincronizar el cuerpo de
#27 con `gh pr edit 27 --body-file specs/006-h5-skill-boe-legislacion/gates/pr-h5.md`.
