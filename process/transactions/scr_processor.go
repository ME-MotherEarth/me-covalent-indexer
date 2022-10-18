package transactions

import (
	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	"github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/smartContractResult"
	covalent "github.com/ME-MotherEarth/me-covalent-indexer"
	"github.com/ME-MotherEarth/me-covalent-indexer/process/utility"
	"github.com/ME-MotherEarth/me-covalent-indexer/schema"
)

type scProcessor struct {
	pubKeyConverter core.PubkeyConverter
}

// NewSCResultsProcessor creates a new instance of smart contracts processor
func NewSCResultsProcessor(pubKeyConverter core.PubkeyConverter) (*scProcessor, error) {
	if check.IfNil(pubKeyConverter) {
		return nil, covalent.ErrNilPubKeyConverter
	}

	return &scProcessor{pubKeyConverter: pubKeyConverter}, nil
}

// ProcessSCRs converts smart contracts data to a specific structure defined by avro schema
func (scp *scProcessor) ProcessSCRs(transactions map[string]data.TransactionHandler, timeStamp uint64) []*schema.SCResult {
	allSCRs := make([]*schema.SCResult, 0, len(transactions))

	for currTxHash, currTx := range transactions {
		currSCR := scp.processSCResult(currTx, currTxHash, timeStamp)

		if currSCR != nil {
			allSCRs = append(allSCRs, currSCR)
		}
	}
	return allSCRs
}

func (scp *scProcessor) processSCResult(tx data.TransactionHandler, txHash string, timeStamp uint64) *schema.SCResult {
	scrTx, castOk := tx.(*smartContractResult.SmartContractResult)
	if !castOk {
		return nil
	}

	var relayerAddress []byte
	if len(scrTx.GetRelayerAddr()) > 0 {
		relayerAddress = utility.EncodePubKey(scp.pubKeyConverter, scrTx.GetRelayerAddr())
	}

	return &schema.SCResult{
		Hash:           []byte(txHash),
		Nonce:          int64(scrTx.GetNonce()),
		GasLimit:       int64(scrTx.GetGasLimit()),
		GasPrice:       int64(scrTx.GetGasPrice()),
		Value:          utility.GetBytes(scrTx.GetValue()),
		Sender:         utility.EncodePubKey(scp.pubKeyConverter, scrTx.GetSndAddr()),
		Receiver:       utility.EncodePubKey(scp.pubKeyConverter, scrTx.GetRcvAddr()),
		RelayerAddr:    relayerAddress,
		RelayedValue:   utility.GetBytes(scrTx.GetRelayedValue()),
		Code:           scrTx.GetCode(),
		Data:           scrTx.GetData(),
		PrevTxHash:     scrTx.GetPrevTxHash(),
		OriginalTxHash: scrTx.GetOriginalTxHash(),
		CallType:       int32(scrTx.GetCallType()),
		CodeMetadata:   scrTx.GetCodeMetadata(),
		ReturnMessage:  scrTx.GetReturnMessage(),
		Timestamp:      int64(timeStamp),
	}
}
