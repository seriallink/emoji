package emoji

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEmojiless(t *testing.T) {

	emojitext := "I'm a 😀 person!"

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	emojiless, err := emojer.Emojiless(emojitext)
	assert.NoError(t, err, "Emojiless method cannot return an error")
	assert.Equal(t, "I'm a :grinning_face: person!", emojiless)

	err = emojer.Close()
	assert.NoError(t, err, "Close method cannot return an error")

}

func TestEmojiness(t *testing.T) {

	emojitext := "We are a :grinning_face: :family_man_woman_girl_girl:!"
	emojifake := ":this_is: :not_an_emoji:, :do_not: : convert_it:!"

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	emojiless, err := emojer.Emojiness(emojitext)
	assert.NoError(t, err, "Emojiness method cannot return an error")
	assert.Equal(t, "We are a 😀 👨‍👩‍👧‍👧!", emojiless)

	emojiless, err = emojer.Emojiness(emojifake)
	assert.NoError(t, err, "Emojiness method cannot return an error")
	assert.Equal(t, emojifake, emojiless)

	err = emojer.Close()
	assert.NoError(t, err, "Close method cannot return an error")

}

func TestGetByUnicode(t *testing.T) {

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	row, err := emojer.Get("ucode","1F37A")
	assert.NoError(t, err, "GetByUnicode method cannot return an error")
	assert.Equal(t, "🍺", row.Emoji)
	assert.Equal(t, ":beer_mug:", row.Alias)

	err = emojer.Close()
	assert.NoError(t, err, "Close method cannot return an error")

}

func TestGetByAlias(t *testing.T) {

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	row, err := emojer.Get("alias",":beer_mug:")
	assert.NoError(t, err, "GetByAlias method cannot return an error")
	assert.Equal(t, "🍺", row.Emoji)
	assert.Equal(t, "1F37A", row.Unicode)

	err = emojer.Close()
	assert.NoError(t, err, "Close method cannot return an error")

}

func TestAll(t *testing.T) {

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	rows, err := emojer.All("alias")
	assert.NoError(t, err, "All method cannot return an error")

	type result struct {
		ok bool
		expected string
		actual string
	}

	semaphore := make(chan *result,len(rows))

	for i:=0; i<len(rows); i++ {

		go func(r *Row) {

			emojer, _ := New()
			emojiless, _ := emojer.Emojiless(r.Emoji)

			semaphore <- &result {
				ok: assert.Equal(t, r.Alias, emojiless),
				expected: r.Alias,
				actual: r.Alias,
			}

		}(&rows[i])

	}

	for i:=0; i<len(rows); i++ {
		if r := <-semaphore; !r.ok {
			assert.Fail(t, fmt.Sprintf("Expected[%s] | Actual[%s]", r.expected, r.actual))
		}
	}

	err = emojer.Close()
	assert.NoError(t, err, "Close method cannot return an error")

}