#!/bin/bash
NETWORK_NAME="tp-nivelador_testing_net"
PORT="5678"
SERVER_NAME="server"
TEST_MESSAGE="Hola_Mundo_$(date +%s)"

# 1. Agregamos 'timeout 1' antes de nc
RESULT=$(docker run --rm --network="$NETWORK_NAME" alpine /bin/sh -c "echo '$TEST_MESSAGE' | timeout 1 nc $SERVER_NAME $PORT")

# 2. Limpiamos posibles caracteres \r de la respuesta del servidor
RESULT=$(echo "$RESULT" | tr -d '\r')

if [ "$RESULT" = "$TEST_MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi