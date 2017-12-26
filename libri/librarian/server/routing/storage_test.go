package routing

import (
	"errors"
	"math/rand"
	"testing"

	"container/heap"

	"github.com/drausin/libri/libri/common/db"
	"github.com/drausin/libri/libri/common/id"
	cstorage "github.com/drausin/libri/libri/common/storage"
	"github.com/drausin/libri/libri/librarian/server/peer"
	sstorage "github.com/drausin/libri/libri/librarian/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestFromStored(t *testing.T) {
	srt := newTestStoredTable(rand.New(rand.NewSource(0)), 128)
	rt := fromStored(srt, NewDefaultParameters())
	assertRoutingTablesEqual(t, rt, srt)
}

func TestToRoutingTable(t *testing.T) {
	rt, _, _ := NewTestWithPeers(rand.New(rand.NewSource(0)), 128)
	srt := toStored(rt)
	assertRoutingTablesEqual(t, rt, srt)
}

func TestRoutingTable_SaveLoad(t *testing.T) {
	rt1, _, _ := NewTestWithPeers(rand.New(rand.NewSource(0)), 8)
	kvdb, cleanup, err := db.NewTempDirRocksDB()
	defer cleanup()
	defer kvdb.Close()
	assert.Nil(t, err)
	ssl := cstorage.NewServerSL(kvdb)

	err = rt1.Save(ssl)
	assert.Nil(t, err)

	rt2, err := Load(ssl, NewDefaultParameters())
	assert.Nil(t, err)

	// check that routing tables are the same
	assert.Equal(t, rt1.SelfID().Bytes(), rt2.SelfID().Bytes())
	assert.Equal(t, len(rt1.(*table).peers), len(rt2.(*table).peers))
	assert.Equal(t, rt1.(*table).NumPeers(), rt2.(*table).NumPeers())
	assert.Equal(t, rt1.(*table).peers, rt2.(*table).peers)

	// descend into the protected table fields to check equality
	for bi, bucket1 := range rt1.(*table).buckets {
		bucket2 := rt2.(*table).buckets[bi]
		assert.Equal(t, bucket1.depth, bucket2.depth)
		assert.Equal(t, bucket1.lowerBound, bucket2.lowerBound)
		assert.Equal(t, bucket1.upperBound, bucket2.upperBound)
		assert.Equal(t, bucket1.containsSelf, bucket2.containsSelf)
		assert.Equal(t, bucket1.Len(), bucket2.Len())

		// the ActivePeers array may have some small differences in
		// ordering (and thus the Position map will also reflect those
		// differences), but here we check that the same peers are popped
		// off at the same time, which is really want we care about
		for bucket2.Len() > 0 {
			peer1, peer2 := heap.Pop(bucket1), heap.Pop(bucket2)
			assert.Equal(t, peer1, peer2)
		}
	}
}

func newTestStoredTable(rng *rand.Rand, n int) *sstorage.RoutingTable {
	rt := &sstorage.RoutingTable{
		SelfId: id.NewPseudoRandom(rng).Bytes(),
		Peers:  make([]*sstorage.Peer, n),
	}
	for i := 0; i < n; i++ {
		rt.Peers[i] = peer.NewTestStoredPeer(rng, i)
	}
	return rt
}

func assertRoutingTablesEqual(t *testing.T, rt Table, srt *sstorage.RoutingTable) {
	assert.Equal(t, srt.SelfId, rt.SelfID().Bytes())
	for _, sp := range srt.Peers {
		spIDStr := id.FromBytes(sp.Id).String()
		if toPeer, exists := rt.(*table).peers[spIDStr]; exists {
			peer.AssertPeersEqual(t, sp, toPeer)
		}
	}
}

func TestLoad_err(t *testing.T) {

	// simulates missing/not stored table
	rt1, err := Load(&cstorage.TestSLD{}, NewDefaultParameters())
	assert.Nil(t, rt1)
	assert.Nil(t, err)

	// simulates loading error
	rt2, err := Load(
		&cstorage.TestSLD{
			Bytes:   []byte("some random bytes"),
			LoadErr: errors.New("some random error"),
		},
		NewDefaultParameters(),
	)
	assert.Nil(t, rt2)
	assert.NotNil(t, err)

	// simulates bad stored table
	rt3, err := Load(
		&cstorage.TestSLD{
			Bytes:   []byte("the wrong bytes"),
			LoadErr: nil,
		},
		NewDefaultParameters(),
	)
	assert.Nil(t, rt3)
	assert.NotNil(t, err)
}
