/*
* Copyright (c) 2008-2023, Hazelcast, Inc. All Rights Reserved.
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

package codec

import (
	proto "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
)

const (
	MCGetClusterMetadataCodecRequestMessageType  = int32(0x200E00)
	MCGetClusterMetadataCodecResponseMessageType = int32(0x200E01)

	MCGetClusterMetadataCodecRequestInitialFrameSize = proto.PartitionIDOffset + proto.IntSizeInBytes

	MCGetClusterMetadataResponseCurrentStateOffset = proto.ResponseBackupAcksOffset + proto.ByteSizeInBytes
	MCGetClusterMetadataResponseClusterTimeOffset  = MCGetClusterMetadataResponseCurrentStateOffset + proto.ByteSizeInBytes
	MCGetClusterMetadataResponseClusterIdOffset    = MCGetClusterMetadataResponseClusterTimeOffset + proto.LongSizeInBytes
)

// Gets the current metadata of a cluster.

func EncodeMCGetClusterMetadataRequest() *proto.ClientMessage {
	clientMessage := proto.NewClientMessageForEncode()
	clientMessage.SetRetryable(true)

	initialFrame := proto.NewFrameWith(make([]byte, MCGetClusterMetadataCodecRequestInitialFrameSize), proto.UnfragmentedMessage)
	clientMessage.AddFrame(initialFrame)
	clientMessage.SetMessageType(MCGetClusterMetadataCodecRequestMessageType)
	clientMessage.SetPartitionId(-1)

	return clientMessage
}

func DecodeMCGetClusterMetadataResponse(clientMessage *proto.ClientMessage) (currentState byte, memberVersion string, jetVersion string, clusterTime int64, clusterId types.UUID, isClusterIdExists bool) {
	frameIterator := clientMessage.FrameIterator()
	initialFrame := frameIterator.Next()

	currentState = DecodeByte(initialFrame.Content, MCGetClusterMetadataResponseCurrentStateOffset)
	clusterTime = DecodeLong(initialFrame.Content, MCGetClusterMetadataResponseClusterTimeOffset)
	if len(initialFrame.Content) >= MCGetClusterMetadataResponseClusterIdOffset+proto.UUIDSizeInBytes {
		clusterId = DecodeUUID(initialFrame.Content, MCGetClusterMetadataResponseClusterIdOffset)
		isClusterIdExists = true
	} else {
		isClusterIdExists = false
	}
	memberVersion = DecodeString(frameIterator)
	jetVersion = DecodeNullableForString(frameIterator)
	return currentState, memberVersion, jetVersion, clusterTime, clusterId, isClusterIdExists
}
