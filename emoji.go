package emoji

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
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

	// check if db file exists
	if _, err := os.Stat(dbname); os.IsNotExist(err) {
		Load(dbname)
	}

	// open db in read-only mode
	e.db, err = bolt.Open(dbname, os.ModePerm, &bolt.Options{
		ReadOnly: true,
		Timeout: time.Second,
	})

	// and create a read-only transaction
	if err == nil {
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

func (e *Emojer) All(bucket string) (rows []Row, err error) {

	// open bucket cursor
	cursor := e.tx.Bucket([]byte(bucket)).Cursor()

	// iterate over rows
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {

		// init row
		row := Row{}

		// unmarshal emoji
		if err = json.Unmarshal(v,&row); err != nil {
			return
		}

		rows = append(rows,row)

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
		if regexp.MustCompile("[^a-zA-Z ]").MatchString(string(c)) {

			// append unicode value
			criteria = append(criteria,fmt.Sprintf("%U",c)[2:])

			// create prefix
			prefix := []byte(strings.Join(criteria," "))

			// search emoji
			if key, _ := cursor.Seek(prefix); key != nil && bytes.HasPrefix(key,prefix) {

				// get row
				row, _ := e.Get("ucode",string(prefix))

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
				criteria = []string{fmt.Sprintf("%U",c)[2:]}

				// set a new prefix
				prefix = []byte(strings.Join(criteria," "))

				// perform a new search
				if key, _ := cursor.Seek(prefix); key != nil && bytes.HasPrefix(key,prefix) {

					row, _ := e.Get("ucode",string(prefix))

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
