package repository

type SQLiteDatabase struct{}

func NewSqliteDatabase() *SQLiteDatabase {
	return &SQLiteDatabase{}
}
