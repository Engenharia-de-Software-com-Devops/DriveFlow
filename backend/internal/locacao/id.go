package locacao

import (
	"crypto/rand"
	"encoding/hex"
)

// IDAleatorio gera um identificador hexadecimal de 16 caracteres.
func IDAleatorio() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic("nao foi possivel gerar id: " + err.Error())
	}
	return hex.EncodeToString(b)
}
