package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/eXpansiiVe/LoLBot/pkg/schemes"
)

const apiKey string = "RGAPI-e719f2a1-8946-46fe-82f8-9c9a56c6eee8"

// Used to get a response from an API
func getRequestData(link, error_message string) io.ReadCloser {
	rawResponseData, err := http.Get(link)
	if err != nil {
		fmt.Println(error_message, err)
		return nil
	}

	return rawResponseData.Body
}

// A simple reader use to read the ResponseDate passed by parameter
func filterRequestData(rawResponseData io.ReadCloser, error_message string) []byte {
	responseData, err := io.ReadAll(rawResponseData)
	if err != nil {
		fmt.Println(error_message, err)
		return nil
	}

	return responseData
}

// Get the summoner id through API
func getAccountID(nickname string) []byte {
	accountLink := "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/" + url.QueryEscape(nickname) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponseAccountData := getRequestData(accountLink, "Got an error retrieving the Account ID api")

	responseAccountData := filterRequestData(rawResponseAccountData, "Got an error reading the account ID body")
	defer rawResponseAccountData.Close()
	return responseAccountData
}

// Get the summoner Data through API
func getRankData(id string) []byte {

	playerDataLink := "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/" + url.QueryEscape(id) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponsePlayerData := getRequestData(playerDataLink, "Got an error retrieving the summoner data")

	responsePlayerData := filterRequestData(rawResponsePlayerData, "Got an error reading the summoner Data body")

	defer rawResponsePlayerData.Close()
	return responsePlayerData
}

// Parse the data to check if there's a rank and return it
func filterRankData(schemeLoLData schemes.LoLAccount) (string, error) {
	var playerRanks string = ""
	var schemeLen int = len(schemeLoLData)
	var noRankError error

	if schemeLen > 0 {
		for i := 0; i < schemeLen; i++ {
			if schemeLoLData[i].QueueType == "RANKED_SOLO_5x5" {
				playerRanks += fmt.Sprintf("Rank SoloQ: %s %s\n", schemeLoLData[i].Tier, schemeLoLData[i].Rank)
			} else {
				playerRanks += fmt.Sprintf("Rank Flex: %s %s\n", schemeLoLData[i].Tier, schemeLoLData[i].Rank)
			}
		}
		return playerRanks, nil
	}
	noRankError = errors.New("no rank found")
	return "", noRankError

}

// Get telegram data through API
func getTelegramApi() []byte {
	rawResponseTelegramData := getRequestData("https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/getUpdates?offset=-1", "Got an error retrieving telegram response")
	responseTelegramData := filterRequestData(rawResponseTelegramData, "Got an error while reading the body")

	defer rawResponseTelegramData.Close()
	return responseTelegramData
}

// Send response to telegram through API and return a response
func sendHttpMessage(chatId int64, messageId int, message string) ([]byte, error) {
	chatIdString := fmt.Sprintf("%d", chatId)
	messageIdString := fmt.Sprintf("%d", messageId)
	formattedUrl := "https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/sendMessage?chat_id=" + url.QueryEscape(chatIdString) + "&reply_to_message_id=" + url.QueryEscape(messageIdString) + "&text=" + url.QueryEscape(message)

	rawResponseData, err := http.Get(formattedUrl)
	if err != nil {
		return nil, err
	}
	defer rawResponseData.Body.Close()
	responseData, err := io.ReadAll(rawResponseData.Body)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func main() {
	// Instanciate every schemes
	var schemeTg schemes.ApiTelegram
	var schemeAccount schemes.AccountRiot
	var schemeLoLData schemes.LoLAccount
	var schemeTgMessageResponse schemes.ApiTelegramMessage

	var updateID int

	for {
		// Get telegram response data
		rawTelegramResponseData := getTelegramApi()
		json.Unmarshal(rawTelegramResponseData, &schemeTg)
		// Sleep the app to not spam too much requests
		fmt.Println("Sleeping")
		time.Sleep(5 * time.Second)

		// Check to not cycle to the same message
		if updateID == schemeTg.Result[0].Message.MessageID {
			fmt.Println("Inside if")
			continue
		}
		// Get the id of the riot account
		responseAccountData := getAccountID(schemeTg.Result[0].Message.Text)
		json.Unmarshal(responseAccountData, &schemeAccount)

		// Get the summoner data response
		rankData := getRankData(schemeAccount.ID)
		json.Unmarshal(rankData, &schemeLoLData)

		// Check if there's a schemeLoLData.QueueType and if there's any take the actual rank 
		rank, err := filterRankData(schemeLoLData)
		message := fmt.Sprintf("%s\n%s", schemeTg.Result[0].Message.Text, rank)
		if err != nil {
			log.Println(err)
			message = fmt.Sprintf("No rank found for %s", schemeTg.Result[0].Message.Text)
		}

		// Send a message to the message sender on telegram with the results
		responseMessage, err := sendHttpMessage(schemeTg.Result[0].Message.Chat.ID, schemeTg.Result[0].Message.MessageID, message)
		if err != nil {
			log.Println(err)
		}

		json.Unmarshal(responseMessage, &schemeTgMessageResponse)
		fmt.Println("Response from telegram message ", schemeTgMessageResponse.Ok)
		// Assign the messageId to the var updateID to not cycle on the same message
		updateID = schemeTg.Result[0].Message.MessageID

	}
}
