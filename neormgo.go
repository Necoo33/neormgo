package neormgo

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const Version = "0.2.0"

type Neorm struct {
	Schema    string
	Query     string
	_Table    string
	Pool      *sql.DB
	_Type     string
	_User     string
	_Password string
	_Scope    string
}

// database connectors:

func (orm Neorm) Connect(connString string) (Neorm, error) {
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return Neorm{}, err
	}

	return Neorm{
		Pool:   db,
		Query:  "",
		_Table: "",
		Schema: "",
	}, nil
}

func (orm *Neorm) Close() {
	orm.Pool.Close()
}

func (orm *Neorm) QueryDrop() error {
	ctx := context.Background()
	getConn, err := orm.Pool.Conn(ctx)

	if err != nil {
		fmt.Println("Error when getting new connection from pool!")
		return err
	}

	if strings.HasPrefix(orm.Query, "CREATE TABLE") && orm.Schema != "" {
		useTable := fmt.Sprintf("USE %s;", orm.Schema)
		_, err = getConn.ExecContext(ctx, useTable)

		if err != nil {
			fmt.Println("Error When Executing Use Query!")
			return err
		}
	}

	_, err = getConn.ExecContext(ctx, orm.Query)

	if err != nil {
		fmt.Println("Error when executing QueryDrop!")
		return err
	}

	defer getConn.Close()

	return nil
}

type Row struct {
	Columns map[string]interface{}
}

func (orm *Neorm) Execute() ([]map[string]interface{}, error) {
	ctx := context.Background()

	newConn, err := orm.Pool.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer newConn.Close()

	if orm._Type == "s" {
		stmt, err := newConn.PrepareContext(ctx, orm.Query)

		if err != nil {
			return nil, err
		}

		rows, err := stmt.Query()
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		var results []map[string]interface{}
		for rows.Next() {
			err := rows.Scan(valuePtrs...)

			if err != nil {
				return nil, err
			}

			rows := make(map[string]interface{})
			for i, col := range columns {
				var v interface{}
				val := values[i]

				b, ok := val.([]byte)
				if ok {
					v = string(b)
				} else {
					v = val
				}

				rows[col] = v
			}

			results = append(results, rows)
		}

		if err = rows.Err(); err != nil {
			return nil, err
		}

		return results, nil
	} else {
		stmt, err := newConn.PrepareContext(ctx, orm.Query)

		if err != nil {
			return nil, err
		}

		_, err = stmt.ExecContext(ctx)

		if err != nil {
			return nil, err
		} else {
			return nil, nil
		}
	}
}

// schema and table builder:

func (orm *Neorm) CreateSchema(name string) Neorm {
	orm.Schema = name

	orm.Query = fmt.Sprintf("CREATE DATABASE %s", name)

	return *orm
}

func (orm *Neorm) Use(schema string) Neorm {
	orm.Query = fmt.Sprintf("USE %s", schema)

	return *orm
}

func (orm *Neorm) CreateTable(name string) Neorm {
	orm._Table = name

	orm.Query = fmt.Sprintf("CREATE TABLE %s", name)

	return *orm
}

func (orm Neorm) IfNotExist() Neorm {
	if strings.HasPrefix(orm.Query, "CREATE DATABASE") {
		splitTheQuery := strings.Split(orm.Query, " DATABASE ")

		orm.Query = fmt.Sprintf("%s DATABASE IF NOT EXISTS %s", splitTheQuery[0], splitTheQuery[1])
	} else if strings.HasPrefix(orm.Query, "CREATE TABLE") {
		splitTheQuery := strings.Split(orm.Query, " TABLE ")

		orm.Query = fmt.Sprintf("%s TABLE IF NOT EXISTS %s", splitTheQuery[0], splitTheQuery[1])
	} else {
		panic("You cannot add 'IF NOT EXISTS' parameter if you don't start to create a schema or table.")
	}

	return orm
}

func (orm *Neorm) AddColumn(name string) Neorm {
	if strings.HasPrefix(orm.Query, "ALTER TABLE") {
		orm.Query = fmt.Sprintf("%s ADD COLUMN %s", orm.Query, name)
	} else {
		splitTheQuery := strings.Split(orm.Query, " ")
		lengthOfTheQuery := len(splitTheQuery)

		failString := fmt.Sprintf("%s (", orm._Table)

		if (splitTheQuery[lengthOfTheQuery-1] == orm._Table) && !strings.Contains(orm.Query, failString) {
			orm.Query = fmt.Sprintf("%s (%s", orm.Query, name)
		} else {
			orm.Query = fmt.Sprintf("%s, %s", orm.Query, name)
		}
	}

	return *orm
}

func (orm *Neorm) Type(typeVal string) Neorm {
	orm.Query = fmt.Sprintf("%s %s", orm.Query, strings.ToUpper(typeVal))

	return *orm
}

func (orm *Neorm) Null() Neorm {
	orm.Query = fmt.Sprintf("%s NULL", orm.Query)

	return *orm
}

func (orm *Neorm) NotNull() Neorm {
	orm.Query = fmt.Sprintf("%s NOT NULL", orm.Query)

	return *orm
}

func (orm *Neorm) AutoIncrement() Neorm {
	orm.Query = fmt.Sprintf("%s AUTO_INCREMENT", orm.Query)

	return *orm
}

func (orm *Neorm) Default(value interface{}) Neorm {
	splitTheQuery := strings.Split(orm.Query, ", ")

	lengthOfTheSplitTheQuery := len(splitTheQuery)

	splitTheSplittedQuery := strings.Split(splitTheQuery[lengthOfTheSplitTheQuery-1], " ")

	if splitTheSplittedQuery[1] == "INT" ||
		splitTheSplittedQuery[1] == "TINYINT" ||
		splitTheSplittedQuery[1] == "SMALLINT" ||
		splitTheSplittedQuery[1] == "MEDIUMINT" ||
		splitTheSplittedQuery[1] == "BIGINT" ||
		splitTheSplittedQuery[1] == "BIT" {
		switch t := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			orm.Query = fmt.Sprintf("%s DEFAULT %d", orm.Query, t)
		default:
			panic("You cannot give a non integer go value to a column if it's a mysql integer variant.")
		}
	}

	if splitTheSplittedQuery[1] == "BOOL" ||
		splitTheSplittedQuery[1] == "BOOLEAN" {
		switch t := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			if t != 0 && t != 1 {
				panic("You cannot give a default value none other than 1 or 0 if a mysql column has the Boolean type")
			} else {
				orm.Query = fmt.Sprintf("%s DEFAULT %d", orm.Query, t)
			}
		case bool:
			orm.Query = fmt.Sprintf("%s DEFAULT %v", orm.Query, t)
		default:
			panic("You cannot give any other default value to a boolean column none other than an integer type or boolean")
		}
	}

	splitTheColumnType := strings.Split(splitTheSplittedQuery[1], "(")[0]

	if splitTheColumnType == "CHAR" ||
		splitTheColumnType == "VARCHAR" ||
		splitTheColumnType == "TEXT" ||
		splitTheColumnType == "TINYTEXT" ||
		splitTheColumnType == "MEDIUMTEXT" ||
		splitTheColumnType == "LONGTEXT" ||
		splitTheColumnType == "BINARY" ||
		splitTheColumnType == "VARBINARY" ||
		splitTheColumnType == "DATETIME" ||
		splitTheColumnType == "TIMESTAMP" {
		switch t := value.(type) {
		case string, map[string]interface{}:
			orm.Query = fmt.Sprintf("%s DEFAULT %s", orm.Query, t)
		default:
			panic("You cannot give any other default value than string or json if your column is a mysql string variant.")
		}
	}

	return *orm
}

func (orm *Neorm) Unique() Neorm {
	orm.Query = fmt.Sprintf("%s UNIQUE", orm.Query)

	return *orm
}

func (orm *Neorm) Check(condition string) Neorm {
	orm.Query = fmt.Sprintf("%s CHECK (%s)", orm.Query, condition)

	return *orm
}

func (orm *Neorm) CharacterSet(characterSet string) Neorm {
	orm.Query = fmt.Sprintf("%s CHARACTER SET %s", orm.Query, characterSet)

	return *orm
}

func (orm *Neorm) PrimaryKey() Neorm {
	if strings.Contains(orm.Query, "PRIMARY KEY") {
		panic("A table cannot has two primary key")
	}

	orm.Query = fmt.Sprintf("%s PRIMARY KEY", orm.Query)

	return *orm
}

func (orm *Neorm) ForeignKey(column string, referenceStruct interface{}) Neorm {
	if strings.HasPrefix(orm.Query, "ALTER TABLE") {
		orm.Query = fmt.Sprintf("%s ADD FOREIGN KEY (%s)", orm.Query, column)
	} else {
		orm.Query = fmt.Sprintf("%s, FOREIGN KEY (%s)", orm.Query, column)
	}

	referencesValues := reflect.ValueOf(referenceStruct)
	if referencesValues.Kind() != reflect.Struct {
		panic("The references of foreign keys must be a struct.")
	}

	referencesFields := referencesValues.Type()

	for i := 0; i < referencesValues.NumField(); i++ {
		fieldValue := referencesValues.Field(i).Interface()
		fieldName := referencesFields.Field(i).Name

		switch t := fieldValue.(type) {
		case string:
			orm.Query = fmt.Sprintf("%s REFERENCES %s(%s)", orm.Query, strings.ToLower(fieldName), t)
		default:
			panic("Any values of referencesStruct argument cannot be other than string.")
		}
	}

	return *orm
}

func (orm *Neorm) ForeignKeyWithConstraint(constraint, column string, referenceStruct interface{}) Neorm {
	if strings.HasPrefix(orm.Query, "ALTER TABLE") {
		orm.Query = fmt.Sprintf("%s ADD CONSTRAINT %s FOREIGN KEY (%s)", orm.Query, constraint, column)
	} else {
		orm.Query = fmt.Sprintf("%s, CONSTRAINT %s FOREIGN KEY (%s)", orm.Query, constraint, column)
	}

	referencesValues := reflect.ValueOf(referenceStruct)
	if referencesValues.Kind() != reflect.Struct {
		panic("The references of foreign keys must be a struct.")
	}

	referencesFields := referencesValues.Type()

	for i := 0; i < referencesValues.NumField(); i++ {
		fieldValue := referencesValues.Field(i).Interface()
		fieldName := referencesFields.Field(i).Name

		switch t := fieldValue.(type) {
		case string:
			orm.Query = fmt.Sprintf("%s REFERENCES %s(%s)", orm.Query, fieldName, t)
		default:
			panic("Any values of referencesStruct argument cannot be other than string.")
		}
	}

	return *orm
}

func (orm *Neorm) Unsigned() Neorm {
	orm.Query = fmt.Sprintf("%s UNSIGNED", orm.Query)

	return *orm
}

func (orm *Neorm) Zerofill() Neorm {
	orm.Query = fmt.Sprintf("%s ZEROFILL", orm.Query)

	return *orm
}

func (orm *Neorm) Enum(values []string) Neorm {
	orm.Query = fmt.Sprintf("%s ENUM(", orm.Query)

	lengthOfValues := len(values)

	for i, value := range values {
		if i+1 == lengthOfValues {
			orm.Query = fmt.Sprintf("%s'%s')", orm.Query, value)
		} else {
			orm.Query = fmt.Sprintf("%s'%s', ", orm.Query, value)
		}
	}

	return *orm
}

func (orm *Neorm) OnUpdate(newValue string) Neorm {
	orm.Query = fmt.Sprintf("%s ON UPDATE %s", orm.Query, newValue)

	return *orm
}

func (orm *Neorm) OnDelete(newValue string) Neorm {
	orm.Query = fmt.Sprintf("%s ON DELETE %s", orm.Query, newValue)

	return *orm
}

func (orm *Neorm) GeneratedAlways(condition string) Neorm {
	orm.Query = fmt.Sprintf("%s GENERATED ALWAYS AS %s", orm.Query, condition)

	return *orm
}

func (orm *Neorm) Virtual() Neorm {
	orm.Query = fmt.Sprintf("%s VIRTUAL", orm.Query)

	return *orm
}

func (orm *Neorm) Stored() Neorm {
	orm.Query = fmt.Sprintf("%s STORED", orm.Query)

	return *orm
}

func (orm *Neorm) Spatial() Neorm {
	orm.Query = fmt.Sprintf("%s SPATIAL", orm.Query)

	return *orm
}

func (orm *Neorm) Generated() Neorm {
	orm.Query = fmt.Sprintf("%s GENERATED", orm.Query)

	return *orm
}

func (orm *Neorm) Index(index interface{}) Neorm {
	switch t := index.(type) {
	case string:
		orm.Query = fmt.Sprintf("%s, INDEX (%s)", orm.Query, t)
	case []string:
		lengthOfT := len(t)

		for i, value := range t {
			if i+1 == lengthOfT {
				orm.Query = fmt.Sprintf("%s%s)", orm.Query, value)

				continue
			}

			if i == 0 {
				orm.Query = fmt.Sprintf("%s, INDEX (%s, ", orm.Query, value)

				continue
			}

			orm.Query = fmt.Sprintf("%s%s, ", orm.Query, value)
		}
	}

	return *orm
}

func (orm *Neorm) Comment(comment string) Neorm {
	orm.Query = fmt.Sprintf("%s COMMENT '%s'", orm.Query, comment)

	return *orm
}

func (orm *Neorm) DefaultOnNull(value interface{}) Neorm {
	switch t := value.(type) {
	case string:
		orm.Query = fmt.Sprintf("%s DEFAULT '%s' ON NULL", orm.Query, t)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		orm.Query = fmt.Sprintf("%s DEFAULT %d ON NULL", orm.Query, t)
	case float32, float64:
		orm.Query = fmt.Sprintf("%s DEFAULT %f ON NULL", orm.Query, t)
	case bool:
		orm.Query = fmt.Sprintf("%s DEFAULT %v ON NULL", orm.Query, t)
	}

	return *orm
}

func (orm *Neorm) Invisible() Neorm {
	orm.Query = fmt.Sprintf("%s INVISIBLE", orm.Query)

	return *orm
}

func (orm *Neorm) CustomKeyword(keywordAndValue string) Neorm {
	orm.Query = fmt.Sprintf("%s %s", orm.Query, keywordAndValue)

	return *orm
}

// altering functions for columns:

func (orm *Neorm) AlterTable(name string) Neorm {
	orm.Query = fmt.Sprintf("ALTER TABLE %s", name)

	return *orm
}

func (orm *Neorm) Add(something string) Neorm {
	orm.Query = fmt.Sprintf("%s ADD %s", orm.Query, something)

	return *orm
}

func (orm *Neorm) Drop(something string) Neorm {
	orm.Query = fmt.Sprintf("%s DROP %s", orm.Query, something)

	return *orm
}

func (orm *Neorm) ModifyColumn(column string) Neorm {
	orm.Query = fmt.Sprintf("%s MODIFY COLUMN %s", orm.Query, column)

	return *orm
}

func (orm *Neorm) ChangeColumn(oldColumn, newColumn string) Neorm {
	orm.Query = fmt.Sprintf("%s CHANGE COLUMN %s %s", orm.Query, oldColumn, newColumn)

	return *orm
}

func (orm *Neorm) After(columnFromAfter string) Neorm {
	orm.Query = fmt.Sprintf("%s AFTER %s", orm.Query, columnFromAfter)

	return *orm
}

func (orm *Neorm) First() Neorm {
	orm.Query = fmt.Sprintf("%s FIRST", orm.Query)

	return *orm
}

func (orm *Neorm) DropColumn(column string) Neorm {
	orm.Query = fmt.Sprintf("%s DROP COLUMN %s", orm.Query, column)

	return *orm
}

func (orm *Neorm) AddIndex(indexName, column string) Neorm {
	orm.Query = fmt.Sprintf("%s ADD INDEX %s (%s)", orm.Query, indexName, column)

	return *orm
}

func (orm *Neorm) AddUniqueIndex(indexName, column string) Neorm {
	orm.Query = fmt.Sprintf("%s ADD UNIQUE INDEX %s (%s)", orm.Query, indexName, column)

	return *orm
}

func (orm *Neorm) DropIndex(index string) Neorm {
	orm.Query = fmt.Sprintf("%s DROP INDEX %s", orm.Query, index)

	return *orm
}

func (orm *Neorm) AddPrimaryKey(column string) Neorm {
	orm.Query = fmt.Sprintf("%s ADD PRIMARY KEY (%s)", orm.Query, column)

	return *orm
}

func (orm *Neorm) DropPrimaryKey() Neorm {
	orm.Query = fmt.Sprintf("%s DROP PRIMARY KEY", orm.Query)

	return *orm
}

func (orm *Neorm) DropForeingKey(foreignKey string) Neorm {
	orm.Query = fmt.Sprintf("%s DROP FOREIGN KEY %s", orm.Query, foreignKey)

	return *orm
}

func (orm *Neorm) RenameColumn(oldName, newName string) Neorm {
	orm.Query = fmt.Sprintf("%s RENAME COLUMN %s TO %s", orm.Query, oldName, newName)

	return *orm
}

func (orm *Neorm) RenameTable(newName string) Neorm {
	orm.Query = fmt.Sprintf("%s RENAME TO %s", orm.Query, newName)

	return *orm
}

func (orm *Neorm) AddConstraint(constraint string) Neorm {
	orm.Query = fmt.Sprintf("%s ADD CONSTRAINT %s", orm.Query, constraint)

	return *orm
}

func (orm *Neorm) DropConstraint(constraint string) Neorm {
	orm.Query = fmt.Sprintf("%s DROP CONSTRAINT %s", orm.Query, constraint)

	return *orm
}

func (orm *Neorm) AddFulltextIndex(column string) Neorm {
	orm.Query = fmt.Sprintf("%s ADD FULLTEXT (%s)", orm.Query, column)

	return *orm
}

func (orm *Neorm) AddSpatialIndex(column string) Neorm {
	orm.Query = fmt.Sprintf("%s ADD SPATIAL INDEX (%s)", orm.Query, column)

	return *orm
}

func (orm *Neorm) DisableKeys() Neorm {
	orm.Query = fmt.Sprintf("%s DISABLE KEYS", orm.Query)

	return *orm
}

func (orm *Neorm) EnableKeys() Neorm {
	orm.Query = fmt.Sprintf("%s ENABLE KEYS", orm.Query)

	return *orm
}

func (orm *Neorm) Engine(engine string) Neorm {
	orm.Query = fmt.Sprintf("%s ENGINE = %s", orm.Query, engine)

	return *orm
}

// user actions:

func (orm *Neorm) CreateUser(name, scope string) Neorm {
	orm.Query = "CREATE USER"

	return *orm
}

func (orm *Neorm) UserInfos(username, scope, password string) Neorm {
	orm._User = username
	orm._Password = password
	orm._Scope = scope
	orm.Query = fmt.Sprintf("%s '%s'@'%s' IDENTIFIED BY %s", orm.Query, username, scope, password)

	return *orm
}

func (orm *Neorm) GrantPrivileges(privileges interface{}, schema string) Neorm {
	orm.Query = "GRANT"

	switch t := privileges.(type) {
	case string:
		{
			switch t {
			case "all":
			case "ALL":
			case "All":
			case "*":
			case "":
				orm.Query = fmt.Sprintf("%s ALL PRIVILEGES", orm.Query)
			}

			if orm.Query == "GRANT ALL PRIVILEGES" {
				break
			}

			splitThePrivileges := strings.Split(t, " ")
			lengthOfThePrivileges := len(splitThePrivileges)

			for i, privilege := range splitThePrivileges {
				switch strings.ToUpper(privilege) {
				case "SELECT":
				case "INSERT":
				case "DELETE":
				case "UPDATE":
				case "CREATE":
				case "DROP":
				case "ALTER":
				case "GRANT OPTION":
					if i+1 != lengthOfThePrivileges {
						orm.Query = fmt.Sprintf("%s %s,", orm.Query, privilege)
					} else {
						orm.Query = fmt.Sprintf("%s %s", orm.Query, privilege)
					}
				default:
					privilegeString := fmt.Sprintf("The privilege type '%s' not supported.", privilege)
					panic(privilegeString)
				}
			}
		}
	case []string:
		{
			lengthOfThePrivileges := len(t)

			for i, privilege := range t {
				switch strings.ToUpper(privilege) {
				case "SELECT":
				case "INSERT":
				case "DELETE":
				case "UPDATE":
				case "CREATE":
				case "DROP":
				case "ALTER":
				case "GRANT OPTION":
					if i+1 != lengthOfThePrivileges {
						orm.Query = fmt.Sprintf("%s %s,", orm.Query, strings.Trim(privilege, " "))
					} else {
						orm.Query = fmt.Sprintf("%s %s", orm.Query, strings.Trim(privilege, " "))
					}
				default:
					privilegeString := fmt.Sprintf("The privilege type '%s' not supported.", privilege)
					panic(privilegeString)
				}
			}
		}
	default:
		panic("privileges has to be either string or string array")
	}

	orm.Query = fmt.Sprintf("%s ON %s TO '%s'@'%s' ", orm.Query, schema, orm._User, orm._Scope)

	return *orm
}

func (orm *Neorm) RevokePrivileges(privileges interface{}, schema string) Neorm {
	orm.Query = "REVOKE"

	switch t := privileges.(type) {
	case string:
		{
			switch t {
			case "all":
			case "ALL":
			case "All":
			case "*":
			case "":
				orm.Query = fmt.Sprintf("%s ALL PRIVILEGES", orm.Query)
			}

			if orm.Query == "REVOKE ALL PRIVILEGES" {
				break
			}

			splitThePrivileges := strings.Split(t, " ")
			lengthOfThePrivileges := len(splitThePrivileges)

			for i, privilege := range splitThePrivileges {
				switch strings.ToUpper(privilege) {
				case "SELECT":
				case "INSERT":
				case "DELETE":
				case "UPDATE":
				case "CREATE":
				case "DROP":
				case "ALTER":
				case "GRANT OPTION":
					if i+1 != lengthOfThePrivileges {
						orm.Query = fmt.Sprintf("%s %s,", orm.Query, privilege)
					} else {
						orm.Query = fmt.Sprintf("%s %s", orm.Query, privilege)
					}
				default:
					privilegeString := fmt.Sprintf("The privilege type '%s' not supported.", privilege)
					panic(privilegeString)
				}
			}
		}
	case []string:
		{
			lengthOfThePrivileges := len(t)

			for i, privilege := range t {
				switch strings.ToUpper(privilege) {
				case "SELECT":
				case "INSERT":
				case "DELETE":
				case "UPDATE":
				case "CREATE":
				case "DROP":
				case "ALTER":
				case "GRANT OPTION":
					if i+1 != lengthOfThePrivileges {
						orm.Query = fmt.Sprintf("%s %s,", orm.Query, strings.Trim(privilege, " "))
					} else {
						orm.Query = fmt.Sprintf("%s %s", orm.Query, strings.Trim(privilege, " "))
					}
				default:
					privilegeString := fmt.Sprintf("The privilege type '%s' not supported.", privilege)
					panic(privilegeString)
				}
			}
		}
	default:
		panic("privileges has to be either string or string array")
	}

	orm.Query = fmt.Sprintf("%s ON %s TO '%s'@'%s' ", orm.Query, schema, orm._User, orm._Scope)

	return *orm
}

func (orm *Neorm) ShowGrants() Neorm {
	orm.Query = fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", orm._User, orm._Scope)

	return *orm
}

func (orm *Neorm) SetPassword(password string) Neorm {
	orm.Query = fmt.Sprintf("SET PASSWORD FOR '%s'@'%s' = PASSWORD(%s)", orm._User, orm._Scope, password)

	return *orm
}

func (orm *Neorm) DropUser(user, scope string) Neorm {
	orm.Query = fmt.Sprintf("DROP USER '%s'@'%s'", user, scope)

	return *orm
}

func (orm *Neorm) AllUsers() Neorm {
	orm.Query = "SELECT user, host FROM mysql.user"

	return *orm
}

func (orm *Neorm) RenameUser(newName, newScope string) Neorm {
	orm.Query = fmt.Sprintf("RENAME USER '%s'@'%s' TO '%s'@'%s'", orm._User, orm._Scope, newName, newScope)

	return *orm
}

func (orm *Neorm) SetDefaultRole(role string) Neorm {
	orm.Query = fmt.Sprintf("SET DEFAULT ROLE '%s' FOR '%s'@'%s'", role, orm._User, orm._Scope)

	return *orm
}

func (orm *Neorm) FlushPrivileges() Neorm {
	orm.Query = "FLUSH PRIVILEGES"

	return *orm
}

func (orm *Neorm) LockUserAccount(user, scope string) Neorm {
	orm.Query = fmt.Sprintf("ALTER USER '%s'@'%s' ACCOUNT LOCK", user, scope)

	return *orm
}

func (orm *Neorm) PasswordExpiration(user, scope, expirationStr string) Neorm {
	orm.Query = fmt.Sprintf("ALTER USER '%s'@'%s' PASSWORD EXPIRE %s", user, scope, expirationStr)

	return *orm
}

// query builder:

func (orm *Neorm) Select(columns interface{}) Neorm {
	query := ""
	orm._Table = ""
	orm.Query = ""
	orm._Type = "s"

	switch t := columns.(type) {
	case string:
		if columns != "*" {
			panic("If you want to use string in columns argument: it has to be '*'")
		}

		query = "SELECT * FROM"
	case []string:
		query = "SELECT"

		for i, column := range t {
			if i+1 == len(t) {
				query = fmt.Sprintf("%s %s FROM", query, column)
			} else {
				query = fmt.Sprintf("%s %s,", query, column)
			}
		}
	}

	orm.Query = query

	return *orm
}

func (orm *Neorm) Insert(columns []string, values interface{}) Neorm {
	orm._Table = ""
	orm.Query = ""
	orm._Type = "i"
	columnValues := "("

	lengthOfColumns := len(columns)
	for i, column := range columns {
		if i+1 != lengthOfColumns {
			columnValues = fmt.Sprintf("%s%s, ", columnValues, column)
		} else {
			columnValues = fmt.Sprintf("%s%s)", columnValues, column)
		}
	}

	newValues := "("

	if slice, ok := values.([]interface{}); ok {
		lengthOfValues := len(slice)

		for i, value := range slice {
			if i+1 != lengthOfValues {
				switch t := value.(type) {
				case string, map[string]interface{}:
					newValues = fmt.Sprintf("%s'%s', ", newValues, t)
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					newValues = fmt.Sprintf("%s%d, ", newValues, t)
				case float32, float64:
					newValues = fmt.Sprintf("%s%f, ", newValues, t)
				case bool:
					newValues = fmt.Sprintf("%s%v, ", newValues, t)
				}
			} else {
				switch t := value.(type) {
				case string, map[string]interface{}:
					newValues = fmt.Sprintf("%s'%s')", newValues, t)
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					newValues = fmt.Sprintf("%s%d)", newValues, value)
				case float32, float64:
					newValues = fmt.Sprintf("%s%f)", newValues, value)
				case bool:
					newValues = fmt.Sprintf("%s%v)", newValues, value)
				}
			}
		}
	} else {
		panic("values argument should be a slice.")
	}

	orm.Query = fmt.Sprintf("INSERT INTO %s VALUES %s", columnValues, newValues)

	return *orm
}

func (orm *Neorm) Update() Neorm {
	orm._Table = ""
	orm._Type = "u"
	orm.Query = "UPDATE"

	return *orm
}

func (orm *Neorm) Delete() Neorm {
	orm._Table = ""
	orm._Type = "u"
	orm.Query = "DELETE FROM"

	return *orm
}

func (orm *Neorm) Table(table string) Neorm {
	if strings.HasPrefix(orm.Query, "INSERT INTO") {
		splittedString := strings.Split(orm.Query, " INTO ")

		orm.Query = fmt.Sprintf("%s INTO %s %s", splittedString[0], table, splittedString[1])
	} else {
		orm.Query = fmt.Sprintf("%s %s", orm.Query, table)
	}

	return *orm
}

func (orm *Neorm) Where(column, mark string, value interface{}) Neorm {
	switch t := value.(type) {
	case string, map[string]interface{}:
		orm.Query = fmt.Sprintf("%s WHERE %s %s '%s'", orm.Query, column, mark, t)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		orm.Query = fmt.Sprintf("%s WHERE %s %s %d", orm.Query, column, mark, t)
	case float32, float64:
		orm.Query = fmt.Sprintf("%s WHERE %s %s %f", orm.Query, column, mark, t)
	case bool:
		orm.Query = fmt.Sprintf("%s WHERE %s %s %v", orm.Query, column, mark, t)
	default:
		orm.Query = fmt.Sprintf("%s WHERE %s %s %s", orm.Query, column, mark, t)
	}

	return *orm
}

func (orm *Neorm) Or(column, mark string, value interface{}) Neorm {
	switch t := value.(type) {
	case string, map[string]interface{}:
		orm.Query = fmt.Sprintf("%s OR %s %s '%s'", orm.Query, column, mark, t)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		orm.Query = fmt.Sprintf("%s OR %s %s %d", orm.Query, column, mark, t)
	case float32, float64:
		orm.Query = fmt.Sprintf("%s OR %s %s %f", orm.Query, column, mark, t)
	case bool:
		orm.Query = fmt.Sprintf("%s OR %s %s %v", orm.Query, column, mark, t)
	default:
		orm.Query = fmt.Sprintf("%s OR %s %s %s", orm.Query, column, mark, t)
	}

	return *orm
}

func (orm *Neorm) And(column, mark string, value interface{}) Neorm {
	switch t := value.(type) {
	case string, map[string]interface{}:
		orm.Query = fmt.Sprintf("%s AND %s %s '%s'", orm.Query, column, mark, t)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		orm.Query = fmt.Sprintf("%s AND %s %s %d", orm.Query, column, mark, t)
	case float32, float64:
		orm.Query = fmt.Sprintf("%s AND %s %s %f", orm.Query, column, mark, t)
	case bool:
		orm.Query = fmt.Sprintf("%s AND %s %s %v", orm.Query, column, mark, t)
	default:
		orm.Query = fmt.Sprintf("%s AND %s %s %s", orm.Query, column, mark, t)
	}

	return *orm
}

func (orm *Neorm) Set(column string, value interface{}) Neorm {
	if strings.Contains(orm.Query, "SET") {
		switch t := value.(type) {
		case string, map[string]interface{}:
			orm.Query = fmt.Sprintf("%s, %s = '%s'", orm.Query, column, t)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			orm.Query = fmt.Sprintf("%s, %s = %d", orm.Query, column, t)
		case float32, float64:
			orm.Query = fmt.Sprintf("%s, %s = %f", orm.Query, column, t)
		case bool:
			orm.Query = fmt.Sprintf("%s, %s = %v", orm.Query, column, t)
		default:
			orm.Query = fmt.Sprintf("%s, %s = %s", orm.Query, column, t)
		}
	} else {
		switch t := value.(type) {
		case string, map[string]interface{}:
			orm.Query = fmt.Sprintf("%s SET %s = '%s'", orm.Query, column, t)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			orm.Query = fmt.Sprintf("%s SET %s = %d", orm.Query, column, t)
		case float32, float64:
			orm.Query = fmt.Sprintf("%s SET %s = %f", orm.Query, column, t)
		case bool:
			orm.Query = fmt.Sprintf("%s SET %s = %v", orm.Query, column, t)
		default:
			orm.Query = fmt.Sprintf("%s SET %s = %s", orm.Query, column, t)
		}
	}

	return *orm
}

func (orm *Neorm) Like(columns []string, operand string) Neorm {
	for i, column := range columns {
		if i == 0 {
			orm.Query = fmt.Sprintf("%s WHERE %s LIKE '", orm.Query, column)
			orm.Query = orm.Query + "%"
			orm.Query = orm.Query + operand
			orm.Query = orm.Query + "%"
			orm.Query = orm.Query + "'"
		} else {
			orm.Query = fmt.Sprintf("%s OR %s LIKE '", orm.Query, column)
			orm.Query = orm.Query + "%"
			orm.Query = orm.Query + operand
			orm.Query = orm.Query + "%"
			orm.Query = orm.Query + "'"
		}
	}

	return *orm
}

func (orm *Neorm) OrderBy(column, ordering string) Neorm {
	switch ordering {
	case "ASC", "Asc", "asc":
		orm.Query = fmt.Sprintf("%s ORDER BY %s ASC", orm.Query, column)
	case "DESC", "Desc", "desc":
		orm.Query = fmt.Sprintf("%s ORDER BY %s DESC", orm.Query, column)
	default:
		panic("Error on OrderBy method: ordering should be either ASC or DESC.")
	}

	return *orm
}

func (orm *Neorm) OrderRandom() Neorm {
	orm.Query = fmt.Sprintf("%s ORDER BY RAND()", orm.Query)

	return *orm
}

func (orm *Neorm) Length(table string) Neorm {
	orm.Query = fmt.Sprintf("SELECT COUNT(*) AS length FROM %s", table)

	return *orm
}

func (orm *Neorm) Limit(limit int) Neorm {
	orm.Query = fmt.Sprintf("%s LIMIT %d", orm.Query, limit)

	return *orm
}

func (orm *Neorm) Offset(offset int) Neorm {
	orm.Query = fmt.Sprintf("%s OFFSET %d", orm.Query, offset)

	return *orm
}

func (orm *Neorm) CustomQuery(query string) Neorm {
	orm.Query = query

	return *orm
}

func (orm *Neorm) Finish() Neorm {
	if strings.HasPrefix(orm.Query, "CREATE TABLE") {
		orm.Query = fmt.Sprintf("%s);", orm.Query)
	} else {
		orm.Query = fmt.Sprintf("%s;", orm.Query)
	}

	return *orm
}
