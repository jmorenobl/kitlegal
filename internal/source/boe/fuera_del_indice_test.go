package boe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmorenobl/kitlegal/internal/cli"
	"github.com/jmorenobl/kitlegal/internal/core"
	"github.com/jmorenobl/kitlegal/internal/core/schema"
)

// La Ley 9/2017, cuyo «Artículo 118» es el bloque a1-30 de su índice grabado, y
// el id que se compone con su número, que el índice no tiene (ADR 0049).
const (
	normaDeLaLCSP        = "BOE-A-2017-12902"
	bloqueDelArticulo118 = "a1-30"
	bloqueCompuesto      = "a118"
)

// TestBloqueFueraDelIndice comprueba que la lectura fallida de un bloque que el
// índice guardado de su norma no tiene es «no encontrado» y nombra el bloque del
// artículo, sin pedir nada más (ADR 0049); y que sin el índice guardado, con el id
// en él o con un fallo que no es de la fuente, el fallo es el de siempre.
func TestBloqueFueraDelIndice(t *testing.T) {
	t.Parallel()

	direccionCompuesta := direccionDelBloque(normaDeLaLCSP, bloqueCompuesto)
	direccionDelIndiceDeLaLCSP := direccionDelIndice(normaDeLaLCSP)
	conElArticulo := "la norma BOE-A-2017-12902 no tiene el bloque a118 en su índice: el «Artículo 118» es el bloque " +
		"a1-30 (" + direccionCompuesta + ")"
	fecha := "2019-06-01"

	casos := []struct {
		nombre string
		// pedidor responde a la lectura del bloque; el índice y lo demás salen de
		// las grabaciones.
		pedidor func(*testing.T, *relojDePrueba) *pedidorDePrueba
		// conIndice dice si el índice de la norma se lee antes, y queda guardado.
		conIndice bool
		ec        schema.Contexto
		consulta  core.Consulta
		clase     schema.Clase
		// mensaje es el esperado; vacío, no se comprueba.
		mensaje string
	}{
		{
			nombre:    "limite-con-el-indice-guardado",
			pedidor:   salvoEn(direccionCompuesta, fallaLaLectura(schema.ClaseLimiteOTos)),
			conIndice: true,
			consulta:  ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: bloqueCompuesto},
			clase:     schema.ClaseNoEncontrado,
			mensaje:   conElArticulo,
		},
		{
			nombre:    "fuente-caida-con-el-indice-guardado",
			pedidor:   salvoEn(direccionCompuesta, fallaLaLectura(schema.ClaseFuenteNoDisponible)),
			conIndice: true,
			consulta:  ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: bloqueCompuesto},
			clase:     schema.ClaseNoEncontrado,
			mensaje:   conElArticulo,
		},
		{
			nombre:    "404-con-el-indice-guardado",
			pedidor:   salvoEn(direccionCompuesta, respondeAlBloqueCon(404)),
			conIndice: true,
			consulta:  ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: bloqueCompuesto},
			clase:     schema.ClaseNoEncontrado,
			mensaje:   conElArticulo,
		},
		{
			nombre:    "con-fecha",
			pedidor:   salvoEn(direccionCompuesta, fallaLaLectura(schema.ClaseLimiteOTos)),
			conIndice: true,
			consulta:  ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: bloqueCompuesto, Fecha: &fecha},
			clase:     schema.ClaseNoEncontrado,
			mensaje:   conElArticulo,
		},
		{
			nombre:    "articulos",
			pedidor:   salvoEn(direccionCompuesta, fallaLaLectura(schema.ClaseLimiteOTos)),
			conIndice: true,
			consulta:  ConsultaArticulos{Norma: normaDeLaLCSP, Bloques: []string{bloqueDelArticulo118, bloqueCompuesto}},
			clase:     schema.ClaseNoEncontrado,
			mensaje:   "no se ha podido resolver el bloque a118, en la posición 2 de 2: " + conElArticulo,
		},
		{
			nombre:    "offline-con-el-indice-guardado",
			pedidor:   reproduce(carpetaDeLasGrabaciones),
			conIndice: true,
			ec:        schema.Contexto{Offline: true},
			consulta:  ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: bloqueCompuesto},
			clase:     schema.ClaseNoEncontrado,
			mensaje:   conElArticulo,
		},
		{
			nombre:    "id-sin-la-forma-de-un-articulo",
			pedidor:   salvoEn(direccionDelBloque(normaDeLaLCSP, "zz9"), fallaLaLectura(schema.ClaseLimiteOTos)),
			conIndice: true,
			consulta:  ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: "zz9"},
			clase:     schema.ClaseNoEncontrado,
			mensaje: "la norma BOE-A-2017-12902 no tiene el bloque zz9 en su índice: el id de un bloque es el campo id " +
				"de kitlegal boe indice BOE-A-2017-12902, no el número del artículo (" +
				direccionDelBloque(normaDeLaLCSP, "zz9") + ")",
		},
		{
			nombre:   "sin-el-indice-guardado",
			pedidor:  salvoEn(direccionCompuesta, fallaLaLectura(schema.ClaseLimiteOTos)),
			consulta: ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: bloqueCompuesto},
			clase:    schema.ClaseLimiteOTos,
		},
		{
			nombre:    "con-el-id-en-el-indice",
			pedidor:   salvoEn(direccionDelBloque(normaDeLaLCSP, bloqueDelArticulo118), fallaLaLectura(schema.ClaseLimiteOTos)),
			conIndice: true,
			consulta:  ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: bloqueDelArticulo118},
			clase:     schema.ClaseLimiteOTos,
		},
		{
			nombre:    "fallo-inesperado",
			pedidor:   salvoEn(direccionCompuesta, fallaLaLectura(schema.ClaseInesperado)),
			conIndice: true,
			consulta:  ConsultaArticulo{Norma: normaDeLaLCSP, Bloque: bloqueCompuesto},
			clase:     schema.ClaseInesperado,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()

			banco := nuevoBanco(t, caso.pedidor)

			if caso.conIndice {
				_, err := banco.resuelve(t, schema.Contexto{}, ConsultaIndice{Norma: normaDeLaLCSP})
				require.NoError(t, err)
			}

			antes := len(banco.pedidor.peticiones)

			_, err := banco.resuelve(t, caso.ec, caso.consulta)

			require.Error(t, err)
			assert.Equal(t, caso.clase, cli.Clasificar(err))

			if caso.mensaje != "" {
				assert.Equal(t, caso.mensaje, err.Error())
			}

			for _, peticion := range banco.pedidor.peticiones[antes:] {
				assert.NotEqual(t, direccionDelIndiceDeLaLCSP, peticion.URL, "el índice no se vuelve a pedir")
			}
		})
	}
}

// TestBloquesDelArticulo comprueba qué entradas del índice son las del artículo
// cuyo número lleva el id: el título empieza por el del artículo y lo que sigue no
// continúa ni el número ni el sufijo.
func TestBloquesDelArticulo(t *testing.T) {
	t.Parallel()

	indice := Indice{Bloques: []EntradaDeIndice{
		{ID: "a1-30", Titulo: "Artículo 118"},
		{ID: "a1-31", Titulo: "Artículo 118 bis"},
		{ID: "a1-32", Titulo: "Artículo 1180"},
		{ID: "a1-33", Titulo: "Artículo 85 bis. Del contrato"},
		{ID: "a1-34", Titulo: " Artículo 7. Definiciones "},
		{ID: "a1-35", Titulo: "Artículo 7 (derogado)"},
		{ID: "a1-36", Titulo: "Artículo único"},
	}}

	casos := []struct {
		bloque string
		titulo string
		ids    []string
	}{
		{bloque: "a118", titulo: "Artículo 118", ids: []string{"a1-30"}},
		{bloque: "a118bis", titulo: "Artículo 118 bis", ids: []string{"a1-31"}},
		{bloque: "a1180", titulo: "Artículo 1180", ids: []string{"a1-32"}},
		{bloque: "a85bis", titulo: "Artículo 85 bis", ids: []string{"a1-33"}},
		{bloque: "a7", titulo: "Artículo 7", ids: []string{"a1-34", "a1-35"}},
		{bloque: "a8", titulo: "Artículo 8"},
		{bloque: "da3"},
		{bloque: "a1-30"},
		{bloque: "A118"},
	}

	for _, caso := range casos {
		t.Run(caso.bloque, func(t *testing.T) {
			t.Parallel()

			titulo, ids := bloquesDelArticulo(indice, caso.bloque)

			assert.Equal(t, caso.titulo, titulo)
			assert.Equal(t, caso.ids, ids)
		})
	}
}

// fallaLaLectura es el Pedidor de prueba que falla con un error de la clase.
func fallaLaLectura(clase schema.Clase) func(*testing.T, *relojDePrueba) *pedidorDePrueba {
	return func(*testing.T, *relojDePrueba) *pedidorDePrueba {
		return &pedidorDePrueba{responde: fallaCon(errorConClaseDePrueba{clase: clase})}
	}
}

// respondeAlBloqueCon es el Pedidor de prueba que responde con el estado y sin
// cuerpo, con el instante del reloj.
func respondeAlBloqueCon(estado int) func(*testing.T, *relojDePrueba) *pedidorDePrueba {
	return func(_ *testing.T, reloj *relojDePrueba) *pedidorDePrueba {
		return &pedidorDePrueba{responde: respondeCon(estado, nil, reloj.ahora())}
	}
}
