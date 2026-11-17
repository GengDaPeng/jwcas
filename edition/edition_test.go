// Copyright 2026 The JWXT Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edition

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"sync"
	"testing"
)

func TestLocalEditionAndMarketplaceEntitlements(t *testing.T) {
	key, signer, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	originalKey := publicKey
	publicKey = key
	t.Cleanup(func() { publicKey = originalKey })

	cases := []struct {
		name        string
		licenseText string
		tier        string
		commercial  bool
		enterprise  bool
	}{
		{name: "missing"},
		{name: "malformed", licenseText: "invalid"},
		{name: "community", tier: Community},
		{name: "commercial", tier: Commercial, commercial: true},
		{name: "enterprise", tier: Enterprise, commercial: true, enterprise: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				loadOnce = sync.Once{}
				current, currentErr, currentText = nil, nil, ""
			})

			text := tc.licenseText
			if tc.tier != "" {
				payload, err := json.Marshal(License{Edition: tc.tier, Features: []string{"abac"}})
				if err != nil {
					t.Fatal(err)
				}
				encoded := base64.RawURLEncoding.EncodeToString(payload)
				signature := ed25519.Sign(signer, []byte(encoded))
				text = encoded + "." + base64.RawURLEncoding.EncodeToString(signature)
			}
			t.Setenv("license", text)

			if Get() != Enterprise || Require(Enterprise, FeatureAdminSeparation) != nil {
				t.Fatal("the self-hosted edition must provide its local features")
			}
			if got := Allows(Commercial, FeaturePaidIntegrations); got != tc.commercial {
				t.Errorf("paid Commercial integration access = %t, want %t", got, tc.commercial)
			}
			if got := Allows(Enterprise, FeaturePaidIntegrations); got != tc.enterprise {
				t.Errorf("paid Enterprise integration access = %t, want %t", got, tc.enterprise)
			}
			wantText := ""
			if tc.commercial {
				wantText = text
			}
			if GetLicenseText() != wantText {
				t.Error("unexpected license sent to the Marketplace")
			}
		})
	}
}
