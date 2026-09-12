// Package ejemplo son los dos applets con los que se demuestra el kernel, y la
// implementación de referencia que copiará cada applet posterior: `echo`, que
// declara un verbo por omisión y sostiene la entrega literal del hito, y
// `contar`, que declara dos verbos y ninguno por omisión y sostiene la rama
// contraria —nombrar el verbo es obligatorio— (FR-009,
// contracts/registro-y-describe.md §3).
//
// Los dos son **ejemplos del patrón, no funcionalidad del binario que se
// publica**: no acceden a la red, no tocan disco y su `data` se deriva de sus
// argumentos. Son un paquete normal del módulo —los comodines de Go lo
// alcanzan, así que pasa por vet, lint, formato y cobertura sin enumerarlo en
// ningún sitio— y lo que impide que el binario distribuido lo enlace, ni por
// descuido, son dos comprobaciones: depguard prohíbe importarlo fuera de este árbol
// y de los tests, e internal/arch_test.go comprueba que no aparece en el cierre
// transitivo de cmd/kitlegal (docs/ADR/0010-applets-de-ejemplo-fuera-de-testdata.md).
//
// Ninguno de los dos declara una sola bandera, un solo código de salida ni una
// sola forma de presentación, y los heredan todos del kernel: eso es lo que
// SC-010 pide demostrar, y por eso este paquete no lleva ninguna excepción de
// lint —ni os.Exit, ni os.Stdout, ni os.Stderr, ni fmt.Print*—. El listón del
// código que se va a copiar no baja.
package ejemplo

import "github.com/jmorenobl/kitlegal/internal/app"

// Applets son los dos applets de ejemplo, en el orden en que se registran. El
// orden no decide nada —el registro ordena los nombres por su cuenta— y existe
// para que quien los recorra, como hace el test que comprueba lo que un applet
// hereda, los vea siempre en la misma sucesión.
func Applets() []app.Applet {
	return []app.Applet{Echo(), Contar()}
}

// Registro es el registro del binario de e2e, construido con **exactamente** el
// mismo mecanismo que el del binario que se publica: lo único que cambia entre
// los dos es qué applets se registran (FR-001,
// contracts/registro-y-describe.md §3).
//
// Devuelve error en lugar de resolverlo por su cuenta porque un registro
// inválido es un defecto de quien escribió el applet, y quien tiene que
// convertirlo en un fallo de arranque es la raíz de composición, no este
// paquete: aquí no se escribe en ningún descriptor ni se termina ningún proceso
// (FR-008).
func Registro() (*app.Registro, error) {
	registro := &app.Registro{}

	for _, applet := range Applets() {
		if err := registro.Registrar(applet); err != nil {
			return nil, err
		}
	}

	return registro, nil
}
