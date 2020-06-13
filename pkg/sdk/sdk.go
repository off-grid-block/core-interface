package sdk

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/pkg/errors"

	cic "github.com/off-grid-block/core-interface/pkg/config"
)

// FabricSetup implementation
type SDKConfig struct {
	ConfigFile      string
	OrgID           string
	OrdererID       string
	ChannelID       string
	initialized     bool
	ChannelConfig   string
	ChaincodeGoPath string
	ChaincodePath   map[string]string
	CollectionPath  map[string]string
	OrgAdmin        string
	OrgName         string
	UserName        string
	Client          *channel.Client
	Mgmt            *resmgmt.Client
	FabricSDK 		*fabsdk.FabricSDK
	Event           *event.Client
	MgmtIdentity	msp.SigningIdentity
}

// Set up DEON Admin SDK
func SetupSDK() (*SDKConfig, error) {

	var ccPath = map[string]string{"vote": "vote/chaincode"}

	fSetup := &SDKConfig {
		OrdererID: 			"orderer.example.com",
		ChannelID: 			"mychannel",
		ChannelConfig:		"/Users/brianli/deon/fabric-samples/first-network/channel-artifacts/channel.tx",
		ChaincodeGoPath:	"/Users/brianli/deon",
		ChaincodePath:		ccPath,
		OrgAdmin:			"Admin",
		OrgName:			"org1",
		ConfigFile:			"config.yaml",
		UserName:			"User1",
	}

	err := fSetup.Initialization()
	if err != nil {
		return fSetup, errors.WithMessage(err, "Unable to initialize the Fabric SDK")
	}

	// Close SDK
	defer fSetup.CloseSDK()

	err = fSetup.AdminSetup()
	if err != nil {
		return fSetup, errors.WithMessage(err, "Failed to set up admin")
	}

	err = fSetup.ChannelSetup()
	if err != nil {
		return fSetup, errors.WithMessage(err, "Failed to set up channel")
	}

	err = fSetup.ClientSetup()
	if err != nil {
		return fSetup, errors.WithMessage(err, "Failed to set up client")
	}

	err = fSetup.ChainCodeInstallationInstantiation()
	if err != nil {
		return fSetup, errors.WithMessage(err, "Failed to set up chaincodes")
	}

	return fSetup, nil
}

// Initialization setups new sdk
func (s *SDKConfig) Initialization() error {

	// Add parameters for the initialization
	if s.initialized {
		return errors.New("sdk is already initialized")
	}

	// Initialize the SDK with the configuration file
	fsdk, err := fabsdk.New(config.FromFile(s.ConfigFile))
	if err != nil {
		return errors.WithMessage(err, "failed to create SDK")
	}
	s.FabricSDK = fsdk
	fmt.Println("SDK is now created")

	fmt.Println("Initialization Successful")
	s.initialized = true

	return nil

}

func (s *SDKConfig) AdminSetup() error {

	// The resource management client is responsible for managing channels (create/update channel)
	resourceManagerClientContext := s.FabricSDK.Context(fabsdk.WithUser(s.OrgAdmin), fabsdk.WithOrg(s.OrgName))
	resMgmtClient, err := resmgmt.New(resourceManagerClientContext)
	if err != nil {
		return errors.WithMessage(err, "failed to create channel management client from Admin identity")
	}
	s.Mgmt = resMgmtClient
	fmt.Println("Resource management client created")

	// The MSP client allow us to retrieve user information from their identity, like its signing identity which we will need to save the channel
	mspClient, err := mspclient.New(s.FabricSDK.Context(), mspclient.WithOrg(s.OrgName))
	if err != nil {
		return errors.WithMessage(err, "failed to create MSP client")
	}

	s.MgmtIdentity, err = mspClient.GetSigningIdentity(s.OrgAdmin)
	if err != nil {
		return errors.WithMessage(err, "failed to get mgmt signing identity")
	}

	return nil
}

func (s *SDKConfig) ChannelSetup() error {

	req := resmgmt.SaveChannelRequest{ChannelID: s.ChannelID, ChannelConfigPath: s.ChannelConfig, SigningIdentities: []msp.SigningIdentity{s.MgmtIdentity}}
	//create channel
	txID, err := s.Mgmt.SaveChannel(req, resmgmt.WithOrdererEndpoint(s.OrdererID))
	if err != nil || txID.TransactionID == "" {
		return errors.WithMessage(err, "failed to save channel")
	}
	fmt.Println("Channel created")

	// Make mgmt user join the previously created channel
	if err = s.Mgmt.JoinChannel(s.ChannelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(s.OrdererID)); err != nil {
		return errors.WithMessage(err, "failed to make mgmt join channel")
	}
	fmt.Println("Channel joined")

	return nil
}

// Setup client and setupt access to channel events
func (s *SDKConfig)  ClientSetup() error {

	// Channel client is used to Query or Execute transactions
	var err error

	clientChannelContext := s.FabricSDK.ChannelContext(s.ChannelID, fabsdk.WithUser(s.UserName))
	s.Client, err = channel.New(clientChannelContext)
	if err != nil {
		return errors.WithMessage(err, "failed to create new channel client")
	}
	fmt.Println("Channel client created")

	// Creation of the client which will enables access to our channel events
	s.Event, err = event.New(clientChannelContext)
	if err != nil {
		return errors.WithMessage(err, "failed to create new event client")
	}
	fmt.Println("Event client created")

	return nil
}

// Close the SDK
func (s *SDKConfig) CloseSDK() {
	s.FabricSDK.Close()
}

// Installs and instantiates chaincode for all apps
func (s *SDKConfig) ChainCodeInstallationInstantiation() error {

	// install & instantiate cc for each of the apps
	for ccID, ccPath := range s.ChaincodePath {

		// Create the chaincode package that will be sent to the peers
		ccPackage, err := packager.NewCCPackage(ccPath, s.ChaincodeGoPath)
		if err != nil {
			return errors.WithMessage(err, "failed to create chaincode package")
		}

		fmt.Printf("Chaincode package %v created\n", ccID)

		// Install the chaincode to org peers
		installCCReq := resmgmt.InstallCCRequest{
			Name: ccID, 
			Path: ccPath, 
			Version: "0", 
			Package: ccPackage}
		_, err = s.Mgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			return errors.WithMessage(err, "failed to install chaincode")
		}

		fmt.Printf("Chaincode %v installed\n", ccID)

		// Set up chaincode policy (***DEFAULT HARD-CODED POLICY***)
		ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP"})

		cfg, err := cic.NewCollectionConfig(ccPath)
		if err != nil {
			return errors.WithMessage(err, "Failed to read collections config information")
		}

		// instantiate chaincode with cc policy and collection configs
		resp, err := s.Mgmt.InstantiateCC(
			// Channel ID
			s.ChannelID, 
			// InstantiateCCRequest struct
			resmgmt.InstantiateCCRequest{
				Name: ccID, 
				Path: s.ChaincodeGoPath, 
				Version: "0", 
				Args: [][]byte{[]byte("init")}, 
				Policy: ccPolicy, 
				CollConfig: cfg,
			},
			// options
			resmgmt.WithRetry(retry.DefaultResMgmtOpts))

		if err != nil || resp.TransactionID == "" {
			return errors.WithMessage(err, "failed to instantiate the chaincode")
		}

		fmt.Printf("Chaincode %v instantiated\n", ccID)
	}

	return nil
}