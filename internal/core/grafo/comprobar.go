package grafo

import (
	"cmp"
	"maps"
	"slices"
	"strings"
	"time"
)

// Las reglas de `graph check` sobre una instantánea del grafo del mundo
// (data-model §6; research.md D15; H7.1 data-model §5 y research.md D5). Son
// dos:
//
//   - version-obsoleta: la deciden las lecturas (H7.1 FR-023, FR-024). Cada
//     bloque con fila de lecturas (U, A) —U, la redacción que vio su última
//     lectura; A, la que vio la anterior— da un hallazgo sobre A si las fechas
//     de vigencia de U y de A son un texto válido para time.Parse con
//     «20060102» —ocho cifras ASCII que nombran un día que existe (FR-064)— y la
//     de U es estrictamente posterior, con la fecha de vigencia y la procedencia
//     de la última observación de U. Un bloque sin fila cuenta con una sola
//     lectura, (R, R), que no da nada (H7.1 FR-026).
//   - fuente-caducada: cada nodo, del tipo que sea, cuya última observación
//     declaró una vigencia y cuya fecha de consulta más esa vigencia es
//     estrictamente anterior al instante de la comprobación, aunque también
//     tenga una version-obsoleta (FR-066).
//
// Los hallazgos van con todas las version-obsoleta antes que todas las
// fuente-caducada y, dentro de cada clase, por id comparando bytes (H7.1
// FR-011), y no dependen del orden en el que la instantánea trae los nodos, las
// aristas y las filas: la misma instantánea en el mismo instante da siempre la
// misma lista (FR-062).

// formatoDeFechaDeVigencia es el de la fecha de vigencia de una BloqueVersion,
// AAAAMMDD, como la publica boe.
const formatoDeFechaDeVigencia = "20060102"

// Comprobar aplica las dos reglas de `graph check` a la instantánea en el
// instante ahora, el del reloj de la invocación (FR-067), y devuelve sus
// hallazgos, cada uno con su explicación (FR-061), en una lista que nunca es
// nula (FR-060).
//
// Una instantánea con una fecha de consulta que no es RFC 3339, que ninguna
// entrega guarda, da un error, sin ningún hallazgo: la regla genérica (H7.1
// research.md D21). El error no nombra el id del nodo, que puede ser de una
// Persona.
func Comprobar(instantanea Instantanea, ahora time.Time) ([]Hallazgo, error) {
	grafoLeido, err := indexar(instantanea)
	if err != nil {
		return nil, err
	}

	hallazgos := slices.Concat(grafoLeido.versionesObsoletas(instantanea.Lecturas), grafoLeido.fuentesCaducadas(ahora))
	slices.SortStableFunc(hallazgos, func(a, b Hallazgo) int {
		return cmp.Or(cmp.Compare(rangoDeClase(a.Clase), rangoDeClase(b.Clase)), strings.Compare(a.ID, b.ID))
	})

	return listaNoNula(hallazgos), nil
}

// rangoDeClase es la posición de la clase en la lista de hallazgos: primero
// version-obsoleta y después fuente-caducada (H7.1 FR-011).
func rangoDeClase(clase ClaseDeHallazgo) int {
	if clase == ClaseVersionObsoleta {
		return 0
	}

	return 1
}

// indice es la instantánea preparada para las reglas: cada nodo por su id, con
// el instante de su última observación; los ids en orden de bytes; y, por cada
// relación, los extremos distintos de las aristas que llegan a un nodo, también
// en orden de bytes.
type indice struct {
	nodos     map[string]nodoLeido
	ids       []string
	entrantes map[extremo][]string
}

// nodoLeido es un nodo de la instantánea con el instante de su última
// observación.
type nodoLeido struct {
	NodoDeInstantanea

	instante time.Time
}

// extremo es un nodo por su id y una relación de las aristas que llegan a él.
type extremo struct {
	relacion string
	id       string
}

// indexar prepara la instantánea para las reglas. Falla si un nodo tiene una
// fecha de consulta que no es RFC 3339.
func indexar(instantanea Instantanea) (indice, error) {
	leido := indice{
		nodos:     make(map[string]nodoLeido, len(instantanea.Nodos)),
		entrantes: map[extremo][]string{},
	}

	for _, nodo := range instantanea.Nodos {
		observado, err := instante(nodo.UltimaObservacion.FechaConsulta)
		if err != nil {
			return indice{}, err
		}

		leido.nodos[nodo.ID] = nodoLeido{NodoDeInstantanea: nodo, instante: observado}
	}

	for _, arista := range instantanea.Aristas {
		llegada := extremo{relacion: arista.Relacion, id: arista.Destino}
		leido.entrantes[llegada] = append(leido.entrantes[llegada], arista.Origen)
	}

	for clave, ids := range leido.entrantes {
		slices.Sort(ids)
		leido.entrantes[clave] = slices.Compact(ids)
	}

	leido.ids = slices.Sorted(maps.Keys(leido.nodos))

	return leido, nil
}

// redaccionFechada es una BloqueVersion de la instantánea con su fecha de
// vigencia válida, como texto y como fecha.
type redaccionFechada struct {
	nodoLeido

	vigente time.Time
	texto   string
}

// versionesObsoletas da un hallazgo version-obsoleta por cada fila de
// lecturas (U, A) en la que U y A tienen una fecha de vigencia válida y la de
// U es estrictamente posterior: sobre A, con la fecha de vigencia y la
// procedencia de la última observación de U (H7.1 FR-023). Una fila con la
// misma redacción en las dos, con la misma fecha o con una anterior en U no da
// nada. Las filas se recorren por bloque, comparando bytes, así que la lista
// no depende de su orden en la instantánea.
func (i indice) versionesObsoletas(lecturas []LecturasDeBloque) []Hallazgo {
	var hallazgos []Hallazgo

	porBloque := slices.SortedFunc(slices.Values(lecturas), func(a, b LecturasDeBloque) int {
		return strings.Compare(a.Bloque, b.Bloque)
	})

	for _, fila := range porBloque {
		ultima, conFechaLaUltima := i.fechada(fila.Ultima)
		anterior, conFechaLaAnterior := i.fechada(fila.Anterior)

		if !conFechaLaUltima || !conFechaLaAnterior || !ultima.vigente.After(anterior.vigente) {
			continue
		}

		hallazgos = append(hallazgos, Hallazgo{
			Clase:                 ClaseVersionObsoleta,
			ID:                    anterior.ID,
			Explicacion:           explicarVersionObsoleta(i.citar(anterior.ID), anterior.texto, ultima.texto, ultima.UltimaObservacion),
			Procedencia:           ultima.UltimaObservacion,
			FechaVigencia:         anterior.texto,
			FechaVigenciaReciente: ultima.texto,
		})
	}

	return hallazgos
}

// fechada es la redacción de ese id con su fecha de vigencia, si la tiene
// válida: un texto que time.Parse lee con «20060102». Una fecha que no es
// texto o que no se lee así —vacía, de otra longitud, con signo o cifras que
// no son ASCII, o de un día que no existe— no es válida, y la redacción ni
// recibe ni provoca un hallazgo (FR-064; research.md D34).
func (i indice) fechada(id string) (redaccionFechada, bool) {
	nodo := i.nodos[id]

	texto, esTexto := nodo.Datos[DatoFechaVigencia].(string)
	if !esTexto {
		return redaccionFechada{}, false
	}

	vigente, err := time.Parse(formatoDeFechaDeVigencia, texto)
	if err != nil {
		return redaccionFechada{}, false
	}

	return redaccionFechada{nodoLeido: nodo, vigente: vigente, texto: texto}, true
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
