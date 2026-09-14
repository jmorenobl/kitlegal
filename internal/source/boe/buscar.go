package boe

import (
	"context"
	"strings"

	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// Lo que solo usa buscar: el separador de sus argumentos y la clave del
// identificador de cada resultado. El título, el rango y su texto,
// vigencia_agotada y estado_consolidacion se leen con las claves de metadatos.go
// y avisos.go (refs/boe.py 288-298; data-model.md §2.3).
const (
	// separadorDelTexto es el espacio con el que buscar une sus argumentos antes
	// de construir la consulta (refs/boe.py 784).
	separadorDelTexto = " "
	// claveDelIdentificador es la del identificador BOE-A-… de cada resultado
	// (refs/boe.py 293 y 298).
	claveDelIdentificador = "identificador"
)

// buscar resuelve buscar <texto>… (contrato verbos-y-salidas §1): une los
// argumentos con un espacio y construye con direccionDeBusqueda la dirección de
// la búsqueda antes de abrir nada —un texto sin operadores y sin ninguna palabra
// es «argumentos» sin procedencia, que firma y fecha el kernel (contrato
// errores-y-codigos, fila 4)— y, dentro de invocar, resuelve la búsqueda con
// consultar sobre su entrada y con la vigencia de buscar, cinco minutos (FR-091):
// servida de la entrada vigente, o pedida en JSON, leída con leerBusqueda y
// escrita, también cuando no hay resultados (FR-032). Ni su 404 ni su data vacío
// dicen que algo no exista: el primero es «fuente no disponible» (fila 9) y el
// segundo, la lista vacía (J3). El resultado lleva la dirección de la búsqueda en
// éxito, en ensayo y en fallo (FR-002, FR-101), y en éxito, la fecha de la
// consulta que lo sostiene (FR-096).
func (f *Fuente) buscar(ctx context.Context, ec schema.Contexto, consulta ConsultaBuscar) (schema.Resultado, error) {
	direccion, err := direccionDeBusqueda(strings.Join(consulta.Texto, separadorDelTexto))
	if err != nil {
		return schema.Resultado{}, err
	}

	return f.invocar(ctx, ec, direccion, func(ctx context.Context, en *invocacion) (schema.Resultado, error) {
		resuelta, err := consultar(ctx, en, claveDeLaBusqueda(direccion), pedidoDeLaBusqueda(direccion), vigenciaCorta,
			leerBusqueda)
		if err != nil {
			return resultadoDelError(err), err
		}

		return resultadoDeLaConsulta(direccion, resuelta), nil
	})
}

// leerBusqueda compone el data de buscar con el data de su respuesta
// (data-model.md §2.3; refs/boe.py 282-298): data vacío es la lista vacía, y no
// nula, que se guarda como cualquier otra (J3, FR-032); y si no, los resultados de
// resultadosDeBusqueda en el orden de la fuente —en lista o uno suelto (J4,
// FR-070), sin los que no son objeto (J8)—. Cada resultado lleva su
// identificador, su título y su vigencia_agotada leídos con las reglas de J9, sin
// el marcador ? de refs/boe.py 289-297 (FR-016); el texto de rango y el de
// estado_consolidacion si son objetos, y la cadena vacía si no; y la dirección
// pública de la norma con el identificador tal como llega, vacío incluido
// (refs/boe.py 298). Lo que no se puede leer es el fallo de la lectura, que nombra
// el primer campo que no es texto.
func leerBusqueda(datos any) ([]ResultadoDeBusqueda, error) {
	resultados := []ResultadoDeBusqueda{}
	if esVacio(datos) {
		return resultados, nil
	}

	for _, resultado := range resultadosDeBusqueda(datos) {
		campos := lectorDeCampos{objeto: resultado}
		identificador := campos.texto(claveDelIdentificador)
		leido := ResultadoDeBusqueda{
			Identificador:       identificador,
			Titulo:              campos.texto(claveDelTitulo),
			Rango:               campos.textoDelObjeto(claveDelRango, subclaveDelTexto),
			VigenciaAgotada:     campos.texto(claveDeVigenciaAgotada),
			EstadoConsolidacion: campos.textoDelObjeto(claveDeEstadoDeConsolidacion, subclaveDelTexto),
			URL:                 direccionPublicaDeLaNorma(identificador),
		}

		if campos.err != nil {
			return nil, campos.err
		}

		resultados = append(resultados, leido)
	}

	return resultados, nil
}
