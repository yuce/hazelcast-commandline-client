/*
 * Copyright (c) 2008-2021, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mapcmd

import (
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/cobra"

	hzcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const MapPutExample = `  # Put key, value pair to map. The unit for ttl/max-idle is one of (ns,us,ms,s,m,h)
  map put --key-type string --key hello --value-type float32 --value 19.94 --name myMap --ttl 1300ms --max-idle 1400ms`

func NewPut(config *hazelcast.Config) *cobra.Command {
	var (
		mapName,
		mapKey,
		mapKeyType,
		mapValue,
		mapValueType,
		mapValueFile string
	)
	var (
		ttl,
		maxIdle time.Duration
		showType bool
	)
	cmd := &cobra.Command{
		Use:     "put [--name mapname | --key keyname | --value-type type | {--value-file file | --value value} | --ttl ttl | --max-idle max-idle]",
		Short:   "Put value to map",
		Example: MapPutExample,
		PreRunE: hzcerrors.RequiredFlagChecker,
		RunE: func(cmd *cobra.Command, args []string) error {
			ot, err := output.TypeStringFor(cmd)
			if err != nil {
				return err
			}
			key, err := internal.ConvertString(mapKey, mapKeyType)
			if err != nil {
				return hzcerrors.NewLoggableError(err, "Conversion error on key %s to type %s, %s", mapKey, mapKeyType, err)
			}
			var (
				ttlE bool
				//maxIdleE bool
			)
			if ttl.Seconds() != 0 {
				if err = validateTTL(ttl); err != nil {
					return hzcerrors.NewLoggableError(err, "ttl is invalid")
				}
				ttlE = true
			}
			/*
				if maxIdle.Seconds() != 0 {
					if err = isNegativeSecond(maxIdle); err != nil {
						return hzcerrors.NewLoggableError(err, "max-idle is invalid")
					}
					maxIdleE = true
				}
			*/
			var normalizedValue interface{}
			if normalizedValue, err = normalizeMapValue(mapValue, mapValueFile, mapValueType); err != nil {
				return err
			}
			ci, err := getClient(cmd.Context(), config)
			if err != nil {
				return err
			}
			keyData, err := ci.EncodeData(key)
			if err != nil {
				return err
			}
			valueData, err := ci.EncodeData(normalizedValue)
			if err != nil {
				return err
			}
			var req *hazelcast.ClientMessage
			switch {
			//case ttlE && maxIdleE:
			//	oldValue, err = m.PutWithTTLAndMaxIdle(cmd.Context(), key, normalizedValue, ttl, maxIdle)
			case ttlE:
				req = codec.EncodeMapPutRequest(mapName, keyData, valueData, 0, ttl.Milliseconds())
			//case maxIdleE:
			//	oldValue, err = m.PutWithMaxIdle(cmd.Context(), key, normalizedValue, maxIdle)
			default:
				req = codec.EncodeMapPutRequest(mapName, keyData, valueData, 0, -1)
			}
			resp, err := ci.InvokeOnKey(cmd.Context(), req, keyData, nil)
			if err != nil {
				var handled bool
				handled, err = isCloudIssue(err, config)
				if handled {
					return err
				}
				return hzcerrors.NewLoggableError(err, "Cannot put given entry to the map %s", mapName)
			}
			raw := codec.DecodeMapPutResponse(resp)
			valueType := raw.Type()
			oldValue, err := ci.DecodeData(raw)
			if err != nil {
				oldValue = serialization.NondecodedType(serialization.TypeToString(valueType))
			}
			return printSingleValue(oldValue, valueType, showType, ot)
		},
	}
	decorateCommandWithMapNameFlags(cmd, &mapName, true, "specify the map name")
	decorateCommandWithMapKeyFlags(cmd, &mapKey, true, "key of the entry")
	decorateCommandWithMapKeyTypeFlags(cmd, &mapKeyType, false)
	decorateCommandWithValueFlags(cmd, &mapValue, &mapValueFile)
	decorateCommandWithMapValueTypeFlags(cmd, &mapValueType, false)
	decorateCommandWithTTL(cmd, &ttl, false, "ttl value of the entry")
	decorateCommandWithMaxIdle(cmd, &maxIdle, false, "max-idle value of the entry")
	decorateCommandWithShowTypesFlag(cmd, &showType)
	return cmd
}
