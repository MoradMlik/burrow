package state

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/genesis"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/dump"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
)

type MockDumpReader struct {
	accounts int
	storage  int
	names    int
	events   int
}

func (m *MockDumpReader) Next() (*dump.Dump, error) {
	// acccounts
	row := dump.Dump{Height: 102}

	if m.accounts > 0 {
		var addr crypto.Address
		binary.PutUint64BE(addr.Bytes(), uint64(m.accounts))

		row.Account = &acm.Account{
			Address: addr,
			Balance: 102,
		}

		if m.accounts%2 > 0 {
			row.Account.Code = make([]byte, rand.Int()%10000)
		} else {
			row.Account.PublicKey = crypto.PublicKey{}
		}
		m.accounts--
	} else if m.storage > 0 {
		var addr crypto.Address
		binary.PutUint64BE(addr.Bytes(), uint64(m.storage))
		storagelen := rand.Int() % 25

		row.AccountStorage = &dump.AccountStorage{
			Address: addr,
			Storage: make([]*dump.Storage, storagelen),
		}

		for i := 0; i < storagelen; i++ {
			row.AccountStorage.Storage[i] = &dump.Storage{}
		}

		m.storage--
	} else if m.names > 0 {
		row.Name = &names.Entry{
			Name:    fmt.Sprintf("name%d", m.names),
			Data:    fmt.Sprintf("data%x", m.names),
			Owner:   crypto.ZeroAddress,
			Expires: 1337,
		}
		m.names--
	} else if m.events > 0 {
		datalen := rand.Int() % 10
		data := make([]byte, datalen*32)
		topiclen := rand.Int() % 5
		topics := make([]binary.Word256, topiclen)
		row.EVMEvent = &dump.EVMEvent{
			ChainID: "MockyChain",
			Event: &exec.LogEvent{
				Address: crypto.ZeroAddress,
				Data:    data,
				Topics:  topics,
			},
		}
		m.events--
	} else {
		return nil, nil
	}

	return &row, nil
}

func BenchmarkLoadDump(b *testing.B) {
	for n := 0; n < b.N; n++ {
		mock := MockDumpReader{
			accounts: 629,
			storage:  620,
			names:    100,
			events:   9990,
		}
		_, err := MakeGenesisState(dbm.NewMemDB(), &genesis.GenesisDoc{}, &mock)
		require.NoError(b, err)
	}
}
