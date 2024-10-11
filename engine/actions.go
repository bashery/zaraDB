package engine

import (
	"fmt"

	"github.com/tidwall/gjson"
)

var aggr Aggregate

// ok
func HandleQueries(query string) string {

	parsedQuery := gjson.Parse(query)

	switch parsedQuery.Get("action").Str { // action

	// aggregate actions
	case "aggregate":
		return aggr.aggrigate(parsedQuery)

	case "count":
		// count(parsedQuery)
		return "count"

	case "sum":
		return "not implemented yet"

	case "avg":
		return "not implemented yet"

	case "min":
		return "not implemented yet"

	case "max":
		return "not implemented yet"

	// database actions
	case "findOne":
		return db.findOne(parsedQuery)

	case "findMany":
		return db.findMany(parsedQuery)

	case "findById":
		return db.findById(parsedQuery)

	case "insert":
		return db.insertOne(parsedQuery)

	case "insertMany":
		return db.insertMany(parsedQuery)

	// update
	case "updateById":
		return db.updateById(parsedQuery)

	case "updateOne":
		return db.updateOne(parsedQuery)

	case "updateMany":
		return db.updateMany(parsedQuery)

	case "deleteById":
		return db.deleteById(parsedQuery)

	case "deleteOne":
		return db.deleteOne(parsedQuery)

	case "deleteMany":
		return db.deleteMany(parsedQuery)

	case "transaction":
		return transaction(parsedQuery)

	// manage database
	case "create_collection":
		return createCollection(parsedQuery.Get("collection"))

	case "delete_collection":
		return deleteCollection(parsedQuery.Get("collection"))

	case "getCollections":
		//return showCollections(db.path)
		return getCollections()

	default:
		return fmt.Errorf("unknown '%s' cation", parsedQuery.Get("action").Str).Error()
	}
}
