package disk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

type Disk struct {
	dbPath string
}

func New(dbPath string) (*Disk, error) {
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}
	return &Disk{
		dbPath: dbPath,
	}, nil
}

func (d *Disk) Close() error {
	return nil
}

func (d *Disk) Write(blockData database.BlockData) error {
	// Mashal the block data for writing to disk in a more human readable format.
	data, err := json.MarshalIndent(blockData, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(d.getPath(blockData.Header.Number), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

func (d *Disk) getPath(blockNum uint64) string {
	name := strconv.FormatUint(blockNum, 10)
	return path.Join(d.dbPath, fmt.Sprintf("%s.json", name))
}

func (d *Disk) GetBlockByNumber(num uint64) (database.BlockData, error) {
	f, err := os.OpenFile(d.getPath(num), os.O_RDONLY, 0600)
	if err != nil {
		return database.BlockData{}, err
	}

	// Decode the contents of the block
	var blockData database.BlockData

	if err := json.NewDecoder(f).Decode(&blockData); err != nil {
		return database.BlockData{}, err
	}

	return blockData, nil
}

func (d *Disk) GetBlock(hash string) (database.BlockData, error) {
	// return d.GetBlock(num)
	return database.BlockData{}, errors.New("not implemented")
}

func (d *Disk) Reset() error {
	if err := os.RemoveAll(d.dbPath); err != nil {
		return err
	}

	return os.MkdirAll(d.dbPath, 0755)
}

func (d *Disk) ForEach() database.Iterator {
	return &diskIterator{
		storage: d,
	}
}

type diskIterator struct {
	storage *Disk  // Access to the Storage API
	current uint64 // Current block number being iterated over.
	eoc     bool   // End of chain.
}

func (di *diskIterator) Next() (database.BlockData, error) {
	if di.eoc {
		return database.BlockData{}, errors.New("end of chain")
	}

	di.current++
	blockData, err := di.storage.GetBlockByNumber(di.current)
	if errors.Is(err, fs.ErrNotExist) {
		di.eoc = true
	}

	return blockData, err
}

func (di *diskIterator) Done() bool {
	return di.eoc
}
