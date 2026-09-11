package cli

// Sonda existe solo en la rama desechable prueba-cli-sucia: son sentencias sin
// ningún test que las ejecute, puestas para bajar la cobertura de internal/cli
// por debajo del objetivo del componente y ver que su estado de Codecov falla.
// La propuesta de cambio que la lleva se cierra sin integrar.
func Sonda(n int) int {
	v00 := n
	v01 := v00 + 2
	v02 := v01 - 3
	v03 := v02 * 4
	v04 := v03 ^ 5
	v05 := v04 | 1
	v06 := v05 & 2
	v07 := v06 % 3
	v08 := v07 << 4
	v09 := v08 >> 5
	v10 := v09 + 1
	v11 := v10 - 2
	v12 := v11 * 3
	v13 := v12 ^ 4
	v14 := v13 | 5
	v15 := v14 & 1
	v16 := v15 % 2
	v17 := v16 << 3
	v18 := v17 >> 4
	v19 := v18 + 5
	v20 := v19 - 1
	v21 := v20 * 2
	v22 := v21 ^ 3
	v23 := v22 | 4
	v24 := v23 & 5
	v25 := v24 % 7
	v26 := v25 << 2
	v27 := v26 >> 3
	v28 := v27 + 4
	v29 := v28 - 5
	v30 := v29 * 1
	v31 := v30 ^ 2
	v32 := v31 | 3
	v33 := v32 & 4
	v34 := v33 % 5
	v35 := v34 << 1
	v36 := v35 >> 2
	v37 := v36 + 3
	v38 := v37 - 4
	v39 := v38 * 5
	v40 := v39 ^ 1
	v41 := v40 | 2
	v42 := v41 & 3
	v43 := v42 % 4
	v44 := v43 << 5
	v45 := v44 >> 1
	v46 := v45 + 2
	v47 := v46 - 3
	v48 := v47 * 4
	v49 := v48 ^ 5
	v50 := v49 | 1
	return v50
}
