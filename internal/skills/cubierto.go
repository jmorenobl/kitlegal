package skills

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// regionDeLoCubierto es la de lo que cubre la base, en el SKILL.md que lleve sus
// marcas, hoy el de la skill de entrada (contracts/datos-y-regiones.md §3;
// data-model §3; research D1 de H33): una línea por fila de lo cubierto de
// data/cobertura.yaml.
var regionDeLoCubierto = regionDeMarcas{
	nombre: "lo que cubre la base",
	inicio: "<!-- lo que cubre la base: lo genera make skills-sync -->",
	fin:    "<!-- fin de lo que cubre la base -->",
}

// regenerarLoCubierto es el SKILL.md de contenido de la skill con lo que hay entre
// las marcas de la región de lo cubierto regenerado (lineasDeLoCubierto), y los
// defectos de la región, con SKILL.md delante —sus marcas que no van en pareja,
// el de la región sin uno de sus dos YAML de data/—, y los de esos YAML, con su
// fichero delante: los de su lector y el de cada materia que no es un vertical.
// Con cualquiera, la región queda como está; sin sus marcas, el SKILL.md no la
// lleva y no hay nada que regenerar. El error queda para lo que no se puede leer.
func (d *datosDeLasRegiones) regenerarLoCubierto(skill string, contenido []byte) (
	regenerado []byte, deLaRegion, deLosDatos []*DefectoDeSkill, err error,
) {
	entre, lleva, err := regionDeLoCubierto.buscar(contenido)

	switch {
	case err != nil:
		return contenido, defectosDelError(skill, ficheroDeLaSkill, err), nil, nil
	case !lleva:
		return contenido, nil, nil, nil
	}

	lineas, err := d.lineasDeLoCubierto()
	if err != nil {
		deLaRegion, deLosDatos, err = defectosDeLosDatosQueFaltan(skill,
			ficheroDeLaSkill+": la región "+regionDeLoCubierto.nombre, err)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s: %s: %w", skill, ficheroDeLaSkill, err)
		}

		return contenido, deLaRegion, deLosDatos, nil
	}

	return entre.sustituir(contenido, lineas), nil, nil, nil
}

// lineasDeLoCubierto son las líneas de la región de lo cubierto: una por fila de
// lo cubierto de data/cobertura.yaml, en su orden, «- **<capacidad>**:
// <alcance>», con la capacidad con su primera letra en mayúscula y sin su
// materia. O lo que lo impide: un *defectoDeLosDatos si falta data/cobertura.yaml
// o data/normas.yaml, si tienen defectos o si alguna fila lleva una materia que
// no es el vertical de ninguna norma —una causa por fila, en su orden (research
// D4 de H33)—, o el error de su lectura.
func (d *datosDeLasRegiones) lineasDeLoCubierto() ([]byte, error) {
	cobertura, err := usar(d.cobertura, ficheroDeLaCobertura)
	if err != nil {
		return nil, err
	}

	normas, err := usar(d.normas, ficheroDeLasNormas)
	if err != nil {
		return nil, err
	}

	var (
		lineas      bytes.Buffer
		sinVertical []error
	)

	for _, fila := range cobertura.Cubierto {
		if fila.Materia != "" && !slices.ContainsFunc(normas, func(norma Norma) bool {
			return norma.Vertical == fila.Materia
		}) {
			sinVertical = append(sinVertical, errors.New("la fila de lo cubierto de la capacidad «"+fila.Capacidad+
				"» lleva la materia «"+fila.Materia+"», que no es el vertical de ninguna norma de "+carpetaDeLosDatos+"/"+
				ficheroDeLasNormas))
		}

		lineas.WriteString("- **" + conLaPrimeraEnMayuscula(fila.Capacidad) + "**: " + fila.Alcance + "\n")
	}

	if len(sinVertical) > 0 {
		return nil, &defectoDeLosDatos{fichero: ficheroDeLaCobertura, causa: errors.Join(sinVertical...)}
	}

	return lineas.Bytes(), nil
}

// conLaPrimeraEnMayuscula es el texto con su primera letra en mayúscula.
func conLaPrimeraEnMayuscula(texto string) string {
	_, tamano := utf8.DecodeRuneInString(texto)

	return strings.ToUpper(texto[:tamano]) + texto[tamano:]
}
