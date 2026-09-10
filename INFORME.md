# Informe — TP Nivelador (Docker, Comunicaciones y Concurrencia)

## Introducción

El sistema implementa una lotería distribuida: uno o más **clientes** (agencias, en Go) leen apuestas desde un archivo CSV y las envían por lotes a un **servidor** central (Python), que las persiste, calcula los ganadores una vez que todas las agencias esperadas terminaron de enviar sus apuestas, y devuelve a cada agencia sus propios ganadores.

A continuación se describe cada ejercicio.


## Ejercicio 1 — Múltiples clientes

Se agrega `crear_dc.py`, un generador interactivo de `docker-compose.yaml`: pide por consola la cantidad de clientes y crea un servicio `client_i` por cada uno, con `AGENCY_ID=i` para diferenciarlos en los logs.

## Ejercicio 2 — Exponer el puerto

Se agrega el mapeo `ports: 5678:5678` al servicio `server` (dentro del generador `crear_dc.py`), permitiendo `echo "Hello World" | nc localhost 5678` desde el host.

## Ejercicio 3 — Leer `INPUT_FILE` y persistir `OUTPUT_FILE`

En el cliente se añaden los campos `InputFile`, `OutputFile`. Luego se lee linea por linea mandando cada una al servidor y persistiendo la respuesta  para liberar los file handles.
En el `docker-compose.yaml` se agregan los volumenes tanto para input como para output lo que evita reconstruir la imagen del cliente al cambiar los CSV.

No hay cambio en el servidor.

## Ejercicio 4 — Short read / short write

Se reescribe `safe_socket` en ambos lenguajes:

- **Servidor**: `recv_all(sock, size)` se acumula en un `bytearray` hasta completar un cierto tamaño. `send_all(sock, data)` reintenta `send()` hasta agotar el buffer.
- **Cliente**: `RecvAll`/`SendAll` hacen el mismo loop acumulativo sobre `io.Reader`/`io.Writer`.

Ademas, el servidor introduce un header de 4 bytes (`_MESSAGE_HEADER_SIZE`) para conocer de antemano el tamaño del mensaje a leer, en lugar de un buffer de tamaño fijo; el cliente hace lo simétrico con `MESSAGE_HEADER_SIZE`.

## Ejercicio 5 — Protocolo cliente-servidor y lógica de lotería

Aparece en ambos lados un módulo `protocol` dedicado, separando el **modelo de dominio** de la **capa de comunicación**, y el servidor integra la clase `Lottery` provista por la cátedra (`src_frozen/lottery/`).

Se incluyen estructuras de datos para facilitar la comunicacion y procesamiento:

- `Bet` — modelo de una apuesta.
- `Lottery`
- Mensajes del protocolo (tags de 1 byte, iguales en ambos lados): `AGENCY=1`, `BET=2` (efímero, reemplazado en el Ej. 6), `DONE=3`, `WINNERS=4`.

En la comunicacion la agencia se identifica una vez (`AGENCY`), envía sus apuestas y cierra con `DONE`. Luego el servidor responde con la lista de ganadores de esa agencia (`WINNERS`).

## Ejercicio 6 — Batches

Se reemplaza el envío de apuestas una por una por lotes (`BATCH`) configurables por la variable de entorno `BATCH_SIZE`.
Se elimina `BET` del protocolo y se agregan `BATCH=5`, `BATCH_OK=6`, `BATCH_ERROR=7`.
En el cliente se acumulan las apuestas en y al llegar a `BatchSize` llama `sendBatch` (manda el frame y espera el ack).
En el servidor se decodifica el batch completo y lo persiste con `Lottery.store_bets`. Se responde `BATCH_OK` o `BATCH_ERROR` por lote, sin abortar la conexión ante un batch inválido. Esto permite reintentar ese lote puntual en vez de perder toda la sesión.

## Ejercicio 7 — Concurrencia y quorum de agencias

El servidor pasa de atender clientes secuencialmente a un **thread por conexión**, y se agrega la sincronización necesaria para no calcular ganadores hasta que todas las agencias esperadas terminaron.

Se añade el `AgencyQuorum` que encapsula un `threading.Condition` y un `set()` de IDs de agencias que ya terminaron. `wait(agency_id) -> bool` agrega el id, notifica (`notify_all`) y bloquea hasta alcanzar el mínimo (`AGENCY_QUORUM_MIN`).

## Ejercicio 8 — Apagado *graceful* ante SIGTERM

Cambios simétricos en cliente y servidor para terminar limpiamente ante `SIGTERM`, liberando sockets, threads/goroutines y archivos antes de finalizar el proceso.