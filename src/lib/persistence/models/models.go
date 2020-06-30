package models

// Slugs holds many types of slugs of a string
type Slugs struct {
	NoDashes      string `json:"noDashes,omitempty" bson:"noDashes,omitempty"`           // thisisaslugwithnodashes
	Dashes        string `json:"dashes,omitempty" bson:"dashes,omitempty"`               // this-is-a-slug-with-dashes
	Year          string `json:"year,omitempty" bson:"year,omitempty"`                   // thisisaslugwithyear2019
	DashesAndYear string `json:"dashesAndYear,omitempty" bson:"dashesAndYear,omitempty"` // this-is-a-slug-with-dashes-and-year-2019
}

// IsEmpty ...
func (s *Slugs) IsEmpty() bool {
	return s.Dashes == "" &&
		s.DashesAndYear == "" &&
		s.NoDashes == "" &&
		s.Year == ""
}

// IsEqual ...
func (s *Slugs) IsEqual(test Slugs) bool {
	return s.Dashes == test.Dashes &&
		s.DashesAndYear == test.DashesAndYear &&
		s.NoDashes == test.NoDashes &&
		s.Year == test.Year
}

// FlagIsSet checks if a given flag is set in flags
func FlagIsSet(flags uint64, flag uint64) bool {
	return flags&flag != 0
}

// FlagIsNotSet checks if a given flag is not in flags
func FlagIsNotSet(flags uint64, flag uint64) bool {
	return !FlagIsSet(flags, flag)
}
