package instalacion

// TipoDeEntrada es lo que hay en una ruta, visto sin seguir un enlace
// simbólico (data-model §3; FR-028).
type TipoDeEntrada int

// Los cinco tipos de entrada. El valor cero es la ruta que no existe.
const (
	// EntradaAusente es la ruta en la que no hay nada.
	EntradaAusente TipoDeEntrada = iota
	// EntradaDirectorio es un directorio real, no un enlace a uno.
	EntradaDirectorio
	// EntradaFichero es un fichero regular.
	EntradaFichero
	// EntradaEnlace es un enlace simbólico, resuelva o no.
	EntradaEnlace
	// EntradaOtra es cualquier otra cosa: una tubería con nombre, un socket o
	// un dispositivo.
	EntradaOtra
)

// Entrada es lo que el dominio sabe de una ruta: su tipo y, si es un enlace,
// su destino y si resuelve. Nunca se sabe nada de lo que hay al otro lado de
// un enlace (FR-028).
type Entrada struct {
	// Tipo es el de la propia entrada, sin seguirla.
	Tipo TipoDeEntrada

	// Destino es, en un enlace, su destino literal, tal como se escribió al
	// crearlo; vacío en los demás tipos.
	Destino string

	// Resuelve dice, en un enlace, si seguirlo, con los enlaces que encadene,
	// llega a algo que existe; es falso si cuelga o está en un ciclo, y en los
	// demás tipos.
	Resuelve bool
}

// Disco es la lectura del sistema de ficheros que necesita el dominio, sin
// seguir nunca un enlace simbólico por debajo de la raíz del ámbito (FR-028;
// research.md D6). Lo implementa el adaptador internal/disco.
//
// Las rutas son las del Ámbito: con /, limpias y como se alcanzan desde el
// directorio de trabajo; lo que haya por encima de la raíz del ámbito se usa
// tal cual (FR-027). Que una ruta no exista no es un error. Cualquier otro
// fallo de entrada y salida sí lo es, y es un fallo de la orden, no un
// conflicto (data-model §3): el dominio lo devuelve tal cual, así que el error
// tiene que nombrar la operación y la ruta.
type Disco interface {
	// Examinar dice qué hay en ruta sin seguirla. Solo para decir si un enlace
	// resuelve se sigue su destino, sin leer nada de lo que encuentre.
	Examinar(ruta string) (Entrada, error)

	// Huella es la huella de los bytes del fichero regular de ruta, con la
	// forma de HuellaDe. Lo que no es un fichero regular no se abre y es un
	// error, también si lo era al examinarlo y ya no lo es al abrirlo.
	Huella(ruta string) (string, error)

	// Leer son los bytes del fichero regular de ruta, con las mismas
	// garantías que Huella.
	Leer(ruta string) ([]byte, error)

	// Nombres son los nombres de las entradas del directorio real de ruta, en
	// cualquier orden; lo que no es un directorio real no se lista y es un
	// error.
	Nombres(ruta string) ([]string, error)
}

// Enlazador crea los enlaces simbólicos de las entradas de host y dice si se
// pueden crear en un directorio (FR-024; research.md D9). En test se sustituye
// por uno que falla.
type Enlazador interface {
	// Disponible dice si en directorio se puede crear un enlace simbólico. El
	// del sistema lo sondea creando y retirando uno de prueba, y deja el
	// directorio con las mismas entradas y los mismos bytes; si no puede
	// retirarlo, es un error, que nombra la ruta de la sonda.
	Disponible(directorio string) (bool, error)

	// Enlazar crea en ruta un enlace simbólico con ese destino literal.
	Enlazar(destino, ruta string) error
}

// Escritor son las operaciones sueltas con las que se aplica un plan, fase a
// fase (data-model §5; research.md D7, D8). No decide nada: el orden y lo que
// se escribe los fija el plan.
type Escritor interface {
	// CrearDirectorio crea el directorio real de ruta, cuyo padre existe.
	CrearDirectorio(ruta string) error

	// EscribirFichero deja en ruta un fichero regular con contenido, de forma
	// atómica —un temporal del mismo directorio renombrado encima— y sin
	// permiso de ejecución (FR-015).
	EscribirFichero(ruta string, contenido []byte) error

	// Retirar retira la entrada de ruta: un fichero, un enlace, sin seguirlo,
	// o un directorio vacío.
	Retirar(ruta string) error

	// Enlazar crea en ruta un enlace simbólico con ese destino literal, con el
	// Enlazador de la invocación.
	Enlazar(destino, ruta string) error
}
