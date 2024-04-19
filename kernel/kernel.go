package main

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////IMPORTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////STRUCTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type BodyIniciarProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	Path string `json:"path"`
}

type ResponseProceso struct {
	Pid    int    `json:"pid"`
	Estado string `json:"estado"`
}

// Puertos e ip's de memoria y cpu (luego mover a config.json)
var Ip_Memory string = "localhost"
var Port_Memory int = 8002
var Ip_CPU string = "localhost"
var Port_CPU int = 8006

// Estructura cuyo formato concuerda con el del archivo config.json del kernel
type KernelConfig struct {
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////MAIN///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	/*
		read(GET, body) => que hay que hacer
		res
		switch res
	*/

	//finalizar_proceso()
	//detener_planificacion()
	//iniciar_planificacion()
	//listar_proceso()

	config := iniciarConfiguracion("config.json")

	printConfig(*config)

	iniciar_planificacion(*config)

}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////FUNCIONES/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Me gustaría hacer que pueda recibir cualquier struct (como cuando usabamos template<typename T> en C++), pero todavía no estoy seguro de como hacerlo.
// De ser así, podríamos usar una sola función para extraer la info de todos los archivos config.json
func iniciarConfiguracion(filePath string) *KernelConfig {
	//En el tp0 usan punteros y guardan la variable en un archivo "globals".
	// No estoy seguro del motivo, y por ahora no lo veo necesario
	var config *KernelConfig

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

// Utilizado para testear "IniciarConfiguracion()"
func printConfig(config KernelConfig) {

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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////API's//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Solamente esqueleto
func iniciar_proceso(config KernelConfig) {

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(BodyIniciarProceso{
		Path: "string",
	})
	// Error Handler de la codificación
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process", config.Ip_Memory, config.Port_Memory)

	// Se crea una request donde se "efectúa" un PUT hacia url, enviando el Body anteriormente mencionado
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", config.Ip_Memory, config.Port_Memory)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", config.Ip_Memory, config.Port_Memory)
		return
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response ResponseProceso

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return
	}

	// Imprime pid (parámetro de la estructura)
	fmt.Printf("pid: %d\n", response.Pid)
}

// Solamente esqueleto
func finalizar_proceso(config KernelConfig) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process/%d", config.Ip_Memory, config.Port_Memory, pid)

	// Se crea una request donde se "efectúa" un GET hacia la url
	req, err := http.NewRequest("DELETE", url, nil)

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", config.Ip_Memory, config.Port_Memory)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", config.Ip_Memory, config.Port_Memory)
		return
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}
}

func estado_proceso(config KernelConfig) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process/%d", config.Ip_Memory, config.Port_Memory, pid)

	// Se crea una request donde se "efectúa" un GET hacia la url
	req, err := http.NewRequest("DELETE", url, nil)

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", config.Ip_Memory, config.Port_Memory)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", config.Ip_Memory, config.Port_Memory)
		return
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}
}

func iniciar_planificacion(config KernelConfig) {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/plani", config.Ip_CPU, config.Port_CPU)

	// Genera una petición HTTP.
	req, err := http.NewRequest("PUT", url, nil)

	// Check error generando una request.
	if err != nil {
		fmt.Printf("Error creando request: %s\n", err.Error())
		return
	}

	// Se envía la request al servidor.
	respuesta, err := cliente.Do(req)

	// Check request enviada.
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", config.Ip_CPU, config.Port_CPU)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	defer respuesta.Body.Close()

	// Check response recibida.
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Todo salió bien, la planificación se inició correctamente.
	fmt.Println("Planificación iniciada exitosamente.")
}

func detener_planificacion() {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/plani", Ip_CPU, Port_CPU)

	// Genera una petición HTTP.
	req, err := http.NewRequest("DELETE", url, nil)

	// Check error generando una request.
	if err != nil {
		fmt.Printf("Error creando request: %s\n", err.Error())
		return
	}

	// Se envía la request al servidor.
	respuesta, err := cliente.Do(req)

	// Check request enviada.
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", Ip_CPU, Port_CPU)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	defer respuesta.Body.Close()

	// Check response recibida.
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Todo salió bien, la planificación se detuvo correctamente.
	fmt.Println("Planificación detenida exitosamente.")
}

/*
Se encargará de mostrar por consola y retornar por la api el listado de procesos
que se encuentran en el sistema con su respectivo estado dentro de cada uno de ellos.
*/
func listar_proceso(config KernelConfig) {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/process", config.Ip_Memory, config.Port_Memory)

	// Genera una petición HTTP.
	req, err := http.NewRequest("GET", url, nil)

	// Check error generando una request.
	if err != nil {
		fmt.Printf("Error creando request: %s\n", err.Error())
		return
	}

	// Se envía la request al servidor.
	respuesta, err := cliente.Do(req)

	// Check request enviada.
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", config.Ip_Memory, config.Port_Memory)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return
	}

	fmt.Println(string(bodyBytes))
}

func dispatch() {

}

func interrupt() {

}
