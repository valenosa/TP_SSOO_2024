package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	/*
		read(GET, body) => que hay que hacer
		res
		switch res
		iniciar_proceso()
	*/
}

type BodyIniciarProceso struct {
	Path string `json:"path"`
}

type ResponseIniciarProceso struct {
	Pid string `json:"pid"`
}

func inciar_proceso() {

	ip := "localhost"
	puerto := 8080

	// Inicializar valor a enviar
	body, err := json.Marshal(BodyIniciarProceso{
		Path: "",
	})

	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
	}

	cliente := &http.Client{}

	url := fmt.Sprintf("http://%s:%d/paquetes", ip, puerto)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("error creado request a ip:%s puerto:%d", ip, puerto)
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		fmt.Printf("error enviando request a ip:%s puerto:%d", ip, puerto)
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d", respuesta.StatusCode)
	}

	var response ResponseIniciarProceso
	err = json.NewDecoder(respuesta.Body).Decode(&response)
	if err != nil {
		fmt.Printf("Error decodificando")
	}

	fmt.Println(response.Pid)
}

func finalizar_proceso() {
	//implementar
}

func estado_proceso() {
	//implementar
}

func iniciar_planificacion() {
	//implementar
}

func detener_planificacion() {
	//implementar
}

func listar_proceso() {
	//implementar
}

func dispatch() {

}

func interrupt() {

}
