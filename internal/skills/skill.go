package skills

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Lo que forma una skill en el repositorio (data-model §1).
const (
	// carpetaDeLasSkills es el directorio de skills, relativo a la raíz del
	// repositorio: cada directorio que tiene dentro es una skill.
	carpetaDeLasSkills = "skills"

	// ficheroDeLaSkill es el fichero con el frontmatter, el protocolo y las
	// reglas de una skill, dentro de su directorio.
	ficheroDeLaSkill = "SKILL.md"
)

// Skill es una skill del repositorio cargada: un directorio del directorio de
// skills y su SKILL.md (data-model §1).
type Skill struct {
	// Nombre es el nombre del directorio de la skill.
	Nombre string

	// Directorio es la ruta del directorio de la skill:
	// <raíz>/skills/<Nombre>.
	Directorio string

	// Contenido son los bytes de SKILL.md, tal cual.
	Contenido []byte
}

// DefectoDeSkill es un defecto de una skill del repositorio: nombra la skill y
// dice qué le falla (FR-040, FR-041).
type DefectoDeSkill struct {
	// Skill es el nombre del directorio de la skill.
	Skill string

	// Defecto dice qué falla, como «falta SKILL.md» o «falta name».
	Defecto string
}

// Error presenta la skill y el defecto, como «boe-legislacion: falta name».
func (d *DefectoDeSkill) Error() string {
	return d.Skill + ": " + d.Defecto
}

// Listar devuelve los nombres de las skills del repositorio de raíz, en orden de
// nombre: todo directorio del directorio de skills es una skill, y ninguna otra
// entrada lo es (data-model §1). Un repositorio sin directorio de skills no
// tiene ninguna, sin error; el error queda para un directorio de skills que no
// se puede listar, y lo nombra.
func Listar(raiz string) ([]string, error) {
	directorio := filepath.Join(raiz, carpetaDeLasSkills)

	entradas, err := os.ReadDir(directorio)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("el directorio de skills %s no se puede listar: %w", directorio, err)
	}

	var nombres []string

	for _, entrada := range entradas {
		if entrada.IsDir() {
			nombres = append(nombres, entrada.Name())
		}
	}

	return nombres, nil
}

// Cargar lee el SKILL.md de la skill de nombre, uno de los que da Listar. Una
// skill sin SKILL.md, o cuyo SKILL.md no es un fichero regular, es un
// DefectoDeSkill; cualquier otro error, el de un fichero que no se puede
// consultar o leer, nombra la skill y lo envuelve.
func Cargar(raiz, nombre string) (Skill, error) {
	directorio := filepath.Join(raiz, carpetaDeLasSkills, nombre)
	ruta := filepath.Join(directorio, ficheroDeLaSkill)

	estado, err := os.Lstat(ruta)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return Skill{}, &DefectoDeSkill{Skill: nombre, Defecto: "falta " + ficheroDeLaSkill}
	case err != nil:
		return Skill{}, fmt.Errorf("%s: %s no se puede consultar: %w", nombre, ficheroDeLaSkill, err)
	case !estado.Mode().IsRegular():
		return Skill{}, &DefectoDeSkill{Skill: nombre, Defecto: ficheroDeLaSkill + " no es un fichero regular"}
	}

	// filepath.Clean es lo que el control de rutas reconoce como saneado antes de
	// abrir un fichero (gosec G304); el nombre es el de un directorio del de
	// skills, sin separadores de ruta.
	contenido, err := os.ReadFile(filepath.Clean(ruta))
	if err != nil {
		return Skill{}, fmt.Errorf("%s: %s no se puede leer: %w", nombre, ficheroDeLaSkill, err)
	}

	return Skill{Nombre: nombre, Directorio: directorio, Contenido: contenido}, nil
}

// ContarLineas es el número de líneas de un contenido (data-model §1.2): el de
// saltos de línea, más uno si no está vacío y no termina en salto.
func ContarLineas(contenido []byte) int {
	lineas := bytes.Count(contenido, []byte("\n"))
	if len(contenido) > 0 && contenido[len(contenido)-1] != '\n' {
		lineas++
	}

	return lineas
}
