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

	inciarServidorInterfaz((configInterfaz.Port))
}

func inciarServidorInterfaz(puerto int) {
	port := ":" + strconv.Itoa(puerto)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}

// Funcion que genera una interfaz, según el tipo establecido en config.json
func crearInterfaz(configJson config.IO) {
	switch configJson.Type {

	case "GENERICA":
		crearInterfazGenerica(configJson)

	case "STDIN":
		//crearInterfazSTDIN()

	case "STDOUT":
		//crearInterfazSTDOUT()

	}
}

func crearInterfazGenerica(configJson config.IO) {

	time.Sleep(time.Duration(configJson.Unit_Work_Time))
}

// func crearInterfazSTDIN(){}
// func crearInterfazSTDOUT(){}

//TODO: Interfaz Genérica desarrollada.
/*
	 En el config.json está la unidad de trabajo,
	 cuyo valor se va a multiplicar
	 por otro valor dado según el tipo de interfaz que tengamos,

	 Al iniciar una Interfaz de I/O la misma deberá recibir 2 parámetros:
 	 -Nombre (id)
	 -Archivo de Configuración.

	 _INTERFACES GENÉRICAS_
	Ante una petición van a esperar una cantidad de unidades de trabajo,
	cuyo valor va a venir dado en la petición desde el Kernel.

	Las instrucciones que aceptan estas interfaces son:
	IO_GEN_SLEEP

	Al leer el archivo de configuración solo le van a importar las propiedades de:
	type
	unit_work_time
	ip_kernel
	port_kernel

*/
