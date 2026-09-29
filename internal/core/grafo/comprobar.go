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
//   - fuente-caducada: cada Norma, cada Bloque y la redacción vista de cada
//     bloque de la instantánea —la BloqueVersion que vio su última lectura o,
//     si no tiene fila, la de RedaccionVistaSinLecturas— cuya última
//     observación declaró una vigencia y cuya fecha de consulta más esa
//     vigencia es estrictamente anterior al instante de la comprobación (FR-066;
//     H7.1 FR-030). Nunca otra BloqueVersion, que ninguna lectura vio la última
//     vez, ni otro tipo de nodo.
//
// Los hallazgos van con todas las version-obsoleta antes que todas las
// fuente-caducada y, dentro de cada clase, por id comparando bytes (H7.1
// FR-011), y no dependen del orden en el que la instantánea trae los nodos, las
// aristas y las filas: la misma instantánea en el mismo instante da siempre la
// misma lista (FR-062). Se calculan todos y se listan los MaximoDeHallazgos
// primeros de ese orden; los totales de cada clase los cuentan todos (H7.1
// FR-010, FR-012).

// formatoDeFechaDeVigencia es el de la fecha de vigencia de una BloqueVersion,
// AAAAMMDD, como la publica boe.
const formatoDeFechaDeVigencia = "20060102"

// Comprobar aplica las dos reglas de `graph check` a todo lo que trae la
// instantánea en el instante ahora, el del reloj de la invocación (FR-067), y
// devuelve la comprobación del ámbito: su norma y sus bloques, copiados tal
// como se pidieron, el total de hallazgos de cada clase, cuántos se omiten y
// los MaximoDeHallazgos primeros, cada uno con su explicación (FR-061), en una
// lista que nunca es nula (FR-060; H7.1 FR-010 a FR-012). Qué entra en el
// ámbito no lo decide Comprobar sino la lectura acotada que da la instantánea
// (H7.1 data-model §6).
//
// Una instantánea con una fecha de consulta que no es RFC 3339, que ninguna
// entrega guarda, da un error, sin ningún hallazgo: la regla genérica (H7.1
// research.md D21). El error no nombra el id del nodo, que puede ser de una
// Persona.
func Comprobar(instantanea Instantanea, ambito Ambito, ahora time.Time) (Comprobacion, error) {
	grafoLeido, err := indexar(instantanea)
	if err != nil {
		return Comprobacion{}, err
	}

	obsoletas := grafoLeido.versionesObsoletas(instantanea.Lecturas)
	caducadas := grafoLeido.fuentesCaducadas(ahora)

	hallazgos := slices.Concat(obsoletas, caducadas)
	slices.SortStableFunc(hallazgos, func(a, b Hallazgo) int {
		return cmp.Or(cmp.Compare(rangoDeClase(a.Clase), rangoDeClase(b.Clase)), strings.Compare(a.ID, b.ID))
	})

	listados := slices.Clip(hallazgos[:min(len(hallazgos), MaximoDeHallazgos)])

	return Comprobacion{
		Norma:           ambito.Norma,
		Bloques:         slices.Clone(ambito.Bloques),
		VersionObsoleta: len(obsoletas),
		FuenteCaducada:  len(caducadas),
		Omitidos:        len(hallazgos) - len(listados),
		Hallazgos:       listaNoNula(listados),
	}, nil
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
// el instante de su última observación; los ids en orden de bytes; por cada
// relación, los extremos distintos de las aristas que llegan a un nodo y de
// las que salen de él, también en orden de bytes; y los ids de las redacciones
// vistas de los bloques.
type indice struct {
	nodos     map[string]nodoLeido
	ids       []string
	entrantes map[extremo][]string
	salientes map[extremo][]string
	vistas    map[string]bool
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
// fecha de consulta que no es RFC 3339.
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

		leido.nodos[nodo.ID] = nodoLeido{NodoDeInstantanea: nodo, instante: observado}
	}

	for _, arista := range instantanea.Aristas {
		llegada := extremo{relacion: arista.Relacion, id: arista.Destino}
		leido.entrantes[llegada] = append(leido.entrantes[llegada], arista.Origen)

		salida := extremo{relacion: arista.Relacion, id: arista.Origen}
		leido.salientes[salida] = append(leido.salientes[salida], arista.Destino)
	}

	for _, extremos := range []map[extremo][]string{leido.entrantes, leido.salientes} {
		for clave, ids := range extremos {
			slices.Sort(ids)
			extremos[clave] = slices.Compact(ids)
		}
	}

	leido.ids = slices.Sorted(maps.Keys(leido.nodos))

	vistas, err := leido.redaccionesVistas(instantanea.Lecturas)
	if err != nil {
		return indice{}, err
	}

	leido.vistas = vistas

	return leido, nil
}

// redaccionesVistas son los ids de la redacción vista de cada Bloque de la
// instantánea (H7.1 data-model §3): la que vio su última lectura, si tiene fila
// de lecturas, o, si no, la de RedaccionVistaSinLecturas entre las
// BloqueVersion a las que llegan sus aristas eli:has_version, si llega a
// alguna, como la calcula la entrega (FR-026).
func (i indice) redaccionesVistas(lecturas []LecturasDeBloque) (map[string]bool, error) {
	ultimas := make(map[string]string, len(lecturas))
	for _, fila := range lecturas {
		ultimas[fila.Bloque] = fila.Ultima
	}

	vistas := map[string]bool{}

	for _, id := range i.ids {
		if i.nodos[id].Tipo != TipoBloque {
			continue
		}

		if ultima, conFila := ultimas[id]; conFila {
			vistas[ultima] = true

			continue
		}

		vista, hay, err := RedaccionVistaSinLecturas(i.versionesDe(id))
		if err != nil {
			return nil, err
		}

		if hay {
			vistas[vista] = true
		}
	}

	return vistas, nil
}

// versionesDe son las BloqueVersion de la instantánea a las que llega una
// arista eli:has_version desde el bloque de ese id.
func (i indice) versionesDe(bloque string) []NodoDeInstantanea {
	var versiones []NodoDeInstantanea

	for _, id := range i.salientes[extremo{relacion: RelacionTieneVersion, id: bloque}] {
		if nodo := i.nodos[id]; nodo.Tipo == TipoBloqueVersion {
			versiones = append(versiones, nodo.NodoDeInstantanea)
		}
	}

	return versiones
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

// fuentesCaducadas da un hallazgo fuente-caducada por cada nodo vigente cuya
// última observación declaró una vigencia y caducó estrictamente antes de
// ahora: uno en el límite exacto o con una fecha de consulta posterior a ahora
// no ha caducado, y uno sin vigencia no caduca (FR-066; H7.1 FR-030).
func (i indice) fuentesCaducadas(ahora time.Time) []Hallazgo {
	var hallazgos []Hallazgo

	for _, id := range i.ids {
		nodo := i.nodos[id]
		if nodo.Vigencia == 0 || !i.vigente(nodo) {
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

// vigente dice si el nodo es de los que pueden dar fuente-caducada: una Norma,
// un Bloque o la BloqueVersion que es la redacción vista de un bloque (H7.1
// FR-030).
func (i indice) vigente(nodo nodoLeido) bool {
	switch nodo.Tipo {
	case TipoNorma, TipoBloque:
		return true
	case TipoBloqueVersion:
		return i.vistas[nodo.ID]
	default:
		return false
	}
}
