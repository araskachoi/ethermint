package tx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/urfave/cli"
)

type Request struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Response struct {
	Error  *RPCError       `json:"error"`
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
}

var (
	HOST = os.Getenv("HOST")

	SendTx = cli.Command{
		Name:   "sendtx",
		Usage:  "Command to send transactions",
		Action: sendTx,
		Flags:  []cli.Flag{},
	}

	GenerateAccts = cli.Command{
		Name:   "genAccts",
		Usage:  "Generate given number of accounts",
		Action: genAccts,
		Flags:  []cli.Flag{},
	}
)

func getRandAcct(min, max) {
	return rand.Intn(max-min+1) + min
}

func createRequest(method string, params interface{}) Request {
	return Request{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
}

func call(method string, params interface{}) (*Response, error) {
	req, err := json.Marshal(createRequest(method, params))
	if err != nil {
		return nil, err
	}

	var rpcRes *Response
	time.Sleep(1 * time.Second)
	/* #nosec */
	res, err := http.Post(HOST, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	rpcRes = new(Response)
	err = decoder.Decode(&rpcRes)
	if err != nil {
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	return rpcRes, nil
}

func sendTx(accts []sdk.AccAddress, value, gasLimit, gasPrice string, maxTx int) error {

	ticker := time.NewTicker(time.Duration(600) * time.Nanosecond)
	defer ticker.Stop()

	eChan := make(chan error)
	txs := 0

	select {
	case <-ticker.C:
		txs++

		if txs >= maxTx {
			ticker.Stop()
		}

		//roundrobin style tx sending
		go func(e chan error) {
			param := make([]map[string]string, 1)
			param[0] = make(map[string]string)
			from := accts[getRandAcct(0,len(accts)]
			to := accts[getRandAcct(0,len(accts)]
			for (from == to) {
				to = accts[getRandAcct(0,len(accts)]
			}
			param[0]["from"] = "0x" + fmt.Sprintf("%x", accts[getRandAcct(0,len(accts)])
			param[0]["to"] = "0x" + fmt.Sprintf("%x", accts[getRandAcct(0,len(accts)])
			param[0]["value"] = "3B9ACA00"     //replace this with value
			param[0]["gasLimit"] = "0x5208"    //replace this with gasLimit
			param[0]["gasPrice"] = "0x15EF3C0" //replace this with gasPrice

			rpcRes, err := call("eth_sendTransaction", param)
			if err != nil {
				eChan <- err
			}

			var hash hexutil.Bytes
			err = json.Unmarshal(rpcRes.Result, &hash)
			if err != nil {
				eChan <- err
			}
		}(eChan)
	}

	return nil
}

func genAccts(noAccts uint64) []sdk.AccAddress {
	out := []sdk.AccAddress{}
	for i := uint64(0); i < noAccts; i++ {
		pubkey := secp256k1.GenPrivKey().PubKey()
		addr := sdk.AccAddress(pubkey.Address())
		out = append(out, addr)
		// baseAcc := auth.NewBaseAccount(addr, pubkey, i, i)
		// ethAcc := types.EthAccount{BaseAccount: baseAcc, CodeHash: []byte{1, 2}}
		// println(ethAcc)
	}

	return out
}
