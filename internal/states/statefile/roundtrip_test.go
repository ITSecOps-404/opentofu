// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statefile

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/opentofu/opentofu/internal/encryption"
)

func TestRoundtrip(t *testing.T) {
	const dir = "testdata/roundtrip"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, info := range entries {
		const inSuffix = ".in.tfstate"
		const outSuffix = ".out.tfstate"

		if info.IsDir() {
			continue
		}
		inName := info.Name()
		if !strings.HasSuffix(inName, inSuffix) {
			continue
		}
		name := inName[:len(inName)-len(inSuffix)]
		outName := name + outSuffix

		t.Run(name, func(t *testing.T) {
			oSrcWant, err := os.ReadFile(filepath.Join(dir, outName))
			if err != nil {
				t.Fatal(err)
			}
			oWant, diags := readStateV4(oSrcWant)
			if diags.HasErrors() {
				t.Fatal(diags.Err())
			}

			ir, err := os.Open(filepath.Join(dir, inName))
			if err != nil {
				t.Fatal(err)
			}
			defer ir.Close()

			f, err := Read(ir, encryption.StateEncryptionDisabled())
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			var buf bytes.Buffer
			err = Write(f, &buf, encryption.StateEncryptionDisabled())
			if err != nil {
				t.Fatal(err)
			}
			oSrcWritten := buf.Bytes()

			oGot, diags := readStateV4(oSrcWritten)
			if diags.HasErrors() {
				t.Fatal(diags.Err())
			}

			problems := deep.Equal(oGot, oWant)
			sort.Strings(problems)
			for _, problem := range problems {
				t.Error(problem)
			}
		})
	}
}
