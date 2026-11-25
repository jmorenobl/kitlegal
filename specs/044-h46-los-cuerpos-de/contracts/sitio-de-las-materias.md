# Contrato: el sitio de las materias, comprobado (FR-005 a FR-009)

## 1. El defecto (FR-005, FR-006)

En `defectosDelCuerpo(skill, contenido)` (`internal/skills/sincronia.go`), sobre el contenido regenerado, detrás de los
dos defectos de hoy:

- Si el cuerpo lleva el par de marcas de los paquetes en pareja (`regionDeLosPaquetes.buscar(cuerpo)` sin error y con
  región), se cuenta `n = runas(cuerpo) − runas(lo que hay entre las dos líneas de las marcas)`. Las marcas, la tabla de
  comandos generada y todo lo demás del cuerpo cuentan.
- Con `n > 9600` (`maximoSinLosPaquetes`), un defecto:

  ```text
  <skill>: SKILL.md: cuerpo sin las líneas de los paquetes de <n> caracteres (máximo 9600)
  ```

  Con 9600, ninguno.
- Sin las marcas, ninguno, con cualquier tamaño (`consulta-juridica`, que lleva las de la base, o una skill nueva sin
  ellas). Con unas marcas que no van en pareja, ninguno nuevo: el de las marcas ya sale de `paquetes.regenerar`.

Cuando salen el de 11 000 y este, el de 11 000 va delante (FR-005). Es un defecto de `skills-check`: `make
skills-check`, y con él `make ci`, sale en rojo; `make skills-sync` no escribe nada mientras haya defectos, como hoy.

**Ejemplo**: `plazos-y-territorio` ensanchado fuera de las marcas hasta 9601 sin su línea de lo administrativo (9695
con ella):

```text
plazos-y-territorio: SKILL.md: cuerpo sin las líneas de los paquetes de 9601 caracteres (máximo 9600)
```

**Quién lo lee y cuánto**: quien cambia el cuerpo de una skill de capacidad o escribe un paquete, persona o run, una vez
por `make ci` o `make skills-check`; una línea de 93 a 103 bytes por skill que se pasa (93 con `normativa`, 103 con
`plazos-y-territorio`), tres como mucho. No depende del uso del binario. **Cuándo deja de darse**: cuando el cuerpo
vuelve a 9600 o menos sin las líneas de los paquetes; con los textos del hito (9498, 9483 y 9473) no se da.

## 2. `casosDePresupuestos` (FR-008)

Subtest `presupuestos` de `TestSkillsDelRepositorio` (`internal/app/skills_test.go`), con los casos de data-model §3:
para las skills con las marcas, 9600 + p sin defectos, 9601 + p con solo el de FR-005 y 11 001 con los dos en su orden;
para las demás, 11 000 y 11 001 como hoy; 499 y 500 líneas como hoy. Lo corre también el paso de sincronía del guion
`h45-paquete-de-mas` en su copia, con el paquete de prueba en `normativa`.

Mutantes (research P1; el registro, `mutantes.md`): el tope a 9601 o a 9599, la regla desactivada, sin restar las líneas
de los paquetes y la regla también sin marcas ponen el subtest en rojo, cada uno en los casos que dice P1.

## 3. `TestVerticalDeEnsayo` (FR-009)

`internal/skills/vertical_de_ensayo_test.go`: fuera `comprobarElSitioDeOtroPaquete`, su llamada, el import
`unicode/utf8` y la frase del comentario del test que la describe. Lo demás —la sincronía sin defectos ni deriva, la eval
que se carga, ningún `.go`, ningún `SKILL.md` cambiado fuera de las marcas, el reparto de la fila con norma, `skills/`
y `data/` del repositorio sin cambios, y sus dos mutantes— no cambia.
