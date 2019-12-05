package dmrtools

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/tidwall/gjson"
)

// CheckUserFile checks to see if the user file is present, and if not
// then downloads it. If the file is present, then check the age. If
// the file is over 7 days old, redownload it.
func CheckUserFile(userFile string, userURL string) {
	userExistence := fileExists(userFile)
	if userExistence != true {
		fmt.Println("Users file is not present.")
		downloadFile(userFile, userURL)
	} else {
		fmt.Println(userFile, "exists. Checking age...")
		fileModTime(userFile, userURL)
	}
}

// CheckRptrFile checks to see if the repeater file is present, and
// if not then downloads it. If the file is present, then check the
// age. If the file is over 7 days old, redownload it.
func CheckRptrFile(rptrFile string, rptrURL string) {
	rptrExistence := fileExists(rptrFile)
	if rptrExistence != true {
		fmt.Println("Repeater file is not present.")
		downloadFile(rptrFile, rptrURL)
	} else {
		fmt.Println(rptrFile, "exists. Checking age...")
		fileModTime(rptrFile, rptrURL)
	}
}

// GetUserCall reads the user call file and returns the callsign of
// the ID that was queried.
func GetUserCall(userFile string, userID string) string {
	input, err := ioutil.ReadFile(userFile)
	if err != nil {
		fmt.Println(err)
	}
	userLookup := string(input)
	uSign := gjson.Get(userLookup, `users.#(id=`+userID+`).callsign`)
	return (uSign.String())
}

// GetRepeaterCall reads the rptr call file and returns the callsign of
// the ID that was queried.
func GetRepeaterCall(rptrFile string, rptID string) string {
	input, err := ioutil.ReadFile(rptrFile)
	if err != nil {
		fmt.Println(err)
	}
	rptrLookup := string(input)
	cSign := gjson.Get(rptrLookup, `rptrs.#(id=`+rptID+`).callsign`)
	return (cSign.String())
}

// GetAliasString returns a full alias string for a given user ID. Needs
// to be cleaned up, for sure. Probably would do better with a variadic
// function but this is my first attempt at it. It works for now.
func GetAliasString(userFile string, userID string) string {
	input, err := ioutil.ReadFile(userFile)
	if err != nil {
		fmt.Println(err)
	}
	userLookup := string(input)
	s := []string{}
	uSign := gjson.Get(userLookup, `users.#(id=`+userID+`).callsign`)
	s = append(s, uSign.String())
	uCity := gjson.Get(userLookup, `users.#(id=`+userID+`).city`)
	s = append(s, uCity.String())
	uState := gjson.Get(userLookup, `users.#(id=`+userID+`).state`)
	s = append(s, uState.String())
	result := strings.Join(s, ", ")
	return result
}

// GetAliasShort returns a user's callsign and name. Needs to be cleaned
// up, for sure. Probably would do better with a variadic function but
// this is my first attempt at it. It works for now.
func GetAliasShort(userFile string, userID string) string {
	input, err := ioutil.ReadFile(userFile)
	if err != nil {
		fmt.Println(err)
	}
	userLookup := string(input)
	s := []string{}
	uSign := gjson.Get(userLookup, `users.#(id=`+userID+`).callsign`)
	s = append(s, uSign.String())
	uName := gjson.Get(userLookup, `users.#(id=`+userID+`).name`)
	s = append(s, uName.String())
	result := strings.Join(s, ", ")
	return result
}

// WriteCounter counts the number of bytes written to it. It implements to the
// io.Writer interface and we can pass this into io.TeeReader() which will
// report progress on each write cycle.
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

// PrintProgress does what it says on the tin. It.. well, it prints the progress.
func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func fileExists(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func fileStale(timestamp time.Time) bool {
	return time.Now().Sub(timestamp) > 168*time.Hour
}

func fileModTime(fileName string, fileURL string) {
	filen, err := os.Stat(fileName)
	if err != nil {
		return
	}
	mtime := filen.ModTime()
	fmt.Printf("Time: %v\n", mtime)
	fstale := fileStale(mtime)
	if fstale {
		downloadFile(fileName, fileURL)
	} else {
		fmt.Println(fileName, "is up to date.")
	}
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	fmt.Printf("Download of %v file started...\n", filepath)
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	fmt.Printf("\nDownload of %v finished!\n", filepath)
	fmt.Print("\n")

	return nil
}
