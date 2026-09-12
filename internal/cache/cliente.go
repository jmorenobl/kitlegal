package cache

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// VariableDirectorio es la variable de entorno con la que se declara dónde vive
// la caché sin tocar ninguna invocación. Se exporta porque quien la documenta,
// la nombra en un mensaje o la fija en un test tiene que escribirla una sola
// vez y en un solo sitio (FR-023).
const VariableDirectorio = "KITLEGAL_CACHE_DIR"

// ajustes recoge lo que las opciones declaran antes de construir el cliente. Es
// privado: las opciones son el único modo de tocarlo, y por eso ninguna de sus
// combinaciones puede llegar a un cliente sin pasar por la validación de New.
//
// El nombre no es «configuración», que es el del contrato, por una razón
// mecánica: misspell (locale US) marca esa palabra suelta como errata inglesa,
// el hito no admite ninguna supresión y esta tarea no declara el fichero de
// configuración del lint, así que rige la reescritura del término que el plan
// deja prevista (supuesto S2, D15, que nombra «ajustes»).
type ajustes struct {
	directorio  string
	soloLectura bool
	reloj       func() time.Time
	registrador *slog.Logger
}

// Opcion es lo que ajusta un cliente antes de construirlo. Devuelve error
// porque la validación ocurre al construir y no al escribir la llamada: New las
// aplica en el orden en que llegan, la primera inválida es «argumentos» (2) y
// la última repetida gana (D2).
type Opcion func(*ajustes) error

// ConDirectorio declara dónde vive la caché y tiene la precedencia máxima: por
// encima de la variable de entorno y del directorio de la cuenta (FR-023). La
// cadena vacía es «argumentos» (2): no declarar la opción y declararla sin nada
// dentro no son lo mismo.
func ConDirectorio(dir string) Opcion {
	return func(a *ajustes) error {
		if dir == "" {
			return errorDeRutaVacia("construir", origenOpcion)
		}

		a.directorio = dir

		return nil
	}
}

// SoloLectura construye un cliente que no crea ni escribe nada: ni el
// directorio, ni la base de datos, ni ninguna entrada. Es el modo que la
// invocación con --offline pide, y se declara al construir porque crear la base
// y migrarla ocurre al abrir y no al operar (FR-015).
func SoloLectura() Opcion {
	return func(a *ajustes) error {
		a.soloLectura = true

		return nil
	}
}

// ConReloj declara de dónde sale «ahora» para escribir la expiración de una
// entrada y para decidir si sigue vigente. Por omisión es time.Now; inyectarlo
// es lo que permite comprobar el borde de la expiración sin esperar tiempo real
// (FR-009). Un reloj nulo es «argumentos» (2).
func ConReloj(ahora func() time.Time) Opcion {
	return func(a *ajustes) error {
		if ahora == nil {
			return errorDeArgumentos("construir", "la opción ConReloj no lleva ningún reloj", nil)
		}

		a.reloj = ahora

		return nil
	}
}

// ConRegistrador declara a dónde van los eventos de nivel debug de la caché,
// que es el mismo registrador que el kernel entrega al applet. Sin la opción se
// descartan. Un registrador nulo es «argumentos» (2): quien pasa un nulo quería
// registrar algo y en silencio no registraría nada (D14).
func ConRegistrador(registrador *slog.Logger) Opcion {
	return func(a *ajustes) error {
		if registrador == nil {
			return errorDeArgumentos("construir", "la opción ConRegistrador no lleva ningún registrador", nil)
		}

		a.registrador = registrador

		return nil
	}
}

// Cliente es la caché: el único objeto del módulo que abre la base de datos y
// el que implementa el puerto del dominio. Un solo tipo para los dos modos —el
// normal y el de solo lectura—, porque partir el puerto o el cliente en dos
// está prohibido (FR-017).
//
// Es seguro usarlo desde varias goroutines a la vez.
type Cliente struct {
	// directorio es el directorio efectivo tal como se resolvió, sin sanear:
	// es el que los mensajes nombran, y nombrarlo de otra forma que como se
	// declaró haría más difícil reconocerlo.
	directorio string
	// ruta es la base de datos, <directorio>/cache.db, saneada por
	// filepath.Join. Se fija al construir y no cambia (FR-020, FR-023).
	ruta string
	// origen es de dónde salió el directorio, solo para los mensajes (FR-022).
	origen origenDeLaRuta
	// soloLectura lo fija SoloLectura() al construir y no cambia después
	// (FR-015).
	soloLectura bool
	// reloj nunca es nulo: time.Now por omisión (FR-009).
	reloj func() time.Time
	// registrador nunca es nulo: descarta por omisión (D14).
	registrador *slog.Logger

	// db es la conexión con la base de datos, y es nula en un solo caso: el
	// cliente de solo lectura que no encontró cache.db, que se construye sin
	// base porque este modo no crea ninguna y para el que toda lectura es una
	// ausencia (FR-015). No se expone de ninguna forma: la imposibilidad de
	// ejecutar SQL desde fuera del paquete es por construcción (FR-005).
	db *sql.DB
	// versionEsquema es la versión que la base tiene aplicada: la que este
	// binario conoce cuando hay esquema, y 0 cuando no lo hay —en solo lectura,
	// que no migra— (FR-024, FR-025).
	versionEsquema int64

	// mu protege el estado de cierre, que es lo único que cambia en la vida del
	// cliente. Lo demás se fija al construir y solo se lee (D10).
	mu      sync.Mutex
	cerrado bool
}

// New construye la caché: aplica las opciones en orden, resuelve de dónde sale
// la ruta y comprueba que el valor sirve. Recibe contexto porque construir la
// caché **es** una operación con entrada y salida, y no crea ninguno por su
// cuenta (FR-003).
//
// El orden de los fallos no es casual. Primero las opciones, porque una opción
// inválida es un error del código que llama y se corrige sin mirar el entorno;
// después la ruta, que es lo que la persona declara; y solo entonces el
// contexto, para que un plazo agotado no tape un fallo de argumentos que
// aparecería igual con todo el tiempo del mundo. Un contexto cancelado o
// vencido es «fuente no disponible» (4), la misma clase con la que el kernel
// trata el plazo agotado.
//
// Construir la caché abre la base de datos, y en modo normal la crea y la
// migra: el esquema se pone al día al abrir y no al operar, que es la razón por
// la que el modo de solo lectura tiene que conocerse aquí y no en cada llamada
// (FR-015, FR-024). Un fallo al abrir no deja ninguna conexión abierta y no
// devuelve ningún cliente.
func New(ctx context.Context, opciones ...Opcion) (*Cliente, error) {
	declarados := ajustes{
		reloj:       time.Now,
		registrador: slog.New(slog.DiscardHandler),
	}

	for _, opcion := range opciones {
		if err := opcion(&declarados); err != nil {
			return nil, err
		}
	}

	directorio, de, err := rutaEfectiva(declarados.directorio, os.LookupEnv)
	if err != nil {
		return nil, err
	}

	if err := compruebaRuta(directorio, de); err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, errorDeFuenteNoDisponible("construir",
			"el contexto terminó antes de construir la caché", err)
	}

	cliente := &Cliente{
		directorio:  directorio,
		ruta:        filepath.Join(directorio, ficheroDeLaBase),
		origen:      de,
		soloLectura: declarados.soloLectura,
		reloj:       declarados.reloj,
		registrador: declarados.registrador,
	}

	cliente.registrador.DebugContext(ctx, "caché: ruta resuelta",
		slog.String("ruta", cliente.ruta),
		slog.String("origen", string(cliente.origen)),
		slog.Bool("solo_lectura", cliente.soloLectura))

	if err := cliente.abre(ctx); err != nil {
		return nil, err
	}

	return cliente, nil
}

// Close cierra la conexión con la base de datos y es idempotente: la primera
// llamada la cierra y devuelve lo que dijera el cierre, y las siguientes no
// hacen nada y devuelven nil (FR-004). Es lo que hace que el defer de quien
// construyó el cliente y un cierre explícito antes de tiempo convivan sin que el
// segundo parezca un error.
//
// Un cliente sin base —el de solo lectura que no encontró cache.db— no tiene
// nada que cerrar, y cerrarlo tampoco falla.
func (c *Cliente) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cerrado {
		return nil
	}

	c.cerrado = true

	if c.db == nil {
		return nil
	}

	if err := c.db.Close(); err != nil {
		return c.falloInesperadoEn("cerrar", fmt.Sprintf("no se pudo cerrar %q", c.ruta), err)
	}

	return nil
}
