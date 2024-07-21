// Copyright 2024 The Ebitengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package convert_test

import (
	"bytes"
	"fmt"
	"io" // TODO: Use math/rand/v2 when the minimum supported version becomes Go 1.22.
	"math/rand"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/audio/internal/convert"
)

func TestStereoI16(t *testing.T) {
	testCases := []struct {
		Name string
		In   []int16
	}{
		{
			Name: "nil",
			In:   nil,
		},
		{
			Name: "-1, 0, 1, 0",
			In:   []int16{-1, 0, 1, 0},
		},
		{
			Name: "8 0s",
			In:   []int16{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			Name: "random 256 values",
			In:   randInt16s(256),
		},
		{
			Name: "random 65536 values",
			In:   randInt16s(65536),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			for _, mono := range []bool{false, true} {
				mono := mono
				t.Run(fmt.Sprintf("mono=%t", mono), func(t *testing.T) {
					var inBytes, outBytes []byte
					for _, v := range tc.In {
						inBytes = append(inBytes, byte(v), byte(v>>8))
						if mono {
							// As the source is mono, the output should be stereo.
							outBytes = append(outBytes, byte(v), byte(v>>8), byte(v), byte(v>>8))
						} else {
							outBytes = append(outBytes, byte(v), byte(v>>8))
						}
					}
					s := convert.NewStereoI16(bytes.NewReader(inBytes), mono, false)
					var got []byte
					for {
						var buf [97]byte
						n, err := s.Read(buf[:])
						got = append(got, buf[:n]...)
						if err != nil {
							if err != io.EOF {
								t.Fatal(err)
							}
							break
						}
						// Shifting by incomplete bytes should not affect the result.
						for i := 0; i < 2*2; i++ {
							if _, err := s.Seek(int64(i), io.SeekCurrent); err != nil {
								if err != io.EOF {
									t.Fatal(err)
								}
								break
							}
						}
					}
					want := outBytes
					if !bytes.Equal(got, want) {
						t.Errorf("got: %v, want: %v", got, want)
					}
				})
			}
		})
	}
}

func randBytes(n int) []byte {
	r := make([]byte, n)
	for i := range r {
		r[i] = byte(rand.Intn(256))
	}
	return r
}

func TestStereoI16EightBits(t *testing.T) {
	testCases := []struct {
		Name string
		In   []byte
	}{
		{
			Name: "nil",
			In:   nil,
		},
		{
			Name: "1, 0, 1, 0",
			In:   []byte{1, 0, 1, 0},
		},
		{
			Name: "8 0s",
			In:   []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			Name: "random 256 values",
			In:   randBytes(256),
		},
		{
			Name: "random 65536 values",
			In:   randBytes(65536),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			for _, mono := range []bool{false, true} {
				mono := mono
				t.Run(fmt.Sprintf("mono=%t", mono), func(t *testing.T) {
					inBytes := tc.In
					var outBytes []byte
					for _, v := range tc.In {
						v16 := int16(int(v)*0x101 - (1 << 15))
						if mono {
							// As the source is mono, the output should be stereo.
							outBytes = append(outBytes, byte(v16), byte(v16>>8), byte(v16), byte(v16>>8))
						} else {
							outBytes = append(outBytes, byte(v16), byte(v16>>8))
						}
					}
					s := convert.NewStereoI16(bytes.NewReader(inBytes), mono, true)
					var got []byte
					for {
						var buf [97]byte
						n, err := s.Read(buf[:])
						got = append(got, buf[:n]...)
						if err != nil {
							if err != io.EOF {
								t.Fatal(err)
							}
							break
						}
						// Shifting by incomplete bytes should not affect the result.
						for i := 0; i < 2*2; i++ {
							if _, err := s.Seek(int64(i), io.SeekCurrent); err != nil {
								if err != io.EOF {
									t.Fatal(err)
								}
								break
							}
						}
					}
					want := outBytes
					if !bytes.Equal(got, want) {
						t.Errorf("got: %v, want: %v", got, want)
					}
				})
			}
		})
	}
}
