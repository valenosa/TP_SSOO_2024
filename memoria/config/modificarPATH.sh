#!/bin/bash

# Códigos de color
ROJO='\033[0;31m'
VERDE='\033[0;32m'
MARRON='\033[0;33m'
AZUL='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
AMARILLO='\033[1;33m'
NEGRO='\033[0;30m'
GRIS_OSCURO='\033[1;30m'
GRIS_CLARO='\033[0;37m'
ROJO_CLARO='\033[1;31m'
VERDE_CLARO='\033[1;32m'
AZUL_CLARO='\033[1;34m'
PURPURA_CLARO='\033[1;35m'
CIAN_CLARO='\033[1;36m'
BLANCO='\033[1;37m'
NC='\033[0m' # No Color

#Las variables de entorno creadas son locales de la ejecución del script actual. Una vez que se deja de ejecutar, muere.
# si ejecutas source ./modificarIP.sh las variables de entorno son creadas localmente en la terminal, por lo que podes matar el script, y al levantarlo nuevamente siguen estando. Mueren cuando ripea la terminal.

# Los configs vienen pre-seteados en localhost y ya no usan el "caso default" que setea a localhost si la variable estaba vacía (se puede cambiar).
# Se puede escribir localhost o l 

modificar() {
    local var_name=$2
    local current_value=$(eval echo \$$var_name)
    echo -e "${GRIS_OSCURO}$2 Actual:${NC} $current_value"
    read -p "$(echo -e ${VERDE}$2 Nuevo:${NC} )" new_value
    #si no escribo nada, se mantiene el valor actual
    if [ -z "$new_value" ]; then
        new_value=$current_value
    fi
    #si escribo l, pone localhost
    if [ "$new_value" = "l" ]; then
      new_value="localhost"
    fi
    eval export $var_name=$new_value
}

escribirMEM_PORT() {
  sed -i "s/\"port\": .*,/\"port\": $MEM_PORT,/" Memoria_IO-FS.json
  sed -i "s/\"port\": .*,/\"port\": $MEM_PORT,/" Memoria_Plani-DL-Mem.json
  sed -i "s/\"port\": .*,/\"port\": $MEM_PORT,/" Memoria_SE.json
}

escribirINST_PATH() {
  sed -i "s|\"instructions_path\": .*,|\"instructions_path\": \"$INST_PATH\",|" Memoria_IO-FS.json
  sed -i "s|\"instructions_path\": .*,|\"instructions_path\": \"$INST_PATH\",|" Memoria_Plani-DL-Mem.json
  sed -i "s|\"instructions_path\": .*,|\"instructions_path\": \"$INST_PATH\",|" Memoria_SE.json
}

export INST_PATH=../../algo-pruebas;
export MEM_PORT=8002;

while true; do
    echo -e "${AMARILLO}1.${NC} Modificar PATH"
    echo -e "${AMARILLO}2.${NC} Modificar Puerto Memoria"
    echo -e "${AMARILLO}p.${NC} Print Settings"
    echo -e "${AMARILLO}d.${NC} Default Settings"
    echo -e "${ROJO}s.${NC} Salir"
    echo 
    read -p "$(echo -e ${AMARILLO}Opción:${NC} )" opcion

    case $opcion in
        1) modificar "PATH" "INST_PATH"
           escribirINST_PATH ;;
        2) modificar "Puerto" "MEM_PORT"
           escribirMEM_PORT ;;
        p) echo -e "${GRIS_OSCURO}INST_PATH:${NC} $INST_PATH"
           echo -e "${GRIS_OSCURO}MEM_PORT:${NC} $MEM_PORT" ;;
        d) export INST_PATH=../../algo-pruebas;
           export MEM_PORT=8002; 
           escribirINST_PATH
           escribirMEM_PORT;;
        s) echo -e "${ROJO}Saliendo...${NC}"; break ;;
        *) echo -e "${ROJO}Opción no válida${NC}" ;;
    esac
    echo
done
