package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	endpoint        = "https://api.nibol.co/v2/app/business/reservation/desk/create"
	authorizeFolder = "/.config/nibblefibble"
)

type authConfig struct {
	DeskID      string `json:"desk_id"`
	SpaceID     string `json:"space_id"`
	BearerToken string `json:"token"`
	Identity    string `json:"identity"`
}

type bookDeskPayload struct {
	Day     string `json:"day"`
	From    int    `json:"from"`
	To      int    `json:"to"`
	DeskID  string `json:"desk_id"`
	SpaceID string `json:"space_id"`
}

/*
	Take the authorities files
	from the .config/nibblefibble folder
*/
func listFileAuthorizations() ([]string, error) {
	var output []string

	homeDIR, err := os.UserHomeDir()
	if err != nil {
		return output, err
	}

	basePath := homeDIR + authorizeFolder

	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		return output, err
	}

	for _, file := range files {
		output = append(output, basePath+"/"+file.Name())
	}

	return output, nil
}

/*
	Read the authorization file
	and take the secret informations
*/
func readAuthorization(filePath string) (authConfig, error) {
	var payload authConfig

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return payload, err
	}

	err = json.Unmarshal([]byte(file), &payload)
	if err != nil {
		return payload, err
	}

	return payload, nil
}

/*
	The day to provide
	must be in the following
	form. For example:
	20210210
*/
func composeNextDay() string {
	now := time.Now().Add(24 * time.Hour)
	year, month, day := now.Date()

	monthAsString := strconv.Itoa(int(month))
	dayAsString := strconv.Itoa(day)
	if int(month) < 10 {
		monthAsString = "0" + monthAsString
	}

	if day < 10 {
		dayAsString = "0" + dayAsString
	}

	return strconv.Itoa(year) + monthAsString + dayAsString
}

func bookDesk(payload bookDeskPayload, bearerToken string) error {
	bodyAsBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyAsBytes))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+bearerToken)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return errors.New(string(responseBody))
	}

	return nil
}

func prepareBookingPayload(auth authConfig) bookDeskPayload {
	return bookDeskPayload{
		To:      1800,
		From:    900,
		Day:     composeNextDay(),
		SpaceID: auth.SpaceID,
		DeskID:  auth.DeskID,
	}
}

func main() {
	wg := new(sync.WaitGroup)

	filesPath, err := listFileAuthorizations()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wg.Add(len(filesPath))

	for _, filePath := range filesPath {
		go func(fp string) {
			defer wg.Done()

			authorization, err := readAuthorization(fp)
			if err != nil {
				return
			}

			payload := prepareBookingPayload(authorization)

			fmt.Printf("Try to booking the desk for the date %s for user %s\n", payload.Day, authorization.Identity)

			err = bookDesk(payload, authorization.BearerToken)
			if err != nil {
				fmt.Println("Error", err.Error())
			}
		}(filePath)
	}

	wg.Wait()
}
