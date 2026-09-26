package instalacion

// ManifiestoIlegible es un kitlegal.json que existe y no se puede usar: no es
// un fichero regular, no se puede leer o no respeta la forma de
// contracts/manifiesto.md (FR-035). Quien no puede leer el manifiesto no sabe
// qué es suyo, así que install lo nombra como conflicto y no toca nada, list y
// doctor salen nombrándolo y el aviso no tiene ningún efecto.
//
// No declara una clase a propósito: nunca llega así al kernel. El conflicto
// de install y el ámbito ilegible de list y doctor son los errores que la
// declaran (data-model §9). Se exporta para reconocerlo con errors.As.
type ManifiestoIlegible struct {
	// causa dice qué lo hace ilegible; puede envolver el error de
	// encoding/json/v2 que rechazó el documento.
	causa error
}

// Error dice que el manifiesto es ilegible y por qué, si se sabe.
func (e *ManifiestoIlegible) Error() string {
	if e.causa == nil {
		return "manifiesto ilegible"
	}

	return "manifiesto ilegible: " + e.causa.Error()
}

// Unwrap devuelve lo que lo hace ilegible.
func (e *ManifiestoIlegible) Unwrap() error {
	return e.causa
}
