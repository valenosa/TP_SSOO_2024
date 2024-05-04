package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// --------------------------| ESTRUCTURAS PARA EXTRAER INFO DEL config.json |-------------------------------------------------------------
type Cpu struct {
	Port               int    `json:"port"`
	Ip_Memory          string `json:"ip_memory"`
	Port_Memory        int    `json:"port_memory"`
	Number_Felling_tlb int    `json:"number_felling_tlb"`
	Algorithm_tlb      string `json:"algorithm_tlb"`
}

type IO struct {
	Port               int    `json:"port"`
	Type               string `json:"type"`
	Unit_Work_Time     int    `json:"unit_work_time"`
	Ip_Kernel          string `json:"ip_kernel"`
	Port_Kernel        int    `json:"port_kernel"`
	Ip_Memory          string `json:"ip_memory"`
	Port_Memory        int    `json:"port_memory"`
	Dialfs_Path        string `json:"dialfs_path"`
	Dialfs_Block_Size  int    `json:"dialfs_block_size"`
	Dialfs_Block_Count int    `json:"dialfs_block_count"`
}

type Kernel struct {
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

type Memoria struct {
	Port              int    `json:"port"`
	Memory_Size       int    `json:"memory_size"`
	Page_Size         int    `json:"page_size"`
	Instructions_Path string `json:"instructions_path"`
	Delay_Response    int    `json:"delay_response"`
}

// --------------------------| FUNCIONES PARA EXTRAER INFO DEL config.json |-------------------------------------------------------------

// Se implementó el uso de interface{} a la función. De esta manera, la misma puede recibir distintos tipos de datos, o en este caso, estructuras (polimorfismo).
// Gracias a esta implementación, luego la función podrá ser trasladada a un paquete aparte y ser utilizada por todos los módulos.
func Decode(filePath string, configJson interface{}) error {
	// Abre el archivo
	configFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	// Cierra el archivo una vez que la función termina (ejecuta el return)
	defer configFile.Close()

	// Decodifica la info del json en la variable config
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(configJson)
	if err != nil {
		return err
	}

	return nil
}

// Decodifica la info del json en la variable configJson y maneja el error en caso de haberlo)
func Iniciar(filePath string, configJson interface{}) {
	err := Decode(filePath, &configJson)

	if err != nil {
		fmt.Println("Error al iniciar configuración: ", err)
	}
}

func Logger(path string) {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

// --------------------------| FUNCIONES PARA TESTEAR |----------------------------------------------------------------

// Utilizado para testear "IniciarConfiguracion()"
func printConfig(configJson Kernel) {

	fmt.Println("port: ", configJson.Port)
	fmt.Println("ip_memory: ", configJson.Ip_Memory)
	fmt.Println("port_memory: ", configJson.Port_Memory)
	fmt.Println("ip_cpu: ", configJson.Ip_CPU)
	fmt.Println("port_cpu: ", configJson.Port_CPU)
	fmt.Println("planning_algorithm: ", configJson.Planning_Algorithm)
	fmt.Println("quantum: ", configJson.Quantum)
	fmt.Println("resources: ", configJson.Resources)
	fmt.Println("resource_instances: ", configJson.Resource_Instances)
	fmt.Println("multiprogramming: ", configJson.Multiprogramming)
}

//------------- ESTO NO VA ACA PERO ES GLOBAL Y LO USAN TODOS LOS MODULOS ------------------------------------------

// retorna true si la request fue exitosa, false en caso contrario
func Request(port int, ip string, metodo string, query string, bodies ...[]byte) *http.Response {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, query)

	body := ifBody(bodies...)

	// Se crea una request donde se "efectúa" el metodo (PUT / DELETE / GET / POST) hacia url, enviando el Body si lo hay
	req, err := http.NewRequest(metodo, url, body)

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", ip, port)
		return nil
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", ip, port)
		return nil
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return nil
	}

	//Todo salió bien
	fmt.Printf("%s %s exitoso \n", metodo, query)
	return respuesta
}

func ifBody(bodies ...[]byte) io.Reader {
	if len(bodies) == 0 {
		return nil
	}
	return bytes.NewBuffer(bodies[0])
}
