package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncodeAgency_ProducesTagAndBigEndianId(t *testing.T) {
	got := EncodeAgency(3)

	want := []byte{tagAgency, 0, 0, 0, 3}
	if !bytes.Equal(got, want) {
		t.Fatalf("EncodeAgency(3) = %v, want %v", got, want)
	}
}

func TestEncodeBet_ProducesExpectedBytes(t *testing.T) {
	bet := Bet{FirstName: "Ana", LastName: "Diaz", Document: 28555666, Birthdate: "1985-05-05", Number: 7574}

	got, err := EncodeBet(bet)
	if err != nil {
		t.Fatalf("EncodeBet returned error: %v", err)
	}

	want := []byte{tagBet, 3}
	want = append(want, []byte("Ana")...)
	want = append(want, 4)
	want = append(want, []byte("Diaz")...)
	documentBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(documentBytes, 28555666)
	want = append(want, documentBytes...)
	want = append(want, []byte("1985-05-05")...)
	numberBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(numberBytes, 7574)
	want = append(want, numberBytes...)

	if !bytes.Equal(got, want) {
		t.Fatalf("EncodeBet() = %v, want %v", got, want)
	}
}

func TestEncodeBet_RejectsWrongBirthdateLength(t *testing.T) {
	bet := Bet{FirstName: "Ana", LastName: "Diaz", Document: 1, Birthdate: "1985-5-5", Number: 1}

	if _, err := EncodeBet(bet); err == nil {
		t.Fatal("EncodeBet with malformed birthdate should return an error")
	}
}

func TestEncodeDone_ProducesJustTheTag(t *testing.T) {
	got := EncodeDone()

	want := []byte{tagDone}
	if !bytes.Equal(got, want) {
		t.Fatalf("EncodeDone() = %v, want %v", got, want)
	}
}

func TestDecodeWinners_ParsesMultipleRecords(t *testing.T) {
	betA := Bet{FirstName: "Ana", LastName: "Diaz", Document: 28555666, Birthdate: "1985-05-05", Number: 7574}
	betB := Bet{FirstName: "Bob", LastName: "Ruiz", Document: 12345678, Birthdate: "1990-01-01", Number: 42}

	fieldsA, err := encodeBetFields(betA)
	if err != nil {
		t.Fatalf("encodeBetFields returned error: %v", err)
	}
	fieldsB, err := encodeBetFields(betB)
	if err != nil {
		t.Fatalf("encodeBetFields returned error: %v", err)
	}

	payload := []byte{tagWinners, 0, 0, 0, 2}
	payload = append(payload, fieldsA...)
	payload = append(payload, fieldsB...)

	got, err := DecodeWinners(payload)
	if err != nil {
		t.Fatalf("DecodeWinners returned error: %v", err)
	}

	want := []Bet{betA, betB}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("DecodeWinners() = %v, want %v", got, want)
	}
}

func TestDecodeWinners_OfEmptyListReturnsEmptySlice(t *testing.T) {
	payload := []byte{tagWinners, 0, 0, 0, 0}

	got, err := DecodeWinners(payload)
	if err != nil {
		t.Fatalf("DecodeWinners returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("DecodeWinners() = %v, want empty slice", got)
	}
}

func TestDecodeWinners_RejectsWrongTag(t *testing.T) {
	payload := EncodeAgency(3)

	if _, err := DecodeWinners(payload); err == nil {
		t.Fatal("DecodeWinners with wrong tag should return an error")
	}
}

func TestDecodeWinners_RejectsPayloadTooShortForHeader(t *testing.T) {
	payload := []byte{tagWinners, 0, 0}

	if _, err := DecodeWinners(payload); err == nil {
		t.Fatal("DecodeWinners with a truncated header should return an error")
	}
}

func TestDecodeWinners_RejectsRecordTruncatedMidName(t *testing.T) {
	payload := []byte{tagWinners, 0, 0, 0, 1, 3, 'A'}

	if _, err := DecodeWinners(payload); err == nil {
		t.Fatal("DecodeWinners with a record truncated mid-name should return an error")
	}
}

func TestDecodeWinners_RejectsRecordTruncatedMidFixedFields(t *testing.T) {
	betFields, err := encodeBetFields(Bet{FirstName: "Ana", LastName: "Diaz", Document: 1, Birthdate: "1985-05-05", Number: 1})
	if err != nil {
		t.Fatalf("encodeBetFields returned error: %v", err)
	}

	truncated := betFields[:len(betFields)-3]
	payload := []byte{tagWinners, 0, 0, 0, 1}
	payload = append(payload, truncated...)

	if _, err := DecodeWinners(payload); err == nil {
		t.Fatal("DecodeWinners with a record truncated mid-fixed-fields should return an error")
	}
}

func TestDecodeWinners_RejectsHugeCountWithoutPanicking(t *testing.T) {
	payload := []byte{tagWinners, 0xFF, 0xFF, 0xFF, 0xFF}

	if _, err := DecodeWinners(payload); err == nil {
		t.Fatal("DecodeWinners with a huge declared count and no records should return an error")
	}
}
