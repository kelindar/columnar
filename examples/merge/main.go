// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package main

import (
	"encoding/json"
	"fmt"

	"github.com/kelindar/column"
)

// Movement represents a movement with a position and velocity
type Movement struct {
	Position [2]float64 `json:"position,omitempty"`
	Velocity [2]float64 `json:"velocity,omitempty"`
}

func main() {

	// A merging function that accepts a velocity vector and updates
	// the movement structure accordingly.
	mergeVectors := func(value, delta string) string {
		movement, ok := parseMovement(value)
		if !ok {
			movement = Movement{
				Position: [2]float64{0, 0},
			}
		}

		// Parse the incoming delta value
		velocity, ok := parseVector(delta)
		if !ok {
			return value
		}

		// Update the current velocity and recalculate the position
		movement.Velocity = velocity
		movement.Position[0] += velocity[0] // Update X
		movement.Position[1] += velocity[1] // Update Y

		// Encode the movement as JSON and return the updated value
		return encode(movement)
	}

	// Create a column with a specified merge function
	db := column.NewCollection()
	db.CreateColumn("location", column.ForString(
		column.WithMerge(mergeVectors), // use our merging function
	))

	// Insert an empty row
	id, _ := db.Insert(func(r column.Row) error {
		r.SetString("location", "{}")
		return nil
	})

	// Update several times
	for i := 0; i < 20; i++ {

		// Move the location by applying a same velocity vector
		db.Query(func(txn *column.Txn) error {
			location := txn.String("location")
			return txn.QueryAt(id, func(r column.Row) error {
				location.Merge(encode([2]float64{1, 2}))
				return nil
			})
		})

		// Print out current location
		db.Query(func(txn *column.Txn) error {
			location := txn.String("location")
			return txn.QueryAt(id, func(r column.Row) error {
				value, _ := location.Get()
				fmt.Printf("%.2d: %v \n", i, value)
				return nil
			})
		})
	}

}

// parseMovement parses a value string into a Movement struct
func parseMovement(value string) (out Movement, ok bool) {
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return Movement{}, false
	}
	return out, true
}

// parseVector parses a value string into 2 dimensional array
func parseVector(value string) (out [2]float64, ok bool) {
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return [2]float64{}, false
	}
	return out, true
}

// encodes encodes the value as JSON
func encode(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
