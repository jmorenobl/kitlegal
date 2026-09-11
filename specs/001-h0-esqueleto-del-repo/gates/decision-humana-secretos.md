# Decisión humana · línea base de la detección de secretos

**Fecha**: 2026-09-10 · **Hito**: H0 · **Escalada por**: paso `reparar` del bucle de tareas, intento 1 de T001
**Criterio aplicado**: constitución, «Criterio de decisión autónoma» §4 (escalar en lugar de adivinar: alcance)

## Qué se escaló

Con T001 implementada, siete de los ocho controles de `make ci` quedaron en verde. `make secrets` no:
gitleaks encontró cuatro hallazgos. El agente no los resolvió porque hacerlo exigía decidir el alcance del
hito, y lo escaló en lugar de decidir. Fue lo correcto: `plan.md` y `research.md` D10 afirmaban que H0 no
crea `.gitleaksignore` porque «no hay ningún falso positivo que excluir», y esa premisa era falsa.

## Hechos verificados

Ejecutando `gitleaks dir .` (v8.30.1) sobre el árbol y consultando los ficheros implicados:

1. Los cuatro hallazgos son **dos** líneas de ejemplo, marcadas `// DON'T`, en
   `skills/golang-security/references/secrets.md` del paquete vendorizado `samber/cc-skills-golang`.
   Son documentación que enseña qué **no** hacer con credenciales. No son credenciales.
2. Aparecían por duplicado porque el repositorio contenía **dos copias completas** del mismo paquete:
   `agent/` y `.agents/`, 297 ficheros cada una, con el mismo contenido en distinto formato de frontmatter.
   Ambas entraron en el commit inicial `812a7d2`, anterior al hito.
3. A `agent/` no la referenciaba nada: `CLAUDE.md` nombra `.agents/`, los enlaces simbólicos de
   `.claude/skills/` apuntan a `.agents/` y `skills-lock.json` no guarda rutas locales.
4. `gitleaks dir` **no respeta `.gitignore`**: dejar de versionar un directorio no lo saca del escaneo.
   Solo borrarlo del disco. Una ruta desnuda en `.gitleaksignore` tampoco excluye nada; la huella
   `fichero:regla:línea` sí.

## Opciones consideradas

| Opción | Qué hacía | Por qué no |
|---|---|---|
| Solo `.gitleaksignore` con 4 huellas | Cambio mínimo, sin borrar nada | Conserva 297 ficheros duplicados que nada usa y duplica las excepciones |
| Allowlist de rutas en `.gitleaks.toml` | Excluye los árboles vendorizados enteros | Deja de mirar ese contenido para siempre y contradice D10, que exige exclusión por hallazgo |
| Sacar las skills del repositorio | Las reinstala `skills-lock.json` | Cambia la configuración de trabajo del usuario y excede con mucho el hito |

## Decisión

**Borrar `agent/` del repositorio y del disco, y crear `.gitleaksignore` con las dos huellas restantes**,
justificada cada una con su comentario y sin desactivar ninguna regla.

Motivos: `agent/` es una copia muerta que nada referencia, de modo que eliminarla no cambia el
comportamiento de ninguna herramienta y reduce las excepciones a la mitad; las dos huellas restantes son
exactamente el procedimiento que `research.md` D10 prescribe, son auditables y vuelven a saltar si el
paquete vendorizado cambia, que es la propiedad deseable. El control no se debilita en ningún punto.

## Efecto sobre el alcance de H0

El hito **sí** entrega `.gitleaksignore`. No es alcance nuevo: un control de secretos que no puede quedar
en verde sobre el árbol existente no está instalado, y H0 existe para instalarlo. Lo que cambia es una
premisa equivocada del plan, no el objetivo del hito. SC-008 se sigue cumpliendo porque el fichero nace
con contenido justificado y no como marcador de posición vacío.

Artefactos alineados con esta decisión: `plan.md` (estructura y párrafo dedicado), `data-model.md`
(inventario y ausencias deliberadas), `research.md` D10, `tasks.md` (rutas de T001 y «Notas») y
`quickstart.md` (nota del escenario 9).

## Riesgo residual

Si la herramienta que instaló las skills vuelve a crear `agent/`, `make secrets` volverá a dar cuatro
hallazgos y `make ci` se pondrá en rojo con un mensaje que nombra los ficheros. Es ruidoso pero no
silencioso, que es la propiedad que se prefiere. Si ocurre de forma recurrente, la decisión a revisar es
si el paquete de skills debe estar versionado en este repositorio.

## Contador de intentos

El intento 1 de T001 se descartó: la tarea no falló por un defecto de implementación, sino por estar mal
delimitada. `gates/tareas-intentos.json` se puso a cero para T001 al aplicar esta decisión.
