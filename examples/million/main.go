// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/kelindar/column"
	"github.com/kelindar/column/fixtures"
)

func main() {
	amount, runs := 10000000, 20
	players := column.NewCollection(column.Options{
		Capacity: amount,
	})

	// insert the data first
	measure("insert", fmt.Sprintf("%v rows", amount), func() {
		createCollection(players, amount)
	}, 1)

	// snapshot the dataset
	measure("snapshot", fmt.Sprintf("%v rows", amount), func() {
		buffer := bytes.NewBuffer(nil)
		players.Snapshot(buffer)
	}, 10)

	// run a full scan
	measure("full scan", "age >= 30", func() {
		players.Query(func(txn *column.Txn) error {
			count := txn.WithFloat("age", func(v float64) bool {
				return v >= 30
			}).Count()
			fmt.Printf("-> result = %v\n", count)
			return nil
		})
	}, runs)

	// run a full scan
	measure("full scan", `class == "rogue"`, func() {
		players.Query(func(txn *column.Txn) error {
			count := txn.WithString("class", func(v string) bool {
				return v == "rogue"
			}).Count()
			fmt.Printf("-> result = %v\n", count)
			return nil
		})
	}, runs)

	// run a query over human mages
	measure("indexed query", "human mages", func() {
		players.Query(func(txn *column.Txn) error {
			fmt.Printf("-> result = %v\n", txn.With("human", "mage").Count())
			return nil
		})
	}, runs*1000)

	// run a query over human mages
	measure("indexed query", "human female mages", func() {
		players.Query(func(txn *column.Txn) error {
			fmt.Printf("-> result = %v\n", txn.With("human", "female", "mage").Count())
			return nil
		})
	}, runs*1000)

	// update everyone
	measure("update", "balance of everyone", func() {
		updates := 0
		players.Query(func(txn *column.Txn) error {
			balance := txn.Float64("balance")
			return txn.Range(func(idx uint32) {
				updates++
				balance.Set(1000.0)
			})
		})
		fmt.Printf("-> updated %v rows\n", updates)
	}, runs)

	// update age of mages
	measure("update", "age of mages", func() {
		updates := 0
		players.Query(func(txn *column.Txn) error {
			age := txn.Int("age")
			return txn.With("mage").Range(func(idx uint32) {
				updates++
				age.Set(99)
			})
		})
		fmt.Printf("-> updated %v rows\n", updates)
	}, runs)
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

	// index for humans
	out.CreateIndex("human", "race", func(r column.Reader) bool {
		return r.String() == "human"
	})

	// index for mages
	out.CreateIndex("mage", "class", func(r column.Reader) bool {
		return r.String() == "mage"
	})

	// index for males
	out.CreateIndex("male", "gender", func(r column.Reader) bool {
		return r.String() == "male"
	})

	// index for females
	out.CreateIndex("female", "gender", func(r column.Reader) bool {
		return r.String() == "female"
	})

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

// measure runs a function and measures it
func measure(action, name string, fn func(), iterations int) {
	defer func(start time.Time, stdout *os.File) {
		os.Stdout = stdout
		elapsed := time.Since(start) / time.Duration(iterations)
		fmt.Printf("-> %v took %v\n", action, elapsed.String())
	}(time.Now(), os.Stdout)

	fmt.Println()
	fmt.Printf("running %v of %v...\n", action, name)

	// Run a few times so the results are more stable
	null, _ := os.Open(os.DevNull)
	for i := 0; i < iterations; i++ {
		if i > 0 { // Silence subsequent runs
			os.Stdout = null
		}

		fn()
	}
}
