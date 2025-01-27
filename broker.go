package main

type Broker struct {
	topics           map[EquipmentName]Topic
	MeasurementsAddr string
	GEAddr           string
}

type Topic struct {
	Name         string
	Measurements []Measurement
}

func ListenMeasurements() error {
	panic("implement me")
}

func ListenGE() error {
	panic("implement me")
}
