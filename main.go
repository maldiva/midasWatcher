package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/juliangruber/go-intersect"
	"github.com/navigante/midas-watch-list/db"
	"github.com/ybbus/jsonrpc"
)

var rpcClient = jsonrpc.NewClientWithOpts("http://127.0.0.1:44445", &jsonrpc.RPCClientOpts{
	CustomHeaders: map[string]string{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("u"+":"+"p")),
	},
})

type transactionInfo struct {
	Value    float64
	Adresses []string
}

type blockInformation map[string]interface{}

func getTransactionInfo(transactionHash string) ([]transactionInfo, error) {

	call, err := rpcClient.Call("getrawtransaction", transactionHash, 1)

	if err != nil {
		return []transactionInfo{}, err
	}

	transaction := call.Result.(map[string]interface{})

	preConvertedVout := transaction["vout"].([]interface{})

	vout := make([]map[string]interface{}, len(preConvertedVout))

	for i, d := range preConvertedVout {
		vout[i] = d.(map[string]interface{})
	}

	var res []transactionInfo

	for i := range vout {

		currentVout := vout[i]

		script := currentVout["scriptPubKey"].(map[string]interface{})

		preConvertedAdresses := script["addresses"]

		if preConvertedAdresses == nil {
			continue
		}

		temp := preConvertedAdresses.([]interface{})

		addresses := make([]string, len(temp))

		for i, d := range temp {
			addresses[i] = fmt.Sprintf("%v", d)
		}

		value, err := currentVout["value"].(json.Number).Float64()

		if err != nil {
			continue
		}

		res = append(res, transactionInfo{value, addresses})

	}

	return res, nil

}

func getBlockInfo(hash interface{}) (blockInformation, error) { // returns block information by hash

	block, err := rpcClient.Call("getblock", hash)

	if err != nil {
		return blockInformation{}, err
	}

	blockData := block.Result.(map[string]interface{})

	return blockData, nil

}

func getTransactionHashes(blockData blockInformation) []string { //returns all transaction hashes from block information

	st := fmt.Sprintf("%v", blockData["tx"])

	st = strings.TrimSuffix(st, "]")
	st = strings.TrimPrefix(st, "[")

	transactions := strings.Fields(st)

	return transactions
}

func findNewTransactions() ([]transactionInfo, error) {

	var res []transactionInfo

	tip, err := rpcClient.Call("getbestblockhash")

	if err != nil {
		return []transactionInfo{}, err
	}

	fmt.Println("Best block hash is ", tip.Result)

	tipHash := tip.Result

	currentHash := tip.Result

	lastVisibleBlockHash := db.GetLastVisibleBlockHash()

	if lastVisibleBlockHash == currentHash {
		return []transactionInfo{}, nil
	}

	for currentHash != lastVisibleBlockHash {

		currentBlockData, err := getBlockInfo(currentHash)
		if err != nil {
			return []transactionInfo{}, nil
		}

		currentBlockTransactions := getTransactionHashes(currentBlockData)

		for _, i := range currentBlockTransactions {

			txInfo, err := getTransactionInfo(i)

			if err != nil {
				fmt.Println(err)
				continue
			}

			if txInfo != nil { // simple check to not operate with non-standard transactions

				for i := range txInfo {

					// find intersection of adresses and convert back to []string
					preConvertedIntersection := intersect.Hash(txInfo[i].Adresses, db.GetWatchAddresses())
					temp := preConvertedIntersection.([]interface{})
					intersection := make([]string, len(temp))
					for k, j := range temp {
						intersection[k] = j.(string)
					}

					if len(intersection) > 0 {
						// unclear how to operate with value and []addresses
						res = append(res, txInfo[i])
					}

				}

			}
		}

		currentHash = currentBlockData["previousblockhash"]

	}

	db.SetLastVisibleBlockHash(tipHash.(string))

	return res, nil

}

func main() {

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			newIncomingTransactions, err := findNewTransactions()

			if err != nil {
				fmt.Println(err)
			}

			if len(newIncomingTransactions) > 0 {
				fmt.Println("Found new transactions on addresses from watchlist", newIncomingTransactions)
				fmt.Println("-------------------------------------------")
			} else {
				fmt.Println("No new transactions detected")
				fmt.Println("-------------------------------------------")
			}
		}
	}()
	select {}
}
