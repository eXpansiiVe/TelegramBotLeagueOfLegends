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

const riotApiKey string = ""
const telegramApiKey string = ""

// Handler is called everytime telegram sends us a webhook event
func Handler(_ http.ResponseWriter, req *http.Request) {
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
	checkCommand(body)
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
	accountLink := "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/" + url.QueryEscape(nickname) + "?api_key=" + url.QueryEscape(riotApiKey)

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

	playerDataLink := "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/" + url.QueryEscape(id) + "?api_key=" + url.QueryEscape(riotApiKey)

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
func filterRankData(schemeLolData schemes.LoLAccount, playerSlice map[string]string) error {
	log.Println("Inside function filterRankData")
	var schemeLen = len(schemeLolData)
	var noRankError error

	if schemeLen > 0 {
		playerSlice["Nickname"] = schemeLolData[0].SummonerName
		for i := 0; i < schemeLen; i++ {
			if schemeLolData[i].QueueType == "RANKED_SOLO_5x5" {
				playerSlice["RankSoloQ"] = fmt.Sprintf("%s %s", schemeLolData[i].Tier, schemeLolData[i].Rank)
				playerSlice["LpQ"] = fmt.Sprintf("%v", schemeLolData[i].LeaguePoints)
				playerSlice["WinsQ"] = fmt.Sprintf("%v", schemeLolData[i].Wins)
				playerSlice["LosesQ"] = fmt.Sprintf("%v", schemeLolData[i].Losses)
			} else {
				playerSlice["RankFlex"] = fmt.Sprintf("%s %s", schemeLolData[i].Tier, schemeLolData[i].Rank)
				playerSlice["LpFlex"] = fmt.Sprintf("%v", schemeLolData[i].LeaguePoints)
				playerSlice["WinsFlex"] = fmt.Sprintf("%v", schemeLolData[i].Wins)
				playerSlice["LosesFlex"] = fmt.Sprintf("%v", schemeLolData[i].Losses)
			}
		}
		return nil
	}
	noRankError = errors.New("rank not found")
	return noRankError
}

// Create a formatted text message for telegram
func messageTextFormatter(playerSlice map[string]string, imgId string, schemeLolData schemes.LoLAccount) (string, string) {
	log.Println("Inside function messageTextFormatter")
	imgLink := "https://ddragon.leagueoflegends.com/cdn/12.16.1/img/profileicon/" + url.QueryEscape(imgId) + ".png"
	schemeLen := len(schemeLolData)
	formattedText := fmt.Sprintf("<b>Nickname:</b> %v\n", playerSlice["Nickname"])
	formattedText += fmt.Sprintf("<b>Level:</b> %v\n\n", playerSlice["Level"])

	if schemeLen == 1 {
		if schemeLolData[0].QueueType == "RANKED_SOLO_5x5" {
			formattedText += "<b>SoloQ</b> \n"
			formattedText += fmt.Sprintf("<b>Rank:</b> %v\n", playerSlice["RankSoloQ"])
			formattedText += fmt.Sprintf("<b>Wins:</b> %v\n", playerSlice["WinsQ"])
			formattedText += fmt.Sprintf("<b>Loses:</b> %v\n", playerSlice["LosesQ"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", playerSlice["LpQ"])
		} else {
			formattedText += "<b>Flex</b>\n"
			formattedText += fmt.Sprintf("<b>Rank:</b> %v\n", playerSlice["RankFlex"])
			formattedText += fmt.Sprintf("<b>Wins:</b> %v\n", playerSlice["WinsFlex"])
			formattedText += fmt.Sprintf("<b>Loses:</b> %v\n", playerSlice["LosesFlex"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", playerSlice["LpFlex"])
		}
	} else {
		if schemeLolData[0].QueueType == "RANKED_SOLO_5x5" || schemeLolData[1].QueueType == "RANKED_SOLO_5x5" {
			formattedText += "<b>SoloQ</b> \n"
			formattedText += fmt.Sprintf("<b>Rank:</b> %v\n", playerSlice["RankSoloQ"])
			formattedText += fmt.Sprintf("<b>Wins:</b> %v\n", playerSlice["WinsQ"])
			formattedText += fmt.Sprintf("<b>Loses:</b> %v\n", playerSlice["LosesQ"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", playerSlice["LpQ"])
		}
		if schemeLolData[0].QueueType == "RANKED_FLEX_SR" || schemeLolData[1].QueueType == "RANKED_FLEX_SR" {
			formattedText += "<b>Flex</b>\n"
			formattedText += fmt.Sprintf("<b>Rank:</b> %v\n", playerSlice["RankFlex"])
			formattedText += fmt.Sprintf("<b>Wins:</b> %v\n", playerSlice["WinsFlex"])
			formattedText += fmt.Sprintf("<b>Loses:</b> %v\n", playerSlice["LosesFlex"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", playerSlice["LpFlex"])
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
	formattedUrl := "https://api.telegram.org/bot" + url.QueryEscape(telegramApiKey) + "/sendPhoto?chat_id=" + url.QueryEscape(chatIdString) + "&reply_to_message_id=" + url.QueryEscape(messageIdString) + "&photo=" + url.QueryEscape(link) + "&caption=" + url.QueryEscape(message) + "&parse_mode=html"

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
	formattedUrl := "https://api.telegram.org/bot" + url.QueryEscape(telegramApiKey) + "/sendMessage?chat_id=" + url.QueryEscape(chatIdString) + "&reply_to_message_id=" + url.QueryEscape(messageIdString) + "&text=" + url.QueryEscape(message)

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

func checkCommand(body *schemes.WebhookBody) {
	// Instantiate every schemes
	var schemeAccount schemes.AccountRiot
	var schemeLoLData schemes.LoLAccount
	var schemeTgMessageResponse schemes.ApiTelegramMessage
	var schemePlayerNotFound schemes.PlayerNotFound

	playerInfo := make(map[string]string)

	// Remove /euw from the Message.Text
	body.Message.Text = strings.ReplaceAll(body.Message.Text, "/euw", "")
	log.Println("Nickname: ", body.Message.Text)

	// Get the id of the riot account
	responseAccountData := getAccountID(body.Message.Text)

	// Check if the player exist, if not exist send a message, assign to updateID the new value and goes to new cycle
	err := json.Unmarshal(responseAccountData, &schemePlayerNotFound)
	if err != nil {
		log.Println("Got an error unmarshal response account data to schemePlayerNotFound")
	}
	if schemePlayerNotFound.Status.StatusCode == 404 {
		message := fmt.Sprintf("Username %s is not valid!", body.Message.Text)
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
		formattedMessage := fmt.Sprintf("<b>Nickname:</b>%v\n", body.Message.Text)
		formattedMessage += fmt.Sprintf("<b>Level:</b> %v\n\n", playerInfo["Level"])
		imgLink := "https://ddragon.leagueoflegends.com/cdn/12.16.1/img/profileicon/" + url.QueryEscape(imgId) + ".png"
		formattedMessage += fmt.Sprintf("<b>Rank of player%s not found!</>", body.Message.Text)
		// Send a message to the message sender on telegram with the results
		responseMessage, err := sendPhotoMessage(body.Message.Chat.ID, body.Message.MessageID, formattedMessage, imgLink)

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
	log.Println("Listening on port 3000")
	err := http.ListenAndServe(":3000", http.HandlerFunc(Handler))
	if err != nil {
		log.Panic(err)
	}
}
