package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

// MOVELO A UTILS (struct tambien usada por kernel.go)
type Interfaz struct {
	Nombre string
	Tipo   string
	Puerto int
}

func main() {

	// Configura el logger
	config.Logger("IO.log")

	log.Printf("Soy un logeano")

	// Crear interfaz (TESTING)
	conectarInterfaz("GENERICO", "config.json")
}

func conectarInterfaz(nombre string, filePath string) {

	// Extrae info de config.json
	var configInterfaz config.IO

	config.Iniciar(filePath, &configInterfaz)

	//Insertar Nombre, Puerto, Tipo de interfaz
	body, err := json.Marshal(Interfaz{
		Nombre: nombre,
		Tipo:   configInterfaz.Type,
		Puerto: configInterfaz.Port,
	})

	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Enviar request al servidor
	respuesta := config.Request(configInterfaz.Port_Kernel, configInterfaz.Ip_Kernel, "POST", "interfaz", body)

	// verificamos si hubo error en la request
	if respuesta == nil {
		return
	}

	iniciarServidorInterfaz(configInterfaz)
}

func iniciarServidorInterfaz(configInterfaz config.IO) {

	http.HandleFunc("POST /IO_GEN_SLEEP", createHandlerIO_GEN_SLEEP(configInterfaz))
	//http.HandleFunc("POST /IO_STDOUT_WRITE", handlerIO_STDOUT_WRITE)
	//http.HandleFunc("POST /IO_STDIN_READ", handlerIO_STDOUT_WRITE)

	port := ":" + strconv.Itoa(configInterfaz.Port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}

// Implemantación de la Interfaz Génerica

func createHandlerIO_GEN_SLEEP(configIO config.IO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//Crea una variable tipo Interfaz (para interpretar lo que se recibe de la unidadesDeTrabajo)
		var unidadesDeTrabajo int

		// Decodifica el request (codificado en formato json)
		err := json.NewDecoder(r.Body).Decode(&unidadesDeTrabajo)

		// Error de la decodificación
		if err != nil {
			fmt.Printf("Error al decodificar request body: ")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Imprime el request por consola (del lado del server)
		fmt.Println("Request path:", unidadesDeTrabajo)

		//Ejecuta IO_GEN_SLEEP
		sleepTime := configIO.Unit_Work_Time * unidadesDeTrabajo
		time.Sleep(time.Duration(sleepTime))

		// Responde al cliente
		w.WriteHeader(http.StatusOK)

	}
}
