package grafo

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

// Las reglas de `graph check` sobre una instantánea del grafo del mundo
// (data-model §6; research.md D15). Son dos:
//
//   - version-obsoleta: las versiones de un Bloque son las BloqueVersion a las
//     que llega desde él una arista eli:has_version y cuya fecha de vigencia es
//     un texto válido para time.Parse con «20060102» —ocho cifras ASCII que
//     nombran un día que existe; las demás ni reciben ni provocan un hallazgo—.
//     Cada versión con otra de su Bloque de fecha de vigencia estrictamente
//     posterior da un hallazgo, una sola vez aunque cuelgue de varios Bloque,
//     con la procedencia y la fecha de vigencia de la más reciente (FR-063,
//     FR-064).
//   - fuente-caducada: cada nodo, del tipo que sea, cuya última observación
//     declaró una vigencia y cuya fecha de consulta más esa vigencia es
//     estrictamente anterior al instante de la comprobación, aunque también
//     tenga una version-obsoleta (FR-066).
//
// Los hallazgos van ordenados por clase y después por id, comparando bytes, y
// no dependen del orden en el que la instantánea trae los nodos y las aristas:
// la misma instantánea en el mismo instante da siempre la misma lista (FR-062).

// formatoDeFechaDeVigencia es el de la fecha de vigencia de una BloqueVersion,
// AAAAMMDD, como la publica boe.
const formatoDeFechaDeVigencia = "20060102"

// Comprobar aplica las dos reglas de `graph check` a la instantánea en el
// instante ahora, el del reloj de la invocación (FR-067), y devuelve sus
// hallazgos, cada uno con su explicación (FR-061), en una lista que nunca es
// nula (FR-060).
//
// Una instantánea con lo que ninguna entrega guarda —una fecha de consulta que
// no es RFC 3339, una vigencia negativa o con fracción de segundo— es un
// defecto de lo guardado y da un error, sin ningún hallazgo; el error no nombra
// el id del nodo, que puede ser de una Persona.
func Comprobar(instantanea Instantanea, ahora time.Time) ([]Hallazgo, error) {
	grafoLeido, err := indexar(instantanea)
	if err != nil {
		return nil, err
	}

	hallazgos := slices.Concat(grafoLeido.versionesObsoletas(), grafoLeido.fuentesCaducadas(ahora))
	slices.SortFunc(hallazgos, func(a, b Hallazgo) int {
		return cmp.Or(strings.Compare(string(a.Clase), string(b.Clase)), strings.Compare(a.ID, b.ID))
	})

	return listaNoNula(hallazgos), nil
}

// indice es la instantánea preparada para las reglas: cada nodo por su id, con
// el instante de su última observación; los ids en orden de bytes; y, por cada
// relación, los extremos distintos de las aristas que llegan a un nodo y de las
// que salen de él, también en orden de bytes.
type indice struct {
	nodos     map[string]nodoLeido
	ids       []string
	entrantes map[extremo][]string
	salientes map[extremo][]string
}

// nodoLeido es un nodo de la instantánea con el instante de su última
// observación.
type nodoLeido struct {
	NodoDeInstantanea

	instante time.Time
}

// extremo es un nodo por su id y una relación de las aristas que llegan a él o
// que salen de él.
type extremo struct {
	relacion string
	id       string
}

// indexar prepara la instantánea para las reglas. Falla si un nodo tiene una
// fecha de consulta que no es RFC 3339 o una vigencia que no es un número
// entero y positivo de segundos, que es como se guarda (FR-065).
func indexar(instantanea Instantanea) (indice, error) {
	leido := indice{
		nodos:     make(map[string]nodoLeido, len(instantanea.Nodos)),
		entrantes: map[extremo][]string{},
		salientes: map[extremo][]string{},
	}

	for _, nodo := range instantanea.Nodos {
		observado, err := instante(nodo.UltimaObservacion.FechaConsulta)
		if err != nil {
			return indice{}, err
		}

		if nodo.Vigencia < 0 || nodo.Vigencia%time.Second != 0 {
			return indice{}, fmt.Errorf("la vigencia guardada %s no es un número entero de segundos positivo", nodo.Vigencia)
		}

		leido.nodos[nodo.ID] = nodoLeido{NodoDeInstantanea: nodo, instante: observado}
	}

	for _, arista := range instantanea.Aristas {
		llegada := extremo{relacion: arista.Relacion, id: arista.Destino}
		salida := extremo{relacion: arista.Relacion, id: arista.Origen}

		leido.entrantes[llegada] = append(leido.entrantes[llegada], arista.Origen)
		leido.salientes[salida] = append(leido.salientes[salida], arista.Destino)
	}

	for _, extremos := range []map[extremo][]string{leido.entrantes, leido.salientes} {
		for clave, ids := range extremos {
			slices.Sort(ids)
			extremos[clave] = slices.Compact(ids)
		}
	}

	leido.ids = slices.Sorted(maps.Keys(leido.nodos))

	return leido, nil
}

// versionFechada es una versión de un Bloque con su fecha de vigencia válida,
// como texto y como fecha.
type versionFechada struct {
	nodoLeido

	vigente time.Time
	texto   string
}

// versionesObsoletas da un hallazgo version-obsoleta por cada versión con otra
// de su Bloque de fecha de vigencia estrictamente posterior. Los Bloque se
// recorren por id, así que una versión que cuelga de varios lo recibe del de id
// menor en el que está superada.
func (i indice) versionesObsoletas() []Hallazgo {
	var hallazgos []Hallazgo

	superadas := map[string]bool{}

	for _, id := range i.ids {
		if i.nodos[id].Tipo != TipoBloque {
			continue
		}

		versiones := i.versionesDe(id)
		if len(versiones) == 0 {
			continue
		}

		reciente := slices.MaxFunc(versiones, compararRecencia)

		for _, version := range versiones {
			if !version.vigente.Before(reciente.vigente) || superadas[version.ID] {
				continue
			}

			superadas[version.ID] = true
			hallazgos = append(hallazgos, Hallazgo{
				Clase:                 ClaseVersionObsoleta,
				ID:                    version.ID,
				Explicacion:           explicarVersionObsoleta(i.citar(version.ID), version.texto, reciente.texto, reciente.UltimaObservacion),
				Procedencia:           reciente.UltimaObservacion,
				FechaVigencia:         version.texto,
				FechaVigenciaReciente: reciente.texto,
			})
		}
	}

	return hallazgos
}

// versionesDe son las del Bloque: las BloqueVersion a las que llega desde él
// una arista eli:has_version con una fecha de vigencia válida. Una fecha que
// no es texto o que time.Parse no lee con «20060102» —vacía, de otra longitud,
// con signo o cifras que no son ASCII, o de un día que no existe— deja la
// versión fuera: ni se ordena ni recibe ni provoca un hallazgo (FR-064;
// research.md D34).
func (i indice) versionesDe(bloque string) []versionFechada {
	var versiones []versionFechada

	for _, id := range i.salientes[extremo{relacion: RelacionTieneVersion, id: bloque}] {
		nodo, existe := i.nodos[id]
		if !existe || nodo.Tipo != TipoBloqueVersion {
			continue
		}

		texto, esTexto := nodo.Datos[DatoFechaVigencia].(string)
		if !esTexto {
			continue
		}

		vigente, err := time.Parse(formatoDeFechaDeVigencia, texto)
		if err != nil {
			continue
		}

		versiones = append(versiones, versionFechada{nodoLeido: nodo, vigente: vigente, texto: texto})
	}

	return versiones
}

// compararRecencia es positivo si a es más reciente que b: de fecha de
// vigencia posterior; a igualdad, de última observación de mayor instante; y a
// igualdad, de id menor, comparando bytes (FR-063).
func compararRecencia(a, b versionFechada) int {
	return cmp.Or(a.vigente.Compare(b.vigente), a.instante.Compare(b.instante), strings.Compare(b.ID, a.ID))
}

// fuentesCaducadas da un hallazgo fuente-caducada por cada nodo cuya última
// observación declaró una vigencia y caducó estrictamente antes de ahora: uno
// en el límite exacto o con una fecha de consulta posterior a ahora no ha
// caducado, y uno sin vigencia no caduca (FR-066).
func (i indice) fuentesCaducadas(ahora time.Time) []Hallazgo {
	var hallazgos []Hallazgo

	for _, id := range i.ids {
		nodo := i.nodos[id]
		if nodo.Vigencia == 0 {
			continue
		}

		caducidad := instanteDeCaducidad(nodo.instante, nodo.Vigencia)
		if !caducidad.Before(ahora) {
			continue
		}

		hallazgos = append(hallazgos, Hallazgo{
			Clase:            ClaseFuenteCaducada,
			ID:               id,
			Explicacion:      explicarFuenteCaducada(i.citar(id), nodo.UltimaObservacion, nodo.Vigencia, caducidad),
			Procedencia:      nodo.UltimaObservacion,
			VigenciaSegundos: int64(nodo.Vigencia / time.Second),
		})
	}

	return hallazgos
}
