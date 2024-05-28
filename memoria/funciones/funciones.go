package funciones

import (
	"fmt"
	"os"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

var configJson config.Memoria

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

// insertData inserta las instrucciones en la memoria de instrucciones asociadas al PID especificado
// e imprime las instrucciones guardadas en memoria junto con su PID correspondiente.
func InsertData(pid uint32, memoriaInstrucciones map[uint32][]string, data []byte) {
	// Separar las instrucciones por medio de tokens
	instrucciones := strings.Split(string(data), "\n")
	// Insertar las instrucciones en la memoria de instrucciones
	memoriaInstrucciones[pid] = instrucciones
	// Imprimir las instrucciones guardadas en memoria
	fmt.Println("Instrucciones guardadas en memoria: ")
	for pid, instrucciones := range memoriaInstrucciones {
		fmt.Printf("PID: %d\n", pid)
		for _, instruccion := range instrucciones {
			fmt.Println(instruccion)
		}
		fmt.Println()
	}
}
