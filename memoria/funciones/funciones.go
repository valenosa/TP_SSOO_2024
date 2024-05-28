package funciones

import (
	"fmt"
	"os"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

var configJson config.Memoria

// Toma de a un archio a la vez y guarda las instrucciones en un map l
func GuardarInstrucciones(pid uint32, path string, memoriaInstrucciones map[uint32][]string) {
	path = configJson.Instructions_Path + "/" + path
	data := extractInstructions(path)
	insertData(pid, memoriaInstrucciones, data)
}

// Abre el arhivo y guarda su contenido en un arrayfunc extractInstructions(path string) []byte
func extractInstructions(path string) []byte {
	// Lee el archivo
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error al leer el archivo de instrucciones")
		return nil
	}

	// Ahora data es un array de bytes con el contenido del archivo
	return file
}

// funciona todo bien con uint?
func insertData(pid uint32, memoriaInstrucciones map[uint32][]string, data []byte) {
	// Separar instrucciones por medio de tokens
	instrucciones := strings.Split(string(data), "\n")
	// Inserta en la memoria de instrucciones
	memoriaInstrucciones[pid] = instrucciones
	fmt.Println("Instrucciones guardadas en memoria: ")
	for pid, instrucciones := range memoriaInstrucciones {
		fmt.Printf("PID: %d\n", pid)
		for _, instruccion := range instrucciones {
			fmt.Println(instruccion)
		}
		fmt.Println()
	}
}
