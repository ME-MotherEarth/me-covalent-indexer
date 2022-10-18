package miniblocks

import (
	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	"github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/block"
	"github.com/ME-MotherEarth/me-core/hashing"
	"github.com/ME-MotherEarth/me-core/marshal"
	"github.com/ME-MotherEarth/me-covalent-indexer"
	"github.com/ME-MotherEarth/me-covalent-indexer/schema"
	logger "github.com/ME-MotherEarth/me-logger"
)

var log = logger.GetOrCreate("covalent/process/block/miniBlocks/miniBlocksProcessor")

type miniBlocksProcessor struct {
	hasher     hashing.Hasher
	marshaller marshal.Marshalizer
}

// NewMiniBlocksProcessor will create a new instance of miniBlocksProcessor
func NewMiniBlocksProcessor(hasher hashing.Hasher, marshaller marshal.Marshalizer) (*miniBlocksProcessor, error) {
	if check.IfNil(marshaller) {
		return nil, covalent.ErrNilMarshaller
	}
	if check.IfNil(hasher) {
		return nil, covalent.ErrNilHasher
	}

	return &miniBlocksProcessor{
		hasher:     hasher,
		marshaller: marshaller,
	}, nil
}

// ProcessMiniBlocks converts mini blocks core data to a specific mini blocks structure array defined by avro schema
func (mbp *miniBlocksProcessor) ProcessMiniBlocks(header data.HeaderHandler, body data.BodyHandler) ([]*schema.MiniBlock, error) {
	moaBody, castOk := body.(*block.Body)
	if !castOk {
		return nil, covalent.ErrBlockBodyAssertion
	}

	moaMiniBlocks := moaBody.GetMiniBlocks()
	miniBlocks := make([]*schema.MiniBlock, 0, len(moaMiniBlocks))

	for _, mb := range moaMiniBlocks {

		miniBlock, err := mbp.processMiniBlock(mb, header)
		if err != nil {
			log.Warn("miniBlocksProcessor.ProcessMiniBlocks cannot process miniBlock", "error", err)
			continue
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return miniBlocks, nil
}

func (mbp *miniBlocksProcessor) processMiniBlock(miniBlock *block.MiniBlock, header data.HeaderHandler) (*schema.MiniBlock, error) {
	miniBlockHash, err := core.CalculateHash(mbp.marshaller, mbp.hasher, miniBlock)
	if err != nil {
		return nil, err
	}

	return &schema.MiniBlock{
		Hash:            miniBlockHash,
		TxHashes:        miniBlock.GetTxHashes(),
		SenderShardID:   int32(miniBlock.SenderShardID),
		ReceiverShardID: int32(miniBlock.ReceiverShardID),
		Type:            int32(miniBlock.Type),
		Timestamp:       int64(header.GetTimeStamp()),
	}, nil
}
