#!/bin/bash

TEST_MESSAGE="Hello, Docker!"

# Enviar mensaje al servidor y capturar la respuesta
# rm -> eliminar el contenedor despues de correrlo
# network -> la red sobre la que corre nuestro servidor
NETWORK="tp0_testing_net"
SERVER_PORT=12345
RESPONSE=$(docker run --rm --network $NETWORK alpine sh -c "echo '$TEST_MESSAGE' | nc server $SERVER_PORT 2>/dev/null")

# Verificamos que el mensaje recibido sea igual al enviado
if [ "$RESPONSE" == "$TEST_MESSAGE" ]; then
  echo "action: test_echo_server | result: success"
  exit 0
else
  echo "action: test_echo_server | result: fail"
  exit 1
fi
