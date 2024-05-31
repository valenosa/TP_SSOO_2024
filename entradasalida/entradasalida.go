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

//*======================================| MAIN |======================================\\

func main() {

	// Configura el logger
	config.Logger("IO.log")

	log.Printf("Soy un logeano")

	// Crear interfaz (TESTING)
	conectarInterfazIO("GENERIC_SHIT", "config.json")
}

//*======================================| FUNCIONES |======================================\\

func conectarInterfazIO(nombre string, filePath string) {

	// Extrae info de config.json
	var configNuevaInterfaz config.IO

	config.Iniciar(filePath, &configNuevaInterfaz)

	//Levanto el server de la nuevaInterfazIO
	serevrErr := iniciarServidorInterfaz(configNuevaInterfaz)
	if serevrErr != nil {
		fmt.Printf("Error al iniciar servidor de interfaz: %s", serevrErr.Error())
		return
	}

	//Creo Interfaz base
	var nuevaInterfazIO = structs.Interfaz{TipoInterfaz: configNuevaInterfaz.Type, PuertoInterfaz: configNuevaInterfaz.Port}

	//Creo y codifico la request de coneccion a Kernel
	var requestConectarIO = structs.RequestConectarInterfazIO{NombreInterfaz: nombre, Interfaz: nuevaInterfazIO}
	body, marshalErr := json.Marshal(requestConectarIO)
	if marshalErr != nil {
		fmt.Printf("error codificando body: %s", marshalErr.Error())
		return
	}

	// Si todo es correcto envio la request de coneccion a Kernel
	respuesta := config.Request(configNuevaInterfaz.Port_Kernel, configNuevaInterfaz.Ip_Kernel, "POST", "interfaz", body)
	if respuesta == nil {
		return
	}
}

func iniciarServidorInterfaz(configInterfaz config.IO) error {

	http.HandleFunc("POST /IO_GEN_SLEEP", handlerIO_GEN_SLEEP(configInterfaz))

	var err = config.IniciarServidor(configInterfaz.Port)
	return err
}

//*======================================| HANDLERS |======================================\\

// Implemantación de la Interfaz Génerica

func handlerIO_GEN_SLEEP(configIO config.IO) http.HandlerFunc {
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
		fmt.Println("Unidades de Trabajo:", unidadesDeTrabajo)

		//Ejecuta IO_GEN_SLEEP
		sleepTime := configIO.Unit_Work_Time * unidadesDeTrabajo
		fmt.Println("Zzzzzz...")
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		fmt.Println("Wakey wakey, its time for school")
		// Responde al cliente
		w.WriteHeader(http.StatusOK)
	}
}
