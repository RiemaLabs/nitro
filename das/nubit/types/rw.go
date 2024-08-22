package types

import (
	"bytes"
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/offchainlabs/nitro/arbstate/daprovider"
)

type SquareData struct {
	RowRoots    [][]byte   `json:"row_roots"`
	ColumnRoots [][]byte   `json:"column_roots"`
	Rows        [][][]byte `json:"rows"`
	SquareSize  uint64     `json:"square_size"` // Refers to original data square size
	StartRow    uint64     `json:"start_row"`
	EndRow      uint64     `json:"end_row"`
}

type NubitBlobReader interface {
	Read(context.Context, *BlobPointer) ([]byte, *SquareData, error)
	GetProof(ctx context.Context, msg []byte) ([]byte, error)
}

func NewReaderForNubit(nubitBlobReader NubitBlobReader) *readerForNubit {
	return &readerForNubit{nubitBlobReader: nubitBlobReader}
}

type readerForNubit struct {
	nubitBlobReader NubitBlobReader
}

func (c *readerForNubit) IsValidHeaderByte(headerByte byte) bool {
	return IsNubitMessageHeaderByte(headerByte)
}

// NubitMessageHeaderFlag indicates that this data is a Blob Pointer
// which will be used to retrieve data from Nubit
const NubitMessageHeaderFlag byte = 0xda

func hasBits(checking byte, bits byte) bool {
	return (checking & bits) == bits
}

func IsNubitMessageHeaderByte(header byte) bool {
	return hasBits(header, NubitMessageHeaderFlag)
}

func (c *readerForNubit) GetProof(ctx context.Context, msg []byte) ([]byte, error) {
	return c.nubitBlobReader.GetProof(ctx, msg)
}

func (c *readerForNubit) RecoverPayloadFromBatch(
	ctx context.Context,
	batchNum uint64,
	batchBlockHash common.Hash,
	sequencerMsg []byte,
	preimageRecorder daprovider.PreimageRecorder,
	validateSeqMsg bool,
) ([]byte, error) {
	return RecoverPayloadFromNubitBatch(ctx, batchNum, sequencerMsg, c.nubitBlobReader, preimageRecorder, validateSeqMsg)
}

func RecoverPayloadFromNubitBatch(
	ctx context.Context,
	batchNum uint64,
	sequencerMsg []byte,
	nubitBlobReader NubitBlobReader,
	preimageRecorder daprovider.PreimageRecorder,
	validateSeqMsg bool,
) ([]byte, error) {
	buf := bytes.NewBuffer(sequencerMsg[40:])

	header, err := buf.ReadByte()
	if err != nil {
		log.Error("Couldn't deserialize Nubit header byte", "err", err)
		return nil, nil
	}
	if !IsNubitMessageHeaderByte(header) {
		log.Error("Couldn't deserialize Nubit header byte", "err", errors.New("tried to deserialize a message that doesn't have the Nubit header"))
		return nil, nil
	}

	blobPointer := BlobPointer{}
	blobBytes := buf.Bytes()
	err = blobPointer.UnmarshalBinary(blobBytes)
	if err != nil {
		log.Error("Couldn't unmarshal Nubit blob pointer", "err", err)
		return nil, nil
	}

	payload, _, err := nubitBlobReader.Read(ctx, &blobPointer)
	if err != nil {
		log.Error("Failed to resolve blob pointer from nubit", "err", err)
		return nil, err
	}

	// we read a batch that is to be discarded, so we return the empty batch
	if len(payload) == 0 {
		return payload, nil
	}

	// if preimageRecorder != nil {
	// 	if squareData == nil {
	// 		log.Error("squareData is nil, read from replay binary, but preimages are empty")
	// 		return nil, err
	// 	}

	// 	odsSize := squareData.SquareSize / 2
	// 	rowIndex := squareData.StartRow
	// 	for _, row := range squareData.Rows {
	// 		treeConstructor := structures.NewConstructor(preimageRecorder, odsSize)
	// 		root, err := structures.ComputeNmtRoot(treeConstructor, uint(rowIndex), row)
	// 		if err != nil {
	// 			log.Error("Failed to compute row root", "err", err)
	// 			return nil, err
	// 		}

	// 		rowRootMatches := bytes.Equal(squareData.RowRoots[rowIndex], root)
	// 		if !rowRootMatches {
	// 			log.Error("Row roots do not match", "eds row root", squareData.RowRoots[rowIndex], "calculated", root)
	// 			log.Error("Row roots", "row_roots", squareData.RowRoots)
	// 			return nil, err
	// 		}
	// 		rowIndex += 1
	// 	}

	// 	rowsCount := len(squareData.RowRoots)
	// 	slices := make([][]byte, rowsCount+rowsCount)
	// 	copy(slices[0:rowsCount], squareData.RowRoots)
	// 	copy(slices[rowsCount:], squareData.ColumnRoots)

	// 	dataRoot := structures.HashFromByteSlices(preimageRecorder, slices)

	// 	dataRootMatches := bytes.Equal(dataRoot, blobPointer.DataRoot[:])
	// 	if !dataRootMatches {
	// 		log.Error("Data Root do not match", "blobPointer data root", blobPointer.DataRoot, "calculated", dataRoot)
	// 		return nil, nil
	// 	}
	// }

	return payload, nil
}

type NubitBlobWriter interface {
	Store(context.Context, []byte) ([]byte, error)
}

func NewWriterForNubit(nubitWriter NubitBlobWriter) *writerForNubit {
	return &writerForNubit{nubitWriter: nubitWriter}
}

type writerForNubit struct {
	nubitWriter NubitBlobWriter
}

func (c *writerForNubit) Store(ctx context.Context, message []byte, timeout uint64, disableFallbackStoreDataOnChain bool) ([]byte, error) {
	msg, err := c.nubitWriter.Store(ctx, message)
	if err != nil {
		if disableFallbackStoreDataOnChain {
			return nil, errors.New("unable to batch to Nubit and fallback storing data on chain is disabled")
		}
		return nil, err
	}
	message = msg
	return message, nil
}

func (c *writerForNubit) Name() string {
	return "Nubit"
}
