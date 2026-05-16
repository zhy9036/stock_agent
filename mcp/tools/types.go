package tools

import (
	"reflect"
	"strings"
)

type PriceInput struct {
	Symbol string `json:"symbol" jsonschema:"stock ticker symbol"`
}

type TradeInput struct {
	Symbol   string `json:"symbol"`
	Quantity int    `json:"quantity"`
}

func StructSchema[T any]() string {
	var t T
	typ := reflect.TypeOf(t)

	var sb strings.Builder
	sb.WriteString("INPUT SCHEMA:\n{\n")

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		jsonName := f.Tag.Get("json")
		if jsonName == "" {
			jsonName = f.Name
		}

		sb.WriteString(
			"  \"" + jsonName + "\": " + f.Type.Name() + ",\n",
		)
	}

	sb.WriteString("}\n")

	return sb.String()
}
