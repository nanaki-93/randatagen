package model

type DataGen struct {
	Columns []Column `json:"columns"`
}

type Column struct {
	Name     string `json:"name"`
	Datatype string `json:"datatype"`
	Length   int    `json:"length"`
}
