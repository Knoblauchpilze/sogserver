package data

// getDuplicatedElementErrorKey :
// Used to retrieve a string describing part of the error
// message issued by the database when trying to insert a
// duplicated element on a unique column. Can be used to
// standardize the definition of this error.
//
// Return part of the error string issued when inserting
// an already existing key.
func getDuplicatedElementErrorKey() string {
	return "SQLSTATE 23505"
}

// getForeignKeyViolationErrorKey :
// Used to retrieve a string describing part of the error
// message issued by the database when trying to insert an
// element that does not match a foreign key constraint.
// Can be used to standardize the definition of this error.
//
// Return part of the error string issued when violating a
// foreign key constraint.
func getForeignKeyViolationErrorKey() string {
	return "SQLSTATE 23503"
}
