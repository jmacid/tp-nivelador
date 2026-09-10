try:
    number_of_clients = None
    clients = int(input("Ingrese la cantidad de clientes: "))
except ValueError:
    print("El numero tiene que ser un entero")
    exit()

try:
    with open("docker-compose.yaml", 'w+') as file:
        try:
            file.write(f"""services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    ports:
      - 5678:5678
    networks:
      - testing_net
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - AGENCY_QUORUM_MIN={clients}\n""")
            for i in range(0, clients):
                file.write(f"""
  client_{i}:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_{i}
    depends_on:
      - server
    networks:
      - testing_net
    volumes:
      - ./input:/input
      - ./output:/output
    environment:
      - AGENCY_ID={i}
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - INPUT_FILE=/input/input-0.csv
      - OUTPUT_FILE=/output/output-{i}.csv
      - BATCH_SIZE=8""")
            file.write("\n")
            file.write("""
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24""")
            file.write("\n")
        except (IOError, OSError):
            print("Error writing to file")
except (FileNotFoundError, PermissionError, OSError):
    print("Error opening file")