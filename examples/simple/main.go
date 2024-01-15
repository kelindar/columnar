package main

import (
	"encoding/json"
	"os"

	"github.com/kelindar/column"
)

func main() {

	// Create a new columnar collection
	players := column.NewCollection()
	players.CreateColumn("serial", column.ForKey())
	players.CreateColumn("name", column.ForString())
	players.CreateColumn("active", column.ForBool())
	players.CreateColumn("class", column.ForEnum())
	players.CreateColumn("race", column.ForEnum())
	players.CreateColumn("age", column.ForInt())
	players.CreateColumn("hp", column.ForInt())
	players.CreateColumn("mp", column.ForInt())
	players.CreateColumn("balance", column.ForFloat64())
	players.CreateColumn("gender", column.ForEnum())
	players.CreateColumn("guild", column.ForEnum())
	players.CreateColumn("location", column.ForRecord(func() *Location {
		return new(Location)
	}))

	// index on humans
	players.CreateIndex("human", "race", func(r column.Reader) bool {
		return r.String() == "human"
	})

	// index for mages
	players.CreateIndex("mage", "class", func(r column.Reader) bool {
		return r.String() == "mage"
	})

	// index for old
	players.CreateIndex("old", "age", func(r column.Reader) bool {
		return r.Int() >= 30
	})

	// Load the items into the collection
	loaded := loadFixture("players.json")
	players.Query(func(txn *column.Txn) error {
		for _, v := range loaded {
			txn.InsertKey(v.Serial, func(r column.Row) error {
				r.SetKey(v.Serial)
				r.SetString("name", v.Name)
				r.SetBool("active", v.Active)
				r.SetEnum("class", v.Class)
				r.SetEnum("race", v.Race)
				r.SetInt("age", v.Age)
				r.SetInt("hp", v.Hp)
				r.SetInt("mp", v.Mp)
				r.SetFloat64("balance", v.Balance)
				r.SetEnum("gender", v.Gender)
				r.SetEnum("guild", v.Guild)
				r.SetRecord("location", &v.Location)
				return nil
			})
		}
		return nil
	})

	// Run an indexed query
	players.Query(func(txn *column.Txn) error {
		name := txn.String("name")
		return txn.With("human", "mage", "old").Range(func(idx uint32) {
			value, _ := name.Get()
			println("old mage, human:", value)
		})
	})
}

// loadFixture loads a fixture by its name
func loadFixture(name string) []Player {
	b, err := os.ReadFile("../../fixtures/" + name)
	if err != nil {
		panic(err)
	}

	var data []Player
	if err := json.Unmarshal(b, &data); err != nil {
		panic(err)
	}

	return data
}

// --------------------------- Player ----------------------------

type Player struct {
	Serial   string   `json:"serial"`
	Name     string   `json:"name"`
	Active   bool     `json:"active"`
	Class    string   `json:"class"`
	Race     string   `json:"race"`
	Age      int      `json:"age"`
	Hp       int      `json:"hp"`
	Mp       int      `json:"mp"`
	Balance  float64  `json:"balance"`
	Gender   string   `json:"gender"`
	Guild    string   `json:"guild"`
	Location Location `json:"location"`
}

type Location struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (l Location) MarshalBinary() ([]byte, error) {
	return json.Marshal(l)
}

func (l *Location) UnmarshalBinary(b []byte) error {
	return json.Unmarshal(b, l)
}
