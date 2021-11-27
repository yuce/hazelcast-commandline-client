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

package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hazelcast/hazelcast-commandline-client/internal"
)

var pncGetCmd = &cobra.Command{
	Use:   "get [--name pncname]",
	Short: "get from PN Counter",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		config, err := internal.MakeConfig(cmd)
		if err != nil { //TODO error look like unhandled although it is handled in MakeConfig.Find a better approach
			return
		}
		p, err := getPNCounter(config, pncName)
		if err != nil {
			return
		}
		value, err := p.Get(ctx)
		if err != nil {
			fmt.Printf("Error: Cannot get value of PN Counter %s\n", pncName)
			return
		}
		fmt.Println(value)
	},
}

func init() {
	pncGetCmd.Flags().StringVarP(&pncName, "name", "m", "", "specify the PN Counter name")
	pncGetCmd.MarkFlagRequired("name")
}
