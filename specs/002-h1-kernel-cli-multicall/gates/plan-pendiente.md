# Pendientes de decisión humana — gate `plan` de H1

**Estado: ninguno.** Los siete motivos del rechazo se han corregido en los artefactos del plan sin
necesidad de intervención humana. Este fichero deja constancia de por qué, para que el siguiente juez
pueda comprobarlo en lugar de suponerlo.

El criterio aplicado es el §4 de «Criterio de decisión autónoma» de la constitución: se escala cuando la
duda afecta a **alcance, frontera humana, privacidad, términos de uso de una fuente, reglas de anomalías o
una decisión ya cerrada de `CLAUDE.md`**. Ninguno de los siete motivos toca nada de eso:

| Motivo del rechazo | Qué clase de decisión era | Por qué no escala |
|---|---|---|
| Excepción de `forbidigo` para el `main` del binario de e2e | Configuración de un control mecánico | No cambia ninguna regla de `docs/ROADMAP.md` §2: la aplica a una raíz de composición, que es lo que la regla ya autoriza para `cmd/`. La excepción resultante es **más estrecha** que la que el plan tenía escrita, porque además retira `internal/cli/**` |
| `*slog.Logger` dentro de `schema.Contexto` | Reparto de tipos entre paquetes | La constitución §IV ya fija el criterio («dominio puro, sin I/O»): no había que elegir un valor nuevo, sino aplicar el que ya estaba escrito. Se aplica y se hace mecánico con `depguard` |
| Canal de la descripción de `--dry-run` | Implementación de un requisito ya cerrado | FR-022 y la respuesta Q4 del `clarify` fijan el comportamiento observable («siempre visible, en la salida de error, sin depender de `--verbose` ni `KITLEGAL_LOG`»); lo que faltaba era el mecanismo que lo cumpliera, que es competencia del plan |
| Cómo se conocen `--json` y `--verbose` antes de la gramática | Mecánica interna de `internal/cli` | No añade interfaz nueva ni cambia comportamiento observable alguno: hace posible el que SC-014 y FR-045 ya exigen |
| Verbo por omisión | Contrato del applet | La entrega literal del hito (`kitlegal echo hola --json`) la fija `docs/ROADMAP.md` §4; el spec delega en el plan la forma de cumplirla. Es la única de las siete que añade un campo al contrato, y por eso queda registrada en *Complexity Tracking* con sus alternativas rechazadas |
| Exclusión de `errcheck` para `fmt.Fprint*` | Endurecimiento de un control | FR-057 prohíbe exclusiones **nuevas** sin justificar; retirar una vieja cuya justificación caducó va en la dirección que la constitución pide (criterio §1), no en contra |
| `python3` en los prerrequisitos del quickstart | Documentación de una herramienta de verificación | No es dependencia del producto ni de `make ci`: solo del guion de validación manual |

**Consecuencia para el workflow:** el gate `plan` puede volver a juzgarse sin pausa humana. Si el juez
siguiente considerase que alguna de estas siete correcciones sí toca alcance —en particular el campo
`PorOmision` del contrato del applet, que es la de mayor superficie—, ese sería el punto en el que
procedería escalar, y este fichero es el lugar donde anotarlo.
