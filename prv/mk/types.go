package mk

import "github.com/hazelcast/hazelcast-commandline-client/prv"

func ValueFromString(value, valueType string) (any, error) {
	return prv.ConvertString(value, valueType)
}
