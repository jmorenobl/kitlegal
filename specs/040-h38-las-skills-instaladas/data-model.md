# Data model: H38 · Las skills instaladas se actualizan con el binario

Ninguna forma persistente cambia: `kitlegal.json`, los hosts y los sobres de los applets siguen como en H19 (FR-016).
Lo nuevo son dos tipos y un puerto del dominio, una función del adaptador de la caché, un fichero vacío en la caché y
dos líneas. El contrato de su comportamiento está en contracts/aviso-y-reinstalacion.md.

## §1 · `instalacion.PermisoDelAviso`

Lo que la invocación permite hacer al aviso. Lo decide el kernel (research.md D2) y lo recibe el avisador.

| Valor | Invocaciones (FR-001, FR-002, FR-009) | Qué hace el aviso |
|---|---|---|
| `SoloAvisar` | análisis fallido (exit 2), `--describe`, `mcp serve` (con o sin `--dry-run`) | la línea de hoy (FR-007) |
| `Ensayar` | análisis bien, decisión ejecutar, `--dry-run`, verbo que no sirve | la de FR-009 si se cumplen FR-003 (a)-(c); si no, la de hoy |
| `Reinstalar` | análisis bien, decisión ejecutar, sin `--dry-run`, verbo que no sirve | reinstala si se cumplen FR-003 (a)-(d) y da la de FR-006; si no, la de hoy |

Las invocaciones que hoy no buscan el aviso —toda ayuda, `version`, `skills …`, lo que falla antes de resolver un
applet con su verbo y cada llamada de herramienta del servidor MCP— siguen sin llamar al avisador y no tienen permiso.

## §2 · `instalacion.Reinstalacion`

Los puertos y los datos con que el aviso decide y reinstala. Un valor por llamada; no guarda nada entre llamadas.

| Campo | Tipo | Lo pone la composición (`internal/app/aviso.go`) con |
|---|---|---|
| `Disco` | `Disco` | `disco.Lector{}` |
| `Enlazador` | `Enlazador` | el de `DependenciasDeSkills`, el mismo de `install` |
| `Escritor` | `Escritor` | `disco.NuevoEscritor(Enlazador)` |
| `Candado` | `TomarCandado` | `cache.TomarCandado("skills-global.lock", 10*time.Minute)` envuelta (§3) |
| `Home` | `string` | `os.Getenv("HOME")` en cada llamada |
| `Version` | `string` | la del binario |
| `Empotradas` | `[]SkillEmpotrada` | lo empotrado, leído como hoy (`skillsEmpotradasDe`) |
| `Retiradas` | `Retiradas` | `skillsRetiradas` (`internal/app/retiradas.go`) |

Método: `Avisar(permiso PermisoDelAviso) (linea string, hay bool)`.

## §3 · El candado

- **Puerto** (`internal/core/instalacion/puertos.go`): `type TomarCandado func() (liberar func(), tomado bool)`.
  Con `tomado == false`, `liberar` es nil y no queda nada escrito.
- **Adaptador** (`internal/cache/candado.go`): `func TomarCandado(nombre string, caducidad time.Duration) (liberar
  func(), tomado bool)`.
- **Fichero**: `<cache.Directorio()>/skills-global.lock`, regular, vacío, modo 0o600, creado en exclusiva. Lo que
  decide es su fecha de modificación: de hace menos de 10 minutos, **tomado**; de hace 10 minutos o más, **huérfano**.
  Existe solo mientras dura una reinstalación; `liberar` lo borra, salga bien o mal.
- **Directorio**: si `cache.Directorio()` no existe, se crea con 0o700 (`permisosDelDirectorio` de la caché).

## §4 · El pedido

Las skills que se reinstalan, en orden de nombre y cada una una vez:

- cada skill que el manifiesto global declara y el binario empotra (`declaradasYEmpotradas`);
- más, por cada declarada con un nombre de `Retiradas`, sus herederas (`Retiradas[nombre]`);
- y ninguna más.

Con el ámbito global de `HOME` y sin `--host`, es un `instalacion.Pedido{Ambito, Skills, Hosts: nil}`, el mismo que
`ValidarInvocacion` da a `kitlegal skills install -g <esas skills>`.

## §5 · Las líneas

Una por invocación como mucho, en la salida de error, con su salto (FR-008).

| Línea | Forma | Ejemplo (bytes con el salto) |
|---|---|---|
| FR-006 | `aviso: kitlegal <binario> ha actualizado las skills globales, que eran de <nombra>; se aplican en la próxima conversación` | `v0.2.0`, `v0.1.0`: 119 |
| FR-009 | `aviso: kitlegal <binario> se actualizarían las skills globales, que eran de <nombra>; se aplican en la próxima conversación` | `v0.2.0`, `v0.1.0`: 122 |
| de hoy | la de `instalacion.Aviso`, sin cambiar un byte (FR-007) | global, `v0.1.0`, `v0.2.0`: 123 |

`<binario>` es la versión del binario y `<nombra>`, la que nombra la línea de hoy en ese estado (`versionQueDifiere`),
las dos tal como están.

## §6 · Estados del global y transiciones

| Estado del global antes de una consulta (`Reinstalar`) | Después | Línea |
|---|---|---|
| sin tocar, alguna versión comparada menor y ninguna mayor | como `install -g` del pedido | FR-006 |
| igual que el anterior con algún fichero que falta | como `install -g` del pedido | FR-006 |
| con un conflicto de `Planificar` o una retirada «sin retirar» | sin cambios | de hoy |
| alguna versión comparada mayor que el binario, o sin forma, o solo metadatos distintos | sin cambios | de hoy |
| candado tomado (menos de 10 minutos) | sin cambios | de hoy |
| una escritura que falla a mitad | lo que `install -g` completa sin conflictos | de hoy (del estado anterior) |
| todas las versiones comparadas iguales a la del binario | sin cambios | ninguna |

Con un manifiesto local que el aviso compara, el global no se mira y nada cambia (FR-014). Con un local que existe y
no se puede usar (`./.agents` o `./.agents/skills` que no es un directorio real, o un `kitlegal.json` ilegible), el
global tampoco se mira, nada cambia y no hay ninguna línea, como en H19 (FR-003 a, «Edge Cases»). Con `Ensayar` y
`SoloAvisar`, ningún estado cambia (FR-013).
