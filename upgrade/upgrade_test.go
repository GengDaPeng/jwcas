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

package upgrade

import "testing"

func TestUpstreamUpgradeIsDisabledByDefault(t *testing.T) {
	t.Setenv("disableVersionCheck", "")
	t.Setenv("disableUpgrade", "")
	t.Setenv("allowVendorUpgrade", "true")

	info := GetUpgradeInfo()
	if !info.IsCheckDisabled || !info.IsUpgradeDisabled || info.CanUpgrade || info.AllowVendorUpgrade || info.VendorOrder != nil {
		t.Fatalf("upstream upgrade is enabled: %+v", info)
	}
	if info.LatestVersion != "" || info.CheckError != "" {
		t.Fatalf("the release lookup was not skipped: %+v", info)
	}
	if err := StartUpgrade("v4.29.0", "test"); err == nil {
		t.Fatal("the running instance accepted an upstream upgrade")
	}
}
