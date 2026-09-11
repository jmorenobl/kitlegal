package schema

// Clase es el vocabulario de clases de error que aparece en el sobre de fallo y
// en el esquema que emite --describe. Vive en el dominio porque forma parte del
// contrato JSON; la traducción de cada clase a código de salida vive en el
// kernel, en un único switch sin rama por defecto (FR-029, FR-030).
type Clase string

// Las seis clases del contrato (contracts/sobre-de-salida.md §6). Cada una
// corresponde a un código de salida, y ClaseInesperado recoge todo fallo que no
// encaja en las demás (FR-031).
const (
	// ClaseArgumentos es la bandera desconocida, el valor inválido, el
	// argumento obligatorio ausente y el applet no registrado.
	ClaseArgumentos Clase = "argumentos"
	// ClaseNoEncontrado es lo pedido que no existe en la fuente.
	ClaseNoEncontrado Clase = "no-encontrado"
	// ClaseFuenteNoDisponible es la fuente que no responde, y también el plazo
	// agotado.
	ClaseFuenteNoDisponible Clase = "fuente-no-disponible"
	// ClaseLimiteOTos es el límite de peticiones alcanzado o la restricción de
	// los términos de uso.
	ClaseLimiteOTos Clase = "limite-o-tos"
	// ClaseIdentidadHumana es la acción que requiere identidad humana y que,
	// por tanto, no se ha realizado.
	ClaseIdentidadHumana Clase = "identidad-humana"
	// ClaseInesperado es cualquier fallo que no encaja en los anteriores.
	ClaseInesperado Clase = "inesperado"
)

// DatosError es lo que ocupa data cuando ok es falso: dos claves y ninguna más
// —ni traza, ni código numérico, ni error envuelto—, porque el detalle técnico
// va al registro de eventos y no al sobre (FR-045, research.md D7).
type DatosError struct {
	// Clase es una de las seis, y corresponde al código de salida emitido.
	Clase Clase `json:"clase"`
	// Mensaje es el mensaje dirigido a la persona, el mismo que va a la salida
	// de error.
	Mensaje string `json:"mensaje"`
}
