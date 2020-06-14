package sdk

import (
	"testing"
	"github.com/off-grid-block/core-interface/pkg/sdk"
)

func TestSetupSDK(t *testing.T) {

	fSetup := &sdk.SDKConfig {
		OrdererID: 			"orderer.example.com",
		ChannelID: 			"mychannel",
		ChannelConfig:		"/Users/brianli/deon/fabric-samples/first-network/channel-artifacts/channel.tx",
		ChaincodeGoPath:	"/Users/brianli/deon",
		ChaincodePath: 		make(map[string]string),
		OrgAdmin:			"Admin",
		OrgName:			"org1",
		ConfigFile:			"config_test.yaml",
		UserName:			"User1",
	}

	t.Run("Initialization", func(t *testing.T) {
		err := fSetup.Initialization()
		if err != nil {
			t.Errorf("Unable to initialize the Fabric SDK: %v\n", err)
			return
		}
	})

	defer fSetup.CloseSDK()

	t.Run("AdminSetup", func(t *testing.T) {
		err := fSetup.AdminSetup()
		if err != nil {
			t.Errorf("Failed to set up network admin: %v\n", err)
			return
		}
	})

	t.Run("ChannelSetup", func(t *testing.T) {
		err := fSetup.ChannelSetup()
		if err != nil {
			t.Errorf("Failed to set up channel: %v\n", err)
			return
		}
	})

	t.Run("ClientSetup", func(t *testing.T) {
		err := fSetup.ClientSetup()
		if err != nil {
			t.Errorf("Failed to set up client: %v\n", err)
			return
		}
	})

}