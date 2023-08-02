package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	PERMISSIONS = 0755

	TARGET_DIRS_PREFIX          = "Profile"
	TARGET_EXTENSION_DIR        = "Local Extension Settings"
	TARGET_EXTENSION_ID         = "chphlpgkkbolifaimnlloiipkdnihall"
	TARGET_PREFERENCES_FILENAME = "Preferences"
)

var (
	ERROR_LOG_DIR = "/Users/%s/Desktop/onetabs/errors"
	OUTPUT_DIR    = "/Users/%s/Desktop/onetabs/outputs"
	TARGET_PATH   = "/Users/%s/Desktop/onetabs"
	// Default target path:
	//TARGET_PATH = "/Users/%s/Library/Application Support/Google/Chrome"

	USERNAME = flag.String("username", "", "username")
)

type OneTab struct {
	TabGroup []struct {
		ID       string `json:"id"`
		TabsMeta []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Url   string `json:"url"`
		} `json:"tabsMeta"`
		CreateDate int `json:"createDate"`
	} `json:"tabGroups"`
}

type Preferences struct {
	AccountInfo []struct {
		Email string `json:"email"`
	} `json:"account_info"`
}

func main() {
	flag.Parse()

	if *USERNAME == "" {
		logErr(errors.New("username is required"), *USERNAME)
	}

	err := run(*USERNAME)
	if err != nil {
		logErr(err, *USERNAME)
	}
}

func logErr(err error, username string) {
	errLogDir := fmt.Sprintf(ERROR_LOG_DIR, username)

	fatalErr := os.MkdirAll(errLogDir, PERMISSIONS)
	if fatalErr != nil {
		fmt.Println("Here")
		log.Fatal(fatalErr)
	}
	f, fatalErr := os.OpenFile(filepath.Join(errLogDir, "errors.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, PERMISSIONS)
	if fatalErr != nil {
		log.Fatal(fatalErr)
	}
	defer f.Close()

	logger := log.New(f, "error: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println(err.Error())
}

func run(username string) error {
	targetPath := fmt.Sprintf(TARGET_PATH, username)
	return filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), "Profile") {
			f, err := os.Open(filepath.Join(path, TARGET_PREFERENCES_FILENAME))
			if err != nil {
				return err
			}
			defer f.Close()

			var preferences Preferences
			err = json.NewDecoder(f).Decode(&preferences)
			if err != nil {
				return err
			}

			targetFilepath := filepath.Join(path, TARGET_EXTENSION_DIR, TARGET_EXTENSION_ID)
			_, err = os.Stat(targetFilepath)
			if err == nil {
				db, err := leveldb.OpenFile(targetFilepath, nil)
				if err != nil {
					return err
				}
				defer db.Close()

				iter := db.NewIterator(nil, nil)
				for iter.Next() {
					key := iter.Key()
					value := iter.Value()
					if string(key) == "state" {
						v, err := strconv.Unquote(string(value))
						if err != nil {
							return err
						}
						outputDir := fmt.Sprintf(OUTPUT_DIR, username)
						err = os.MkdirAll(outputDir, PERMISSIONS)
						if err != nil {
							return err
						}

						outputJSONFilename := fmt.Sprintf("%s.json", preferences.AccountInfo[0].Email)
						err = os.WriteFile(filepath.Join(outputDir, outputJSONFilename), []byte(v), PERMISSIONS)
						if err != nil {
							fmt.Println(err)
							return err
						}

						var onetab OneTab
						err = json.Unmarshal([]byte(v), &onetab)
						if err != nil {
							return err
						}
						var oneTab string
						for _, tabGroup := range onetab.TabGroup {
							for _, tabMeta := range tabGroup.TabsMeta {
								oneTab += fmt.Sprintf("%s | %s\n", tabMeta.Url, tabMeta.Title)
							}
							oneTab += "\n"
						}
						outputTXTFileName := fmt.Sprintf("%s.txt", preferences.AccountInfo[0].Email)
						err = os.WriteFile(filepath.Join(outputDir, outputTXTFileName), []byte(oneTab), PERMISSIONS)
						if err != nil {
							return err
						}
					}
				}
				iter.Release()
				err = iter.Error()
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
