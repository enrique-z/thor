// Copyright (c) 2018 The VeChainThor developers

// Distributed under the GNU Lesser General Public License v3.0 software license, see the accompanying
// file LICENSE or <https://www.gnu.org/licenses/lgpl-3.0.html>

package chain_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/vechain/thor/block"
	"github.com/vechain/thor/chain"
	"github.com/vechain/thor/genesis"
	"github.com/vechain/thor/muxdb"
	"github.com/vechain/thor/state"
	"github.com/vechain/thor/tx"
)

func M(args ...interface{}) []interface{} {
	return args
}

func initRepo() *chain.Repository {
	db := muxdb.NewMem()
	g := genesis.NewDevnet()
	b0, _, _, _ := g.Build(state.NewStater(db))

	repo, err := chain.NewRepository(db, b0)
	if err != nil {
		panic(err)
	}
	return repo
}

var privateKey, _ = crypto.GenerateKey()

func newBlock(parent *block.Block, score uint64, txs ...*tx.Transaction) *block.Block {
	builder := new(block.Builder).
		ParentID(parent.Header().ID()).
		TotalScore(parent.Header().
			TotalScore() + score)

	for _, tx := range txs {
		builder.Transaction(tx)
	}

	b := builder.Build()
	sig, _ := crypto.Sign(b.Header().SigningHash().Bytes(), privateKey)
	return b.WithSignature(sig)
}

func TestRepository(t *testing.T) {
	repo := initRepo()

	assert.Equal(t, repo.GenesisBlock(), repo.BestBlock())
	assert.Equal(t, repo.GenesisBlock().Header().ID()[31], repo.ChainTag())

	tx1 := new(tx.Builder).Build()

	receipt1 := &tx.Receipt{}
	b1 := newBlock(repo.GenesisBlock(), 1, tx1)
	assert.Nil(t, repo.AddBlock(b1, tx.Receipts{receipt1}))

	// best block not set, so still 0
	assert.Equal(t, uint32(0), repo.BestBlock().Header().Number())

	repo.SetBestBlockID(b1.Header().ID())
	assert.Equal(t, b1, repo.BestBlock())

	h, _, err := repo.GetBlockHeader(b1.Header().ID())
	assert.Nil(t, err)
	assert.Equal(t, b1.Header(), h)

	assert.Equal(t, M(b1, nil), M(repo.GetBlock(b1.Header().ID())))

	assert.Equal(t, M(tx.Transactions{tx1}, nil), M(repo.GetBlockTransactions(b1.Header().ID())))
	assert.Equal(t, M(tx.Receipts{receipt1}, nil), M(repo.GetBlockReceipts(b1.Header().ID())))

}