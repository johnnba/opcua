// Copyright 2018-2019 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package services

import (
	"testing"
	"time"

	"github.com/gopcua/opcua/ua"
	"github.com/gopcua/opcua/utils/codectest"
)

func TestReadResponse(t *testing.T) {
	cases := []codectest.Case{
		{
			Name: "read response with single float value",
			Struct: NewReadResponse(
				NewResponseHeader(
					time.Date(2018, time.August, 10, 23, 0, 0, 0, time.UTC),
					1, 0, ua.NewNullDiagnosticInfo(), []string{}, NewNullAdditionalHeader(),
				),
				[]*ua.DiagnosticInfo{
					ua.NewNullDiagnosticInfo(),
				},
				ua.NewDataValue(
					true, false, false, false, false, false,
					ua.MustVariant(float32(2.5001559257507324)),
					0, time.Time{}, 0, time.Time{}, 0,
				),
			),
			Bytes: []byte{
				// Timestamp
				0x00, 0x98, 0x67, 0xdd, 0xfd, 0x30, 0xd4, 0x01,
				// RequestHandle
				0x01, 0x00, 0x00, 0x00,
				// ServiceResult
				0x00, 0x00, 0x00, 0x00,
				// ServiceDiagnostics
				0x00,
				// StringTable
				0x00, 0x00, 0x00, 0x00,
				// AdditionalHeader
				0x00, 0x00, 0x00,
				// Results
				// ArraySize
				0x01, 0x00, 0x00, 0x00,
				// EncodingMask
				0x01,
				// Value
				0x0a, 0x8e, 0x02, 0x20, 0x40, 0x01, 0x00, 0x00, 0x00, 0x00,
			},
		},
	}
	codectest.Run(t, cases)
}
