// Package kitlegal es la raíz del módulo y lleva dentro del binario las skills
// del repositorio: por cada skills/<skill>/, su SKILL.md y todo fichero de su
// references/, byte a byte como están en el árbol al compilar (FR-001 a FR-004;
// contracts/skills-e-invocacion.md §1; H19 research.md D3).
//
// La directiva //go:embed vive aquí, en un único fichero de la raíz, porque un
// patrón no puede subir de directorio y skills/ no lleva código: ningún paquete
// de internal/ alcanzaría los ficheros (FR-002). El paquete no interpreta nada:
// entrega lo empotrado tal cual, y quien lo lee a skills es internal/app, que
// ignora todo directorio sin SKILL.md (FR-004). Que lo empotrado sea exactamente
// el árbol lo fija TestSkillsEmpotradas.
package kitlegal

import (
	"embed"
	"io/fs"
)

// skills es lo empotrado. El patrón no empotra scripts/ ni nada más de cada
// skill, y references/* empotra también sus ficheros con punto; que no quede
// fuera nada de lo que FR-001 pide, ni entre nada de un directorio sin SKILL.md,
// lo comprueba TestSkillsEmpotradas contra el árbol.
//
//go:embed skills/*/SKILL.md skills/*/references/*
var skills embed.FS

// Skills devuelve lo empotrado tal cual, con rutas skills/<skill>/…. Es una
// función y no una variable exportada para que nada pueda sustituir lo
// empotrado desde fuera del paquete.
func Skills() fs.FS {
	return skills
}
