import socket
import struct


def recv_all(sock: socket.socket, size: int) -> bytearray:
    data = bytearray()
    while len(data) < size:
        chunk = sock.recv(size - len(data))
        if not chunk:
            raise ConnectionError("La conexión se cerró antes de recibir todos los datos")
        data.extend(chunk)
    return data

_MAX_CONSECUTIVE_EMPTY_SENDS = 10


def send_all(sock: socket.socket, data: bytes) -> None:
    total_sent = 0
    consecutive_empty_sends = 0
    while total_sent < len(data):
        sent = sock.send(data[total_sent:])
        if sent == 0:
            consecutive_empty_sends += 1
            if consecutive_empty_sends >= _MAX_CONSECUTIVE_EMPTY_SENDS:
                raise ConnectionError("La conexión se cerró antes de enviar todos los datos")
            continue
        consecutive_empty_sends = 0
        total_sent += sent


_FRAME_HEADER_SIZE = 4


def send_frame(sock: socket.socket, payload: bytes) -> None:
    header = struct.pack(">I", len(payload))
    send_all(sock, header)
    send_all(sock, payload)


def recv_frame(sock: socket.socket) -> bytes:
    header = recv_all(sock, _FRAME_HEADER_SIZE)
    size = struct.unpack(">I", header)[0]
    return bytes(recv_all(sock, size))
