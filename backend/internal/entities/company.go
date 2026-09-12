// Package entities concentra o dominio da plataforma multi-tenant de locacao
// de veiculos: empresas, frota e contratos de locacao.
//
// E a camada mais interna da arquitetura: nao importa nenhuma outra camada do
// projeto e nao conhece banco, HTTP nem framework.
package entities

import "time"

// Company e o tenant da plataforma. Cada empresa enxerga apenas a propria frota.
//
// As tags JSON preservam o contrato ja consumido pelo frontend, por isso
// continuam em portugues mesmo com os campos em ingles.
type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"nome"`
	TaxID     string    `json:"cnpj"` // CNPJ, apenas digitos
	CreatedAt time.Time `json:"criada_em"`
}
