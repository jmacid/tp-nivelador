package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	tagAgency  = 1
	tagBet     = 2
	tagDone    = 3
	tagWinners = 4
)

const birthdateSize = 10 // YYYY-MM-DD

type Bet struct {
	FirstName string
	LastName  string
	Document  uint32
	Birthdate string
	Number    uint32
}

func EncodeAgency(agencyId uint32) []byte {
	payload := make([]byte, 5)
	payload[0] = tagAgency
	binary.BigEndian.PutUint32(payload[1:], agencyId)
	return payload
}

func encodeName(name string) ([]byte, error) {
	nameBytes := []byte(name)
	if len(nameBytes) > 255 {
		return nil, fmt.Errorf("name too long to encode: %d bytes", len(nameBytes))
	}
	return append([]byte{byte(len(nameBytes))}, nameBytes...), nil
}

func encodeBetFields(bet Bet) ([]byte, error) {
	if len(bet.Birthdate) != birthdateSize {
		return nil, fmt.Errorf("birthdate must be %d chars, got %q", birthdateSize, bet.Birthdate)
	}

	firstName, err := encodeName(bet.FirstName)
	if err != nil {
		return nil, err
	}
	lastName, err := encodeName(bet.LastName)
	if err != nil {
		return nil, err
	}

	fields := make([]byte, 0, len(firstName)+len(lastName)+4+birthdateSize+4)
	fields = append(fields, firstName...)
	fields = append(fields, lastName...)

	document := make([]byte, 4)
	binary.BigEndian.PutUint32(document, bet.Document)
	fields = append(fields, document...)

	fields = append(fields, []byte(bet.Birthdate)...)

	number := make([]byte, 4)
	binary.BigEndian.PutUint32(number, bet.Number)
	fields = append(fields, number...)

	return fields, nil
}

func EncodeBet(bet Bet) ([]byte, error) {
	fields, err := encodeBetFields(bet)
	if err != nil {
		return nil, err
	}
	return append([]byte{tagBet}, fields...), nil
}

func EncodeDone() []byte {
	return []byte{tagDone}
}

func decodeName(payload []byte, offset int) (string, int, error) {
	if offset >= len(payload) {
		return "", 0, fmt.Errorf("payload too short to decode name at offset %d", offset)
	}
	length := int(payload[offset])
	start := offset + 1
	end := start + length
	if end > len(payload) {
		return "", 0, fmt.Errorf("payload too short to decode name of length %d", length)
	}
	return string(payload[start:end]), end, nil
}

func decodeBetFields(payload []byte, offset int) (Bet, int, error) {
	firstName, offset, err := decodeName(payload, offset)
	if err != nil {
		return Bet{}, 0, err
	}
	lastName, offset, err := decodeName(payload, offset)
	if err != nil {
		return Bet{}, 0, err
	}
	if offset+4+birthdateSize+4 > len(payload) {
		return Bet{}, 0, fmt.Errorf("payload too short to decode bet fields")
	}

	document := binary.BigEndian.Uint32(payload[offset : offset+4])
	offset += 4
	birthdate := string(payload[offset : offset+birthdateSize])
	offset += birthdateSize
	number := binary.BigEndian.Uint32(payload[offset : offset+4])
	offset += 4

	return Bet{
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}, offset, nil
}

const minBetFieldsSize = 1 + 1 + 4 + birthdateSize + 4

func DecodeWinners(payload []byte) ([]Bet, error) {
	if len(payload) < 5 || payload[0] != tagWinners {
		return nil, fmt.Errorf("expected message tag %d, got %v", tagWinners, payload)
	}

	count := binary.BigEndian.Uint32(payload[1:5])

	maxPossibleRecords := (len(payload) - 5) / minBetFieldsSize
	capacityHint := int(count)
	if capacityHint > maxPossibleRecords {
		capacityHint = maxPossibleRecords
	}

	offset := 5
	winners := make([]Bet, 0, capacityHint)
	for i := uint32(0); i < count; i++ {
		bet, newOffset, err := decodeBetFields(payload, offset)
		if err != nil {
			return nil, err
		}
		offset = newOffset
		winners = append(winners, bet)
	}
	return winners, nil
}
