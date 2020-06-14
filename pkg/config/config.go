package config

import (
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"	
	"github.com/spf13/viper"
	// "fmt"
)

// type MyCollectionConfigMap struct {
// 	list []MyCollectionConfig
// }

// type MyCollectionConfig struct {
// 	name 				string 	`json:"name"`
// 	policy 				string 	`json:"policy"`
// 	requiredPeerCount 	int32 	`json:"requiredPeerCount"`
// 	maxPeerCount 		int32 	`json:"maxPeerCount"`
// 	blockToLive 		uint64 	`json:"blockToLive"`
// 	memberOnlyRead 		bool 	`json:"memberOnlyRead"`
// }


// TODO: Dynamically read from collections_config.json
// Create collection config to for chaincode instantiation
func NewCollectionConfig(collectionPath string) (*peer.CollectionConfig, error) {

	viper.SetConfigName("collections_config")
	viper.SetConfigType("json")
	viper.AddConfigPath(collectionPath)
	// err := viper.ReadInConfig()
	// if err != nil {
	// 	panic(fmt.Errorf("Fatal error reading in collections config: %s\n", err))
	// }

	name := viper.GetString("name")
	policy := viper.GetString("policy")
	requiredPeerCount := viper.Get("requiredPeerCount").(int32)
	maxPeerCount := viper.Get("maxPeerCount").(int32)
	blockToLive := viper.Get("blockToLive").(uint64)
	

	// // unmarshal config parameters into locally defined struct myCollectionConfig
	// var config MyCollectionConfig

	// err = viper.Unmarshal(&config)
	// if err != nil {
	// 	panic(fmt.Errorf("Failed to unmarshal collections config: %s\n", err))
	// }

	// marshal parameters into the externally defined CollectionConfig struct format
	// configList := make([]*peer.CollectionConfig, 4)

	// for _, config := range configs.list {

	// fmt.Println("***" + config.name + "***")
	p, err := cauthdsl.FromString(policy)
	if err != nil {
        return nil, err
    }

    cpc := &peer.CollectionPolicyConfig{
        Payload: &peer.CollectionPolicyConfig_SignaturePolicy{
            SignaturePolicy: p,
        },
    }

    return &peer.CollectionConfig{
        Payload: &peer.CollectionConfig_StaticCollectionConfig{
            StaticCollectionConfig: &peer.StaticCollectionConfig{
                Name:              name,
                MemberOrgsPolicy:  cpc,
                RequiredPeerCount: requiredPeerCount,
                MaximumPeerCount:  maxPeerCount,
                BlockToLive:       blockToLive}}}, nil
}