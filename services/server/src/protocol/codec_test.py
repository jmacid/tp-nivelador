import struct
import unittest

from . import (
    AGENCY,
    BET,
    DONE,
    WINNERS,
    ProtocolError,
    encode_agency,
    decode_agency,
    encode_bet,
    decode_bet,
    encode_winners,
    decode_winners,
)


class AgencyCodecTest(unittest.TestCase):
    def test_decode_agency_round_trips_encode_agency(self):
        payload = encode_agency(3)
        self.assertEqual(decode_agency(payload), 3)

    def test_decode_agency_rejects_wrong_tag(self):
        with self.assertRaises(ProtocolError):
            decode_agency(encode_bet("A", "B", 1, "2000-01-01", 1))


class BetCodecTest(unittest.TestCase):
    def test_encode_bet_produces_expected_bytes(self):
        payload = encode_bet("Ana", "Diaz", 28555666, "1985-05-05", 7574)

        expected = (
            bytes([BET, 3])
            + b"Ana"
            + bytes([4])
            + b"Diaz"
            + struct.pack(">I", 28555666)
            + b"1985-05-05"
            + struct.pack(">I", 7574)
        )
        self.assertEqual(payload, expected)

    def test_decode_bet_round_trips_encode_bet(self):
        payload = encode_bet("Ana", "Diaz", 28555666, "1985-05-05", 7574)

        self.assertEqual(
            decode_bet(payload), ("Ana", "Diaz", 28555666, "1985-05-05", 7574)
        )

    def test_decode_bet_rejects_wrong_tag(self):
        with self.assertRaises(ProtocolError):
            decode_bet(encode_agency(3))

    def test_encode_bet_rejects_wrong_birthdate_length(self):
        with self.assertRaises(ProtocolError):
            encode_bet("Ana", "Diaz", 1, "1985-5-5", 1)


class WinnersCodecTest(unittest.TestCase):
    def test_decode_winners_parses_multiple_records(self):
        bet_a_fields = encode_bet("Ana", "Diaz", 28555666, "1985-05-05", 7574)[1:]
        bet_b_fields = encode_bet("Bob", "Ruiz", 12345678, "1990-01-01", 42)[1:]
        payload = bytes([WINNERS]) + struct.pack(">I", 2) + bet_a_fields + bet_b_fields

        winners = decode_winners(payload)

        self.assertEqual(
            winners,
            [
                ("Ana", "Diaz", 28555666, "1985-05-05", 7574),
                ("Bob", "Ruiz", 12345678, "1990-01-01", 42),
            ],
        )

    def test_encode_winners_round_trips_decode_winners(self):
        bets = [
            ("Ana", "Diaz", 28555666, "1985-05-05", 7574),
            ("Bob", "Ruiz", 12345678, "1990-01-01", 42),
        ]

        payload = encode_winners(bets)

        self.assertEqual(decode_winners(payload), bets)

    def test_decode_winners_of_empty_list_round_trips(self):
        payload = encode_winners([])

        self.assertEqual(decode_winners(payload), [])

    def test_decode_winners_rejects_wrong_tag(self):
        with self.assertRaises(ProtocolError):
            decode_winners(encode_agency(3))


if __name__ == "__main__":
    unittest.main()
