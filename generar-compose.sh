#!/bin/bash

# Validación de argumentos de entrada
if [ $# -lt 2 ]; then
  echo "Uso: $0 <archivo_salida> <cantidad_clientes>"
  exit 1
fi

OUTPUT_FILE="$1"
CLIENT_COUNT="$2"

cat > "$OUTPUT_FILE" <<EOL
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    volumes:
      - ./server:/config
    environment:
      - PYTHONUNBUFFERED=1
    networks:
      - testing_net
EOL

# Agregacion dinamica de clientes
for ((i=1; i<=CLIENT_COUNT; i++)); do
cat >> "$OUTPUT_FILE" <<EOL

  client$i:
    container_name: client$i
    image: client:latest
    entrypoint: /client
    volumes:
      - ./client:/config
    environment:
      - CLI_ID=$i
    networks:
      - testing_net
    depends_on:
      - server
EOL
done


cat >> "$OUTPUT_FILE" <<EOL
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOL

echo "Archivo '$OUTPUT_FILE' creado exitosamente con $CLIENT_COUNT clientes."
