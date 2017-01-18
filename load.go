package emoji

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"github.com/boltdb/bolt"
)

const dataUrl = "http://unicode.org/Public/emoji/5.0/emoji-test.txt"

func CleanAlias(data, replacer string) string {
	data = strings.Replace(data, "u.s.", "us", -1)
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")
	rep := reg.ReplaceAllString(data, replacer)
	return rep
}

func NoExtraSpaces(data string) string {
	reg, _ := regexp.Compile(" {2,}")
	rep := reg.ReplaceAllString(data, " ")
	return rep
}

func Load(dbname string){

	// open bolt db
	db, err := bolt.Open(dbname, os.ModePerm, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// start a writable transaction
	tx, err := db.Begin(true)
	if err != nil {
		panic(err)
	}

	// refresh buckets
	for _, name := range []string{"ucode","alias"} {
		if tx.Bucket([]byte(name)) != nil {
			if err := tx.DeleteBucket([]byte(name)); err != nil {
				panic(err)
			}
		}
	}

	// create bucket using unicode as key
	ucode, err := tx.CreateBucket([]byte("ucode"))
	if err != nil {
		panic(err)
	}

	// create bucket using alias as key
	alias, err := tx.CreateBucket([]byte("alias"))
	if err != nil {
		panic(err)
	}

	// get emoji data
	res, err := http.Get(dataUrl)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// scan emoji data
	scanner := bufio.NewScanner(res.Body)

	// loop over lines
	for scanner.Scan() {

		// get line
		line := scanner.Text()

		// ignore comment line
		if line != "" && line[:1] != "#" {

			// init row
			row := &Row{}

			// dismember emoji information
			data := strings.Split(line,";")

			// get unicode
			row.Unicode = strings.TrimRight(data[0]," ")

			// get emoji details
			detail := strings.Split(strings.Split(data[1],"# ")[1], " ")

			// get emoji itself
			row.Emoji = strings.TrimRight(detail[0],"️")

			// get emoji name
			row.Name = NoExtraSpaces(CleanAlias(strings.ToLower(strings.Join(detail[1:]," "))," "))

			// create alias
			row.Alias = strings.Replace(fmt.Sprintf(":%s:",row.Name), " ", "_", -1)

			// marshal row
			value, _ := json.Marshal(row)

			// index emoji by unicode
			if err := ucode.Put([]byte(row.Unicode),value); err != nil {
				panic(err)
			}

			// index emoji by alias
			if err := alias.Put([]byte(row.Alias),value); err != nil {
				panic(err)
			}

			// fmt.Printf("key[%s] | emoji[%s] | name[%s] | alias[%s]\n", row.Unicode, row.Emoji, row.Name, row.Alias)

		}

	}

	// commit the transaction
	if err := tx.Commit(); err != nil {
		panic(err)
	}

}