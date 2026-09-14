package skills

// destinoDeLosEnlaces es el destino literal de todo enlace de scripts/ de una
// skill (data-model §3; research.md D5): desde skills/<skill>/scripts/, el
// enlace bin/instalado/kitlegal de la raíz del repositorio, que make install
// apunta al binario instalado. Es relativo para que resuelva igual en cualquier
// clon, y no resuelve hasta que se instala.
const destinoDeLosEnlaces = "../../../bin/instalado/kitlegal"

// Enlace es un enlace simbólico de scripts/ de una skill (data-model §3).
type Enlace struct {
	// Nombre es el del applet y el del enlace dentro de scripts/: invocado por
	// él, el binario multicall ejecuta ese applet.
	Nombre string

	// Destino es el destino literal del enlace.
	Destino string
}

// EnlacesEsperados son los enlaces de scripts/ de una skill con la declaración
// de kitlegal dada (data-model §3; research.md D4 y D5; FR-035, FR-036): uno por
// cada applet de kitlegal-applets, en el orden de la declaración, con el nombre
// del applet y el destino literal ../../../bin/instalado/kitlegal. Salen de la
// declaración y no de lo que haya en scripts/, de modo que un enlace borrado se
// sigue esperando. Sin applets declarados no hay ninguno: nil.
func EnlacesEsperados(deKitlegal DeclaracionDeKitlegal) []Enlace {
	var enlaces []Enlace

	for _, applet := range deKitlegal.Applets {
		enlaces = append(enlaces, Enlace{Nombre: applet, Destino: destinoDeLosEnlaces})
	}

	return enlaces
}
