# T003 — resuelta con intervención humana (2026-09-10)

**Estado final**: `[X]`. Entregada y verificada.

## Qué la bloqueó

Los tres intentos del modo desatendido fallaron por la misma causa, ajena al código: el guardián de
ficheros sensibles del harness deniega toda escritura sobre `lefthook.yml` desde una sesión no
interactiva, y `lefthook.yml` es el entregable central de la tarea.

```
Claude requested permissions to edit /Users/jorge/Projects/kitlegal/lefthook.yml
which is a sensitive file.
```

No era una regla del repositorio: `.claude/settings.json` permite `Write` y `Edit`, y su `deny` no
menciona ninguna ruta. Que el guardián cubra este fichero es coherente, porque `lefthook.yml` declara
órdenes que se ejecutan solas en cada `git commit`.

**El agente no lo rodeó, y ese fue el comportamiento correcto.** Descartó por escrito escribirlo desde
`Bash`, escribirlo con otro nombre y renombrarlo, generarlo desde una receta del `Makefile` y añadirse a
sí mismo el permiso en `.claude/settings.json`. El guardián actúa por herramienta, pero su intención es
que el agente no escriba ese fichero sin aprobación; colarlo por otra puerta lo habría eludido en lugar
de cumplirlo.

## Cómo se resolvió

La sesión principal, con la persona presente, escribió `lefthook.yml` con el contenido que el propio
agente había dejado redactado y fundamentado en `research.md` D16 y FR-022: bloque `pre-commit` con
cuatro trabajos que invocan **solo** órdenes del `Makefile` (`fmt` con `stage_fixed`, `lint-fast`,
`secrets`, `mod-tidy-check`). Ni una orden fuera del `Makefile`, para que lo que corrige el gancho y lo
que verifica la integración continua no puedan divergir.

## Verificado

- `make hooks` instala el gancho: `sync hooks: ✔️ (pre-commit)`, y aparece `.git/hooks/pre-commit`.
- **Escenario 8 de `quickstart.md`, completo.** Con `func  main( )  {` (espaciado roto) y una línea de
  comentario en el índice, `git commit` disparó los cuatro trabajos, los cuatro en verde. El gancho
  devolvió la declaración a su forma canónica y el commit resultante contiene **solo** la línea de
  comentario, ya formateada. `git diff HEAD -- cmd/kitlegal/main.go` quedó vacío: lo confirmado es lo que
  hay en el árbol.
- **Deshacer anclado y quirúrgico**, tal como prescribe el escenario: `reset --soft "$antes"` más
  `restore --source="$antes"` acotado a `cmd/kitlegal/main.go`. `HEAD` volvió a la revisión de partida y
  los diffs de árbol e índice quedaron vacíos, sin tocar los ficheros de estado del workflow.
- `make ci` en verde con el gancho instalado.

## Obligación de verificación resuelta

La segunda de las cinco obligaciones que `research.md` dejaba abiertas queda cerrada: **`stage_fixed: true`
sí vuelve a preparar los ficheros corregidos en lefthook v1.13.6**, aunque `run` no reciba
`{staged_files}`. No hace falta la alternativa con `git add {staged_files}` explícito. Registrado en D16.

## Efecto colateral, a tener en cuenta

A partir de aquí el gancho corre en **cada** commit, incluidos los que hace el propio workflow `hito` en
su paso `commit_tarea`. Es lo pretendido, y significa que una tarea que deje el árbol sin formatear o con
un secreto detectable no llegará a confirmarse.
