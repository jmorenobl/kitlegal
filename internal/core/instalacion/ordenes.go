package instalacion

import (
	"cmp"
	"slices"
	"strings"
)

// retirada es cómo empieza la orden de un hallazgo: sin retirar nada, o
// retirando la entrada del hallazgo con rm, o con rm -r si es un directorio
// real (FR-066).
type retirada int

// Las tres formas de empezar una orden. Ninguna lleva -f.
const (
	// sinRetirar es la orden que solo reinstala.
	sinRetirar retirada = iota
	// conRm retira un fichero, un enlace, sin seguirlo, u otra entrada.
	conRm
	// conRmR retira un directorio real entero.
	conRmR
)

// ordenQueArregla es la orden que arregla un hallazgo, una sola línea de shell
// POSIX (FR-066; contracts/applet-skills.md §6):
//
//	[rm [-r] -- '<ruta>' && ]kitlegal skills install <skill>… [-g | --dir '<ruta>'] [--host claude]
//
// quitar dice si retira antes la entrada de ruta, la del hallazgo, detrás de
// --, para que ninguna ruta se lea como una opción. install nombra las skills
// —ninguna, que son todas las empotradas, si no hay ninguna— con las banderas
// del ámbito, y --host claude si alguna de ellas tiene una entrada de host
// declarada, que nunca con --dir, cuyo ámbito no tiene hosts. Cada ruta va
// entre comillas simples, con cada comilla simple escapada.
func ordenQueArregla(ambito Ambito, quitar retirada, ruta string, skills []string, conHost bool) string {
	partes := make([]string, 0, len(skills)+10)

	switch quitar {
	case conRm:
		partes = append(partes, "rm", "--", entreComillas(ruta), "&&")
	case conRmR:
		partes = append(partes, "rm", "-r", "--", entreComillas(ruta), "&&")
	case sinRetirar:
	}

	partes = append(partes, "kitlegal", "skills", "install")
	partes = append(partes, skills...)

	if banderas := ambito.Banderas(); banderas != "" {
		partes = append(partes, banderas)
	}

	if conHost && ambito.ConHosts() {
		partes = append(partes, "--host", hostClaude)
	}

	return strings.Join(partes, " ")
}

// pendiente es un hallazgo con lo que decide su lugar en la lista: si su orden
// retira algo con rm.
type pendiente struct {
	hallazgo Hallazgo
	retira   bool
}

// grupo es 0 para los que llevan rm, que van primero, y 1 para los demás.
func (p pendiente) grupo() int {
	if p.retira {
		return 0
	}

	return 1
}

// enSuOrden son los hallazgos en el orden de FR-066: primero los que llevan
// rm, después los demás; en cada grupo, por ruta comparada byte a byte y, a
// igual ruta, por el número de su clase. Así, ejecutar las órdenes en ese orden
// retira primero lo que está fuera de su sitio y reinstala después.
func enSuOrden(pendientes []pendiente) []Hallazgo {
	ordenados := slices.Clone(pendientes)
	slices.SortStableFunc(ordenados, func(a, b pendiente) int {
		return cmp.Or(
			cmp.Compare(a.grupo(), b.grupo()),
			cmp.Compare(a.hallazgo.Ruta, b.hallazgo.Ruta),
			cmp.Compare(a.hallazgo.Clase.numero(), b.hallazgo.Clase.numero()),
		)
	})

	hallazgos := make([]Hallazgo, 0, len(ordenados))
	for _, p := range ordenados {
		hallazgos = append(hallazgos, p.hallazgo)
	}

	return hallazgos
}
