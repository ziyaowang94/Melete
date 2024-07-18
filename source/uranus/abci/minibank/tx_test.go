package minibank

import (
	"emulator/uranus/types"
	"testing"
)

func BenchmarkNewTransferTxMustLen(b *testing.B) {
	froms := []string{"from1", "from2"}
	fromMoney := []uint32{1, 1}
	tos := []string{"to1", "to2"}
	toMoney := []uint32{1, 1}
	b.Run("mustLen=512", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tx := NewTransferTxMustLen(froms, fromMoney, tos, toMoney, 512)
			if err := ValidateTransferTx(tx); err != nil {
				panic(err)
			}
			types.MustProtoBytes(tx)
		}
	})
	b.Run("mustLen=256", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tx := NewTransferTxMustLen(froms, fromMoney, tos, toMoney, 256)
			if err := ValidateTransferTx(tx); err != nil {
				panic(err)
			}
			types.MustProtoBytes(tx)
		}
	})
	b.Run("mustLen=128", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tx := NewTransferTxMustLen(froms, fromMoney, tos, toMoney, 128)
			if err := ValidateTransferTx(tx); err != nil {
				panic(err)
			}
			types.MustProtoBytes(tx)
		}
	})

}

func BenchmarkNewTransferTxSize(b *testing.B) {
	froms := []string{"from144444444444231", "from122"}
	fromMoney := []uint32{100, 100}
	tos := []string{"to5551", "to5552"}
	toMoney := []uint32{100, 155}
	b.Run("mustLen=512", func(b *testing.B) {
		b.ReportAllocs()
		tx := NewTransferTxMustLen(froms, fromMoney, tos, toMoney, 509)
		size := TransferTxSize(tx)
		b.ReportMetric(float64(size), "bytes")
	})
	b.Run("mustLen=256", func(b *testing.B) {
		b.ReportAllocs()
		tx := NewTransferTxMustLen(froms, fromMoney, tos, toMoney, 253)
		size := TransferTxSize(tx)
		b.ReportMetric(float64(size), "bytes")
	})
	b.Run("mustLen=128", func(b *testing.B) {
		b.ReportAllocs()
		tx := NewTransferTxMustLen(froms, fromMoney, tos, toMoney, 126)
		size := TransferTxSize(tx)
		b.ReportMetric(float64(size), "bytes")
	})

}
