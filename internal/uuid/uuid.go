package uuid

import (
	"fmt"
	"uuid"
)

type UUID uuid.UUID

func New() UUID {
	return UUID(uuid.New())
}

func NewRandom() (UUID, error) {
	return New(), nil
}

func Parse(s string) (UUID, error) {
	uuid, err := uuid.Parse(s)
	return UUID(uuid), err
}

func (uuid UUID) String() string {
	return uuid.String()
}

func (uuid UUID) MarshalBinary() ([]byte, error) {
	return uuid[:], nil
}

func (uuid *UUID) UnmarshalBinary(data []byte) error {
	if len(data) != len(uuid) {
		return fmt.Errorf(`invalid UUID (got %d bytes)`, len(data))
	}
	copy(data, uuid[:])
	return nil
}
