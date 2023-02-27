package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/alecthomas/kong"
)

// Elgato defined
const MIN_COLOR_TEMPERATURE = 143
const MAX_COLOR_TEMPERATURE = 344
const MIN_BRIGHTNESS = 0
const MAX_BRIGHTNESS = 100

const LIGHT_IP = "192.168.4.40"

type LightStatus struct {
	On          int `json:"on"`
	Brightness  int `json:"brightness"`
	Temperature int `json:"temperature"`
}

type AllLightStatus struct {
	NumberOfLights int           `json:"numberOfLights"`
	Lights         []LightStatus `json:"lights"`
}

func getLightStatus() LightStatus {
	resp, err := http.Get(fmt.Sprintf("http://%s:9123/elgato/lights", LIGHT_IP))
	if err != nil {
		log.Fatalln(err)
	}

	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var status AllLightStatus
	json.Unmarshal(body, &status)

	return status.Lights[0]
}

func setLightStatus(status LightStatus) LightStatus {

	client := &http.Client{}

	// Wrap the individual light status in the AllLightStatus structure
	var allStatus AllLightStatus
	allStatus.NumberOfLights = 1
	allStatus.Lights = append(allStatus.Lights, status)

	jsonBody, err := json.Marshal(allStatus)
	if err != nil {
		panic(err)
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:9123/elgato/lights", LIGHT_IP), bytes.NewBuffer(jsonBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Make the request
	resp, err := client.Do(req)

	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}

	defer resp.Body.Close()
	
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var outStatus AllLightStatus
	json.Unmarshal(body, &outStatus)

	return outStatus.Lights[0]

}



////////////////////////////////////////////// CLI
var CLI struct {
  On struct {
  } `cmd:"" help:"Turn light On"`

  Off struct {
  } `cmd:"" help:"Turn light Off"`

	Warmer struct {
		} `cmd:"" help:"Increase color temperature"`
	
	Cooler struct {
	} `cmd:"" help:"Decrease color temperature"`
	Brighter struct {
		} `cmd:"" help:"Increase brightness"`
	
	Dimmer struct {
	} `cmd:"" help:"Decrease brightness"`
}

func main() {

	var startingStatus = getLightStatus()

	ctx := kong.Parse(&CLI)
  switch ctx.Command() {
  case "on":
		startingStatus.On = 1
  case "off":
		startingStatus.On = 0
	case "warmer":
		startingStatus.Temperature += 10
		if startingStatus.Temperature > MAX_COLOR_TEMPERATURE {
			startingStatus.Temperature = MAX_COLOR_TEMPERATURE
		}
	case "cooler":
		startingStatus.Temperature -= 10
		if startingStatus.Temperature < MIN_COLOR_TEMPERATURE {
			startingStatus.Temperature = MIN_COLOR_TEMPERATURE
		}
	case "brighter":
		if startingStatus.Brightness < 15 {
			// At low brightness levels we want to increase the sensitivity of adjustment
			startingStatus.Brightness += 2
		} else {
			startingStatus.Brightness += 10
		}
		
		if startingStatus.Brightness > MAX_BRIGHTNESS {
			startingStatus.Brightness = MAX_BRIGHTNESS
		}
	case "dimmer":
		if startingStatus.Brightness < 15 {
			// At low brightness levels we want to increase the sensitivity of adjustment
			startingStatus.Brightness -= 2
		} else {
			startingStatus.Brightness -= 10
		}
		
		if startingStatus.Brightness < MIN_BRIGHTNESS {
			startingStatus.Brightness = MIN_BRIGHTNESS
		}
  default:
    panic(ctx.Command())
  }
	
	var newStatus = setLightStatus(startingStatus)
	
	fmt.Printf("%+v\n", newStatus)
}
