package territorio

import (
	"github.com/jmorenobl/kitlegal/internal/core/grafo"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// prefijoDeMunicipio antecede al código INE de cinco cifras en el id de un
// Municipio (FR-043).
const prefijoDeMunicipio = "ine:"

// Observado es lo que el territorio resuelto observa del mundo, con los
// identificadores naturales que ya lleva su data (contracts/emision.md §2;
// research.md D21):
//
//   - el Municipio, «ine:» y su código INE de cinco cifras, con ese código y su
//     nombre oficial, siempre (FR-043);
//   - si la respuesta trae el DIR3 del ayuntamiento, el Organo por ese DIR3, con
//     él en sus datos, y la arista lb:pertenece_a del Organo al Municipio; sin
//     DIR3, ni Organo ni arista: nunca un código derivado (FR-044).
//
// No declara vigencia, porque territorio no consulta ninguna fuente en
// ejecución (FR-065), y no depende de la cobertura: un municipio cubierto y uno
// no cubierto observan lo mismo (FR-045). Cada llamada devuelve operaciones
// nuevas, que quien las recibe puede cambiar sin cambiar las de nadie más.
func (t Territorio) Observado() schema.Observado {
	municipio := prefijoDeMunicipio + t.CodigoINE.Codigo

	operaciones := []schema.Operacion{
		schema.Nodo{ID: municipio, Tipo: grafo.TipoMunicipio, Datos: map[string]any{
			grafo.DatoCodigoINE: t.CodigoINE.Codigo,
			grafo.DatoNombre:    t.Municipio.Nombre,
		}},
	}

	if dir3 := t.DIR3.Codigo; dir3 != "" {
		operaciones = append(operaciones,
			schema.Nodo{ID: dir3, Tipo: grafo.TipoOrgano, Datos: map[string]any{grafo.DatoDIR3: dir3}},
			schema.Arista{Origen: dir3, Relacion: grafo.RelacionPerteneceA, Destino: municipio},
		)
	}

	return schema.Observado{Operaciones: operaciones}
}
