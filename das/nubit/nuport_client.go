package nubit

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/offchainlabs/nitro/util/pretty"

	"github.com/offchainlabs/nitro/das/nubit/structures"
	nTypes "github.com/offchainlabs/nitro/das/nubit/types"

	da "github.com/rollkit/go-da"
	"github.com/rollkit/go-da/proxy"
)

type NuportClient struct {
	client     da.DA
	namespace  da.Namespace
	commitTime time.Time
	cfg        NubitConfig
}

func NewNuportClient(
	cfg NubitConfig,
) (*NuportClient, error) {
	if cfg.Enable {
		// To avoid the linter.
	}
	cn, err := proxy.NewClient(cfg.Url, cfg.Authkey)
	if err != nil {
		return nil, err
	}

	hexStr := hex.EncodeToString([]byte(cfg.Namespace))
	name, err := hex.DecodeString(strings.Repeat("0", int(structures.NamespaceSize*2)-len(hexStr)) + hexStr)
	if err != nil {
		log.Error("decode NubitDA namespace failed", "err", err)
		return nil, err
	}
	log.Info("NubitDABackend", "namespace", hex.EncodeToString(name))

	return &NuportClient{
		client:     cn,
		namespace:  name,
		commitTime: time.Now(),
		cfg:        cfg,
	}, nil
}

func (c *NuportClient) String() string {
	return fmt.Sprintf("NubitDASClient{url:%s}", c.cfg.Url)
}

// TODO(Nubit): Version 2: Implement the Store method to return the blobPointer.
func (c *NuportClient) Store(ctx context.Context, message []byte) ([]byte, error) {
	log.Trace("nubit.NuportClient.Store(...)", "message", pretty.FirstFewBytes(message))
	blobIDs, err := c.client.Submit(ctx, [][]byte{message}, -1.0, c.namespace)
	// Ensure only a single blobID returned
	if err != nil || len(blobIDs) != 1 {
		log.Error("submit batch data with NubitDA client failed", "err", err)
		return nil, err
	} else {
		log.Info("submit batch data with NubitDA client succeed", "blobCommitment", blobIDs[0])
	}
	return blobIDs[0], nil
}

// TODO(Nubit): Version 2: Implement the Read method to return SquareData.
func (c *NuportClient) Read(ctx context.Context, blobPointer *nTypes.BlobPointer) ([]byte, *nTypes.SquareData, error) {
	log.Trace("nubit.NuportClient.Read(...)", "blobPointer", blobPointer)
	return []byte{}, nil, nil
}

// TODO(Nubit): Version 2: Implement the GetProof method to return the proof.
func (c *NuportClient) GetProof(ctx context.Context, msg []byte) ([]byte, error) {
	return []byte{}, nil
}
