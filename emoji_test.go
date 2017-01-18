package emoji

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEmojiless(t *testing.T) {

	emojitext := "🏈"

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	emojiless, err := emojer.Emojiless(emojitext)
	assert.NoError(t, err, "Emojiless method cannot return an error")
	assert.Equal(t, ":american_football:", emojiless)

	assert.NoError(t, emojer.Close(), "Close method cannot return an error")

}

func TestEmojiness(t *testing.T) {

	emojitext := "We are a :grinning_face: :family_man_woman_girl_girl:!"
	emojifake := "this is :not_an_emoji:, :do_not: convert it!"

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	emojiless, err := emojer.Emojiness(emojitext)
	assert.NoError(t, err, "Emojiness method cannot return an error")
	assert.Equal(t, "We are a 😀 👨‍👩‍👧‍👧!", emojiless)

	emojiless, err = emojer.Emojiness(emojifake)
	assert.NoError(t, err, "Emojiness method cannot return an error")
	assert.Equal(t, emojifake, emojiless)

	assert.NoError(t, emojer.Close(), "Close method cannot return an error")

}

func TestGetByUnicode(t *testing.T) {

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	row, err := emojer.Get("ucode","1F37A")
	assert.NoError(t, err, "GetByUnicode method cannot return an error")
	assert.Equal(t, "🍺", row.Emoji)
	assert.Equal(t, ":beer_mug:", row.Alias)

	assert.NoError(t, emojer.Close(), "Close method cannot return an error")

}

func TestGetByAlias(t *testing.T) {

	emojer, err := New()
	assert.NoError(t, err, "New method cannot return an error")

	row, err := emojer.Get("alias",":beer_mug:")
	assert.NoError(t, err, "GetByAlias method cannot return an error")
	assert.Equal(t, "🍺", row.Emoji)
	assert.Equal(t, "1F37A", row.Unicode)

	assert.NoError(t, emojer.Close(), "Close method cannot return an error")

}