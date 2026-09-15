// Package id gera os identificadores usados nos registros da plataforma.
//
// Fica em pkg/ por ser um utilitario generico: nao depende de nenhuma camada
// da aplicacao e poderia ser reaproveitado por outro servico.
package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New devolve um identificador hexadecimal de 16 caracteres.
func New() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic("nao foi possivel gerar id: " + err.Error())
	}
	return hex.EncodeToString(b)
}
