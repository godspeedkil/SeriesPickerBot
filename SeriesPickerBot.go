package SeriesPickerBot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"time"
)

const ANIME_URL string = "https://api.myjson.com/bins/pi4b3"
const TELEGRAM_API_KEY string = "placeholder"
const PICKS int = 10000

type Anime struct {
	Name   string
	Weight float64
}

type Result struct {
	Name string
	Hits int
}

func main() {
	bot, err := tgbotapi.NewBotAPI(TELEGRAM_API_KEY)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch command := update.Message.Command(); command {
		case "list":
			if err := listCommand(bot, update); err != nil {
				log.Panic(err)
			}
		case "select":
			if err := selectCommand(bot, update); err != nil {
				log.Panic(err)
			}
		case "ayaya":
			if err := simpleTextCommand(bot, update, "AYAYA!"); err != nil {
				log.Panic(err)
			}
		case "waifu":
			if err := simpleTextCommand(bot, update, "Rei is trash"); err != nil {
				log.Panic(err)
			}
		default:
			continue
		}
	}
}

func listCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) (error){
	animeList, err := showList()
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, animeList)
	bot.Send(msg)
	return nil
}

func selectCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) (error){
	topFive, err := showTopFive()
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, topFive)
	bot.Send(msg)
	return nil
}

func simpleTextCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, text string) (error){
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)
	return nil
}

func RandomWeightedSelect(animes []Anime, totalWeight float64) (Anime, error) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Float64() * totalWeight
	for _, a := range animes {
		r -= a.Weight
		if r <= 0 {
			return a, nil
		}
	}
	return Anime{}, errors.New("No anime selected")
}

func showList() (string, error) {
	jsonArray, err := fetchJSONArray(ANIME_URL)
	if err != nil {
		return "", err
	}

	sort.Slice(jsonArray, func(i, j int) bool {
		return jsonArray[i].Weight > jsonArray[j].Weight
	})

	return formatListString(jsonArray), nil
}

func showTopFive() (string, error) {
	jsonArray, err := fetchJSONArray(ANIME_URL)
	if err != nil {
		return "", err
	}

	shuffleSlice(jsonArray)

	results, err := getResults(jsonArray, getTotalWeight(jsonArray))
	if err != nil {
		return "", err
	}

	resultsStruct := translateMapToResult(results)

	sort.Slice(resultsStruct, func(i, j int) bool {
		return resultsStruct[i].Hits > resultsStruct[j].Hits
	})

	return formatTopFiveString(resultsStruct), nil
}

func shuffleSlice (slice []Anime) {
	for i := len(slice) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func fetchJSONArray (url string) ([]Anime, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []Anime{}, errors.New("url error")
	}

	jsonArray := make([]Anime, 0)
	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if err = decoder.Decode(&jsonArray); err != nil {
		return []Anime{}, errors.New("decoder error")
	}

	return jsonArray, nil
}

func getTotalWeight (slice []Anime) (float64) {
	var totalWeight float64
	for _, a := range slice {
		totalWeight += a.Weight
	}
	return totalWeight
}

func getResults (slice []Anime, totalWeight float64) (map[string]int, error) {
	results := map[string]int{}
	for i := 0; i < PICKS; i++ {
		a, err := RandomWeightedSelect(slice, totalWeight)
		if err != nil {
			return map[string]int{}, errors.New("select error")
		}
		if _, ok := results[a.Name]; ok {
			results[a.Name]++
		} else {
			results[a.Name] = 1
		}
	}
	return results, nil
}

func formatListString (slice []Anime) (string) {
	var buffer bytes.Buffer
	for n, a := range slice {
		buffer.WriteString(fmt.Sprintf("%d. %s, Weight: %.2f\n", n + 1, a.Name, a.Weight))
	}
	return buffer.String()
}

func formatTopFiveString (slice []Result) (string) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintln("Top 5"))
	for n, a := range slice[:5] {
		buffer.WriteString(fmt.Sprintf("%d. %s, Hits: %d\n", n + 1, a.Name, a.Hits))
	}
	return buffer.String()
}

func translateMapToResult (resultsMap map[string]int) ([]Result) {
	var resultsStruct []Result
	for a, h := range resultsMap {
		resultsStruct = append(resultsStruct, Result{a, h})
	}
	return resultsStruct
}