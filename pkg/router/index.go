// All of the route functions go here
package router

import (
	"theoguessr/site/pkg/db"
	"theoguessr/site/pkg/parser"
	"theoguessr/site/pkg/storage"
	"strconv"

	"math/rand"
	"net/http"


	"github.com/labstack/echo/v4"
)

// The function corresponding with the "/" route
func Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

// The function corresponding with the "/play" route
func Play(c echo.Context) error {
	return c.Render(http.StatusOK, "play.html", nil)
}

// The function corresponding with the "/newgame" route
func NewGame(c echo.Context) error {
	return c.Render(http.StatusOK, "newgame.html", nil)
}

// The function corresponding with the "/api/random" route
// Returns a random Bible verse
func Random(c echo.Context) error {
    // Get the number of context verses the user is asking for
    // Added so the game is not insanely hard
	context, err := strconv.Atoi(c.QueryParam("contextVerses"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "contextVerses must be a number",
		})
	}

	// Get number of Bible verses
	count := db.Db.QueryRow(`SELECT COUNT(*) FROM kjv`)
	var length int
	if err := count.Scan(&length); err != nil {
		panic(err.Error())
	}

	// Generate a random verse id [1, max id number]
	verse_id := rand.Intn(length) + 1

	// Get a verse based on the random id number
	verse := db.Verse{}
	content := db.Db.QueryRow(`SELECT * FROM kjv WHERE id=?`, verse_id)
	if err := content.Scan(&verse.Id, &verse.Book_id, &verse.Book_name, &verse.Chapter, &verse.Verse, &verse.Text); err != nil {
		panic(err.Error())
	}

    // Get the range for the context verses
	startContext := verse_id - context
	if startContext <= 0 {
		startContext = 1
	}

	endContext := verse_id + context
	if endContext > length {
		endContext = length
	}

    // Loop through adding the context verses to the message to be returned to the user
	for i := startContext; i <= endContext; i++ {
		contextVerse := db.Verse{}
		content := db.Db.QueryRow(`SELECT * FROM kjv WHERE id=?`, i)
		if err := content.Scan(&contextVerse.Id, &contextVerse.Book_id, &contextVerse.Book_name, &contextVerse.Chapter, &contextVerse.Verse, &contextVerse.Text); err != nil {
			panic(err.Error())
		}
		verse.Context = append(verse.Context, contextVerse)
	}

	// Return the random bible verse in json format
	return c.JSON(http.StatusOK, verse)
}

// The function corresponding with the "/api/manifest" route
// Returns the Bible manifest for a given version (curently only kjv)
func Manifest(c echo.Context) error {
	kjvManifest, err := storage.OpenManifest("kjv")
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, kjvManifest)
}

// Get a verse by ID
// I dont believe this is currently used anywhere but it is useful none the less
func apiById(c echo.Context, query string) error {
	id := c.QueryParam("id")

	// Get a verse based on the id number
	verse := db.Verse{}
	content := db.Db.QueryRow(`SELECT * FROM kjv WHERE id=?`, id)
	if err := content.Scan(&verse.Id, &verse.Book_id, &verse.Book_name, &verse.Chapter, &verse.Verse, &verse.Text); err != nil {
        // If there is no verse at that id something is wrong
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Verse not found"})
	}
	return c.JSON(http.StatusOK, verse)
}

// The function corresponding with the "/api/:query" route
// Allows for getting a verse, range of verses, or entire chapter
func apiByVerse(c echo.Context, query string) error {
    // Split the query up with the parsing function
	bookQuery, chapterQuery, verseQuery := parser.SplitQuery(query)
    // This is for a single verse
	if bookQuery != "" && chapterQuery != "" && len(verseQuery) == 1 && verseQuery[0] != "" {
		verse := db.Verse{}
		verseText := db.Db.QueryRow(`SELECT * FROM kjv WHERE book_id=? AND chapter=? AND verse=?`, bookQuery, chapterQuery, verseQuery[0])

		if err := verseText.Scan(&verse.Id, &verse.Book_id, &verse.Book_name, &verse.Chapter, &verse.Verse, &verse.Text); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Verse not found"})
		}

		// Return the random bible verse in json format
		return c.JSON(http.StatusOK, verse)
	}

    // This is for a range of verses
	if bookQuery != "" && chapterQuery != "" && len(verseQuery) == 2 {
		verseGroup := db.VerseGroup{}
		startVerse, err := strconv.Atoi(verseQuery[0])
		if err != nil {
			panic(err.Error())
		}
		endVerse, err := strconv.Atoi(verseQuery[1])
		if err != nil {
			panic(err.Error())
		}
		for i := startVerse; i <= endVerse; i++ {

			verse := db.Verse{}
			verseText := db.Db.QueryRow(`SELECT * FROM kjv WHERE book_id=? AND chapter=? AND verse=?`, bookQuery, chapterQuery, i)

			if err := verseText.Scan(&verse.Id, &verse.Book_id, &verse.Book_name, &verse.Chapter, &verse.Verse, &verse.Text); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Verse not found"})
			}
			verseGroup.Verses = append(verseGroup.Verses, verse)
		}
		// Return the random bible verse in json format
		return c.JSON(http.StatusOK, verseGroup)
	}

    // This is for a chapter
	if bookQuery != "" && chapterQuery != "" && len(verseQuery) == 0 {
		verseGroup := db.VerseGroup{}
		verseTexts, err := db.Db.Query(`SELECT * FROM kjv WHERE book_id=? AND chapter=?`, bookQuery, chapterQuery)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Verse not found"})
		}
		for verseTexts.Next() {
			verse := db.Verse{}

			if err := verseTexts.Scan(&verse.Id, &verse.Book_id, &verse.Book_name, &verse.Chapter, &verse.Verse, &verse.Text); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Verse not found"})
			}
			verseGroup.Verses = append(verseGroup.Verses, verse)
		}
		// Return the random bible verse in json format
		return c.JSON(http.StatusOK, verseGroup)
	}

	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Requests should be in the format 'book chapter:verse-verse' or 'byid?id=1234'"})
}

// Decide which type of query is getting run
func Api(c echo.Context) error {
	query := c.Param("query")
	if query == "byid" {
		err := apiById(c, "query")
		if err != nil {
			panic(err.Error())
		}

	} else {
		err := apiByVerse(c, query)
		if err != nil {
			panic(err.Error())
		}
	}
	return nil
}
