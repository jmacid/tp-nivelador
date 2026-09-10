import struct
import unittest

from . import (
    AGENCY,
    DONE,
    WINNERS,
    BATCH,
    BATCH_OK,
    BATCH_ERROR,
    ProtocolError,
    encode_agency,
    decode_agency,
    encode_done,
    encode_winners,
    decode_winners,
    encode_batch,
    decode_batch,
    encode_batch_ok,
    encode_batch_error,
)


class AgencyCodecTest(unittest.TestCase):
    def test_decode_agency_round_trips_encode_agency(self):
        payload = encode_agency(3)
        self.assertEqual(decode_agency(payload), 3)

    def test_decode_agency_rejects_wrong_tag(self):
        with self.assertRaises(ProtocolError):
            decode_agency(encode_done())


class BatchCodecTest(unittest.TestCase):
    def test_encode_batch_produces_expected_bytes(self):
        payload = encode_batch([("Ana", "Diaz", 28555666, "1985-05-05", 7574)])

        expected = (
            bytes([BATCH])
            + struct.pack(">I", 1)
            + bytes([3])
            + b"Ana"
            + bytes([4])
            + b"Diaz"
            + struct.pack(">I", 28555666)
            + b"1985-05-05"
            + struct.pack(">I", 7574)
        )
        self.assertEqual(payload, expected)

    def test_decode_batch_round_trips_encode_batch(self):
        bets = [
            ("Ana", "Diaz", 28555666, "1985-05-05", 7574),
            ("Bob", "Ruiz", 12345678, "1990-01-01", 42),
        ]

        payload = encode_batch(bets)

        self.assertEqual(decode_batch(payload), bets)

    def test_decode_batch_of_empty_list_round_trips(self):
        payload = encode_batch([])

        self.assertEqual(decode_batch(payload), [])

    def test_decode_batch_rejects_wrong_tag(self):
        with self.assertRaises(ProtocolError):
            decode_batch(encode_agency(3))

    def test_encode_batch_rejects_malformed_bet(self):
        with self.assertRaises(ProtocolError):
            encode_batch([("Ana", "Diaz", 1, "bad-date", 1)])


class BatchAckCodecTest(unittest.TestCase):
    def test_encode_batch_ok_produces_just_the_tag(self):
        self.assertEqual(encode_batch_ok(), bytes([BATCH_OK]))

    def test_encode_batch_error_produces_just_the_tag(self):
        self.assertEqual(encode_batch_error(), bytes([BATCH_ERROR]))


class WinnersCodecTest(unittest.TestCase):
    def test_decode_winners_parses_multiple_records(self):
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
