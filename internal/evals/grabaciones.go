package evals

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/source/boe"
)

// Los conjuntos de grabaciones de data-model §7.3: el directorio de la fuente
// boe de cada uno, relativo al directorio de este paquete, que es donde go test
// ejecuta los tests desde los que se usan (research.md V46).
const (
	// GrabacionesDeH4 son las respuestas que grabó H4, dueño del conjunto; H5
	// solo las lee.
	GrabacionesDeH4 = "../../internal/source/boe/testdata/" + boe.NombreDeLaFuente

	// GrabacionesDeH5 son las que graba una persona con TestGrabarEvals en la
	// pausa de la tarea del manifiesto.
	GrabacionesDeH5 = "../../testdata/evals/" + boe.NombreDeLaFuente
)

// UnionDeGrabaciones son los dos conjuntos en el orden en que Preparar los copia
// a un mismo directorio para reproducir: H4 y, encima, H5. Solo comparten el
// nombre de GET_https_www.boe.es_robots.txt.json, que la reproducción no usa
// (data-model §7.3; contrato evals-y-grabaciones §3.4 y §5.1). Cada llamada
// devuelve una lista nueva.
func UnionDeGrabaciones() []string {
	return []string{GrabacionesDeH4, GrabacionesDeH5}
}

// Manifiesto es el manifiesto de grabación de H5 (data-model §7.2).
type Manifiesto struct {
	// Fuente es la fuente de lo que se graba, boe.NombreDeLaFuente.
	Fuente string `json:"fuente"`

	// Normas son las normas que se graban, en el orden del documento.
	Normas []EntradaDelManifiesto `json:"normas"`
}

// EntradaDelManifiesto es una norma que se graba: con qué se busca, por qué
// prefijo se reconoce su resultado, qué bloques se graban de ella y para qué
// eval y requisito.
type EntradaDelManifiesto struct {
	// Busqueda son los términos con los que boe buscar encuentra la norma.
	Busqueda string `json:"busqueda"`

	// TituloEmpiezaPor es el prefijo por el que empieza el título del único
	// resultado de la búsqueda que es la norma; de ese resultado sale su
	// identificador.
	TituloEmpiezaPor string `json:"titulo_empieza_por"`

	// Bloques son los ids de bloque que se graban de la norma; ninguno si
	// ninguna eval la cita.
	Bloques []string `json:"bloques"`

	// Para dice por qué está la entrada: la eval y el requisito que la
	// necesitan.
	Para string `json:"para"`
}

// LeerManifiesto es el único lector del manifiesto de grabación: lo usan el
// arnés que graba, la verificación de los identificadores de data/normas.yaml y
// la comprobación del propio manifiesto, así que los tres leen el mismo documento
// con las mismas reglas (research.md D11). Comprueba, en este orden, y devuelve
// el primer defecto como error, con un Manifiesto vacío (contrato
// evals-y-grabaciones §3.1):
//
//  1. el documento: un único valor JSON con la forma de Manifiesto, sin nada
//     detrás, sin miembros que no sean campos suyos y sin claves repetidas en
//     ningún objeto. encoding/json/v2 rechaza la clave repetida por omisión y,
//     con RejectUnknownMembers, el miembro desconocido, y nombra los dos con su
//     ruta JSON; encoding/json se quedaría con el último valor de una clave
//     repetida sin error (research.md V55);
//  2. la fuente es boe.NombreDeLaFuente;
//  3. normas tiene al menos una entrada;
//  4. cada entrada tiene busqueda, titulo_empieza_por y para no vacíos, y
//     bloques válidos para boe.ValidarBloque y sin repetir; bloques puede faltar;
//  5. ninguna norma está repetida: dos entradas con el mismo
//     titulo_empieza_por, o uno que empieza por el otro, podrían resolver el
//     mismo título. Se compara byte a byte con strings.HasPrefix, como al elegir
//     el resultado de la búsqueda.
//
// Todo error empieza por «manifiesto de grabación: »; los de los puntos 4 y 5
// nombran cada entrada por su posición, desde 1, y su titulo_empieza_por.
func LeerManifiesto(contenido []byte) (Manifiesto, error) {
	var manifiesto Manifiesto
	if err := json.Unmarshal(contenido, &manifiesto, json.RejectUnknownMembers(true)); err != nil {
		return Manifiesto{}, fmt.Errorf("manifiesto de grabación: no es un documento válido: %w", err)
	}

	if err := comprobarManifiesto(manifiesto); err != nil {
		return Manifiesto{}, fmt.Errorf("manifiesto de grabación: %w", err)
	}

	return manifiesto, nil
}

// comprobarManifiesto aplica los puntos 2 a 5 de LeerManifiesto a un documento
// ya decodificado.
func comprobarManifiesto(manifiesto Manifiesto) error {
	if manifiesto.Fuente != boe.NombreDeLaFuente {
		return fmt.Errorf("la fuente es %q y tiene que ser %q", manifiesto.Fuente, boe.NombreDeLaFuente)
	}

	if len(manifiesto.Normas) == 0 {
		return errors.New("normas no tiene ninguna entrada")
	}

	for indice, entrada := range manifiesto.Normas {
		if err := comprobarEntrada(entrada); err != nil {
			return fmt.Errorf("%s: %w", nombrarEntrada(indice, entrada), err)
		}
	}

	return comprobarNormasRepetidas(manifiesto.Normas)
}

// comprobarEntrada aplica el punto 4 a una entrada: primero los tres campos de
// texto, en el orden del documento, y después cada bloque.
func comprobarEntrada(entrada EntradaDelManifiesto) error {
	campos := []struct{ nombre, valor string }{
		{nombre: "busqueda", valor: entrada.Busqueda},
		{nombre: "titulo_empieza_por", valor: entrada.TituloEmpiezaPor},
		{nombre: "para", valor: entrada.Para},
	}

	for _, campo := range campos {
		if campo.valor == "" {
			return fmt.Errorf("el campo %s está vacío", campo.nombre)
		}
	}

	for posicion, bloque := range entrada.Bloques {
		// El error de boe nombra el bloque y la forma que se espera de él, sin
		// que este paquete copie su gramática.
		if err := boe.ValidarBloque(bloque); err != nil {
			return err
		}

		if slices.Contains(entrada.Bloques[:posicion], bloque) {
			return fmt.Errorf("el bloque %q está repetido", bloque)
		}
	}

	return nil
}

// comprobarNormasRepetidas aplica el punto 5 a cada pareja de entradas, en el
// orden del documento, y nombra las dos de la primera pareja repetida.
func comprobarNormasRepetidas(normas []EntradaDelManifiesto) error {
	for primera, anterior := range normas {
		for desplazamiento, posterior := range normas[primera+1:] {
			var motivo string

			switch {
			case anterior.TituloEmpiezaPor == posterior.TituloEmpiezaPor:
				motivo = "con el mismo prefijo"
			case strings.HasPrefix(anterior.TituloEmpiezaPor, posterior.TituloEmpiezaPor),
				strings.HasPrefix(posterior.TituloEmpiezaPor, anterior.TituloEmpiezaPor):
				motivo = "porque el prefijo de una empieza por el de la otra"
			default:
				continue
			}

			return fmt.Errorf("%s y %s: la misma norma repetida, %s",
				nombrarEntrada(primera, anterior), nombrarEntrada(primera+1+desplazamiento, posterior), motivo)
		}
	}

	return nil
}

// nombrarEntrada nombra la entrada de índice indice en normas por su posición,
// desde 1, y su prefijo, que puede estar vacío.
func nombrarEntrada(indice int, entrada EntradaDelManifiesto) string {
	return fmt.Sprintf("entrada %d (%q)", indice+1, entrada.TituloEmpiezaPor)
}
