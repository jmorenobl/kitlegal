package evals

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"fmt"
	"strings"
)

// MedidaDelJuez es lo que se lee de la medida versionada del juez de una skill,
// medida.json de su carpeta del juez (contracts/medida-del-juez.md §1 y
// data-model §4 de H24; FR-040): contra qué casos etiquetados se midió al juez,
// con qué modelo y qué versión de Claude Code, y qué dio. Corresponde si sus dos
// huellas, su modelo y su versión son los de lo que hay, y se cumple si sus dos
// recuentos son 0 (FR-041): lo dice comprobarLaMedida.
type MedidaDelJuez struct {
	// Clase es la clase que decide a la que se refiere.
	Clase string

	// Fecha es cuándo se midió.
	Fecha string

	// ModeloDelJuez y VersionDeClaudeCode son el id del modelo y la versión de
	// Claude Code de los votos de la medida.
	ModeloDelJuez       string
	VersionDeClaudeCode string

	// HuellaDeLaRubrica y HuellaDeLosCasos son las huellas SHA-256, en
	// hexadecimal, de la rúbrica y de los casos con los que se midió.
	HuellaDeLaRubrica string
	HuellaDeLosCasos  string

	// Defectos y Correctos son los casos etiquetados como defecto y como
	// correctos, con los que no dieron lo que dice su etiqueta.
	Defectos  DefectosDeLaMedida
	Correctos CorrectosDeLaMedida
}

// DefectosDeLaMedida son los casos etiquetados como defecto de una medida del
// juez.
type DefectosDeLaMedida struct {
	// Casos es cuántos son.
	Casos int

	// SinMarcar es cuántos de ellos no quedaron marcados: con alguno, la medida
	// no se cumple.
	SinMarcar int
}

// CorrectosDeLaMedida son los casos etiquetados como correctos de una medida del
// juez.
type CorrectosDeLaMedida struct {
	// Casos es cuántos son.
	Casos int

	// Marcados es cuántos de ellos quedaron marcados: con alguno, la medida no
	// se cumple.
	Marcados int
}

// clavesDeLaMedida son las claves de medida.json que lee leerMedidaDelJuez; las
// demás —skill, votos, origen y el fichero de cada huella— no se leen. Cada una
// es un puntero, que distingue la clave que falta, o que está a null, de la que
// lleva el valor cero: un recuento a 0 es justo el que se cumple.
type clavesDeLaMedida struct {
	Clase               *string `json:"clase"`
	Fecha               *string `json:"fecha"`
	ModeloDelJuez       *string `json:"modelo_del_juez"`
	VersionDeClaudeCode *string `json:"version_de_claude_code"`

	Rubrica struct {
		SHA256 *string `json:"sha256"`
	} `json:"rubrica"`

	Casos struct {
		SHA256 *string `json:"sha256"`
	} `json:"casos"`

	Defectos struct {
		Casos     *int `json:"casos"`
		SinMarcar *int `json:"sin_marcar"`
	} `json:"defectos"`

	Correctos struct {
		Casos    *int `json:"casos"`
		Marcados *int `json:"marcados"`
	} `json:"correctos"`
}

// leerMedidaDelJuez lee la medida del juez de esa ruta
// (contracts/medida-del-juez.md §1 de H24). No tiene esquema publicado: se lee
// como un objeto JSON del que solo se miran las diez claves de MedidaDelJuez,
// cada una con su tipo, un texto o un entero; una clave repetida es un error,
// nunca la última que gana. Es un error que no se pueda leer, que no sea ese
// documento o que le falte alguna de las diez, y una clave a null es una clave
// que falta: el error las nombra todas, en el orden del contrato, con las de un
// objeto detrás de la suya y de un punto, como rubrica.sha256. Ningún error
// lleva delante de qué medida habla: lo pone quien la lee.
func leerMedidaDelJuez(ruta string) (MedidaDelJuez, error) {
	contenido, err := leerFichero(ruta)
	if err != nil {
		return MedidaDelJuez{}, fmt.Errorf("no se puede leer: %w", err)
	}

	var leidas clavesDeLaMedida

	if err := json.Unmarshal(contenido, &leidas); err != nil {
		return MedidaDelJuez{}, fmt.Errorf("no se puede leer como JSON: %w", err)
	}

	var faltan []string

	medida := MedidaDelJuez{
		Clase:               valorDeLaClave(leidas.Clase, "clase", &faltan),
		Fecha:               valorDeLaClave(leidas.Fecha, "fecha", &faltan),
		ModeloDelJuez:       valorDeLaClave(leidas.ModeloDelJuez, "modelo_del_juez", &faltan),
		VersionDeClaudeCode: valorDeLaClave(leidas.VersionDeClaudeCode, "version_de_claude_code", &faltan),
		HuellaDeLaRubrica:   valorDeLaClave(leidas.Rubrica.SHA256, "rubrica.sha256", &faltan),
		HuellaDeLosCasos:    valorDeLaClave(leidas.Casos.SHA256, "casos.sha256", &faltan),
		Defectos: DefectosDeLaMedida{
			Casos:     valorDeLaClave(leidas.Defectos.Casos, "defectos.casos", &faltan),
			SinMarcar: valorDeLaClave(leidas.Defectos.SinMarcar, "defectos.sin_marcar", &faltan),
		},
		Correctos: CorrectosDeLaMedida{
			Casos:    valorDeLaClave(leidas.Correctos.Casos, "correctos.casos", &faltan),
			Marcados: valorDeLaClave(leidas.Correctos.Marcados, "correctos.marcados", &faltan),
		},
	}

	if len(faltan) > 0 {
		return MedidaDelJuez{}, fmt.Errorf("claves que faltan: %s", strings.Join(faltan, ", "))
	}

	return medida, nil
}

// valorDeLaClave es el valor leído de una clave de la medida. Si la clave no
// estaba, o estaba a null, la anota en faltan y devuelve el valor cero.
func valorDeLaClave[T any](leido *T, clave string, faltan *[]string) T {
	if leido == nil {
		*faltan = append(*faltan, clave)

		var cero T

		return cero
	}

	return *leido
}

// Las líneas de comprobarLaMedida: las de contracts/medida-del-juez.md §2 de
// H24, carácter a carácter, y las dos de lo que no se puede leer. Cada fichero
// lleva el nombre que tiene dentro de la carpeta de evals de la skill,
// juez/<fichero>, como en los errores del conjunto.
const (
	// medidaQueNoCorresponde es la de la medida que no se puede leer o a la que
	// le falta alguna de sus claves; sigue con el error de leerMedidaDelJuez.
	medidaQueNoCorresponde = "la medida versionada (" + carpetaDelJuez + "/" + ficheroDeMedidaDelJuez + ") no corresponde: %v"

	rubricaDeOtraMedida = "la rúbrica (" + carpetaDelJuez + "/" + ficheroDeRubricaDelJuez + ") no es la de la medida versionada"
	casosDeOtraMedida   = "los casos (" + carpetaDelJuez + "/" + ficheroDeCasosDelJuez + ") no son los de la medida versionada"

	// casosQueNoSeLeen es la de los casos que han dejado de poder leerse desde
	// que se leyó el juez; sigue con el error.
	casosQueNoSeLeen = "los casos (" + carpetaDelJuez + "/" + ficheroDeCasosDelJuez + ") no se pueden leer: %v"

	medidaDeOtroModelo  = "la medida versionada es del modelo %s y el fijado para el juez es %s"
	medidaDeOtraVersion = "la medida versionada es de la versión %s de Claude Code y la fijada para los votos del juez es %s"
	medidaDeOtraClase   = "la clase %s decide y la medida versionada es de %s"

	medidaConDefectosSinMarcar = "la medida versionada no se cumple: %d defectos sin marcar"
	medidaConCorrectosMarcados = "la medida versionada no se cumple: %d correctos marcados"
)

// comprobarLaMedida dice si la medida versionada del juez de una skill
// corresponde y se cumple (contracts/medida-del-juez.md §2 de H24; FR-041):
// devuelve una línea por lo que falla, en este orden, y ninguna si no falla
// nada.
//
//   - La huella SHA-256 de la rúbrica del juez no es la de la medida.
//   - La de sus casos no lo es.
//   - El modelo de la medida no es el fijado para el juez.
//   - Su versión de Claude Code no es la fijada para los votos del juez.
//   - Una clase que decide no es la de la medida: una línea por cada una, que
//     una medida es de una sola clase (research D19 de H24).
//   - Su recuento de defectos sin marcar no es 0.
//   - Su recuento de correctos marcados no es 0.
//
// Las cuatro primeras son las de una medida que no corresponde, y las dos
// últimas, las de una que no se cumple. La medida que no se puede leer, o sin
// alguna de las claves que se leen, tampoco corresponde, y da una sola línea,
// con el error de leerMedidaDelJuez: sin la medida entera no hay con qué
// comparar lo demás.
//
// El modelo y la versión fijados los da quien la llama, y el juez es el de la
// skill, nunca nil. Lee los casos y la medida de sus rutas, y nada más: no usa
// ningún modelo ni abre ningún proceso.
func comprobarLaMedida(juez *Juez, modeloFijado, versionFijada string) []string {
	medida, err := leerMedidaDelJuez(juez.Medida)
	if err != nil {
		return []string{fmt.Sprintf(medidaQueNoCorresponde, err)}
	}

	lineas := lineasDeLasHuellas(juez, medida)
	lineas = append(lineas, lineasDeLoFijado(medida, modeloFijado, versionFijada)...)
	lineas = append(lineas, lineasDeLasClases(juez.Clases, medida.Clase)...)

	return append(lineas, lineasDeLosRecuentos(medida)...)
}

// lineasDeLasHuellas compara las huellas SHA-256 de la rúbrica del juez, que es
// la que va tal cual como sus instrucciones, y de sus casos, leídos de su ruta,
// con las de la medida. Unos casos que no se pueden leer no tienen huella que
// comparar: dan su línea, con el error.
func lineasDeLasHuellas(juez *Juez, medida MedidaDelJuez) []string {
	var lineas []string

	if huellaSHA256([]byte(juez.Rubrica)) != medida.HuellaDeLaRubrica {
		lineas = append(lineas, rubricaDeOtraMedida)
	}

	casos, err := leerFichero(juez.Casos)

	switch {
	case err != nil:
		lineas = append(lineas, fmt.Sprintf(casosQueNoSeLeen, err))
	case huellaSHA256(casos) != medida.HuellaDeLosCasos:
		lineas = append(lineas, casosDeOtraMedida)
	}

	return lineas
}

// lineasDeLoFijado compara el modelo y la versión de Claude Code de la medida
// con los fijados para el juez y para sus votos.
func lineasDeLoFijado(medida MedidaDelJuez, modeloFijado, versionFijada string) []string {
	var lineas []string

	if medida.ModeloDelJuez != modeloFijado {
		lineas = append(lineas, fmt.Sprintf(medidaDeOtroModelo, medida.ModeloDelJuez, modeloFijado))
	}

	if medida.VersionDeClaudeCode != versionFijada {
		lineas = append(lineas, fmt.Sprintf(medidaDeOtraVersion, medida.VersionDeClaudeCode, versionFijada))
	}

	return lineas
}

// lineasDeLasClases compara la clase de la medida con cada clase del juez que
// decide, en su orden: las que solo se publican no necesitan medida.
func lineasDeLasClases(clases []ClaseDelJuez, deLaMedida string) []string {
	var lineas []string

	for _, clase := range clases {
		if clase.Decide && clase.Nombre != deLaMedida {
			lineas = append(lineas, fmt.Sprintf(medidaDeOtraClase, clase.Nombre, deLaMedida))
		}
	}

	return lineas
}

// lineasDeLosRecuentos compara con 0 los dos recuentos de la medida: los
// defectos que no quedaron marcados y los correctos que sí.
func lineasDeLosRecuentos(medida MedidaDelJuez) []string {
	var lineas []string

	if medida.Defectos.SinMarcar != 0 {
		lineas = append(lineas, fmt.Sprintf(medidaConDefectosSinMarcar, medida.Defectos.SinMarcar))
	}

	if medida.Correctos.Marcados != 0 {
		lineas = append(lineas, fmt.Sprintf(medidaConCorrectosMarcados, medida.Correctos.Marcados))
	}

	return lineas
}

// huellaSHA256 es la huella SHA-256 de un contenido, en hexadecimal y en
// minúsculas, como la lleva una medida.
func huellaSHA256(contenido []byte) string {
	suma := sha256.Sum256(contenido)

	return hex.EncodeToString(suma[:])
}
