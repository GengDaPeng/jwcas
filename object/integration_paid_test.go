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

package object

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casdoor/casdoor/edition"
)

func TestPaidMarketplaceDownloadRequiresVendorLicense(t *testing.T) {
	t.Setenv("license", "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("the download contacted the Marketplace without a license")
	}))
	defer server.Close()

	data, err := fetchMarketplaceFileWithLicense(server.URL, true)
	var required *edition.RequiredError
	if data != nil || !errors.As(err, &required) || required.Feature != edition.FeaturePaidIntegrations {
		t.Fatalf("wanted a vendor license error, got %q, %v", data, err)
	}
}
