// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/kelindar/async"
	"github.com/kelindar/column"
	"github.com/kelindar/column/fixtures"
	"github.com/kelindar/xxrand"
)

var (
	classes = []string{"fighter", "mage", "rogue"}
	races   = []string{"human", "elf", "dwarf", "orc"}
)

func main() {
	amount := 1000000
	players := column.NewCollection(column.Options{
		Capacity: amount,
	})
	createCollection(players, amount)

	// This runs point query benchmarks
	runBenchmark("Point Reads/Writes", func(writeTxn bool) (reads int, writes int) {

		// To avoid task granuarity problem, load up a bit more work on each
		// of the goroutines, a few hundred reads should be enough to amortize
		// the cost of scheduling goroutines, so we can actually test our code.
		for i := 0; i < 1000; i++ {
			offset := xxrand.Uint32n(uint32(amount - 1))
			if writeTxn {
				players.QueryAt(offset, func(r column.Row) error {
					r.SetFloat64("balance", 0)
					return nil
				})
				writes++
			} else {
				players.QueryAt(offset, func(r column.Row) error {
					_, _ = r.Float64("balance")
					return nil
				})
				reads++
			}
		}
		return
	})
}

// runBenchmark runs a benchmark
func runBenchmark(name string, fn func(bool) (int, int)) {
	fmt.Printf("Benchmarking %v ...\n", name)
	fmt.Printf("%7v\t%6v\t%17v\t%13v\n", "WORK", "PROCS", "READ RATE", "WRITE RATE")
	for _, workload := range []int{0, 10, 50, 90, 100} {

		// Iterate over various concurrency levels
		for _, n := range []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512} {
			work := make(chan async.Task, n)
			pool := async.Consume(context.Background(), n, work)

			var reads, writes int64
			var wg sync.WaitGroup
			start := time.Now()
			for time.Since(start) < time.Second {
				wg.Add(1)
				work <- async.NewTask(func(ctx context.Context) (interface{}, error) {
					defer wg.Done()

					r, w := fn(xxrand.Intn(100) < workload)
					atomic.AddInt64(&reads, int64(r))
					atomic.AddInt64(&writes, int64(w))
					return nil, nil
				})
			}

			wg.Wait()
			pool.Cancel()

			elapsed := time.Since(start)
			fmt.Printf("%v%%-%v%%\t%6v\t%17v\t%13v\n", 100-workload, workload, n,
				humanize.Comma(int64(float64(reads)/elapsed.Seconds()))+" txn/s",
				humanize.Comma(int64(float64(writes)/elapsed.Seconds()))+" txn/s",
			)
		}
	}
}

// createCollection loads a collection of players
func createCollection(out *column.Collection, amount int) *column.Collection {
	out.CreateColumn("serial", column.ForEnum())
	out.CreateColumn("name", column.ForEnum())
	out.CreateColumn("active", column.ForBool())
	out.CreateColumn("class", column.ForEnum())
	out.CreateColumn("race", column.ForEnum())
	out.CreateColumn("age", column.ForInt())
	out.CreateColumn("hp", column.ForInt())
	out.CreateColumn("mp", column.ForInt())
	out.CreateColumn("balance", column.ForFloat64())
	out.CreateColumn("gender", column.ForEnum())
	out.CreateColumn("guild", column.ForEnum())

	for _, v := range classes {
		class := v
		out.CreateIndex(class, "class", func(r column.Reader) bool {
			return r.String() == class
		})
	}

	for _, v := range races {
		race := v
		out.CreateIndex(race, "race", func(r column.Reader) bool {
			return r.String() == race
		})
	}
	// Load the data in
	data := fixtures.Players()
	for i := 0; i < amount/len(data); i++ {
		insertPlayers(out, data)
	}

	return out
}

// insertPlayers inserts players
func insertPlayers(dst *column.Collection, data []fixtures.Player) error {
	return dst.Query(func(txn *column.Txn) error {
		for _, v := range data {
			txn.Insert(func(r column.Row) error {
				r.SetEnum("serial", v.Serial)
				r.SetEnum("name", v.Name)
				r.SetBool("active", v.Active)
				r.SetEnum("class", v.Class)
				r.SetEnum("race", v.Race)
				r.SetInt("age", v.Age)
				r.SetInt("hp", v.Hp)
				r.SetInt("mp", v.Mp)
				r.SetFloat64("balance", v.Balance)
				r.SetEnum("gender", v.Gender)
				r.SetEnum("guild", v.Guild)
				return nil
			})
		}
		return nil
	})
}
