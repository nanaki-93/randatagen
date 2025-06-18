package model

type DataGen struct {
	Columns        []Column `json:"columns"`
	DbType         string   `json:"dbType"`
	Rows           int      `json:"rows"`
	DbHost         string   `json:"dbHost"`
	DbPort         string   `json:"dbPort"`
	DbUser         string   `json:"dbUser"`
	DbPassword     string   `json:"dbPassword"`
	DbName         string   `json:"dbName"`
	DbSchema       string   `json:"dbSchema"`
	DbTable        string   `json:"dbTable"`
	OutputFilePath string   `json:"-"`
}

type Column struct {
	Name     string `json:"name"`
	Datatype string `json:"datatype"`
	Length   int    `json:"length"`
	Now      bool   `json:"now"`
}
