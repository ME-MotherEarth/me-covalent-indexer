package mock

import (
	"github.com/ME-MotherEarth/me-core/data/indexer"
	"github.com/ME-MotherEarth/me-covalent-indexer/schema"
)

type DataHandlerStub struct {
	ProcessDataCalled func(args *indexer.ArgsSaveBlockData) (*schema.BlockResult, error)
}

func (dhs *DataHandlerStub) ProcessData(args *indexer.ArgsSaveBlockData) (*schema.BlockResult, error) {
	if dhs.ProcessDataCalled != nil {
		return dhs.ProcessDataCalled(args)
	}
	return nil, nil
}
