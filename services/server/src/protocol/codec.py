import struct

AGENCY = 1
BET = 2
DONE = 3
WINNERS = 4

_BIRTHDATE_SIZE = 10  # YYYY-MM-DD


class ProtocolError(Exception):
    """El payload de un frame no respeta el formato de mensaje esperado."""


def _assert_tag(payload: bytes, expected_tag: int) -> None:
    if not payload or payload[0] != expected_tag:
        actual = payload[0] if payload else None
        raise ProtocolError(f"expected message tag {expected_tag}, got {actual}")


def encode_agency(agency_id: int) -> bytes:
    return bytes([AGENCY]) + struct.pack(">I", agency_id)


def decode_agency(payload: bytes) -> int:
    _assert_tag(payload, AGENCY)
    return struct.unpack(">I", payload[1:5])[0]


def _encode_name(name: str) -> bytes:
    encoded = name.encode("utf-8")
    if len(encoded) > 255:
        raise ProtocolError(f"name too long to encode: {len(encoded)} bytes")
    return bytes([len(encoded)]) + encoded


def _decode_name(payload: bytes, offset: int) -> tuple[str, int]:
    length = payload[offset]
    start = offset + 1
    end = start + length
    return payload[start:end].decode("utf-8"), end


def _encode_bet_fields(
    first_name: str, last_name: str, document: int, birthdate: str, number: int
) -> bytes:
    if len(birthdate) != _BIRTHDATE_SIZE:
        raise ProtocolError(
            f"birthdate must be {_BIRTHDATE_SIZE} chars, got {birthdate!r}"
        )
    return (
        _encode_name(first_name)
        + _encode_name(last_name)
        + struct.pack(">I", document)
        + birthdate.encode("ascii")
        + struct.pack(">I", number)
    )


def _decode_bet_fields(payload: bytes, offset: int) -> tuple[tuple, int]:
    first_name, offset = _decode_name(payload, offset)
    last_name, offset = _decode_name(payload, offset)
    document = struct.unpack(">I", payload[offset : offset + 4])[0]
    offset += 4
    birthdate = payload[offset : offset + _BIRTHDATE_SIZE].decode("ascii")
    offset += _BIRTHDATE_SIZE
    number = struct.unpack(">I", payload[offset : offset + 4])[0]
    offset += 4
    return (first_name, last_name, document, birthdate, number), offset


def encode_bet(
    first_name: str, last_name: str, document: int, birthdate: str, number: int
) -> bytes:
    return bytes([BET]) + _encode_bet_fields(
        first_name, last_name, document, birthdate, number
    )


def decode_bet(payload: bytes) -> tuple:
    _assert_tag(payload, BET)
    fields, _ = _decode_bet_fields(payload, 1)
    return fields


def encode_done() -> bytes:
    return bytes([DONE])


def encode_winners(bets: list[tuple]) -> bytes:
    encoded = bytes([WINNERS]) + struct.pack(">I", len(bets))
    for first_name, last_name, document, birthdate, number in bets:
        encoded += _encode_bet_fields(first_name, last_name, document, birthdate, number)
    return encoded


def decode_winners(payload: bytes) -> list[tuple]:
    _assert_tag(payload, WINNERS)
    count = struct.unpack(">I", payload[1:5])[0]
    offset = 5
    winners = []
    for _ in range(count):
        fields, offset = _decode_bet_fields(payload, offset)
        winners.append(fields)
    return winners
