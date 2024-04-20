package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type inputOutpConfig struct {
	Port               int      `json:"port"`
	Ip_Memory          string   `json:"ip_memory"`
	Port_Memory        int      `json:"port_memory"`
	Ip_CPU             string   `json:"ip_cpu"`
	Port_CPU           int      `json:"port_cpu"`
	Planning_Algorithm string   `json:"planning_algorithm"`
	Quantum            int      `json:"quantum"`
	Resources          []string `json:"resources"`          // Está bien el tipo de dato?
	Resource_Instances []int    `json:"resource_instances"` // Está bien el tipo de dato?
	Multiprogramming   int      `json:"multiprogramming"`
}

func main() {

	config := iniciarConfiguracion("config.json")
	printConfig(*config)

	http.HandleFunc("GET", entradaSalida)

	port := ":" + strconv.Itoa(config.Port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
	/*if err == nil {
		fmt.Println("Estoy escuchando en el puerto " + port)
	} else {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
	PARA LOS LOGS*/
}

func iniciarConfiguracion(filePath string) *inputOutpConfig {
	//En el tp0 usan punteros y guardan la variable en un archivo "globals".
	// No estoy seguro del motivo, y por ahora no lo veo necesario
	var config *inputOutpConfig

	// Abre el archivo
	configFile, err := os.Open(filePath)
	if err != nil {
		// log.Fatal(err.Error())
		fmt.Println("Error: ", err)
	}
	// Cierra el archivo una vez que la función termina (ejecuta el return)
	defer configFile.Close()

	// Decodifica la info del json en la variable config
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	// Devuelve config
	return config
}

func printConfig(config inputOutpConfig) {

	fmt.Println("port: ", config.Port)
	fmt.Println("ip_memory: ", config.Ip_Memory)
	fmt.Println("port_memory: ", config.Port_Memory)
	fmt.Println("ip_cpu: ", config.Ip_CPU)
	fmt.Println("port_cpu: ", config.Port_CPU)
	fmt.Println("planning_algorithm: ", config.Planning_Algorithm)
	fmt.Println("quantum: ", config.Quantum)
	fmt.Println("resources: ", config.Resources)
	fmt.Println("resource_instances: ", config.Resource_Instances)
	fmt.Println("multiprogramming: ", config.Multiprogramming)
}

func entradaSalida(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello world! Soy una consola de I/O")
}
