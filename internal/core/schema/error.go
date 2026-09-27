package schema

// Clase es el vocabulario de clases de error que aparece en el sobre de fallo y
// en el esquema que emite --describe. Vive en el dominio porque forma parte del
// contrato JSON; la traducción de cada clase a código de salida vive en el
// kernel, en un único switch sin rama por defecto (FR-029, FR-030).
type Clase string

// Las siete clases del contrato (contracts/sobre-de-salida.md §6; ADR 0023).
// Cada una corresponde a un código de salida, y ClaseInesperado recoge todo
// fallo que no encaja en las demás (FR-031).
//
// Son clases de fallo: la orden no hizo lo que se le pidió. Una verificación
// que encuentra algo sí lo hizo, así que sus hallazgos no son una clase sino
// datos de un resultado correcto (ADR 0023).
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
	// ClaseConflicto es la orden que no se ejecuta porque el estado local lo
	// impide —una entrada que no es suya, un manifiesto ilegible— y que la
	// persona puede resolver. Es una situación prevista, no un defecto: por
	// eso no es inesperada (ADR 0023).
	ClaseConflicto Clase = "conflicto"
	// ClaseInesperado es cualquier fallo que no encaja en los anteriores: un
	// defecto del programa o del entorno.
	ClaseInesperado Clase = "inesperado"
)

// Clases es el vocabulario completo, en el orden del contrato: las siete
// clases y ninguna más. Es de lo que el kernel deriva el `enum` con el que
// --describe restringe `clase` en el sobre de fallo, de modo que el esquema y las
// constantes no puedan decir cosas distintas (FR-017, FR-048). Devuelve una
// lista nueva en cada llamada: nadie puede alterar el vocabulario desde fuera.
func Clases() []Clase {
	return []Clase{
		ClaseArgumentos,
		ClaseNoEncontrado,
		ClaseFuenteNoDisponible,
		ClaseLimiteOTos,
		ClaseIdentidadHumana,
		ClaseConflicto,
		ClaseInesperado,
	}
}

// ConClase lo implementa un error que declara él mismo la clase con la que hay
// que traducirlo a código de salida. Es el puerto por el que un adaptador —el
// cliente HTTP, mañana una fuente— nombra su clase sin importar el kernel: la
// interfaz vive en el dominio y es el kernel quien la reconoce, de modo que la
// dependencia sigue yendo hacia dentro y no entre adaptadores (research.md D4).
//
// Embebe error, así que quien la implementa es un error y se devuelve como tal.
// Y como el kernel la busca con errors.As, que sigue la cadena de Unwrap,
// envolver con %w el error que declara la clase no cambia la clase (FR-031).
//
// Clase debe devolver una de las de Clases(). Quien clasifica no se fía: una
// clase de fuera del vocabulario no podría ir en el sobre —el esquema de
// --describe la restringe a las siete—, así que la trata como lo que es, un
// fallo no previsto (FR-063).
type ConClase interface {
	error
	Clase() Clase
}

// DatosError es lo que ocupa data cuando ok es falso: dos claves y ninguna más
// —ni traza, ni código numérico, ni error envuelto—, porque el detalle técnico
// va al registro de eventos y no al sobre (FR-045, research.md D7).
//
// La etiqueta `jsonschema` de Mensaje es, como las del sobre, la descripción
// formal del contrato escrita junto a la clave (contracts/sobre-de-salida.md
// §6); la de Clase no hace falta, porque el kernel describe el tipo Clase
// entero con el vocabulario de Clases.
type DatosError struct {
	// Clase es una de las siete, y corresponde al código de salida emitido.
	Clase Clase `json:"clase"`
	// Mensaje es el mensaje dirigido a la persona, el mismo que va a la salida
	// de error. Nunca va vacío: un fallo sin mensaje no le dice nada a nadie.
	Mensaje string `json:"mensaje" jsonschema:"minLength=1"`
}
