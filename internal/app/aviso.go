package app

import (
	"os"

	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/disco"
)

// Avisador da la línea del aviso de versión y si hay que darla. Lo registra la
// raíz de composición en el Registro (Registro.Avisar) y lo llama el kernel en
// las invocaciones que lo buscan, que la escribe en la salida de error
// (contracts/aviso.md §1; research.md D5). Nunca devuelve error: lo que no le
// permite decidir es que no hay aviso (data-model §7).
type Avisador func() (linea string, hay bool)

// AvisoDeVersion es el avisador del binario distribuido y del de e2e: el Aviso
// del dominio sobre el disco real de internal/disco, con HOME tal como está en
// el entorno en cada llamada, y con la versión y lo empotrado de las
// dependencias de skills, las mismas con las que el applet instala, de modo que
// el aviso compara con lo que una instalación dejaría (contracts/aviso.md §2-§4;
// FR-070, FR-071). Componerlo no lee nada.
//
// Una versión sin forma SemVer es la de un binario de desarrollo, que no compara
// (FR-073): el avisador lo decide antes de nada, así que ese binario no lee ni lo
// empotrado ni el disco. Unas dependencias sin lo empotrado, o con lo empotrado
// que no se puede leer, son un defecto de composición que el applet skills da
// como fallo inesperado al ejecutarse y que TestSkillsEmpotradas impide
// publicar; el aviso no tiene con qué comparar y no avisa, porque nunca es un
// error ni cambia la invocación que lo busca (FR-072).
func AvisoDeVersion(dependencias DependenciasDeSkills) Avisador {
	return func() (string, bool) {
		if !instalacion.FormaSemVer(dependencias.Version) || dependencias.Skills == nil {
			return "", false
		}

		empotradas, err := skillsEmpotradasDe(dependencias.Skills)
		if err != nil {
			return "", false
		}

		return instalacion.Aviso(disco.Lector{}, os.Getenv(variableHome), dependencias.Version, empotradas)
	}
}
