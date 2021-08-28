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
	"text/template"
	"time"
)

const (
	generalConfigurationFileName = "conf.json"
	endpoint                     = "https://api.nibol.co/v2/app/business/reservation/desk/create"
	authorizeFolder              = "/.config/nibblefibble"
)

type authConfig struct {
	DeskID        string `json:"desk_id"`
	SpaceID       string `json:"space_id"`
	BearerToken   string `json:"token"`
	Identity      string `json:"identity"`
	ExcludingDays []int  `json:"excluding_days"`
}

type generalConfig struct {
	SlackHook     string      `json:"slack_hook"`
	SlackTemplate interface{} `json:"slack_template"`
}

type bookDeskPayload struct {
	Day     string `json:"day"`
	WeekDay int    `json:"-"`
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
		if file.Name() != generalConfigurationFileName {
			output = append(output, basePath+"/"+file.Name())
		}
	}

	return output, nil
}

/*
  Read a JSON file.
  The function return the bytes obtained
*/
func readGeneralConfig() (generalConfig, error) {
	var unstructuredPayload map[string]interface{}
	var payload generalConfig

	homeDIR, err := os.UserHomeDir()
	if err != nil {
		return payload, err
	}

	basePath := homeDIR + authorizeFolder

	file, err := ioutil.ReadFile(basePath + "/" + generalConfigurationFileName)
	if err != nil {
		return payload, err
	}

	err = json.Unmarshal([]byte(file), &unstructuredPayload)
	if err != nil {
		return payload, err
	}

	payload.SlackHook = fmt.Sprintf("%v", unstructuredPayload["slack_hook"])
	payload.SlackTemplate = unstructuredPayload["slack_template"]

	return payload, nil
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
func composeNextDay(date time.Time) string {
	year, month, day := date.Date()

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

func sendNotification(message string, hook string) error {
	var jsonPayload map[string]interface{}
	err := json.Unmarshal([]byte(message), &jsonPayload)
	if err != nil {
		return err
	}

	bytesPayload, err := json.Marshal(jsonPayload)
	if err != nil {
		return err
	}

	_, err = http.Post(hook, "application/json", bytes.NewBuffer(bytesPayload))
	if err != nil {
		return err
	}

	return nil
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
	tomorrow := time.Now().Add(24 * time.Hour)
	tomorrowWeekdayAsInt := int(tomorrow.Weekday())

	return bookDeskPayload{
		To:      1800,
		From:    900,
		Day:     composeNextDay(tomorrow),
		WeekDay: tomorrowWeekdayAsInt,
		SpaceID: auth.SpaceID,
		DeskID:  auth.DeskID,
	}
}

func renderNotificationTemplate(templateRaw interface{}, identity string) (string, error) {
	templateAsString, err := json.Marshal(templateRaw)
	if err != nil {
		return "", err
	}

	templateInstance, err := template.New("notification").Parse(string(templateAsString))
	if err != nil {
		return "", err
	}

	var templateAsBuffer bytes.Buffer
	err = templateInstance.Execute(&templateAsBuffer, map[string]interface{}{
		"Identity": identity,
	})
	if err != nil {
		return "", err
	}

	return templateAsBuffer.String(), nil
}

func abortTheBooking(weekday int, excludingDays []int) bool {
	matchExludedDay := false

	for _, exludedDay := range excludingDays {
		if weekday == exludedDay {
			matchExludedDay = true
			break
		}
	}

	return matchExludedDay
}

func main() {
	// Read configuration file general
	conf, err := readGeneralConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	filesPath, err := listFileAuthorizations()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(filesPath))

	for _, filePath := range filesPath {
		go func(fp string) {
			defer wg.Done()

			authorization, err := readAuthorization(fp)
			if err != nil {
				fmt.Println(err)
				return
			}

			payload := prepareBookingPayload(authorization)

			fmt.Printf("Try to booking the desk for the date %s for user %s\n", payload.Day, authorization.Identity)

			if abortTheBooking(payload.WeekDay, authorization.ExcludingDays) == true {
				fmt.Println("This day is excluded from the booking process")
				return
			}

			err = bookDesk(payload, authorization.BearerToken)
			if err != nil {
				fmt.Println("Error", err.Error())

				templateRendered, err := renderNotificationTemplate(conf.SlackTemplate, authorization.Identity)
				if err != nil {
					fmt.Println("Error", err.Error())
					return
				}

				// Notify to a channel if something wrong
				// occured
				sendNotification(templateRendered, conf.SlackHook)
			}
		}(filePath)
	}

	wg.Wait()
}
