package dtree

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	pbc "github.com/davedotdev/proto/eddie/proto_common"
	pbn "github.com/davedotdev/proto/eddie/proto_node"
)

// DKVIDs returns KV pairs for the node.
// MUST BE INITIALISED.
var DKVIDs KVIDs

// string = listname
// int = jsonid
var sliceOrderMap map[string]map[int]interface{}

// TreeNode is the type for building trees out of our JSON
type TreeNode struct {
	Node     pbn.NODE
	IsSlice  bool
	Branches []*TreeNode
	KVIDs    KVIDs
}

// KVIDs adds in properties to our TreeNode
type KVIDs interface {
	Get() []*pbc.KV
}

// Branch adds a new branch to the current TN
func (tn *TreeNode) Branch(NTR *TreeNode) {
	// De-dup (just in case)
	for _, v := range tn.Branches {
		if reflect.DeepEqual(v, NTR) {
			return
		}
	}
	tn.Branches = append(tn.Branches, NTR)
}

// TODO: Make this work (requires passing a string around as a ptr)
func (tn *TreeNode) String() string {
	returnStr := ""
	UnFoldCount = 0
	UnFoldTN(tn, "", &returnStr)
	return returnStr
}

// JSONTN is a map ready for JSONification
type JSONTN map[string]interface{}

// CreateJSONStructure is a func that spits out a JSON node
func (tn *TreeNode) CreateJSONStructure() *JSONTN {
	ROOT := make(JSONTN)

	UnFoldCount = 0
	CreateJSONUnFoldTN(tn, "", ROOT)
	return &ROOT
}

// CreateJSONUnFoldTN unfolds the tree
func CreateJSONUnFoldTN(tn *TreeNode, father string, JTN JSONTN) {
	// This is the tn top layer
	// If we're a slice, we need to do something else

	if tn.IsSlice {
		// Our JTN value for the key is slice with a specific type.
		// 1.	We need to check how long the KV pairs are for node properties.
		//		If 1, then this is a slice of a single scalar type
		//		If >1 then this is a slice of a set.
		IsSlice := true

		// Nodes still have their jsonid values, we need to figure out if this is append or prepend
		tmptn := []*pbc.KV{}
		jsonid := 0
		var err error
		for _, v := range tn.Node.Kv {
			if v.Key != "jsonid" {
				tmptn = append(tmptn, v)
			} else {
				jsonid, err = strconv.Atoi(v.Value)
				if err != nil {
					fmt.Print(err)
				}
			}
		}
		// At this point, we have our jsonid and have created a new Kv list
		// for testing slices of size 1

		// if len(tn.Node.Kv) > 1 {
		if len(tmptn) > 1 {
			IsSlice = false
		}

		// Now we have the jsonid, remove it
		tn.Node.Kv = tmptn

		// This is for a slice of a single type
		if IsSlice == true {
			// Let's add the KV to the sliceOrderMap
			if sliceOrderMap[tn.Node.Kv[0].Key] == nil {
				sliceOrderMap[tn.Node.Kv[0].Key] = make(map[int]interface{})
			}
			sliceOrderMap[tn.Node.Kv[0].Key][jsonid] = tn.Node.Kv[0].Value
			fmt.Printf("Obtained JSONID(%d) for value: %s\n", jsonid, tn.Node.Kv[0])
			fmt.Printf("sliceOrderMap: %+v\n", sliceOrderMap)
			// Check if our value type has been set to the scalar kind
			typeOfValue := tn.Node.Kv[0].Vtype
			typeOfEntry := reflect.TypeOf(JTN[tn.Node.Type])
			fmt.Printf("Node: Slice of scalars Kv -> %s\n", tn.Node.Kv)
			fmt.Println("Node: I am a scalar slice of type: ", typeOfValue)

			if typeOfEntry == nil {
				fmt.Println("Node: First pass, setting type of slice")
				switch typeOfValue {
				case pbc.T_BOOL:
					JTN[tn.Node.Type] = []bool{}
				case pbc.T_FLOAT64:
					JTN[tn.Node.Type] = []float64{}
				case pbc.T_INT64:
					JTN[tn.Node.Type] = []int64{}
				case pbc.T_STRING:
					JTN[tn.Node.Type] = []string{}
				}
			}
			typeOfEntry = reflect.TypeOf(JTN[tn.Node.Type])
			fmt.Println("JTN entry is of type: ", typeOfEntry)

			// Next we have to do an append based on type
			switch typeOfValue {
			case pbc.T_BOOL:
				boolValue, err := strconv.ParseBool(tn.Node.Kv[0].Value)
				if err != nil {
					fmt.Println("ERR: Couldn't parse bool")
				}
				// Let's copy the array and create a new one based on order
				JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]bool), boolValue)

				// We need to sort the map keys to an order
				ids := make([]int, 0, len(sliceOrderMap[tn.Node.Kv[0].Key]))
				for id := range sliceOrderMap[tn.Node.Kv[0].Key] {
					ids = append(ids, id)
				}
				sort.Ints(ids)
				// Now we have a sorted slice of jsonids
				JTN[tn.Node.Type] = []bool{}

				for _, v := range ids {
					JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]bool), sliceOrderMap[tn.Node.Kv[0].Key][v].(bool))
				}

			case pbc.T_FLOAT64:
				floatValue, err := strconv.ParseFloat(tn.Node.Kv[0].Value, 64)
				if err != nil {
					fmt.Println("ERR: Couldn't parse float")
				}
				JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]float64), floatValue)

				// We need to sort the map keys to an order
				ids := make([]int, 0, len(sliceOrderMap[tn.Node.Kv[0].Key]))
				for id := range sliceOrderMap[tn.Node.Kv[0].Key] {
					ids = append(ids, id)
				}
				sort.Ints(ids)
				// Now we have a sorted slice of jsonids
				JTN[tn.Node.Type] = []float64{}

				for _, v := range ids {
					JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]float64), sliceOrderMap[tn.Node.Kv[0].Key][v].(float64))
				}

			case pbc.T_INT64:
				intValue, err := strconv.ParseInt(tn.Node.Kv[0].Value, 10, 64)
				if err != nil {
					fmt.Println("ERR: Couldn't parse float")
				}
				JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]int64), intValue)

				// We need to sort the map keys to an order
				ids := make([]int, 0, len(sliceOrderMap[tn.Node.Kv[0].Key]))
				for id := range sliceOrderMap[tn.Node.Kv[0].Key] {
					ids = append(ids, id)
				}
				sort.Ints(ids)
				// Now we have a sorted slice of jsonids
				JTN[tn.Node.Type] = []int64{}

				for _, v := range ids {
					JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]int64), sliceOrderMap[tn.Node.Kv[0].Key][v].(int64))
				}

			case pbc.T_STRING:
				JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]string), tn.Node.Kv[0].Value)

				// We need to sort the map keys to an order
				ids := make([]int, 0, len(sliceOrderMap[tn.Node.Kv[0].Key]))
				for id := range sliceOrderMap[tn.Node.Kv[0].Key] {
					ids = append(ids, id)
				}
				sort.Ints(ids)
				// Now we have a sorted slice of jsonids
				JTN[tn.Node.Type] = []string{}

				for _, v := range ids {
					JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]string), sliceOrderMap[tn.Node.Kv[0].Key][v].(string))
				}
			}

		} else {
			fmt.Println("SLICE SET NAME: ", tn.Node.Type)
			if sliceOrderMap[tn.Node.Type] == nil {
				sliceOrderMap[tn.Node.Type] = make(map[int]interface{})
			}

			fmt.Printf("Obtained JSONID(%d) for value: %s\n", jsonid, tn.Node.Kv)
			fmt.Printf("sliceOrderMap: %+v\n", sliceOrderMap)
			// This is for a slice of sets
			fmt.Printf("Node: Slice of sets Kv -> %s\n", tn.Node.Kv)
			JTN[tn.Node.Type] = []interface{}{}
			// Next create a type for this set[x]

			JTNProps := make(JSONTN)

			for _, v1 := range tn.Node.Kv {
				switch v1.Vtype {
				case pbc.T_BOOL:
					Value, err := strconv.ParseBool(v1.Value)
					if err != nil {
						fmt.Println("ERR: Couldn't parse bool")
					}
					JTNProps[v1.Key] = Value
				case pbc.T_FLOAT64:
					Value, err := strconv.ParseFloat(v1.Value, 64)
					if err != nil {
						fmt.Println("ERR: Couldn't parse float")
					}
					JTNProps[v1.Key] = Value

				case pbc.T_INT64:
					Value, err := strconv.ParseInt(v1.Value, 10, 64)
					if err != nil {
						fmt.Println("ERR: Couldn't parse float")
					}
					JTNProps[v1.Key] = Value
				case pbc.T_STRING:
					JTNProps[v1.Key] = v1.Value
				}
			}

			sliceOrderMap[tn.Node.Type][jsonid] = JTNProps

			// We need to sort the map keys to an order
			ids := make([]int, 0, len(sliceOrderMap[tn.Node.Type]))
			for id := range sliceOrderMap[tn.Node.Type] {
				ids = append(ids, id)
			}
			sort.Ints(ids)
			// Now we have a sorted slice of jsonids

			for _, v := range ids {
				JTN[tn.Node.Type] = append(JTN[tn.Node.Type].([]interface{}), sliceOrderMap[tn.Node.Type][v].(interface{}))
			}
		}

	} else {
		// Unwind the properties first
		fmt.Println("Normal Node (non-Slice)")
		fmt.Printf("Node:Kv -> %s\n", tn.Node.Kv)
		JTNProps := make(JSONTN)
		if len(tn.Node.Kv) > 0 {
			for _, v1 := range tn.Node.Kv {
				if v1.Key == "jsonid" {
					continue
				}
				switch v1.Vtype {
				case pbc.T_BOOL:
					Value, err := strconv.ParseBool(v1.Value)
					if err != nil {
						fmt.Println("ERR: Couldn't parse bool")
					}
					JTNProps[v1.Key] = Value
				case pbc.T_FLOAT64:
					Value, err := strconv.ParseFloat(v1.Value, 64)
					if err != nil {
						fmt.Println("ERR: Couldn't parse float")
					}
					JTNProps[v1.Key] = Value

				case pbc.T_INT64:
					Value, err := strconv.ParseInt(v1.Value, 10, 64)
					if err != nil {
						fmt.Println("ERR: Couldn't parse float")
					}
					JTNProps[v1.Key] = Value
				case pbc.T_STRING:
					JTNProps[v1.Key] = v1.Value
				}
			}

			// Add this node's type to JTN with properties
			JTN[tn.Node.Type] = JTNProps
		}

		fmt.Printf("Node:Type -> %s\n", tn.Node.Type)
		fmt.Printf("Node:Father -> %s\n", father)
		// Now let's recurse through each branch
		// It's also worth stating, empty but nested nodes are here
		for _, v := range tn.Branches {
			fmt.Println("Branch Name: ", v.Node.Type)
			fmt.Println()

			// create new Node
			JTNNew := make(JSONTN)

			if JTN[tn.Node.Type] == nil {
				// Case for empty nodes
				JTN[tn.Node.Type] = JTNNew
				CreateJSONUnFoldTN(v, tn.Node.Type, JTNNew)
			} else {
				// Case for non-empty nodes
				CreateJSONUnFoldTN(v, tn.Node.Type, JTNProps)
			}
		}
	}
}
