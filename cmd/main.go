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

func getRequestData(link, error_message string) io.ReadCloser {
	rawResponseData, err := http.Get(link)
	if err != nil {
		fmt.Println(error_message, err)
		return nil
	}

	return rawResponseData.Body
}

func filterRequestData(rawResponseData io.ReadCloser, error_message string) []byte {
	responseData, err := io.ReadAll(rawResponseData)
	if err != nil {
		fmt.Println(error_message, err)
		return nil
	}

	return responseData
}

func getAccountID(nickname string) []byte {
	accountLink := "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/" + url.QueryEscape(nickname) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponseAccountData := getRequestData(accountLink, "Got an error retrieving the Account ID api")

	responseAccountData := filterRequestData(rawResponseAccountData, "Got an error reading the account ID body")
	defer rawResponseAccountData.Close()
	return responseAccountData
}

func getRankData(id string) []byte {

	playerDataLink := "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/" + url.QueryEscape(id) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponsePlayerData := getRequestData(playerDataLink, "Got an error retrieving the rank")

	responsePlayerData := filterRequestData(rawResponsePlayerData, "Got an error reading the lolData body")

	defer rawResponsePlayerData.Close()
	return responsePlayerData
}

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

func getTelegramApi() []byte {
	rawResponseTelegramData := getRequestData("https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/getUpdates?offset=-1", "Got an error retrieving telegram response")
	responseTelegramData := filterRequestData(rawResponseTelegramData, "Got an error while reading the body")

	defer rawResponseTelegramData.Close()
	return responseTelegramData
}

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
	var schemeTg schemes.ApiTelegram
	var schemeAccount schemes.AccountRiot
	var schemeLoLData schemes.LoLAccount
	var schemeTgMessageResponse schemes.ApiTelegramMessage

	var updateID int

	for {
		// TODO: Refactor json.Unmarshal inside func filterRequestData
		rawTelegramResponseData := getTelegramApi()
		json.Unmarshal(rawTelegramResponseData, &schemeTg)

		fmt.Println("Sleeping")
		time.Sleep(5 * time.Second)

		if updateID == schemeTg.Result[0].Message.MessageID {
			fmt.Println("Inside if")
			continue
		}
		responseAccountData := getAccountID(schemeTg.Result[0].Message.Text)
		json.Unmarshal(responseAccountData, &schemeAccount)

		rankData := getRankData(schemeAccount.ID)
		json.Unmarshal(rankData, &schemeLoLData)

		rank, err := filterRankData(schemeLoLData)
		message := fmt.Sprintf("%s\n%s", schemeTg.Result[0].Message.Text, rank)
		if err != nil {
			log.Println(err)
			message = fmt.Sprintf("No rank found for %s", schemeTg.Result[0].Message.Text)
		}

		responseMessage, err := sendHttpMessage(schemeTg.Result[0].Message.Chat.ID, schemeTg.Result[0].Message.MessageID, message)
		if err != nil {
			log.Println(err)
		}
		json.Unmarshal(responseMessage, &schemeTgMessageResponse)
		updateID = schemeTg.Result[0].Message.MessageID

	}
}
