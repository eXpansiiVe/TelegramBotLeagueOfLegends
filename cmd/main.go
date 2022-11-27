package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/eXpansiiVe/LoLBot/pkg/schemes"
)

const apiKey string = "RGAPI-fd2cdcde-e835-44da-8648-4c691175c919"

// This handler is called everytime telegram sends us a webhook event
func Handler(res http.ResponseWriter, req *http.Request) {
	// First, decode the JSON response body
	body := &schemes.WebhookBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		log.Println("could not decode request body", err)
		return
	}
	if !isValidCommand(body) {
		log.Println("Message ignored")
		return
	}
	doEverything(body)
}

func isValidCommand(b *schemes.WebhookBody) bool {
	return strings.HasPrefix(b.Message.Text, "/euw")
}

// Used to get a response from an API
func getRequestData(link, errorMessage string) io.ReadCloser {
	rawResponseData, err := http.Get(link)
	fmt.Printf("Status code: %v\n", rawResponseData.StatusCode)
	if err != nil {
		log.Println(errorMessage, err)
		return nil
	}

	return rawResponseData.Body
}

// A simple reader use to read the ResponseDate passed by parameter
func filterRequestData(rawResponseData io.ReadCloser, errorMessage string) []byte {
	responseData, err := io.ReadAll(rawResponseData)
	if err != nil {
		log.Println(errorMessage, err)
		return nil
	}

	return responseData
}

// Get the summoner id through API
func getAccountID(nickname string) []byte {
	log.Println("Inside function getAccountID")
	accountLink := "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/" + url.QueryEscape(nickname) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponseAccountData := getRequestData(accountLink, "Got an error retrieving the Account ID api")

	responseAccountData := filterRequestData(rawResponseAccountData, "Got an error reading the account ID body")
	defer func(rawResponseAccountData io.ReadCloser) {
		err := rawResponseAccountData.Close()
		if err != nil {
			log.Println("Got an error closing the response account data")
		}
	}(rawResponseAccountData)
	return responseAccountData
}

// Get the summoner Data through API
func getSummonerData(id string) []byte {
	log.Println("Inside function getSummonerData")

	playerDataLink := "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/" + url.QueryEscape(id) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponsePlayerData := getRequestData(playerDataLink, "Got an error retrieving the summoner data")

	responsePlayerData := filterRequestData(rawResponsePlayerData, "Got an error reading the summoner Data body")

	log.Println("response from getting summoner data:\n", string(responsePlayerData))

	defer func(rawResponsePlayerData io.ReadCloser) {
		err := rawResponsePlayerData.Close()
		if err != nil {
			log.Println("Got an error closing the response account data")
		}
	}(rawResponsePlayerData)
	return responsePlayerData
}

// Parse the data to check if there's a rank and return it
func filterRankData(s schemes.LoLAccount, m map[string]string) error {
	log.Println("Inside function filterRankData")
	var schemeLen = len(s)
	var noRankError error

	if schemeLen > 0 {
		m["Nickname"] = s[0].SummonerName
		for i := 0; i < schemeLen; i++ {
			if s[i].QueueType == "RANKED_SOLO_5x5" {
				m["RankSoloQ"] = fmt.Sprintf("%s %s", s[i].Tier, s[i].Rank)
				m["LpQ"] = fmt.Sprintf("%v", s[i].LeaguePoints)
				m["WinsQ"] = fmt.Sprintf("%v", s[i].Wins)
				m["LosesQ"] = fmt.Sprintf("%v", s[i].Losses)
			} else {
				m["RankFlex"] = fmt.Sprintf("%s %s", s[i].Tier, s[i].Rank)
				m["LpFlex"] = fmt.Sprintf("%v", s[i].LeaguePoints)
				m["WinsFlex"] = fmt.Sprintf("%v", s[i].Wins)
				m["LosesFlex"] = fmt.Sprintf("%v", s[i].Losses)
			}
		}
		return nil
	}
	noRankError = errors.New("rank not found")
	return noRankError

}

// Create a formatted text message for telegram
func messageTextFormatter(m map[string]string, imgId string, s schemes.LoLAccount) (string, string) {
	log.Println("Inside function messageTextFormatter")
	imgLink := "https://ddragon.leagueoflegends.com/cdn/12.16.1/img/profileicon/" + url.QueryEscape(imgId) + ".png"
	schemeLen := len(s)
	formattedText := fmt.Sprintf("<b>Nome:</b> %v\n", m["Nickname"])
	formattedText += fmt.Sprintf("<b>Livello:</b> %v\n\n", m["Level"])

	if schemeLen == 1 {
		if s[0].QueueType == "RANKED_SOLO_5x5" {
			formattedText += "<b>SoloQ</b> \n"
			formattedText += fmt.Sprintf("<b>Lega:</b> %v\n", m["RankSoloQ"])
			formattedText += fmt.Sprintf("<b>Vittorie:</b> %v\n", m["WinsQ"])
			formattedText += fmt.Sprintf("<b>Sconfitte:</b> %v\n", m["LosesQ"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", m["LpQ"])
		} else {
			formattedText += "<b>Flex</b>\n"
			formattedText += fmt.Sprintf("<b>Lega:</b> %v\n", m["RankFlex"])
			formattedText += fmt.Sprintf("<b>Vittorie:</b> %v\n", m["WinsFlex"])
			formattedText += fmt.Sprintf("<b>Sconfitte:</b> %v\n", m["LosesFlex"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", m["LpFlex"])
		}
	} else {
		if s[0].QueueType == "RANKED_SOLO_5x5" || s[1].QueueType == "RANKED_SOLO_5x5" {
			formattedText += "<b>SoloQ</b> \n"
			formattedText += fmt.Sprintf("<b>Lega:</b> %v\n", m["RankSoloQ"])
			formattedText += fmt.Sprintf("<b>Vittorie:</b> %v\n", m["WinsQ"])
			formattedText += fmt.Sprintf("<b>Sconfitte:</b> %v\n", m["LosesQ"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", m["LpQ"])
		}
		if s[0].QueueType == "RANKED_FLEX_SR" || s[1].QueueType == "RANKED_FLEX_SR" {
			formattedText += "<b>Flex</b>\n"
			formattedText += fmt.Sprintf("<b>Lega:</b> %v\n", m["RankFlex"])
			formattedText += fmt.Sprintf("<b>Vittorie:</b> %v\n", m["WinsFlex"])
			formattedText += fmt.Sprintf("<b>Sconfitte:</b> %v\n", m["LosesFlex"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", m["LpFlex"])
		}
	}
	return formattedText, imgLink
}

// Send photo message to telegram through API and return a response
func sendPhotoMessage(chatId int64, messageId int, message string, link string) ([]byte, error) {
	log.Println("Inside sendPhotoMessage")
	chatIdString := fmt.Sprintf("%d", chatId)
	messageIdString := fmt.Sprintf("%d", messageId)
	fmt.Printf("ChatId: %v\nMessageId:%v\nMessage: %v\nLink: %v\n", chatIdString, messageIdString, message, link)
	formattedUrl := "https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/sendPhoto?chat_id=" + url.QueryEscape(chatIdString) + "&reply_to_message_id=" + url.QueryEscape(messageIdString) + "&photo=" + url.QueryEscape(link) + "&caption=" + url.QueryEscape(message) + "&parse_mode=html"

	rawResponseData, err := http.Get(formattedUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Error closing Body of sendPhotoMessage", err)
		}
	}(rawResponseData.Body)
	responseData, err := io.ReadAll(rawResponseData.Body)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

// Send message to telegram through API and return a response
func sendMessage(chatId int64, messageId int, message string) ([]byte, error) {
	log.Println("Inside function sendMessage")
	chatIdString := fmt.Sprintf("%d", chatId)
	messageIdString := fmt.Sprintf("%d", messageId)
	fmt.Printf("ChatId %v\nMessageId: %v\n", chatIdString, messageIdString)
	formattedUrl := "https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/sendMessage?chat_id=" + url.QueryEscape(chatIdString) + "&reply_to_message_id=" + url.QueryEscape(messageIdString) + "&text=" + url.QueryEscape(message)

	rawResponseData, err := http.Get(formattedUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Error closing body sendMessage", err)
		}
	}(rawResponseData.Body)
	responseData, err := io.ReadAll(rawResponseData.Body)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func doEverything(body *schemes.WebhookBody) {
	// Instantiate every schemes
	var schemeAccount schemes.AccountRiot
	var schemeLoLData schemes.LoLAccount
	var schemeTgMessageResponse schemes.ApiTelegramMessage
	var schemePlayerNotFound schemes.PlayerNotFound

	playerInfo := make(map[string]string)

	// Remove /euw from the Message.Text
	body.Message.Text = strings.ReplaceAll(body.Message.Text, "/euw", "")
	log.Println("Nickname: ",body.Message.Text)

	// Get the id of the riot account
	responseAccountData := getAccountID(body.Message.Text)

	// Check if the player exist, if not exist send a message, assign to updateID the new value and goes to new cycle
	err := json.Unmarshal(responseAccountData, &schemePlayerNotFound)
	if err != nil {
		log.Println("Got an error unmarshal response account data to schemePlayerNotFound")
	}
	if schemePlayerNotFound.Status.StatusCode == 404 {
		message := fmt.Sprintf("L'username %s non è valido!", body.Message.Text)
		responseMessage, err := sendMessage(body.Message.Chat.ID, body.Message.MessageID, message)
		if err != nil {
			log.Println(err)
		}

		err = json.Unmarshal(responseMessage, &schemeTgMessageResponse)
		if err != nil {
			log.Println("Got error unmarshal response message")
		}
		log.Println("Response from telegram message ", schemeTgMessageResponse.Ok)
		//Assign the messageId to the var updateID to not cycle on the same message
		return
	}

	err = json.Unmarshal(responseAccountData, &schemeAccount)
	if err != nil {
		log.Println("Got error unmarshal response account data")
	}

	// Write summoner level to the map
	playerInfo["Level"] = fmt.Sprintf("%v", schemeAccount.SummonerLevel)
	// Write imgId to var
	imgId := fmt.Sprintf("%v", schemeAccount.ProfileIconID)

	// Get the summoner data response
	rankData := getSummonerData(schemeAccount.ID)
	err = json.Unmarshal(rankData, &schemeLoLData)
	if err != nil {
		log.Println("Got error unmarshal rank data")
	}

	// Check if there's a schemeLoLData.QueueType and if there's any take the actual rank
	err = filterRankData(schemeLoLData, playerInfo)

	// If player rank not found
	if err != nil {
		log.Println(err)
		message := fmt.Sprintf("Il rank del player %s non è stato trovato!", body.Message.Text)
		// Send a message to the message sender on telegram with the results
		responseMessage, err := sendMessage(body.Message.Chat.ID, body.Message.MessageID, message)

		if err != nil {
			log.Println(err)
		}

		err = json.Unmarshal(responseMessage, &schemeTgMessageResponse)
		if err != nil {
			log.Println("Got error unmarshal response message")
		}
		log.Println("Response from telegram message ", schemeTgMessageResponse.Ok)
		fmt.Println("------------------------------")
		return
	}

	// format the message with the playerInfo data
	message, imgLink := messageTextFormatter(playerInfo, imgId, schemeLoLData)

	// Send a message to the message sender on telegram with the results
	responseMessage, err := sendPhotoMessage(body.Message.Chat.ID, body.Message.MessageID, message, imgLink)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(responseMessage, &schemeTgMessageResponse)
	if err != nil {
		log.Println("Got error unmarshal response message")
	}
	log.Println("Response from telegram message ", schemeTgMessageResponse.Ok)
	fmt.Println("------------------------------")

}
func main() {
	log.Println("Listening")
	http.ListenAndServe(":9002", http.HandlerFunc(Handler))

}
