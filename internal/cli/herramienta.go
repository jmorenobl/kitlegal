package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"reflect"
	"slices"
	"strconv"

	"github.com/invopop/jsonschema"
)

// EsquemasDeHerramienta da los dos esquemas con los que el servidor MCP anuncia
// un verbo como herramienta: el de lo que una llamada puede dar y el del sobre
// que devuelve. Salen del mismo generador que el documento de --describe y son
// sus dos partes, con tres diferencias que pide quien los recibe
// (contracts/servidor-mcp.md §2; research.md D6 de H21):
//
//   - la entrada no lleva las ocho banderas globales: son del servidor, que las
//     recibe una vez y las aplica a todas las llamadas, de modo que el nombre de
//     una de ellas es en una llamada una propiedad de más (FR-020);
//   - cada esquema es un documento que se vale solo, con las definiciones que
//     referencia en sus propios `$defs`, porque el protocolo los entrega por
//     separado y una referencia a la raíz de otro documento no resolvería;
//   - no llevan `$schema` y van en JSON compacto: viajan dentro de un mensaje.
//
// Un verbo cuyo documento de --describe no se puede construir tampoco tiene
// esquemas de herramienta, y por el mismo fallo: no se entrega uno a medias.
func EsquemasDeHerramienta(def Verbo) (entrada, salida []byte, err error) {
	campos, err := camposDeLosArgumentos(def.Argumentos)
	if err != nil {
		return nil, nil, err
	}

	g := nuevoGenerador()

	esquemaDeEntrada := g.entrada(campos)
	esquemaDeEntrada.Definitions = g.retirarDefiniciones()

	esquemaDeSalida, err := g.salida(def.Salida)
	if err != nil {
		return nil, nil, err
	}

	esquemaDeSalida.Definitions = g.retirarDefiniciones()

	// Como en --describe, después de reflejar las dos partes: el generador es
	// uno solo y vigila los nombres de las dos, de modo que dos tipos distintos
	// que se llamarían igual no dan dos esquemas que parecen hablar de lo mismo.
	if g.colision != nil {
		return nil, nil, g.colision
	}

	textoDeEntrada, err := codificar(esquemaDeEntrada, "")
	if err != nil {
		return nil, nil, err
	}

	textoDeSalida, err := codificar(esquemaDeSalida, "")
	if err != nil {
		return nil, nil, err
	}

	return []byte(textoDeEntrada), []byte(textoDeSalida), nil
}

// retirarDefiniciones entrega las definiciones reunidas hasta ahora y deja el
// generador sin ninguna, para que la parte que se refleje después reúna las
// suyas. Lo que una parte ha reunido es exactamente lo que referencia: la
// reflexión de cada tipo trae las definiciones que alcanza y ninguna más.
//
// Sin definiciones devuelve nulo y no un mapa vacío, que es lo que hace que un
// esquema que no referencia nada no lleve `$defs`.
func (g *generador) retirarDefiniciones() jsonschema.Definitions {
	reunidas := g.definiciones
	g.definiciones = jsonschema.Definitions{}

	if len(reunidas) == 0 {
		return nil
	}

	return reunidas
}

// Las dos marcas de una línea de órdenes que escribe la conversión: la que
// precede al nombre de una bandera y el terminador, que separa las banderas de lo
// que solo puede ser un argumento de posición (research.md V24 de H21).
const (
	prefijoDeBandera     = "--"
	terminadorDeBanderas = "--"
)

// LineaDeLlamada convierte los argumentos de una llamada a la herramienta de un
// verbo —el objeto JSON de `arguments`— en la línea de órdenes de ese verbo:
// cada campo que no es de posición como `--<nombre>=<valor>`, el terminador `--`
// y, detrás, los de posición en el orden de sus campos, que es el de la orden y
// no el del objeto. Detrás del terminador nada es una bandera, así que ningún
// argumento de una llamada puede nombrar una del servidor (FR-020).
//
// Pasa por los mismos campos que el esquema de entrada y exige de cada valor el
// tipo que ese esquema declara, de modo que no hay dos descripciones de lo que
// una llamada puede dar (research.md D5 de H21). Rechaza, con la clase
// `argumentos` y los mensajes de contracts/servidor-mcp.md §3:
//
//   - unos argumentos que no son un objeto JSON (sin `arguments` son `{}`);
//   - una propiedad que el verbo no declara, también el nombre de una bandera
//     global; de varias, la primera por orden alfabético, para que el mismo
//     rechazo dé siempre el mismo sobre;
//   - un valor que no es del tipo de su campo;
//   - un argumento de posición sin el que le precede, que no tiene orden
//     equivalente.
//
// Que falte un argumento obligatorio no lo dice esta función: lo dice después el
// analizador de la orden, con el mensaje de la orden.
func LineaDeLlamada(def Verbo, argumentos []byte) ([]string, error) {
	campos, err := camposDeLosArgumentos(def.Argumentos)
	if err != nil {
		return nil, err
	}

	herramienta := def.Applet + "_" + def.Verbo

	propiedades, err := propiedadesDeLaLlamada(herramienta, argumentos)
	if err != nil {
		return nil, err
	}

	if err := sinPropiedadesDeMas(herramienta, campos, propiedades); err != nil {
		return nil, err
	}

	var (
		banderas     []string
		posicionales []string
		// ausente es el argumento de posición inmediatamente anterior, si no se
		// ha dado: el que impide escribir el siguiente en su sitio.
		ausente string
	)

	g := nuevoGenerador()

	for _, campo := range campos {
		nombre := nombreEnLaInvocacion(campo)
		_, dePosicion := campo.Tag.Lookup("arg")

		valor, dado := propiedades[nombre]
		if !dado {
			if dePosicion {
				ausente = nombre
			}

			continue
		}

		textos, err := g.textosDelArgumento(herramienta, campo, valor)
		if err != nil {
			return nil, err
		}

		if !dePosicion {
			for _, texto := range textos {
				banderas = append(banderas, prefijoDeBandera+nombre+"="+texto)
			}

			continue
		}

		if ausente != "" {
			return nil, fmt.Errorf("%w: el argumento %q de %s no se puede dar sin %q",
				ErrArgumentos, nombre, herramienta, ausente)
		}

		posicionales = append(posicionales, textos...)
	}

	return slices.Concat(banderas, []string{terminadorDeBanderas}, posicionales), nil
}

// propiedadesDeLaLlamada lee los argumentos de una llamada, que tienen que ser un
// único objeto JSON. Una llamada sin `arguments` es una llamada sin ninguno. Los
// números se conservan como se escribieron: pasarlos por un número de coma
// flotante cambiaría las cifras de un entero grande antes de que nadie lo lea.
func propiedadesDeLaLlamada(herramienta string, argumentos []byte) (map[string]any, error) {
	if len(argumentos) == 0 {
		return map[string]any{}, nil
	}

	var valor any

	decodificador := json.NewDecoder(bytes.NewReader(argumentos))
	decodificador.UseNumber()

	// json.Valid es lo que exige que sea un solo documento: el decodificador lee
	// el primero y no mira lo que haya detrás.
	if json.Valid(argumentos) && decodificador.Decode(&valor) == nil {
		if propiedades, esObjeto := valor.(map[string]any); esObjeto {
			return propiedades, nil
		}
	}

	return nil, fmt.Errorf("%w: los argumentos de %s no son un objeto JSON", ErrArgumentos, herramienta)
}

// sinPropiedadesDeMas rechaza la llamada que da una propiedad que no es ningún
// campo del verbo. El esquema de entrada es cerrado y esta es la comprobación que
// lo hace cierto para quien llama sin validar contra él.
func sinPropiedadesDeMas(
	herramienta string, campos []reflect.StructField, propiedades map[string]any,
) error {
	declarados := make(map[string]bool, len(campos))
	for _, campo := range campos {
		declarados[nombreEnLaInvocacion(campo)] = true
	}

	// En orden alfabético y no en el del mapa, que cambia de una vez a otra: de
	// varias que sobren se nombra siempre la misma.
	for _, nombre := range slices.Sorted(maps.Keys(propiedades)) {
		if !declarados[nombre] {
			return fmt.Errorf("%w: %s no tiene el argumento %q", ErrArgumentos, herramienta, nombre)
		}
	}

	return nil
}

// Los tipos del vocabulario de JSON Schema con los que el generador describe lo
// que una llamada puede escribir en una línea de órdenes.
const (
	tipoDeCadena   = "string"
	tipoDeBooleano = "boolean"
	tipoDeEntero   = "integer"
	tipoDeLista    = "array"
)

// escritura es cómo llega a la línea de órdenes un valor de uno de esos tipos:
// cómo se nombra en el mensaje de quien da otra cosa —solo y en una lista— y
// cómo se lee del valor JSON el texto que se escribe.
type escritura struct {
	nombre        string
	nombreDeLista string
	texto         func(valor any) (string, bool)
}

// escrituras son los tipos que una llamada sabe escribir (research.md D5 de H21):
// una cadena, un booleano y un número entero, y las listas de ellos.
func escrituras() map[string]escritura {
	return map[string]escritura{
		tipoDeCadena: {
			nombre: "una cadena", nombreDeLista: "una lista de cadenas", texto: textoDeCadena,
		},
		tipoDeBooleano: {
			nombre: "un booleano", nombreDeLista: "una lista de booleanos", texto: textoDeBooleano,
		},
		tipoDeEntero: {
			nombre: "un número entero", nombreDeLista: "una lista de números enteros", texto: textoDeEntero,
		},
	}
}

// textosDelArgumento convierte el valor que una llamada da a un campo en los
// textos con los que se escribe en la línea de órdenes: uno, o uno por elemento
// si el campo es una lista.
//
// El tipo que se exige no se deduce del tipo Go del campo sino del esquema con
// el que el generador lo describe, que es el que ve quien llama: una duración,
// por ejemplo, es un entero en Go y una cadena en la línea de órdenes. Un campo
// cuyo esquema no es de los que una llamada sabe escribir no es un error de
// quien llama, sino de quien declaró el verbo.
func (g *generador) textosDelArgumento(
	herramienta string, campo reflect.StructField, valor any,
) ([]string, error) {
	esquema := g.reflejar(campo.Type)

	tipo, lista := esquema.Type, false
	if tipo == tipoDeLista && esquema.Items != nil {
		tipo, lista = esquema.Items.Type, true
	}

	forma, sabida := escrituras()[tipo]
	if !sabida {
		return nil, fmt.Errorf("%w: una llamada no sabe escribir el campo %s, de tipo %s",
			errEsquemaImposible, campo.Name, campo.Type)
	}

	leer, esperado := forma.deUnValor, forma.nombre
	if lista {
		leer, esperado = forma.deUnaLista, forma.nombreDeLista
	}

	textos, valido := leer(valor)
	if !valido {
		return nil, fmt.Errorf("%w: el argumento %q de %s tiene que ser %s",
			ErrArgumentos, nombreEnLaInvocacion(campo), herramienta, esperado)
	}

	return textos, nil
}

// deUnValor lee un valor JSON de ese tipo.
func (e escritura) deUnValor(valor any) ([]string, bool) {
	texto, valido := e.texto(valor)
	if !valido {
		return nil, false
	}

	return []string{texto}, true
}

// deUnaLista lee una lista JSON cuyos elementos son todos de ese tipo.
func (e escritura) deUnaLista(valor any) ([]string, bool) {
	elementos, esLista := valor.([]any)
	if !esLista {
		return nil, false
	}

	textos := make([]string, 0, len(elementos))

	for _, elemento := range elementos {
		texto, valido := e.texto(elemento)
		if !valido {
			return nil, false
		}

		textos = append(textos, texto)
	}

	return textos, true
}

func textoDeCadena(valor any) (string, bool) {
	cadena, esCadena := valor.(string)

	return cadena, esCadena
}

func textoDeBooleano(valor any) (string, bool) {
	booleano, esBooleano := valor.(bool)

	return strconv.FormatBool(booleano), esBooleano
}

// textoDeEntero lee un número JSON de valor entero y lo escribe con todas sus
// cifras y sin nada más, que es como lo lee el analizador de la orden. Para el
// esquema es entero todo número sin parte fraccionaria, se escriba como se
// escriba —`3`, `3.0`, `3e0`—, y por eso se mira el valor y no la forma.
func textoDeEntero(valor any) (string, bool) {
	numero, esNumero := valor.(json.Number)
	if !esNumero {
		return "", false
	}

	racional, legible := new(big.Rat).SetString(numero.String())
	if !legible || !racional.IsInt() {
		return "", false
	}

	return racional.Num().String(), true
}
