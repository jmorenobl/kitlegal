package httpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCabecerasGet comprueba las dos reglas de Cabeceras.Get (data-model.md §3):
// canonicaliza el nombre antes de buscarlo —para que un adaptador pueda pedir
// «content-type» sin conocer la forma canónica ni importar la biblioteca HTTP
// (R2)— y devuelve la cadena vacía cuando la cabecera no está, en lugar de
// obligar a quien llama a distinguir el caso (D2).
func TestCabecerasGet(t *testing.T) {
	t.Parallel()

	// Las cabeceras de una respuesta llegan siempre con los nombres canónicos:
	// es la forma en que el paquete las entrega.
	deLaRespuesta := Cabeceras{
		"Content-Type": {"application/xml; charset=UTF-8"},
		"Retry-After":  {"120", "240"},
		"X-Sin-Valor":  {},
	}

	casos := []struct {
		nombre    string
		cabeceras Cabeceras
		busca     string
		valor     string
	}{
		{
			nombre:    "nombre ya canónico",
			cabeceras: deLaRespuesta,
			busca:     "Content-Type",
			valor:     "application/xml; charset=UTF-8",
		},
		{
			nombre:    "nombre en minúsculas: se canonicaliza",
			cabeceras: deLaRespuesta,
			busca:     "content-type",
			valor:     "application/xml; charset=UTF-8",
		},
		{
			nombre:    "nombre en mayúsculas: se canonicaliza",
			cabeceras: deLaRespuesta,
			busca:     "RETRY-AFTER",
			valor:     "120",
		},
		{
			nombre:    "varios valores: el primero",
			cabeceras: deLaRespuesta,
			busca:     "Retry-After",
			valor:     "120",
		},
		{
			nombre:    "cabecera ausente: cadena vacía",
			cabeceras: deLaRespuesta,
			busca:     "Location",
			valor:     "",
		},
		{
			nombre:    "cabecera presente sin ningún valor: cadena vacía",
			cabeceras: deLaRespuesta,
			busca:     "x-sin-valor",
			valor:     "",
		},
		{
			nombre:    "sin cabeceras: cadena vacía y ningún acceso a un mapa nulo",
			cabeceras: nil,
			busca:     "Content-Type",
			valor:     "",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.valor, caso.cabeceras.Get(caso.busca))
		})
	}
}

// TestRespuestaDescripcion comprueba que la descripción es la de la petición
// **pedida** y no la de la dirección final: es la línea que el adaptador copia
// en schema.Resultado.Ensayo bajo --dry-run, donde nada se ha pedido todavía y
// por tanto no hay ninguna otra dirección que describir (D2, D6, FR-065).
func TestRespuestaDescripcion(t *testing.T) {
	t.Parallel()

	casos := []struct {
		nombre      string
		respuesta   Respuesta
		descripcion string
	}{
		{
			nombre: "método y dirección de la petición",
			respuesta: Respuesta{
				Peticion:  Peticion{Metodo: "GET", URL: "http://fuente.prueba/norma?id=BOE-A-2015-10565"},
				URL:       "http://fuente.prueba/norma?id=BOE-A-2015-10565",
				Estado:    200,
				Cabeceras: Cabeceras{"Content-Type": {"application/xml"}},
				Cuerpo:    []byte("<norma/>"),
			},
			descripcion: "GET http://fuente.prueba/norma?id=BOE-A-2015-10565",
		},
		{
			nombre: "HEAD se describe como HEAD",
			respuesta: Respuesta{
				Peticion: Peticion{Metodo: "HEAD", URL: "http://fuente.prueba/norma"},
				URL:      "http://fuente.prueba/norma",
				Estado:   200,
			},
			descripcion: "HEAD http://fuente.prueba/norma",
		},
		{
			nombre: "tras una redirección describe la pedida, no la final",
			respuesta: Respuesta{
				Peticion: Peticion{Metodo: "GET", URL: "http://fuente.prueba/antigua"},
				URL:      "http://fuente.prueba/norma",
				Estado:   200,
			},
			descripcion: "GET http://fuente.prueba/antigua",
		},
		{
			nombre: "en ensayo, donde no hay estado ni cuerpo",
			respuesta: Respuesta{
				Peticion: Peticion{Metodo: "GET", URL: "http://fuente.prueba/norma"},
				URL:      "http://fuente.prueba/norma",
				Ensayo:   true,
			},
			descripcion: "GET http://fuente.prueba/norma",
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, caso.descripcion, caso.respuesta.Descripcion())
		})
	}
}
