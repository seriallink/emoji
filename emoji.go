package emoji

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
	"github.com/boltdb/bolt"
)
const dbname = "emoji.db"

type Emojer struct {
	db *bolt.DB
	tx *bolt.Tx
}

type Row struct {
	Unicode   string   `json:"unicode"`
	Alias     string   `json:"alias"`
	Name      string   `json:"name"`
	Emoji     string   `json:"emoji"`
}

type Replacer struct {
	Index     int
	Length    int
	Value     string
}

func New() (e Emojer, err error) {

	// check if db exists
	if _, err := os.Stat(dbname); os.IsNotExist(err) {
		Load(dbname)
	}

	// open db
	if e.db, err = bolt.Open(dbname, 0600, nil); err == nil {

		// and create a read-only transaction
		e.tx, err = e.db.Begin(false)

	}

	return

}

func (e *Emojer) Close() error {
	return e.db.Close()
}

func (e *Emojer) Get(bucket, key string) (row Row, err error) {

	// open bucket
	b := e.tx.Bucket([]byte(bucket))

	// get value
	if value := b.Get([]byte(key)); value != nil {

		// unmarshal emoji
		err = json.Unmarshal(value,&row)

	}

	return

}

func (e *Emojer) Emojiless(emojiness string) (emojiless string, err error){

	// copy emoji text
	if emojiless = emojiness; emojiless == "" {
		return
	}

	// create bucket cursor
	cursor := e.tx.Bucket([]byte("ucode")).Cursor()

	// criteria to seek for
	var criteria []string

	// it will be used to replace emojis
	var replacers []Replacer

	// control the number of occurrences found
	count := 0

	// loop over text
	for i, c := range emojiness {

		// check if rune is a possible emoji
		if unicode.IsSymbol(c) || unicode.Is(unicode.Join_Control,c) {

			// append unicode value
			criteria = append(criteria,fmt.Sprintf("%X",c))

			// create prefix
			prefix := []byte(strings.Join(criteria," "))

			// search emoji
			if key, value := cursor.Seek(prefix); key != nil && bytes.HasPrefix(key,prefix) {

				// init emoji db row
				row := &Row{}

				// unmarshal value
				if err = json.Unmarshal(value,row); err != nil {
					return
				}

				// set replacer
				if count == 0 {
					replacers = append(replacers,Replacer{Index:i,Length:len(row.Emoji),Value:row.Alias})
				} else {
					replacers[len(replacers)-1].Length = len(row.Emoji)
					replacers[len(replacers)-1].Value = row.Alias
				}

				// check for more occurrences
				for ; key != nil && bytes.HasPrefix(key,prefix); key, _ = cursor.Next() {
					if count++; count > 1 {
						break
					}
				}

			} else {

				// start a new criteria
				criteria = []string{fmt.Sprintf("%X",c)}

				// set a new prefix
				prefix = []byte(strings.Join(criteria," "))

				// perform a new search
				if key, value := cursor.Seek(prefix); key != nil && bytes.HasPrefix(key,prefix) {

					// init emoji db row
					row := &Row{}

					// unmarshal value
					if err = json.Unmarshal(value,row); err != nil {
						return
					}

					// save the first occurrence
					replacers = append(replacers,Replacer{Index:i,Length:len(row.Emoji),Value:row.Alias})

				}

			}

		} else {

			// reset counter
			count = 0

			// reset criteria
			criteria = []string{}

		}

	}

	// replace emojis
	for i:=len(replacers)-1; i>=0; i-- {
		emojiless = emojiless[0:replacers[i].Index] + replacers[i].Value + emojiless[replacers[i].Index+replacers[i].Length:]
	}

	return

}

func (e *Emojer) Emojiness(emojiless string) (string, error){

	// validate text
	if emojiless == "" {
		return "", nil
	}

	// it will be used in NewReplacer
	var pairs []string

	// create regex pattern
	regex := regexp.MustCompile(`:{1}\w+:{1}`)

	// match emoji aliases
	matches := regex.FindAllString(emojiless,-1)

	for _, match := range matches {

		// get emoji by its alias
		row, err := e.Get("alias",match)
		if err != nil {
			return "", err
		}

		// is it a valid alias?
		if row.Alias == match {

			// create replace pair
			pairs = append(pairs, match, row.Emoji)

		}

	}

	// replace all pairs
	emojiness := strings.NewReplacer(pairs...).Replace(emojiless)

	return emojiness, nil

}
