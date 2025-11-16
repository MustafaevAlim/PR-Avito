package handlers

import "github.com/google/uuid"

func StringToUUID(s string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(s))
}
