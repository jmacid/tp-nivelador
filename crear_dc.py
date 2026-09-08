try:
    number_of_clients = None
    clients = int(input("Ingrese la cantidad de clientes: "))
except ValueError:
    print("El numero tiene que ser un entero")
    exit()

try:
    with open("docker-compose.yaml", 'w+') as file:
        try:
            file.write("""services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678\n""")
            for i in range(0, clients):
                file.write(f"""
  client_{i}:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_{i}
    depends_on:
      - server
    environment:
      - AGENCY_ID={i}
      - SERVER_HOST=server
      - SERVER_PORT=5678""")
            file.write("\n")
        except (IOError, OSError):
            print("Error writing to file")
except (FileNotFoundError, PermissionError, OSError):
    print("Error opening file")

print(f"El numero ingresado es: {clients}")