package lib

import (
	"net/url"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func getDocumentTitle(p string) string {
	// Read file to check for the content
	file_content, err := os.ReadFile(p)
	if err != nil {
		log.Fatal(err) // TODO: return err instead of fatal
	}
	// Convert []byte to string and print to screen
	text := string(file_content)

	// check if there is a
	e := `^#\s+(.+)`
	r := regexp.MustCompile(e)
	result := r.FindStringSubmatch(text)
	if len(result) > 1 {
		// assign the Title to the matching group
		return result[1]
	}

	return ""
}

func isUrl(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
