package shared

var IsDebug bool
var DbType string

func SetDebug(condition bool) {
	IsDebug = condition
}

func GetDebug() bool {
	return IsDebug
}

func SetDbType(text string) {
	DbType = text
}

func GetDbType() string {
	return DbType
}
