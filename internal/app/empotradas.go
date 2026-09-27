package app

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal"
	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
)

const (
	// carpetaDeSkills es la de las skills dentro de lo empotrado, con una
	// carpeta por skill: skills/<skill>/.
	carpetaDeSkills = "skills"

	// ficheroDeLaSkill es el que hace de una carpeta de skills/ una skill
	// (FR-004).
	ficheroDeLaSkill = "SKILL.md"

	// carpetaDeReferencias es la de las referencias de una skill, relativa a su
	// carpeta.
	carpetaDeReferencias = "references"
)

// skillsEmpotradas son las skills que el binario lleva dentro (kitlegal.Skills),
// leídas como las recibe el dominio de la instalación: los bytes de lo empotrado,
// que no dependen de ningún clon ni del directorio de trabajo (FR-003).
func skillsEmpotradas() ([]instalacion.SkillEmpotrada, error) {
	return skillsEmpotradasDe(kitlegal.Skills())
}

// skillsEmpotradasDe lee de un árbol con la forma de lo empotrado una skill por
// cada carpeta skills/<skill>/ cuyo SKILL.md es un fichero, en orden de nombre;
// una carpeta sin él no es una skill y no entra (FR-004). Los ficheros de cada
// skill son su SKILL.md y todo fichero bajo su references/, a cualquier
// profundidad, con la ruta relativa a la carpeta de la skill, en orden de ruta y
// con su huella (contracts/skills-e-invocacion.md §1); nada más de la carpeta
// entra (FR-001).
//
// Recibe el árbol para que las ramas que lo empotrado no da nunca —una carpeta
// sin SKILL.md, una lectura que falla— se puedan ejercer sin tocarlo. Lo que no
// se puede leer es un error que nombra la ruta, nunca una skill o un fichero de
// menos.
func skillsEmpotradasDe(arbol fs.FS) ([]instalacion.SkillEmpotrada, error) {
	// fs.ReadDir las da en orden de nombre.
	carpetas, err := fs.ReadDir(arbol, carpetaDeSkills)
	if err != nil {
		return nil, fmt.Errorf("app: no se puede listar %s en lo empotrado: %w", carpetaDeSkills, err)
	}

	var empotradas []instalacion.SkillEmpotrada

	for _, carpeta := range carpetas {
		if !carpeta.IsDir() {
			continue
		}

		directorio := path.Join(carpetaDeSkills, carpeta.Name())

		esSkill, err := tieneSkillMd(arbol, directorio)
		if err != nil {
			return nil, err
		}

		if !esSkill {
			continue
		}

		ficheros, err := ficherosDeLaSkill(arbol, directorio)
		if err != nil {
			return nil, fmt.Errorf("app: no se puede leer la skill %s en lo empotrado: %w", carpeta.Name(), err)
		}

		empotradas = append(empotradas, instalacion.SkillEmpotrada{Nombre: carpeta.Name(), Ficheros: ficheros})
	}

	return empotradas, nil
}

// tieneSkillMd dice si la carpeta tiene un SKILL.md que es un fichero. Que no
// exista es un no; que no se pueda examinar, un error.
func tieneSkillMd(arbol fs.FS, directorio string) (bool, error) {
	ruta := path.Join(directorio, ficheroDeLaSkill)

	estado, err := fs.Stat(arbol, ruta)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("app: no se puede examinar %s en lo empotrado: %w", ruta, err)
	}

	return estado.Mode().IsRegular(), nil
}

// ficherosDeLaSkill son el SKILL.md y los ficheros de references/ de la carpeta
// de una skill, en orden de ruta. El recorrido no basta para ese orden: da
// references/a/b.md antes que references/a-b.md, y «-» es menor que «/».
func ficherosDeLaSkill(arbol fs.FS, directorio string) ([]instalacion.FicheroEmpotrado, error) {
	var ficheros []instalacion.FicheroEmpotrado

	err := fs.WalkDir(arbol, directorio, func(ruta string, entrada fs.DirEntry, err error) error {
		if err != nil || ruta == directorio {
			return err
		}

		relativa := strings.TrimPrefix(ruta, directorio+"/")

		switch {
		case entrada.IsDir() && relativa != carpetaDeReferencias && !esReferencia(relativa):
			return fs.SkipDir
		case entrada.IsDir() || (relativa != ficheroDeLaSkill && !esReferencia(relativa)):
			return nil
		}

		contenido, err := fs.ReadFile(arbol, ruta)
		if err != nil {
			return err
		}

		ficheros = append(ficheros, instalacion.NuevoFicheroEmpotrado(relativa, contenido))

		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(ficheros, porRuta)

	return ficheros, nil
}

// esReferencia dice si la ruta, relativa a la carpeta de una skill, está bajo su
// references/.
func esReferencia(relativa string) bool {
	return strings.HasPrefix(relativa, carpetaDeReferencias+"/")
}

// porRuta ordena los ficheros de una skill por su ruta, byte a byte.
func porRuta(a, b instalacion.FicheroEmpotrado) int {
	return strings.Compare(a.Ruta, b.Ruta)
}
