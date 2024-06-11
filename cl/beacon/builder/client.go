package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/cl/cltypes"
	"github.com/ledgerwatch/erigon/turbo/engineapi/engine_types"
	"github.com/ledgerwatch/log/v3"
)

var _ BuilderClient = &builderClient{}

type builderClient struct {
	// ref: https://ethereum.github.io/builder-specs/#/
	httpClient *http.Client
	url        *url.URL
}

func NewBlockBuilderClient(baseUrl string) *builderClient {
	u, err := url.Parse(baseUrl)
	if err != nil {
		panic(err)
	}
	c := &builderClient{
		httpClient: &http.Client{},
		url:        u,
	}
	if err := c.GetStatus(context.Background()); err != nil {
		log.Error("cannot connect to builder client", "url", baseUrl, "error", err)
		panic("cannot connect to builder client")
	}
	log.Info("Builder client is ready", "url", baseUrl)
	return c
}

func (b *builderClient) RegisterValidator(ctx context.Context, registers []*cltypes.ValidatorRegistration) error {
	// https://ethereum.github.io/builder-specs/#/Builder/registerValidator
	path := "/eth/v1/builder/validators"
	url := b.url.JoinPath(path).String()
	payload, err := json.Marshal(registers)
	if err != nil {
		return err
	}
	_, err = httpCall[json.RawMessage](ctx, b.httpClient, http.MethodPost, url, nil, bytes.NewBuffer(payload))
	if err != nil {
		log.Warn("[mev builder] httpCall error", "err", err)
	} else {
		log.Trace("[mev builder] RegisterValidator", "payload", string(payload))
	}
	return err
}

func (b *builderClient) GetExecutionPayloadHeader(ctx context.Context, slot int64, parentHash common.Hash, pubKey common.Bytes48) (*ExecutionPayloadHeader, error) {
	// https://ethereum.github.io/builder-specs/#/Builder/getHeader
	path := fmt.Sprintf("/eth/v1/builder/header/%d/%s/%s", slot, parentHash.Hex(), pubKey.Hex())
	url := b.url.JoinPath(path).String()
	header, err := httpCall[ExecutionPayloadHeader](ctx, b.httpClient, http.MethodGet, url, nil, nil)
	if err != nil {
		log.Warn("[mev builder] httpCall error", "err", err, "path", path)
		return nil, err
	}
	builderHeaderBytes, err := json.Marshal(header)
	if err != nil {
		log.Warn("[mev builder] json.Marshal error", "err", err)
		return nil, err
	} else {
		log.Info("[mev builder] builderHeaderBytes", "builderHeaderBytes", string(builderHeaderBytes))
	}
	return header, nil
}

func (b *builderClient) SubmitBlindedBlocks(ctx context.Context, block *cltypes.SignedBlindedBeaconBlock) (*cltypes.Eth1Block, *engine_types.BlobsBundleV1, error) {
	// https://ethereum.github.io/builder-specs/#/Builder/submitBlindedBlocks
	path := "/eth/v1/builder/blinded_blocks"
	url := b.url.JoinPath(path).String()
	payload, err := json.Marshal(block)
	if err != nil {
		return nil, nil, err
	}
	headers := map[string]string{
		"Eth-Consensus-Version": block.Version().String(),
	}
	resp, err := httpCall[BlindedBlockResponse](ctx, b.httpClient, http.MethodPost, url, headers, bytes.NewBuffer(payload))
	if err != nil {
		log.Warn("[mev builder] httpCall error", "headers", headers, "err", err, "payload", string(payload))
		return nil, nil, err
	}

	var eth1Block *cltypes.Eth1Block
	var blobsBundle *engine_types.BlobsBundleV1
	switch resp.Version {
	case "bellatrix", "capella":
		eth1Block = &cltypes.Eth1Block{}
		if err := json.Unmarshal(resp.Data, block); err != nil {
			return nil, nil, err
		}
	case "deneb":
		denebResp := &struct {
			ExecutionPayload *cltypes.Eth1Block          `json:"execution_payload"`
			BlobsBundle      *engine_types.BlobsBundleV1 `json:"blobs_bundle"`
		}{}
		if err := json.Unmarshal(resp.Data, denebResp); err != nil {
			return nil, nil, err
		}
		eth1Block = denebResp.ExecutionPayload
		blobsBundle = denebResp.BlobsBundle
	}
	// log
	eth1blockBytes, err := json.Marshal(eth1Block)
	if err != nil {
		log.Warn("[mev builder] json.Marshal error", "err", err)
		return nil, nil, err
	} else {
		log.Info("[mev builder] eth1blockBytes", "eth1blockBytes", string(eth1blockBytes))
	}
	blobsBundleBytes, err := json.Marshal(blobsBundle)
	if err != nil {
		log.Warn("[mev builder] json.Marshal error", "err", err)
		return nil, nil, err
	} else {
		log.Info("[mev builder] blobsBundleBytes", "blobsBundleBytes", string(blobsBundleBytes))
	}
	return eth1Block, blobsBundle, nil
}

func (b *builderClient) GetStatus(ctx context.Context) error {
	path := "/eth/v1/builder/status"
	url := b.url.JoinPath(path).String()
	_, err := httpCall[json.RawMessage](ctx, b.httpClient, http.MethodGet, url, nil, nil)
	return err
}

func httpCall[T any](ctx context.Context, client *http.Client, method, url string, headers map[string]string, payloadReader io.Reader) (*T, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, payloadReader)
	if err != nil {
		log.Warn("[mev builder] http.NewRequest failed", "err", err, "url", url, "method", method)
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	// send request
	response, err := client.Do(request)
	if err != nil {
		log.Warn("[mev builder] client.Do failed", "err", err, "url", url, "method", method)
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		// read response body
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Warn("[mev builder] io.ReadAll failed", "err", err, "url", url, "method", method)
		}
		return nil, fmt.Errorf("status code: %d. Response content %v", response.StatusCode, string(bytes))
	}
	// read response body
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Warn("[mev builder] io.ReadAll failed", "err", err, "url", url, "method", method)
		return nil, err
	}
	log.Info("[mev builder] httpCall success", "url", url, "method", method, "response", string(bytes), "statusCode", response.StatusCode)

	var body T
	if len(bytes) == 0 {
		return &body, nil
	}
	if err := json.Unmarshal(bytes, &body); err != nil {
		log.Warn("[mev builder] json.Unmarshal error", "err", err, "content", string(bytes))
		return nil, err
	}
	return &body, nil
}
