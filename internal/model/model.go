package model

type RanData struct {
	Columns        []Column `json:"columns"`
	Rows           int      `json:"rows"`
	DbType         string   `json:"dbType"`
	Target         DbStruct `json:"target"`
	OutputFilePath string   `json:"-"`
}

type DbStruct struct {
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

type MigrationData struct {
	DbType string   `json:"dbType"`
	Source DbStruct `json:"source"`
	Target DbStruct `json:"target"`
}

type CreateQuery struct {
	Table      string `json:"table"`
	Index      string `json:"index"`
	PrimaryKey string `json:"primaryKey"`
}
