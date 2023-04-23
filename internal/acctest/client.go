package acctest

import (
	"github.com/labd/go-apollostudio-sdk/pkg/apollostudio"
	"os"
)

func GetClient() (*apollostudio.Client, error) {
	key := os.Getenv("APOLLO_API_KEY")
	ref := os.Getenv("APOLLO_GRAPH_REF")

	client, err := apollostudio.NewClient(
		apollostudio.ClientOpts{
			APIKey:   key,
			GraphRef: ref,
		},
	)

	if err != nil {
		return nil, err
	}

	return client, nil
}
