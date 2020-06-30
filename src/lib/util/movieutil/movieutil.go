package movieutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/models"
	"github.com/dsbezerra/amenic-lambda/src/lib/util/stringutil"
)

// ShouldUpdate ...
func ShouldUpdate(src, test *models.Movie) (bool, models.Movie) {
	should := false
	result := *src

	// NOTE: If these IDs change we will need to implement some routine
	// that checks if the new ID match the movie.
	if src.TmdbID == 0 && test.TmdbID != 0 {
		result.TmdbID = test.TmdbID
		should = true
	}

	if src.ImdbID == "" && test.ImdbID != "" {
		result.ImdbID = test.ImdbID
		should = true
	}

	if src.ClaqueteID == 0 && test.ClaqueteID != 0 {
		result.ClaqueteID = test.ClaqueteID
		should = true
	}

	// NOTE: These two can always change because poster and backdrop images will never contain the same
	// address due to it being uploaded in Cloudinary whenever the admin system starts or we implement
	// the images service
	if src.BackdropURL == "" && test.BackdropURL != "" {
		result.BackdropURL = test.BackdropURL
		should = true
	}

	if src.PosterURL == "" && test.PosterURL != "" {
		result.PosterURL = test.PosterURL
		should = true
	}

	// Update slug if we don't have one or it changed.
	if !test.Slugs.IsEmpty() && !src.Slugs.IsEqual(test.Slugs) {
		result.Slugs = test.Slugs
		should = true
	}

	// Update release date if we don't have one or it changed.
	if src.ReleaseDate == nil && test.ReleaseDate != nil ||
		(test.ReleaseDate != nil && !test.ReleaseDate.IsZero() && test.ReleaseDate.Unix() != src.ReleaseDate.Unix()) {
		result.ReleaseDate = test.ReleaseDate
		should = true
	}

	// Update cast if we don't have one or it changed.
	if SliceCountDifferent(len(src.Cast), len(test.Cast)) {
		result.Cast = test.Cast
		should = true
	}

	// Update genres if we don't have one or it changed.
	if SliceCountDifferent(len(src.Genres), len(test.Genres)) {
		result.Genres = test.Genres
		should = true
	}

	// Update original title if we don't have one or it changed.
	if models.FlagIsNotSet(src.LockFlags, models.MovieLockOriginalTitle) {
		if StringIsNewOrChanged(src.OriginalTitle, test.OriginalTitle) {
			result.OriginalTitle = test.OriginalTitle
			should = true
		}
	}

	// Update title if we don't have one or it changed.
	if models.FlagIsNotSet(src.LockFlags, models.MovieLockTitle) {
		if StringIsNewOrChanged(src.Title, test.Title) {
			result.Title = test.Title
			should = true
		}
	}

	// Update synopsis if we don't have one or it changed.
	if models.FlagIsNotSet(src.LockFlags, models.MovieLockSynopsis) {
		if StringIsNewOrChanged(src.Synopsis, test.Synopsis) {
			result.Synopsis = test.Synopsis
			should = true
		}
	}

	// Update trailer if we don't have one or it changed.
	if StringIsNewOrChanged(src.Trailer, test.Trailer) {
		result.Trailer = test.Trailer
		should = true
	}

	// Update distributor if we don't have one or it changed.
	if StringIsNewOrChanged(src.Distributor, test.Distributor) {
		result.Distributor = test.Distributor
		should = true
	}

	// Update runtime if we don't have one or it changed.
	if IntIsNewOrChanged(src.Runtime, test.Runtime) {
		result.Runtime = test.Runtime
		should = true
	}

	// Update rating if we don't have one or it changed.
	if IntIsNewOrChanged(src.Rating, test.Rating) {
		result.Rating = test.Rating
		should = true
	}

	return should, result
}

// SliceCountDifferent ...
func SliceCountDifferent(src, test int) bool {
	return test > 0 && src != test
}

// IntIsNew checks if the test int is new relative to src int
// ex: zero src and non-zero test means the content is new
func IntIsNew(src, test int) bool {
	return src == 0 && test != 0
}

// IntIsNewOrChanged checks if the test int is new relative to src
// int or it has changed
// ex: empty or non-zero src and non-zero test not equals to src
// means the content has updated
func IntIsNewOrChanged(src, test int) bool {
	return IntIsNew(src, test) || (test != 0 && src != test)
}

// StringIsNew checks if the test string is new relative to src string
// ex: empty src and non-empty test means the content is new
func StringIsNew(src, test string) bool {
	return src == "" && test != ""
}

// StringIsNewOrChanged checks if the test string is new relative to src
// string or it has changed
// ex: empty or non-empty src and non-empty test not equals to src
// means the content has updated
func StringIsNewOrChanged(src, test string) bool {
	return StringIsNew(src, test) || (test != "" && src != test)
}

// UnromanTitle this looks for a roman numeral and convert to arabic.
func UnromanTitle(title string) string {

	var result strings.Builder

	// Most common in titles.
	romap := map[string]string{
		"I":    "1",
		"II":   "2",
		"III":  "3",
		"IV":   "4",
		"V":    "5",
		"VI":   "6",
		"VII":  "7",
		"VIII": "8",
		"IX":   "9",
		"X":    "10",
		"XI":   "11",
		"XII":  "12",
		"XIII": "13",
	}

	convert := func(start int) strings.Builder {
		b := strings.Builder{}
		numeral, remainder := stringutil.BreakBySpaces(title[start:])
		n, ok := romap[numeral]
		if ok {
			b.WriteString(title[0 : start-1])
			b.WriteString(" ")
			b.WriteString(n)
			if remainder != "" {
				b.WriteString(" ")
				b.WriteString(remainder)
			}
			fmt.Println(b.String())
		}
		return b
	}

	i := 0
	processing := true
	for processing {

		if i >= len(title) {
			break
		}

		r := title[i]
		switch r {
		case 'I':
			result = convert(i)
			processing = false
			break
		case 'V':
			result = convert(i)
			processing = false
			break
		case 'X':
			result = convert(i)
			processing = false
			break
		}

		i++
	}

	return result.String()
}

// GenerateSlugs ...
func GenerateSlugs(m *models.Movie) (*models.Slugs, error) {
	if m == nil {
		return nil, errors.New("invalid movie")
	}

	str := m.Title
	size := len(str)
	if size == 0 {
		return nil, errors.New("empty string")
	}

	var noDashes strings.Builder
	var dashes strings.Builder

	start := 0
	end := size - 1

	// Skip left spaces
	for start < size {
		if !stringutil.IsWhitespace(str[start]) {
			break
		}
		start++
	}

	// Skip right spaces
	for end >= start {
		if !stringutil.IsWhitespace(str[end]) {
			break
		}
		end--
	}

	i := start
	for i <= end {
		c := str[i]
		i++

		if stringutil.IsWhitespace(c) {
			curr := dashes.String()
			currSize := len(curr)
			end := currSize - 1
			if end >= 0 {
				if curr[end] != 45 {
					dashes.WriteString("-")
				}
			}
			continue
		}

		// We add only if it is a-z0-9
		if stringutil.IsAlpha(c) || stringutil.IsCharacter(c) {
			if stringutil.IsUppercase(c) {
				c += 32
			}
			noDashes.WriteByte(c)
			dashes.WriteByte(c)
		}
	}

	noDashesStr := noDashes.String()
	dashesStr := dashes.String()
	slugs := &models.Slugs{
		NoDashes: noDashesStr,
		Dashes:   dashesStr,
	}

	if m.ReleaseDate != nil && !m.ReleaseDate.IsZero() {
		y := m.ReleaseDate.Year()
		slugs.Year = fmt.Sprintf("%s%d", slugs.NoDashes, y)
		slugs.DashesAndYear = fmt.Sprintf("%s-%d", slugs.Dashes, y)
	}

	return slugs, nil
}

// FillSlugs ...
func FillSlugs(movie *models.Movie) {
	// NOTE(diego):
	// Since slugs depends on title we need to make sure our title is correct.
	CorrectTitle(movie)
	slugs, _ := GenerateSlugs(movie)
	if slugs != nil {
		movie.Slugs = *slugs
	}
}

// GenerateSlug generate a movie slug separated by hyphens.
// NOTE(diego): We don't care about accents.
func GenerateSlug(str string, withDashes bool) string {
	size := len(str)
	if size == 0 {
		return ""
	}

	var b strings.Builder

	start := 0
	end := size - 1

	// Skip left spaces
	for start < size {
		if !stringutil.IsWhitespace(str[start]) {
			break
		}
		start++
	}

	// Skip right spaces
	for end >= start {
		if !stringutil.IsWhitespace(str[end]) {
			break
		}
		end--
	}

	i := start
	for i <= end {
		c := str[i]
		i++

		// NOTE(diego): We don't replace whitespaces with hyphen anymore due to typos or
		// missing whitespaces
		if stringutil.IsWhitespace(c) && withDashes {
			// Add hyphen in the place of whitespaces making sure we have only one hyphen
			// between words.
			curr := b.String()
			currSize := len(curr)
			end := currSize - 1
			if end >= 0 {
				if curr[end] != 45 {
					b.WriteString("-")
				}
			}
			continue
		}

		// We add only if it is a-z0-9
		if stringutil.IsAlpha(c) || stringutil.IsCharacter(c) {
			if stringutil.IsUppercase(c) {
				c += 32
			}
			b.WriteByte(c)
		}
	}

	return b.String()
}

var prepList = []string{
	"e",
	"o", "os",
	"a", "as", "à", "às",
	"um", "uns", "uma", "umas",
	"de", "do", "dos", "da", "das", "dum", "duns", "duma", "dumas",
	"em", "no", "nos", "na", "nas", "num", "nuns", "numa", "numas",
	"por", "pelo", "pelos", "pela", "pelas",
}

func CapTitle(str string) string {
	result := strings.Builder{}

	size := len(str)
	if size == 0 {
		return ""
	}

	words := strings.Split(str, " ")
	for i, w := range words {

		cap := true
		if stringutil.Contains(prepList, w) {
			if i > 0 {
				prev := words[i-1]
				if !strings.HasSuffix(prev, ":") && !strings.HasSuffix(prev, "-") {
					cap = false
				}
			}
		}

		if cap {
			w = strings.Title(w)
		}

		result.WriteString(w)

		if i < len(words)-1 {
			result.WriteString(" ")
		}
	}

	return result.String()
}

// CorrectTitleMap TODO(diego): Upload a txt file to S3, download it and construct a map mapping lowercased title to correct title.
// Or use a third-party service (TMDb or Claquete) to find correct movie title for the given title.
var CorrectTitleMap = map[string]string{
	"arlequina em aves de rapina": "Aves de Rapina",
	"frozenii":                    "Frozen 2",
}

// CorrectTitle is used to correct any title that may be wrong.
// Ex: Arlequina em Aves de Rapina is a title used by theaters, but not official, then we need to change it to,
// correct one which is: Aves de Rapina
func CorrectTitle(movie *models.Movie) {
	title := stringutil.ToLowerTrimmed(movie.Title)
	if title == "" {
		return
	}
	correct, ok := CorrectTitleMap[title]
	if ok {
		movie.Title = correct
	}
}
