package instalacion

import "slices"

// Las operaciones del Escritor que pueden parar la aplicación, como las nombra
// su fallo (contracts/applet-skills.md §5). Enlazar no está: que no se pueda
// crear un enlace no para la orden (FR-024).
const (
	operacionRetirar  = "retirar"
	operacionCrear    = "crear"
	operacionEscribir = "escribir"
)

// Aplicar lleva a cabo el plan a través del escritor, fase a fase y en el
// orden de research.md D7, y devuelve las skills pedidas tal como quedan, que
// son la salida de install (FR-051):
//
//  1. retira, en su orden, lo que el plan retira;
//  2. crea lo que falta hasta el directorio de skills de cada host y cada
//     enlace de host; si uno no se puede crear, su entrada pasa al recurso de
//     copia, que se escribe en la fase 4, y los demás enlaces se intentan
//     igual (FR-024);
//  3. crea lo que falta hasta el directorio neutro y escribe, de una vez, el
//     manifiesto final, con cada entrada cuyo enlace no se pudo crear en
//     copia, si cambia;
//  4. crea cada directorio que falta y escribe cada fichero: los del plan y,
//     detrás, los de cada recurso de copia.
//
// La salida es la del plan, con la entrada de host de cada enlace que no se
// pudo crear en modo copia; el plan no cambia. Aplicar no decide nada más: lo
// que se retira, se crea y se escribe, y en qué orden, lo fija el plan.
//
// Se para en la primera operación que falla, sin pedir ninguna más y sin
// salida, con un error que nombra la operación y la ruta y envuelve el del
// escritor (FR-044; contracts/applet-skills.md §5). Por el orden de las fases,
// lo que queda en el disco es un estado que volver a planificar completa sin
// ningún conflicto.
func Aplicar(plan Plan, escritor Escritor) ([]SkillInstalada, error) {
	for _, ruta := range plan.Retirar {
		if err := escritor.Retirar(ruta); err != nil {
			return nil, falloAlAplicar(operacionRetirar, ruta, err)
		}
	}

	if err := crearDirectorios(escritor, plan.Enlazar.DirectoriosQueFaltan); err != nil {
		return nil, err
	}

	enCopia, escribir := enlazarLosHosts(escritor, plan)

	if err := escribirElManifiesto(escritor, plan.Manifiesto, enCopia); err != nil {
		return nil, err
	}

	if err := escribirLasEscrituras(escritor, escribir); err != nil {
		return nil, err
	}

	return salidaAplicada(plan.Skills, enCopia), nil
}

// crearDirectorios crea cada directorio de rutas, en orden.
func crearDirectorios(escritor Escritor, rutas []string) error {
	for _, ruta := range rutas {
		if err := escritor.CrearDirectorio(ruta); err != nil {
			return falloAlAplicar(operacionCrear, ruta, err)
		}
	}

	return nil
}

// enlazarLosHosts crea, en orden, cada enlace de host del plan y devuelve las
// rutas de las entradas cuyo enlace no se pudo crear y lo que se escribe en la
// fase 4: lo del plan y, detrás, el recurso de copia de cada una de ellas. El error de un
// enlace que no se puede crear no se propaga porque no es un fallo de la
// orden: FR-024 manda copiar en su lugar y nombrar la entrada en copia, que
// es lo que decide.
func enlazarLosHosts(escritor Escritor, plan Plan) ([]string, Escrituras) {
	escribir := Escrituras{
		DirectoriosQueFaltan: slices.Clone(plan.Escribir.DirectoriosQueFaltan),
		Ficheros:             slices.Clone(plan.Escribir.Ficheros),
	}

	var enCopia []string

	for _, enlace := range plan.Enlazar.Enlaces {
		if escritor.Enlazar(enlace.Destino, enlace.Ruta) == nil {
			continue
		}

		enCopia = append(enCopia, enlace.Ruta)
		escribir.DirectoriosQueFaltan = append(escribir.DirectoriosQueFaltan, enlace.Copia.DirectoriosQueFaltan...)
		escribir.Ficheros = append(escribir.Ficheros, enlace.Copia.Ficheros...)
	}

	return enCopia, escribir
}

// escribirElManifiesto crea lo que falta hasta el directorio neutro y escribe
// el manifiesto final, con cada entrada de host de enCopia en copia, si
// cambia.
func escribirElManifiesto(escritor Escritor, manifiesto ManifiestoFinal, enCopia []string) error {
	if err := crearDirectorios(escritor, manifiesto.DirectoriosQueFaltan); err != nil {
		return err
	}

	contenido, err := manifiesto.Contenido(enCopia)
	if err != nil {
		return falloAlAplicar(operacionEscribir, manifiesto.Ruta, err)
	}

	if contenido == nil {
		return nil
	}

	if err := escritor.EscribirFichero(manifiesto.Ruta, contenido); err != nil {
		return falloAlAplicar(operacionEscribir, manifiesto.Ruta, err)
	}

	return nil
}

// escribirLasEscrituras crea cada directorio que falta y, después, escribe
// cada fichero, en orden.
func escribirLasEscrituras(escritor Escritor, escribir Escrituras) error {
	if err := crearDirectorios(escritor, escribir.DirectoriosQueFaltan); err != nil {
		return err
	}

	for _, fichero := range escribir.Ficheros {
		if err := escritor.EscribirFichero(fichero.Ruta, fichero.Contenido); err != nil {
			return falloAlAplicar(operacionEscribir, fichero.Ruta, err)
		}
	}

	return nil
}

// salidaAplicada son las skills previstas, en una lista nueva, con cada
// entrada de host de enCopia —la de un enlace que no se pudo crear— en copia.
// Es una lista, aunque esté vacía.
func salidaAplicada(previstas []SkillInstalada, enCopia []string) []SkillInstalada {
	skills := make([]SkillInstalada, 0, len(previstas))

	for _, skill := range previstas {
		skill.Enlaces = slices.Clone(skill.Enlaces)

		for i, enlace := range skill.Enlaces {
			if slices.Contains(enCopia, enlace.Ruta) {
				skill.Enlaces[i].Modo = ModoCopia
			}
		}

		skills = append(skills, skill)
	}

	return skills
}
