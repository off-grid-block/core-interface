package config

import (
	"testing"
	"github.com/off-grid-block/core-interface/pkg/config"
	// "github.com/hyperledger/fabric-protos-go/peer"
	// "encoding/json"
)

func TestCollectionsConfig(t *testing.T) {

	t.Run("NewCollectionConfig", func(t *testing.T) {

		config, err := config.NewCollectionConfig("/Users/brianli/deon/core-interface/pkg/config")
		if err != nil {
			t.Errorf("Failed to create new collection config: %v\n", err)
			return
		}

		t.Logf(config.GetStaticCollectionConfig().Name)

	})

}