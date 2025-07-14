package minibank

import (
	"emulator/utils"
	"emulator/utils/store"
	"fmt"
	"testing"
)

func TestMemDB(t *testing.T) {
	db := store.NewPrefixStore("test", ".")
	memDB := NewMemoryDB(db)

	memDB.Set("a", 12)
	memDB.Set("b", 32)
	memDB.CommitWrite(true)
	memDB.Write()

	bz, err := db.Get([]byte("a"))
	fmt.Println(bz)
	fmt.Println(utils.Uint32ToBytes(12))
	if err != nil {
		panic(err)
	}
	if r := utils.BytesToUint32(bz); r != 12 {
		panic("1")
	}

	iter, err := db.Iterator(nil, []byte("z"))
	if err != nil {
		panic(err)
	}
	fmt.Println(iter.Valid())
	for iter.Valid() {
		fmt.Println(iter.Key(), iter.Value())
		iter.Next()
	}
}
