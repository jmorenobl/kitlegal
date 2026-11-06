# Contrato: las variantes de las tres familias

FR-020 a FR-023; SC-006, SC-008. Research D8, D9, V1; data-model §1 a §3.

## 1. Dónde viven

```text
laboratorio/
├── variantes.yaml, variantes/, evals/                 # H39, sin cambios
├── normativa/                                         # H40: las cuatro de hoy, sin cambiar un byte
│   └── normativa-{control,remisiones-en-orden,cuerpo-corto,testigo-en-referencia}/skills/normativa/…
├── plazos-y-territorio/                               # NUEVO
│   └── plazos-y-territorio-control/skills/plazos-y-territorio/
│       ├── SKILL.md
│       └── references/{administrativo,calculo_de_plazos,jerarquia_normativa,municipios}.md
└── jurisprudencia/                                    # NUEVO
    └── jurisprudencia-control/skills/jurisprudencia/
        ├── SKILL.md
        └── references/{consulta_preparada,documento_aportado}.md
```

Las dos `<skill>-control` nuevas son copias de `skills/<skill>` del árbol de partida, que es el de `v0.9.0` (research V1),
hechas antes de cambiar ninguna skill: `cp -R skills/<skill>/. laboratorio/<skill>/<skill>-control/skills/<skill>/`.
Ninguna `<skill>-producto` tiene directorio: es `skills/<skill>` (data-model §3). Nada de esto lo empotra el binario, lo
instala `kitlegal skills install`, lo regenera `make skills-sync` ni lo elige o mide el job de evals (FR-023), y el tope
de 11 000 caracteres no se aplica a ninguna variante versionada.

## 2. Controles

- **`TestVariantesDeControl`** (`internal/evals/laboratorio_familias_test.go`; FR-020, FR-022, SC-006, SC-008), por
  familia del producto:
  - la carpeta `laboratorio/<skill>/` tiene un directorio por variante versionada de la familia y nada más; ninguno es
    `<skill>-producto`, que la familia declara y no se versiona;
  - los ficheros de `<skill>-control` son, ruta a ruta, los de la tabla de huellas SHA-256 del test, escrita al copiarlos
    (17 ficheros: 9 de `normativa`, 5 de `plazos-y-territorio`, 3 de `jurisprudencia`): falla con la ruta del que
    difiere, falta o sobra;
  - los nombres de las variantes de las cuatro familias no se repiten.
- **`TestVariantesDeNormativa`** (H40; FR-021): sus subtests de las variantes declaradas, sus remisiones, las
  referencias de la familia y los nombres, sin cambiar ningún valor esperado. Sale el subtest «control es
  skills/normativa», que pasa a ser la huella de `TestVariantesDeControl` (FR-020: `normativa-control` «deja de
  compararse con `skills/normativa`»), y el de los directorios de la familia, que hace `TestVariantesDeControl` para
  las tres.
- **`TestLaboratorioFueraDelProducto`** (H39 y H40, generalizado; FR-023, SC-008): los nombres que no pueden ser un tramo
  de una ruta de lo que el binario empotra, de lo que deja `kitlegal skills install --dir`, de una skill que ejecute el
  job ni de una carpeta de `evals/` son los de las skills de H39 y los de los directorios de variante de las carpetas de
  las tres familias del producto; ninguna ruta de `laboratorio/` tiene un tramo `<skill>-producto` ni mide
  (`rutaQueMide`). Lo demás, como hoy: los ficheros de `laboratorio/` que son copia de uno de `skills/` quedan fuera de
  la comparación de bytes (research D12 de H40); tras el hito lo son las referencias que no cambian.
- **`TestSkillsDelProductoDesdeControl`** (contracts/remisiones.md §7; FR-007, FR-009, FR-010).

## 3. Uso, de fuera adentro

Las lee el agente de una sesión del sondeo: el `SKILL.md` de su variante una vez, al activar la skill, y cada referencia
cuando la abre, que es lo que se mide; en la aceptación, 520 sesiones, una variante por sesión. Su tamaño es fijo —el de
contracts/remisiones.md §1 y los ficheros de hoy— y no crece con el uso del producto: cientos de normas y miles de
bloques consultados no las tocan. No son una señal. Ninguna skill del producto ni el binario que se distribuye las leen.
