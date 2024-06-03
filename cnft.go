package mint_cnft

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/omarkilani/fluffle"
)

const SHYFT_CNFT_ENDPOINT = "https://api.shyft.to/sol/v1/nft/compressed/mint"

type ShyftCNFTPayload struct {
	Network           string  `json:"network"`
	Creator           string  `json:"creator_wallet"`
	MetadataURI       string  `json:"metadata_uri"`
	MerkleTree        string  `json:"merkle_tree"`
	CollectionAddress *string `json:"collection_address,omitempty"`
	Receiver          *string `json:"receiver,omitempty"`
	PriorityFee       uint64  `json:"priority_fee,omitempty"`
}

type ShyftCNFTResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Result  ShyftCNFTResult `json:"result"`
}

type ShyftCNFTResult struct {
	EncodedTransaction string   `json:"encoded_transaction"`
	Mint               string   `json:"mint"`
	Signers            []string `json:"signers"`
}

type CNFT struct {
	Signature string `json:"signature"`
	Mint      string `json:"mint"`
}

func MintCNFT(apiKey string, endpoint string, creator string, metadata_uri string, merkle_tree string, collection_address *string, receiver *string, priorityFee uint64) (cnft CNFT, err error) {
	headers := map[string]string{
		"x-api-key": apiKey,
	}

	payload := ShyftCNFTPayload{
		Network:           "mainnet-beta",
		Creator:           creator,
		MetadataURI:       metadata_uri,
		MerkleTree:        merkle_tree,
		CollectionAddress: collection_address,
		Receiver:          receiver,
		PriorityFee:       priorityFee,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return cnft, err
	}

	log.Println("Minting CNFT with body:", string(body))
	resp := fluffle.HTTPPostWithHeaders(SHYFT_CNFT_ENDPOINT, "application/json", []byte(body), headers)
	log.Println("Response:", string(resp))

	var shyftCNFTResponse ShyftCNFTResponse
	err = json.Unmarshal(resp, &shyftCNFTResponse)
	if err != nil {
		return cnft, err
	}

	decoded, err := base64.StdEncoding.DecodeString(shyftCNFTResponse.Result.EncodedTransaction)
	if err != nil {
		return cnft, err
	}

	tx, err := types.TransactionDeserialize(decoded)
	if err != nil {
		return cnft, err
	}
	accountJSON := os.Getenv("CNFT_MINT_ACCOUNT")
	if accountJSON == "" {
		return cnft, fmt.Errorf("CNFT_MINT_ACCOUNT not set")
	}

	var bytes []byte
	if err := json.Unmarshal([]byte(accountJSON), &bytes); err != nil {
		return cnft, err
	}

	account, err := types.AccountFromBytes(bytes)
	if err != nil {
		return cnft, err
	}

	rawMsg, err := tx.Message.Serialize()
	if err != nil {
		return cnft, err
	}

	tx.Signatures[0] = account.Sign(rawMsg)

	sig, err := client.NewClient(endpoint).SendTransaction(context.Background(), tx)
	if err != nil {
		log.Printf("MintCNFT: failed to send tx, err: %v", err)
		return cnft, err
	}

	cnft.Signature = sig
	cnft.Mint = shyftCNFTResponse.Result.Mint

	log.Printf("MintCNFT: sig: %v, mint %v", cnft.Signature, cnft.Mint)

	return cnft, nil
}
