package funciones

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var ConfigJson config.Memoria

// Funciones auxiliares
// Toma de a un archivo a la vez y guarda las instrucciones en un map l
func GuardarInstrucciones(pid uint32, path string, memoriaInstrucciones map[uint32][]string) {
	path = ConfigJson.Instructions_Path + "/" + path
	data := ExtractInstructions(path)
	InsertData(pid, memoriaInstrucciones, data)
}

// Abre el archivo especificado por la ruta 'path' y guarda su contenido en un slice de bytes.
// Retorna el contenido del archivo como un slice de bytes.
func ExtractInstructions(path string) []byte {
	// Lee el archivo
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error al leer el archivo de instrucciones")
		return nil
	}

	// Ahora 'file' es un slice de bytes con el contenido del archivo
	return file
}

// insertData inserta las instrucciones en la memoria de instrucciones asociadas al PID especificado,
// e imprime las instrucciones guardadas en memoria junto con su PID correspondiente.
func InsertData(pid uint32, memoriaInstrucciones map[uint32][]string, data []byte) {

	// Separa las instrucciones por medio de tokens.
	instrucciones := strings.Split(string(data), "\n")

	// Inserta las instrucciones en la memoria de instrucciones.
	memoriaInstrucciones[pid] = instrucciones

	// Imprime las instrucciones guardadas en memoria.
	fmt.Println("Instrucciones guardadas en memoria: ")
	for pid, instrucciones := range memoriaInstrucciones {
		fmt.Printf("PID: %d\n", pid)
		for _, instruccion := range instrucciones {
			fmt.Println(instruccion)
		}
		fmt.Println()
	}
}

func AsignarTabla(pid uint32, tablaDePaginas map[uint32]structs.Tabla) {

	//Genera una tabla de páginas en base a un pid.
	tablaDePaginas[pid] = structs.Tabla{}

}

func BuscarMarco(pid uint32, pagina uint32, tablaDePaginas map[uint32]structs.Tabla) string {
	if len(tablaDePaginas[pid]) <= int(pagina) {
		return ""
	}

	marco := tablaDePaginas[pid][pagina]

	marcoStr := strconv.Itoa(marco)

	return marcoStr
}

func LiberarMarcos(marcosALiberar []int, bitMap []bool) {
	for _, marco := range marcosALiberar {
		bitMap[marco] = false
	}
}

func ReasignarPaginas(pid uint32, tablaDePaginas *map[uint32]structs.Tabla, bitMap []bool, size uint32) string {

	lenOriginal := len((*tablaDePaginas)[pid]) //!

	cantidadDePaginas := int(math.Ceil(float64(size) / float64(ConfigJson.Page_Size)))

	//*CASO AGREGAR PAGINAS
	//?Hace falta devolver algo?
	// Itera n cantidad de veces, siendo n la cantidad de paginas a agregar
	//? Funcionan los punteros así?
	for len((*tablaDePaginas)[pid]) < cantidadDePaginas {

		// Por cada página a agregar, si no hay marcos disponibles, se devuelve un error OUT OF MEMORY
		outOfMemory := true

		// Recorre el bitMap buscando un marco desocupado
		for marco, ocupado := range bitMap {
			//?optimizar? (no se si es necesario recorrer todo el bitMap)

			if !ocupado {
				// Guarda en la tabla de páginas del proceso el marco asignado a una página
				(*tablaDePaginas)[pid] = append((*tablaDePaginas)[pid], marco)
				// Marca el marco como ocupado
				bitMap[marco] = true

				// Notifica que por ahora no está OUT OF MEMORY
				outOfMemory = false
			}
		}

		//Si no hubo ningun marco desocupado para la página anterior, devuelve OUT OF MEMORY
		if outOfMemory {
			return "OUT OF MEMORY" //?
			//!OUT OF MEMORY
		}
	}

	//*CASO QUITAR PAGINAS
	//?Hace falta devolver algo?
	if len((*tablaDePaginas)[pid]) > cantidadDePaginas {

		marcosALiberar := (*tablaDePaginas)[pid][cantidadDePaginas:]

		(*tablaDePaginas)[pid] = (*tablaDePaginas)[pid][:cantidadDePaginas]

		LiberarMarcos(marcosALiberar, bitMap)
	}

	fmt.Printf("Se pasó de %d a %d páginas\n", lenOriginal, len((*tablaDePaginas)[pid]))

	return "OK" //?
}

// TODO: Probar
// ! CAMBIAR, PASARLE TAMBIEN UNA TABLA DE PAGINA
func LeerEnMemoria(direccionFisica uint64, tamanioRegistro uint64, espacioUsuario []byte) []byte {

	// Luego, lee el dato desde la dirección física.
	dato := (espacioUsuario)[direccionFisica : direccionFisica+tamanioRegistro]

	// Devuelve el dato como una cadena.
	return dato

}

// TODO: Probar
// ! CAMBIAR, PASARLE TAMBIEN UNA TABLA DE PAGINA
func EscribirEnMemoria(direccionFisica uint64, dato []byte, espacioUsuario *[]byte, bitMap *[]bool) {
	// Verifica si la dirección física está dentro de los límites del espacio de usuario
	if direccionFisica+uint64(len(dato)) > uint64(len(*espacioUsuario)) {
		return
	}
	// Escribe los datos en el espacio de usuario
	copy((*espacioUsuario)[direccionFisica:], dato)

	// Calcula el índice del marco afectado
	marco := direccionFisica / uint64(ConfigJson.Page_Size)

	// Marca el marco como lleno en el bitMap
	(*bitMap)[marco] = true
}
