package config

import (
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"	
	"github.com/spf13/viper"
	"fmt"
)

type myCollectionConfig struct {
	name 				string 	`json:"name"`
	policy 				string 	`json:"policy"`
	requiredPeerCount 	int32 	`json:"requiredPeerCount"`
	maxPeerCount 		int32 	`json:"maxPeerCount"`
	blockToLive 		uint64 	`json:"blockToLive"`
	memberOnlyRead 		bool 	`json:"memberOnlyRead"`
}

// Create collection config to for chaincode instantiation
func NewCollectionConfig(collectionPath string) ([]*peer.CollectionConfig, error) {

	viper.SetConfigName("collections_config")
	viper.SetConfigType("json")
	viper.AddConfigPath(collectionPath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error reading in collections config: %s\n", err))
	}

	// unmarshal config parameters into locally defined struct myCollectionConfig
	var configs []myCollectionConfig

	err = viper.Unmarshal(&configs)
	if err != nil {
		panic(fmt.Errorf("Failed to unmarshal collections config: %s\n", err))
	}

	// marshal parameters into the externally defined CollectionConfig struct format
	configList := make([]*peer.CollectionConfig, 0)

	for _, config := range configs {

		p, err := cauthdsl.FromString(config.policy)
		if err != nil {
	        return configList, err
	    }

	    cpc := &peer.CollectionPolicyConfig{
	        Payload: &peer.CollectionPolicyConfig_SignaturePolicy{
	            SignaturePolicy: p,
	        },
	    }

	    configList = append(configList, &peer.CollectionConfig{
	        Payload: &peer.CollectionConfig_StaticCollectionConfig{
	            StaticCollectionConfig: &peer.StaticCollectionConfig{
	                Name:              config.name,
	                MemberOrgsPolicy:  cpc,
	                RequiredPeerCount: config.requiredPeerCount,
	                MaximumPeerCount:  config.maxPeerCount,
	                BlockToLive:       config.blockToLive}}})
	}

	return configList, nil
}