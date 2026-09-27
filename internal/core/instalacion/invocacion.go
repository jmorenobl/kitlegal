package instalacion

import "slices"

// Invocacion es lo que pide una invocación de skills install, list o doctor,
// tal como llega de la línea de órdenes (contracts/applet-skills.md §1). list
// y doctor no declaran nombres ni --host, así que su invocación es la de
// install sin ellos.
type Invocacion struct {
	// Skills son los nombres posicionales de install, tal cual y en el orden
	// en que se escribieron; vacía sin ninguno.
	Skills []string

	// Global dice si se pasó -g/--global.
	Global bool

	// Host es el valor de --host, o nil si no se pasó; un valor vacío
	// también se pasó.
	Host *string

	// Dir es la ruta de --dir tal como se pasó, con /, o nil si no se pasó;
	// una ruta vacía también se pasó.
	Dir *string
}

// Pedido es una invocación ya validada: dónde actúa y qué pide.
type Pedido struct {
	// Ambito es el de la invocación (data-model §2).
	Ambito Ambito

	// Skills son las pedidas, en orden de nombre y cada una una vez: sin
	// nombres, todas las empotradas, lo que instala install sin nombres
	// (FR-010).
	Skills []string

	// HostClaude dice si se pidió --host claude, que enlaza en el host aunque
	// .claude no exista (FR-023).
	HostClaude bool
}

// ValidarInvocacion comprueba la invocación antes de tocar el disco —no
// recibe el puerto Disco, y HOME le llega como un valor, home, vacío si no
// está definido o está vacío— con la precedencia de FR-052, en el orden de
// contracts/applet-skills.md §2; la primera que se cumple decide:
//
//  1. -g junto a --dir, en los tres verbos (FR-013);
//  2. --host junto a --dir (FR-013);
//  3. --host con un valor distinto de claude (FR-020);
//  4. un nombre que no es de ninguna de las skills empotradas, también el de
//     una que el manifiesto declara y el binario no empotra (FR-010, FR-036),
//     nombrando la primera y las disponibles;
//  5. y solo sin ninguna de las anteriores, -g sin HOME (FR-012).
//
// Las cuatro primeras son de clase «argumentos» (código 2) y ganan a la
// quinta, de clase «inesperado» (código 1), y a cualquier exit 1 del ámbito,
// que no llega a comprobarse. Con un rechazo, el Pedido es el valor cero. Sin
// ninguno, devuelve el ámbito de la invocación y las skills pedidas.
func ValidarInvocacion(invocacion Invocacion, home string, empotradas []SkillEmpotrada) (Pedido, error) {
	if err := comprobarBanderas(invocacion); err != nil {
		return Pedido{}, err
	}

	pedidas, err := skillsPedidas(invocacion.Skills, empotradas)
	if err != nil {
		return Pedido{}, err
	}

	ambito, err := ambitoDe(invocacion, home)
	if err != nil {
		return Pedido{}, err
	}

	return Pedido{Ambito: ambito, Skills: pedidas, HostClaude: invocacion.Host != nil}, nil
}

// comprobarBanderas aplica las filas 1 a 3, las de las banderas, en orden.
func comprobarBanderas(invocacion Invocacion) error {
	switch {
	case invocacion.Global && invocacion.Dir != nil:
		return globalConDir()
	case invocacion.Host != nil && invocacion.Dir != nil:
		return hostConDir()
	case invocacion.Host != nil && *invocacion.Host != hostClaude:
		return hostNoAdmitido(*invocacion.Host)
	}

	return nil
}

// skillsPedidas aplica la fila 4 y devuelve las pedidas: sin nombres, todas
// las empotradas; con nombres, esos, si todos son de una empotrada. En los
// dos casos, en orden de nombre y cada una una vez, sobre una copia: los
// nombres que recibe no cambian.
func skillsPedidas(nombres []string, empotradas []SkillEmpotrada) ([]string, error) {
	disponibles := make([]string, 0, len(empotradas))
	for _, skill := range empotradas {
		disponibles = append(disponibles, skill.Nombre)
	}

	slices.Sort(disponibles)
	disponibles = slices.Compact(disponibles)

	if len(nombres) == 0 {
		return disponibles, nil
	}

	for _, nombre := range nombres {
		if _, empotrada := slices.BinarySearch(disponibles, nombre); !empotrada {
			return nil, skillDesconocida(nombre, disponibles)
		}
	}

	pedidas := slices.Clone(nombres)
	slices.Sort(pedidas)

	return slices.Compact(pedidas), nil
}

// ambitoDe es el ámbito de una invocación cuyas banderas ya se comprobaron:
// --dir, -g, que necesita HOME (fila 5), o el local.
func ambitoDe(invocacion Invocacion, home string) (Ambito, error) {
	switch {
	case invocacion.Dir != nil:
		return NuevoAmbitoDir(*invocacion.Dir), nil
	case invocacion.Global:
		return NuevoAmbitoGlobal(home)
	}

	return NuevoAmbitoLocal(), nil
}
