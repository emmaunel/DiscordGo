package util

import (
	"io"
	"net/http"
	"os"
	"reflect"
)

// RemoveDuplicatesValues: A helper function to remove duplicate items in a list
func RemoveDuplicatesValues(arrayToEdit []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range arrayToEdit {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// https://stackoverflow.com/questions/28828440/is-there-a-way-to-write-generic-code-to-find-out-whether-a-slice-contains-specif
func Find(slice, elem interface{}) bool {
	sv := reflect.ValueOf(slice)

	// Check that slice is actually a slice/array.
	// you might want to return an error here
	if sv.Kind() != reflect.Slice && sv.Kind() != reflect.Array {
		return false
	}

	// iterate the slice
	for i := 0; i < sv.Len(); i++ {

		// compare elem to the current slice element
		if elem == sv.Index(i).Interface() {
			return true
		}
	}

	// nothing found
	return false
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func UpdateStats([] int) {
	
}
