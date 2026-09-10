import socket

# TODO: Complete with a short-read/short-write tolerant implementation


def recv_all(sock: socket.socket, size: int) -> bytearray:
    data = bytearray()
    while len(data) < size:
        chunk = sock.recv(size - len(data))
        if not chunk:
            raise ConnectionError("La conexión se cerró antes de recibir todos los datos")
        data.extend(chunk)
    return data


def send_all(sock: socket.socket, data: bytes) -> None:
    total_sent = 0
    while total_sent < len(data):
        sent = sock.send(data[total_sent:])
        if sent == 0:
            raise ConnectionError("La conexión se cerró antes de enviar todos los datos")
        total_sent += sent
