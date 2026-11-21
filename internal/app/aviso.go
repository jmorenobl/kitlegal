package app

import (
	"os"
	"time"

	"github.com/jmorenobl/kitlegal/internal/cache"
	"github.com/jmorenobl/kitlegal/internal/core/instalacion"
	"github.com/jmorenobl/kitlegal/internal/disco"
)

// El candado de la reinstalación de las skills globales: un fichero en el
// directorio de la caché, y nunca en ~/.agents/skills, que cualquier
// invocación que lo encuentre de hace menos de la caducidad respeta y que, de
// hace la caducidad o más, toma por huérfano (FR-010 de H38;
// contracts/aviso-y-reinstalacion.md §4; research.md D4 y D5 de H38).
const (
	// ficheroDelCandado es su nombre en el directorio de la caché.
	ficheroDelCandado = "skills-global.lock"

	// caducidadDelCandado es la edad a partir de la cual es huérfano.
	caducidadDelCandado = 10 * time.Minute
)

// Avisador da la línea del aviso de versión y si hay que darla. Lo registra la
// raíz de composición en el Registro (Registro.Avisar) y lo llama el kernel en
// las invocaciones que lo buscan, que la escribe en la salida de error
// (contracts/aviso.md §1; research.md D5). Nunca devuelve error: lo que no le
// permite decidir es que no hay aviso (data-model §7).
//
// permiso es lo que la invocación permite hacer al aviso, y lo decide el
// kernel con lo que ya analizó: solo avisar, ensayar o reinstalar las skills
// globales (contracts/aviso-y-reinstalacion.md §1; data-model §1 y research.md
// D2 y D12 de H38). El avisador no hace más de lo que su permiso le deja.
type Avisador func(permiso instalacion.PermisoDelAviso) (linea string, hay bool)

// AvisoDeVersion es el avisador del binario distribuido y del de e2e: la
// Reinstalacion del dominio sobre el disco real de internal/disco, con HOME tal
// como está en el entorno en cada llamada, y con la versión, lo empotrado y el
// creador de enlaces de las dependencias de skills, las mismas con las que el
// applet instala, más los nombres retirados con los que instala: así el aviso
// compara con lo que una instalación dejaría y reinstala lo que kitlegal skills
// install -g instalaría (contracts/aviso.md §2-§4; FR-070, FR-071;
// contracts/aviso-y-reinstalacion.md §2; FR-005 y research.md D12 de H38). El
// candado es el de la caché, con ficheroDelCandado y caducidadDelCandado.
// Componerlo no lee nada.
//
// Con SoloAvisar da la línea de hoy; con Ensayar, la de la reinstalación que
// haría, sin escribir nada; y con Reinstalar, reinstala las skills globales que
// se pueden reinstalar y lo dice (FR-001, FR-006, FR-009 de H38). Lo que falla
// se queda en la línea de hoy (FR-012 de H38).
//
// Una versión sin forma SemVer es la de un binario de desarrollo, que no compara
// (FR-073): el avisador lo decide antes de nada, así que ese binario no lee ni lo
// empotrado ni el disco. Unas dependencias sin lo empotrado, o con lo empotrado
// que no se puede leer, son un defecto de composición que el applet skills da
// como fallo inesperado al ejecutarse y que TestSkillsEmpotradas impide
// publicar; el aviso no tiene con qué comparar y no avisa, porque nunca es un
// error ni cambia la invocación que lo busca (FR-072). Y unas sin Enlazador,
// otro defecto que el applet da como inesperado (errSinEnlazador), no tienen
// con qué reinstalar: el aviso da solo la línea de hoy, con cualquier permiso,
// sin entrar en pánico en una ruta de usuario (constitución IV).
func AvisoDeVersion(dependencias DependenciasDeSkills) Avisador {
	return func(permiso instalacion.PermisoDelAviso) (string, bool) {
		if !instalacion.FormaSemVer(dependencias.Version) || dependencias.Skills == nil {
			return "", false
		}

		empotradas, err := skillsEmpotradasDe(dependencias.Skills)
		if err != nil {
			return "", false
		}

		home := os.Getenv(variableHome)

		if dependencias.Enlazador == nil {
			return instalacion.Aviso(disco.Lector{}, home, dependencias.Version, empotradas)
		}

		reinstalacion := instalacion.Reinstalacion{
			Disco:     disco.Lector{},
			Enlazador: dependencias.Enlazador,
			Escritor:  disco.NuevoEscritor(dependencias.Enlazador),
			Candado: func() (func(), bool) {
				return cache.TomarCandado(ficheroDelCandado, caducidadDelCandado)
			},
			Home:       home,
			Version:    dependencias.Version,
			Empotradas: empotradas,
			Retiradas:  skillsRetiradas,
		}

		return reinstalacion.Avisar(permiso)
	}
}
