package grafo

import (
	"cmp"
	"strings"
	"time"
)

// Una lectura de un bloque es la llegada al grafo de un bloque que devuelve
// `boe articulo` o `boe articulos`: la entrega de su resultado, que el grafo
// confirma (H7.1 FR-020; data-model §2). El grafo no guarda la historia de las
// lecturas, sino, por bloque, las dos únicas que deciden algo —qué redacción
// vio la última y cuál la anterior—, y las lecturas de un bloque se ordenan
// por el orden en que se aplican sus entregas (FR-021; research.md D1).

// Lectura es la de un bloque en una entrega: el Bloque y la BloqueVersion que
// vio (data-model §2).
type Lectura struct {
	// Bloque es el id del Bloque leído, el origen de la arista eli:has_version.
	Bloque string
	// Version es el id de la BloqueVersion que vio, el destino de esa arista.
	Version string
}

// Lecturas son las del lote consolidado: una por arista eli:has_version, del
// Bloque que la origina a la BloqueVersion a la que llega, en el orden de sus
// claves (FR-020; research.md D2). Solo boe articulo y boe articulos emiten esa
// arista, una por bloque distinto de la invocación, así que la lectura sale de
// lo que ya emiten, sin que ningún emisor diga nada más. Un lote sin ninguna no
// tiene lecturas.
func (c Consolidado) Lecturas() []Lectura {
	var lecturas []Lectura

	for _, arista := range c.Aristas {
		if arista.Relacion == RelacionTieneVersion {
			lecturas = append(lecturas, Lectura{Bloque: arista.Origen, Version: arista.Destino})
		}
	}

	return lecturas
}

// LecturasDeBloque es la fila de lecturas de un bloque: qué BloqueVersion vio
// su última lectura y cuál la anterior (data-model §1 y §3). En la primera
// lectura de un bloque, las dos son la misma: no hay cambio que decir.
type LecturasDeBloque struct {
	// Bloque es el id del Bloque.
	Bloque string
	// Ultima es el id de la BloqueVersion que vio su última lectura.
	Ultima string
	// Anterior es el id de la que vio la lectura anterior; en la primera, la
	// misma que Ultima.
	Anterior string
}

// Leida es la fila tras una lectura que ve version: (version, Ultima). La
// última pasa a ser la anterior, así que una lectura que ve lo mismo que las
// dos anteriores deja la fila como estaba (data-model §4).
func (l LecturasDeBloque) Leida(version string) LecturasDeBloque {
	return LecturasDeBloque{Bloque: l.Bloque, Ultima: version, Anterior: l.Ultima}
}

// RedaccionVistaSinLecturas es la redacción vista de un bloque que no tiene
// fila de lecturas —el de un world.db de H7, o uno que H7 observó y nadie ha
// vuelto a leer—, que cuenta con una lectura, la de su redacción observada por
// última vez (FR-026; research.md D3): de versiones, que son las BloqueVersion
// del bloque, el id de la de última observación más reciente, comparando los
// instantes de sus fechas de consulta; a igualdad de instante, la de id menor,
// comparando bytes. El orden de versiones no cuenta. Dice si hay alguna: sin
// ninguna no hay redacción vista.
//
// Una fecha de consulta que no es RFC 3339, que ninguna entrega guarda, da un
// error, que no nombra el id del nodo (research.md D21).
func RedaccionVistaSinLecturas(versiones []NodoDeInstantanea) (string, bool, error) {
	var (
		vista       string
		observadaEl time.Time
	)

	for i, version := range versiones {
		observada, err := instante(version.UltimaObservacion.FechaConsulta)
		if err != nil {
			return "", false, err
		}

		if i == 0 || cmp.Or(observada.Compare(observadaEl), strings.Compare(vista, version.ID)) > 0 {
			vista, observadaEl = version.ID, observada
		}
	}

	return vista, len(versiones) > 0, nil
}
