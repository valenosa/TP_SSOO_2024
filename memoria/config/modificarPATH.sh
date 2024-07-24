#!/bin/bash

# Check and set INST_PATH (if needed)
if [ -z "$INST_PATH" ]; then
  echo "The INST_PATH is not set"
  echo "Using default path (../../algo-pruebas)"
  INST_PATH=../../algo-pruebas
fi

# Update parametro en archivo Memoria_IO-FS.json
sed -i "s|\"instructions_path\": .*,|\"instructions_path\": \"$INST_PATH\",|" Memoria_IO-FS.json

# Update parametro en archivo Memoria_Plani-DL-Mem.json
sed -i "s|\"instructions_path\": .*,|\"instructions_path\": \"$INST_PATH\",|" Memoria_Plani-DL-Mem.json

# Update parametro en archivo Memoria_SE.json
sed -i "s|\"instructions_path\": .*,|\"instructions_path\": \"$INST_PATH\",|" Memoria_SE.json