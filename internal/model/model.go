package model

type GenerateData struct {
	Columns        []Column `json:"columns"`
	Rows           int      `json:"rows"`
	Target         DbStruct `json:"target"`
	OutputFilePath string   `json:"-"`
}

type DbStruct struct {
	DbType     string `json:"dbType"`
	DbHost     string `json:"dbHost"`
	DbPort     int    `json:"dbPort"`
	DbUser     string `json:"dbUser"`
	DbPassword string `json:"dbPassword"`
	DbName     string `json:"dbName"`
	DbSchema   string `json:"dbSchema"`
	DbTable    string `json:"dbTable,omitempty"`
}

type Column struct {
	Name     string `json:"name"`
	Datatype string `json:"datatype"`
	Length   int    `json:"length"`
	Now      bool   `json:"now"`
}

type MigrateData struct {
	Source DbStruct `json:"source"`
	Target DbStruct `json:"target"`
}
