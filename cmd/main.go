package main

import (
	"encoding/json"
	"fmt"
	"io"
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

func filterRankData(schemeLoLData schemes.LoLAccount) string {
	var playerRanks string = ""
	var schemeLen int = len(schemeLoLData)

	for i := 0; i < schemeLen; i++ {
		if schemeLoLData[i].QueueType == "RANKED_SOLO_5x5" {
			playerRanks += fmt.Sprintf("Rank SoloQ: %s %s\n", schemeLoLData[i].Tier, schemeLoLData[i].Rank)
		} else {
			playerRanks += fmt.Sprintf("Rank Flex: %s %s\n", schemeLoLData[i].Tier, schemeLoLData[i].Rank)
		}
	}
	return playerRanks
}

func getTelegramApi() []byte {
	rawResponseTelegramData := getRequestData("https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/getUpdates?offset=-1", "Got an error retrieving telegram response")
	responseTelegramData := filterRequestData(rawResponseTelegramData, "Got an error while reading the body")

	defer rawResponseTelegramData.Close()
	return responseTelegramData
}

func main() {
	var schemeTg schemes.ApiTelegram
	var schemeAccount schemes.AccountRiot
	var schemeLoLData schemes.LoLAccount

	var updateID int

	for {
		// TODO: Refactor json.Unmarshal inside func filterRequestData
		rawTelegramResponseData := getTelegramApi()
		json.Unmarshal(rawTelegramResponseData, &schemeTg)

		fmt.Println("Sleeping")
		time.Sleep(5 * time.Second)

		if updateID == schemeTg.Result[0].Message.MessageID {
			continue
		}
		responseAccountData := getAccountID(schemeTg.Result[0].Message.Text)
		json.Unmarshal(responseAccountData, &schemeAccount)

		rankData := getRankData(schemeAccount.ID)
		json.Unmarshal(rankData, &schemeLoLData)

		rank := filterRankData(schemeLoLData)

		fmt.Printf("%s\n%s", schemeTg.Result[0].Message.Text, rank)
		updateID = schemeTg.Result[0].Message.MessageID

	}
}
