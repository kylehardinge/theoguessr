package storage

import (
	"os"
	"fmt"
	"encoding/json"
	"theoguessr/site/pkg/db"
)

// OpenManifest gets the manifest JSON file for a given bible version
// It returns the bible manifest in a struct of type db.BibleManifest
func OpenManifest(version string) (db.BibleManifest, error) {
	contents, err := os.ReadFile(fmt.Sprintf("assets/%s/manifest.json", version)) 
	if err != nil {
		panic(err.Error())
	}
	manifest := db.BibleManifest{}	
    // Unmarshall the data into a struct usable by go
	json.Unmarshal(contents, &manifest)
	return manifest, nil
}
