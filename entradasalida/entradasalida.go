package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//================================| MAIN |================================\\

func main() {

	// Configura el logger
	config.Logger("IO.log")

	log.Printf("Soy un logeano")

	// Crear interfaz (TESTING)
	conectarInterfaz("GENERICO", "config.json")
}

// conectarInterfaz se encarga de conectar una interfaz con el servidor.
// Recibe el nombre de la interfaz y la ruta del archivo de configuración.
func conectarInterfaz(nombre string, filePath string) {

	// Extrae info de config.json
	var configInterfaz config.IO

	config.Iniciar(filePath, &configInterfaz)

	//Insertar Nombre, Puerto, Tipo de interfaz
	body, err := json.Marshal(structs.NuevaInterfaz{
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

	// Verificamos si hubo error en la request
	if respuesta == nil {
		return
	}

	// Iniciar el servidor de la interfaz
	iniciarServidorInterfaz(configInterfaz)
}

// iniciarServidorInterfaz inicia el servidor HTTP para la interfaz con la configuración proporcionada.
func iniciarServidorInterfaz(configInterfaz config.IO) {

	// Manejadores para las rutas de la interfaz
	http.HandleFunc("POST /IO_GEN_SLEEP", createHandlerIO_GEN_SLEEP(configInterfaz))
	//http.HandleFunc("POST /IO_STDOUT_WRITE", handlerIO_STDOUT_WRITE)
	//http.HandleFunc("POST /IO_STDIN_READ", handlerIO_STDOUT_WRITE)

	//inicio el servidor de la Interfaz (IO)
	config.IniciarServidor(configInterfaz.Port)
}

// createHandlerIO_GEN_SLEEP crea un manejador HTTP para la ruta "/IO_GEN_SLEEP".
// Recibe la configuración de la interfaz.
func createHandlerIO_GEN_SLEEP(configIO config.IO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Crea una variable para almacenar el número de unidades de trabajo.
		var unidadesDeTrabajo int

		// Decodifica el request (codificado en formato JSON).
		err := json.NewDecoder(r.Body).Decode(&unidadesDeTrabajo)

		// Manejo de error de la decodificación.
		if err != nil {
			fmt.Printf("Error al decodificar request body: ")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Imprime el request por consola (del lado del servidor).
		fmt.Println("Request path:", unidadesDeTrabajo)

		// Ejecuta IO_GEN_SLEEP
		sleepTime := configIO.Unit_Work_Time * unidadesDeTrabajo
		time.Sleep(time.Duration(sleepTime))

		// Responde al cliente con un código de estado HTTP 200 OK.
		w.WriteHeader(http.StatusOK)

	}
}
