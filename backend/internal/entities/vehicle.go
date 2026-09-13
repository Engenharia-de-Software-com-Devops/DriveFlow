package entities

// VehicleStatus representa a situacao operacional de um veiculo da frota.
//
// Os valores sao gravados no banco e validados pela constraint CHECK da
// migration 0002, entao permanecem em portugues.
type VehicleStatus string

const (
	VehicleAvailable   VehicleStatus = "disponivel"
	VehicleMaintenance VehicleStatus = "manutencao"
)

// Vehicle e um item da frota e pertence sempre a uma unica empresa.
type Vehicle struct {
	ID        string        `json:"id"`
	CompanyID string        `json:"empresa_id"`
	Plate     string        `json:"placa"`
	Model     string        `json:"modelo"`
	Category  string        `json:"categoria"`
	DailyRate int64         `json:"tarifa_diaria"` // em centavos
	Status    VehicleStatus `json:"status"`
}
