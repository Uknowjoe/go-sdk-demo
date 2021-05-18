package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var (
	cc          = ""
	user        = ""
	secret      = ""
	channelName = ""
	lvl         = logging.INFO
)


func main() {
	fmt.Println("Reading connection profile..")
	c := config.FromFile("./connection-profile.yaml")
	sdk, err := fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	setupLogLevel()
	enrollUser(sdk)

	clientChannelContext := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
	ledgerClient, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", channelName, err)
		os.Exit(1)
	}

	fmt.Printf("\n===== Channel: %s ===== \n", channelName)
	queryChannelInfo(ledgerClient)
	queryChannelConfig(ledgerClient)

	fmt.Println("\n====== Chaincode =========")

	client, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:", channelName, err)
	}

	invokeCC(client, "100")
	old := queryCC(client, []byte("john"))

	oldInt, _ := strconv.Atoi(old)
	invokeCC(client, strconv.Itoa(oldInt+1))

	queryCC(client, []byte("john"))

	fmt.Println("===============")
	fmt.Println("Done.")
}

func queryInstalledCC(sdk *fabsdk.FabricSDK) {
	userContext := sdk.Context(fabsdk.WithUser(user))

	resClient, err := resmgmt.New(userContext)
	if err != nil {
		fmt.Println("Failed to create resmgmt: ", err)
	}

	resp2, err := resClient.QueryInstalledChaincodes()
	if err != nil {
		fmt.Println("Failed to query installed cc: ", err)
	}
	fmt.Println("Installed cc: ", resp2.GetChaincodes())
}

func queryCC(client *channel.Client, name []byte) string {
	var queryArgs = [][]byte{name}
	response, err := client.Query(channel.Request{
		ChaincodeID: cc,
		Fcn:         "query",
		Args:        queryArgs,
	})

	if err != nil {
		fmt.Println("Failed to query: ", err)
	}

	ret := string(response.Payload)
	fmt.Println("Chaincode status: ", response.ChaincodeStatus)
	fmt.Println("Payload: ", ret)
	return ret
}

func invokeCC(client *channel.Client, newValue string) {
	fmt.Println("Invoke cc with new value:", newValue)
	invokeArgs := [][]byte{[]byte("john"), []byte(newValue)}

	_, err := client.Execute(channel.Request{
		ChaincodeID: cc,
		Fcn:         "set",
		Args:        invokeArgs,
	})

	if err != nil {
		fmt.Printf("Failed to invoke: %+v\n", err)
	}
}

func enrollUser(sdk *fabsdk.FabricSDK) {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)
	if err != nil {
		fmt.Printf("Failed to create msp client: %s\n", err)
	}

	_, err = mspClient.GetSigningIdentity(user)
	if err == msp.ErrUserNotFound {
		fmt.Println("Going to enroll user")
		err = mspClient.Enroll(user, msp.WithSecret(secret))

		if err != nil {
			fmt.Printf("Failed to enroll user: %s\n", err)
		} else {
			fmt.Printf("Success enroll user: %s\n", user)
		}

	} else if err != nil {
		fmt.Printf("Failed to get user: %s\n", err)
	} else {
		fmt.Printf("User %s already enrolled, skip enrollment.\n", user)
	}
}

func queryChannelConfig(ledgerClient *ledger.Client) {
	resp1, err := ledgerClient.QueryConfig()
	if err != nil {
		fmt.Printf("Failed to queryConfig: %s", err)
	}
	fmt.Println("ChannelID: ", resp1.ID())
	fmt.Println("Channel Orderers: ", resp1.Orderers())
	fmt.Println("Channel Versions: ", resp1.Versions())
}

func queryChannelInfo(ledgerClient *ledger.Client) {
	resp, err := ledgerClient.QueryInfo()
	if err != nil {
		fmt.Printf("Failed to queryInfo: %s", err)
	}
	fmt.Println("BlockChainInfo:", resp.BCI)
	fmt.Println("Endorser:", resp.Endorser)
	fmt.Println("Status:", resp.Status)
}

func setupLogLevel() {
	logging.SetLevel("fabsdk", lvl)
	logging.SetLevel("fabsdk/common", lvl)
	logging.SetLevel("fabsdk/fab", lvl)
	logging.SetLevel("fabsdk/client", lvl)
}
