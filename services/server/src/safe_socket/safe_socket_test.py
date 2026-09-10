import struct
import unittest

from safe_socket import send_frame, recv_frame


class _FakeSocket:

    _MAX_CHUNK = 3

    def __init__(self, initial_bytes: bytes = b""):
        self._inbox = bytearray(initial_bytes)
        self.sent = bytearray()

    def send(self, data: bytes) -> int:
        chunk = data[: self._MAX_CHUNK]
        self.sent.extend(chunk)
        return len(chunk)

    def recv(self, size: int) -> bytes:
        to_read = min(self._MAX_CHUNK, size)
        chunk = bytes(self._inbox[:to_read])
        del self._inbox[:to_read]
        return chunk


class SendFrameTest(unittest.TestCase):
    def test_send_frame_writes_length_prefixed_payload_despite_short_writes(self):
        sock = _FakeSocket()

        send_frame(sock, b"hello")

        self.assertEqual(bytes(sock.sent), struct.pack(">I", 5) + b"hello")


class RecvFrameTest(unittest.TestCase):
    def test_recv_frame_reads_length_prefixed_payload_despite_short_reads(self):
        sock = _FakeSocket(struct.pack(">I", 8) + b"abcdefgh")

        payload = recv_frame(sock)

        self.assertEqual(payload, b"abcdefgh")

    def test_recv_frame_returns_empty_payload_for_zero_length_frame(self):
        sock = _FakeSocket(struct.pack(">I", 0))

        payload = recv_frame(sock)

        self.assertEqual(payload, b"")


if __name__ == "__main__":
    unittest.main()
