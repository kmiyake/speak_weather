package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

type Coordination struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type Weather struct {
	ID          uint   `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Main struct {
	Temp     float64 `json:"temp"`
	Pressure float64 `json:"pressure"`
	Humidity float64 `json:"humidity"`
	TempMin  float64 `json:"temp_min"`
	TempMax  float64 `json:"temp_max"`
}

type WeatherAPIResponse struct {
	Coord       Coordination `json:"coord"`
	WeatherList []Weather    `json:"weather"`
	Main        Main         `json:"main"`
}

func getWeather(body []byte) (*WeatherAPIResponse, error) {
	var weather = new(WeatherAPIResponse)
	err := json.Unmarshal(body, &weather)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return weather, err
}

type WebhookMessage struct {
	Username string `json:"username,omitempty"`
	IconURL  string `json:"icon_url,omitempty"`
	Text     string `json:"text,omitempty"`
}

type Slack struct {
	URL    string
	Params WebhookMessage
}

func postMessage(webhookURL string, username string, text string) (*http.Response, error) {
	params, _ := json.Marshal(WebhookMessage{
		Username: username,
		Text:     text,
	})

	resp, err := http.PostForm(
		webhookURL,
		url.Values{"payload": {string(params)}},
	)

	return resp, err
}

func speakWeather() (string, error) {
	cityName := os.Getenv("CITY_NAME")
	appID := os.Getenv("OPEN_WEATHER_API_ID")
	url := "https://api.openweathermap.org/data/2.5/weather?q=" + cityName + "&appid=" + appID + "&units=metric"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Get weather is failed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	weather, err := getWeather([]byte(body))
	message := cityName + " will have " + weather.WeatherList[0].Main + " today. The highest temperature will be " + fmt.Sprintf("%.1f", weather.Main.TempMax) + ". The lowest temperature will be " + fmt.Sprintf("%.1f", weather.Main.TempMin) + "."
	resp, err = postMessage(os.Getenv("SLACK_WEBHOOK_URL"), "Weather Bot", message)
	return fmt.Sprintf("%s", resp.Body), err
}

func main() {
	env := os.Getenv("GO_ENV")
	if env == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
		speakWeather()
	} else {
		lambda.Start(speakWeather)
	}
}
