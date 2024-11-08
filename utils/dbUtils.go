package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"

	"github.com/joijoku/PR/shared"
)

type NullString sql.NullString
type NullFloat64 sql.NullFloat64

var mapValidation map[string]string

type DbConnInfo struct {
	DbLocation string `json:"dbLocation" validate:"required,required"`
	DbPort     int    `json:"dbPort" validate:"required,required"`
	DbUser     string `json:"dbUser" validate:"required,required"`
	DbPass     string `json:"dbPass" validate:"required,required"`
	DbName     string `json:"dbName" validate:"required,required"`
	DbSvc      string `json:"dbSvc"`
	DbTypeName string `json:"dbType" validate:"required,required"`
}

func SetDBConf(
	dbLocation string,
	dbPort int,
	dbUser string,
	dbPass string,
	dbName string,
	dbSvc string,
	dbTypeName string,
) DbConnInfo {
	var dbConn DbConnInfo

	shared.Block{
		Try: func() {
			dbConn.DbLocation = dbLocation
			dbConn.DbPort = dbPort
			dbConn.DbUser = dbUser
			dbConn.DbPass = dbPass
			dbConn.DbName = dbName
			dbConn.DbSvc = dbSvc
			dbConn.DbTypeName = dbTypeName
		},
		Catch: func(e shared.Exception) {
			dbConn.DbLocation = ""
			dbConn.DbPort = 0
			dbConn.DbUser = ""
			dbConn.DbPass = ""
			dbConn.DbName = ""
			dbConn.DbSvc = ""
			dbConn.DbTypeName = ""
		},
	}.Do()

	return dbConn
}

func MapToObject(mp map[string]any) DbConnInfo {
	var dbConn DbConnInfo
	shared.Block{
		Try: func() {
			dbConn.DbLocation = mp["dbLocation"].(string)
			dbConn.DbPort = int(math.Round(mp["dbPort"].(float64)))
			dbConn.DbUser = mp["dbUser"].(string)
			dbConn.DbPass = mp["dbPass"].(string)
			dbConn.DbName = mp["dbName"].(string)
			dbConn.DbSvc = mp["dbSvc"].(string)
			dbConn.DbTypeName = mp["dbType"].(string)
		},
		Catch: func(e shared.Exception) {
			dbConn.DbLocation = ""
			dbConn.DbPort = 0
			dbConn.DbUser = ""
			dbConn.DbPass = ""
			dbConn.DbName = ""
			dbConn.DbSvc = ""
			dbConn.DbTypeName = ""
		},
	}.Do()

	return dbConn

}

func ObjectToMap(dbModel DbConnInfo) map[string]any {
	mp := make(map[string]any)

	shared.Block{
		Try: func() {
			mp["dbLocation"] = dbModel.DbLocation
			mp["dbPort"] = dbModel.DbPort
			mp["dbUser"] = dbModel.DbUser
			mp["dbPass"] = dbModel.DbPass
			mp["dbName"] = dbModel.DbName
			mp["dbSvc"] = dbModel.DbSvc
			mp["dbType"] = dbModel.DbTypeName
		},
		Catch: func(e shared.Exception) {

		},
	}.Do()

	return mp
}

func CreateConnection(dbModel DbConnInfo) (*sql.DB, error) {
	var dbinfo string

	log.Println(dbModel.DbTypeName)
	switch dbModel.DbTypeName {
	case "postgresql":
		dbinfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbModel.DbLocation,
			dbModel.DbPort,
			dbModel.DbUser,
			dbModel.DbPass,
			dbModel.DbName)
	case "oracle":
		dbinfo = fmt.Sprintf("%s://%s:%s@%s:%d/%s", dbModel.DbTypeName,
			dbModel.DbUser,
			dbModel.DbPass,
			dbModel.DbLocation,
			dbModel.DbPort,
			dbModel.DbSvc)
	case "sqlserver":
		dbinfo = fmt.Sprintf("server=%s;port=%d;database=%s;user id=%s;password=%s;trustservercertificate=true;encrypt=DISABLE", dbModel.DbLocation,
			dbModel.DbPort,
			dbModel.DbName,
			dbModel.DbUser,
			dbModel.DbPass)
	case "odbc":
		dbinfo = fmt.Sprintf("DSN=%s;Uid=%s;Pwd=%s;", dbModel.DbSvc, dbModel.DbUser, dbModel.DbPass)
	case "mysql":
		dbinfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbModel.DbUser,
			dbModel.DbPass,
			dbModel.DbLocation,
			dbModel.DbPort,
			dbModel.DbName)
	}

	log.Println(dbinfo)
	db, err := sql.Open(dbModel.DbTypeName, dbinfo)
	shared.CheckErr(err)

	err = db.Ping()

	if err != nil {
		log.Println("Connection Failed to Open")
	} else {
		log.Println("Connection Established")

		// db.Close()
	}

	return db, err
}

func CreateConnectionFromToken(token string) (*sql.DB, error) {
	var err error
	var db *sql.DB

	shared.Block{
		Try: func() {
			mpConnectionInfo := DecryptMapConnectionInfo(token)

			db, err = CreateConnection(MapToObject(mpConnectionInfo))
			shared.CheckErr(err)
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
		},
	}.Do()

	return db, err
}

func ReadIntefaceVal(face interface{}) string {
	res, err := json.Marshal(face)
	shared.CheckErr(err)
	var body map[string]interface{}
	json.Unmarshal(res, &body)

	if body["Valid"] == false {
		return string([]byte(nil))
	}

	return body["String"].(string)
}

func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ns = NullString{s.String, false}
	} else {
		*ns = NullString{s.String, true}
	}

	return nil
}

func (ni *NullString) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(ni.String)
}

func (ni *NullFloat64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(ni.Float64)
}

func GetDBResultRow(resultList []interface{}, idx int) map[string]interface{} {
	return resultList[idx].(map[string]interface{})
}

func GetFieldStringValue(row interface{}, idxString string) string {
	return (row.(map[string]interface{}))[idxString].(string)
}

func Select(db *sql.DB, query string, keepCn bool, limit map[string]any) ([]any, error) {
	resList := make([]any, 0)
	var err error
	var resInterface []interface{}

	shared.Block{
		Try: func() {
			err := db.Ping()
			shared.CheckErr(err)
			// resList = make([]any, 0)

			if len(limit) > 0 {
				mapValidation = map[string]string{
					"offset": "required|num",
					"limit":  "required|num",
					"dbType": "required",
				}

				v := SetValidation(limit, mapValidation)
				if v.Validate() {
					switch dbType := limit["dbType"].(string); dbType {
					case "mysql":
						query += fmt.Sprintf("limit %d, %d", limit["offset"].(int), limit["limit"].(int))
					case "postgres":
						query += fmt.Sprintf("limit %d offset %d", limit["limit"].(int), limit["offset"].(int))
					case "sqlserver":
						query += fmt.Sprintf("offset %d rows fetch next %d rows only", limit["offset"].(int), limit["limit"].(int))
					case "oracle":
						query += fmt.Sprintf("offset %d rows fetch next %d rows only", limit["offset"].(int), limit["limit"].(int))
					}
				}
			}

			ShowDebug("Run Select Query : " + query)
			rows, err := db.Query(query)
			shared.CheckErr(err)
			colNames, err := rows.Columns()
			shared.CheckErr(err)

			colLen := len(colNames)
			resInterface = make([]interface{}, colLen)

			for rows.Next() {
				mapRes := make(map[string]any)

				for i := 0; i < colLen; i++ {
					resInterface[i] = new(sql.NullString)
				}

				err := rows.Scan(resInterface...)
				shared.CheckErr(err)
				for i, col := range colNames {
					mapRes[col] = ReadIntefaceVal(resInterface[i])
				}

				resList = append(resList, mapRes)

			}

			rows.Close()

			if !keepCn {
				db.Close()
			}

			err = nil
		},
		Catch: func(e shared.Exception) {
			// panic(e.(error))
			err = e.(error)
		},
	}.Do()

	return resList, err
}

func SelectWithParam(db *sql.DB, query string, params []interface{}, keepCn bool, limit map[string]any) ([]any, error) {
	resList := make([]any, 0)
	var err error

	shared.Block{
		Try: func() {
			err := db.Ping()
			shared.CheckErr(err)

			if len(limit) > 0 {
				mapValidation = map[string]string{
					"offset": "required|num",
					"limit":  "required|num",
					"dbType": "required",
				}

				v := SetValidation(limit, mapValidation)
				if v.Validate() {
					switch dbType := limit["dbType"].(string); dbType {
					case "mysql":
						query += fmt.Sprintf("limit %d, %d", limit["offset"].(int), limit["limit"].(int))
					case "postgres":
						query += fmt.Sprintf("limit %d offset %d", limit["limit"].(int), limit["offset"].(int))
					case "sqlserver":
						query += fmt.Sprintf("offset %d rows fetch next %d rows only", limit["offset"].(int), limit["limit"].(int))
					case "oracle":
						query += fmt.Sprintf("offset %d rows fetch next %d rows only", limit["offset"].(int), limit["limit"].(int))
					}
				}
			}

			// log.Printf("Run Select Query : %s\n", query)
			ShowDebug("Run Select Query : " + query)
			rows, err := db.Query(query, params...)
			shared.CheckErr(err)
			colNames, err := rows.Columns()
			shared.CheckErr(err)

			colLen := len(colNames)
			resInterface := make([]interface{}, colLen)

			for rows.Next() {
				mapRes := make(map[string]any)

				for i := 0; i < colLen; i++ {
					resInterface[i] = new(sql.NullString)
				}

				err := rows.Scan(resInterface...)
				shared.CheckErr(err)
				for i, col := range colNames {
					mapRes[col] = ReadIntefaceVal(resInterface[i])
				}

				resList = append(resList, mapRes)

			}

			defer rows.Close()

			if !keepCn {
				db.Close()
			}

			err = nil

		},
		Catch: func(e shared.Exception) {
			err = e.(error)
			log.Println("error " + err.Error())
		},
	}.Do()

	return resList, err
}

func SelectOne(db *sql.DB, query string, col []string, keepCn bool) (map[string]any, error) {
	var result map[string]any
	var err error
	var resInterface []interface{}

	shared.Block{
		Try: func() {
			err := db.Ping()
			shared.CheckErr(err)
			// resList = make([]any, 0)

			// log.Printf("Run Select Query : %s\n", query)
			ShowDebug("Run Select Query : " + query)
			row := db.QueryRow(query)

			resInterface = make([]interface{}, len(col))
			for i := 0; i < len(col); i++ {
				resInterface[i] = new(sql.NullString)
			}

			err = row.Scan(resInterface...)
			shared.CheckErr(err)

			result = make(map[string]any)
			for i, value := range resInterface {
				result[col[i]] = ReadIntefaceVal(value)
			}

			if !keepCn {
				db.Close()
			}
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
		},
	}.Do()

	return result, err
}

func ExecQueryWithParams(db *sql.DB, query string, params []interface{}, keepCn bool) error {
	var err error

	shared.Block{
		Try: func() {
			err := db.Ping()
			shared.CheckErr(err)

			log.Printf("Execute query : %s\n", query)
			_, err = db.Exec(query, params...)
			shared.CheckErr(err)

			if !keepCn {
				db.Close()
			}

			err = nil
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
		},
	}.Do()

	return err
}

func ExecQuery(db *sql.DB, query string, keepCn bool) error {
	var err error

	shared.Block{
		Try: func() {
			err := db.Ping()
			shared.CheckErr(err)

			log.Printf("Execute query : %s\n", query)
			_, err = db.Exec(query)
			shared.CheckErr(err)

			if !keepCn {
				db.Close()
			}

			err = nil
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
		},
	}.Do()

	return err
}
