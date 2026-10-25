package acctest

import (
	"os"

	"github.com/labd/apollostudio-go-sdk/apollostudio"
)

func GetClient() (*apollostudio.Client, error) {
	key := os.Getenv("APOLLO_API_KEY")
	ref := os.Getenv("APOLLO_GRAPH_REF")

	client, err := apollostudio.NewClient(key, ref)

	if err != nil {
		return nil, err
	}

	return client, nil
}
