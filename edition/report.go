// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

// Modified by the JWXT project to keep license reports offline.

package edition

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Reporter gives what a licensed Casdoor tells the license server about itself
type Reporter struct {
	InstanceId   func() string
	Version      func() string
	UserCount    func() int64
	Deployment   func() string
	DbVersion    func() string
	UpgradeState func() interface{}
}

var (
	reporterMu sync.RWMutex
	reporter   *Reporter
)

// SetReporter is called once the database is ready
func SetReporter(r *Reporter) {
	reporterMu.Lock()
	reporter = r
	reporterMu.Unlock()
}

func getReporter() *Reporter {
	reporterMu.RLock()
	defer reporterMu.RUnlock()
	return reporter
}

func getInstanceId() string {
	r := getReporter()
	if r == nil || r.InstanceId == nil {
		return ""
	}
	return r.InstanceId()
}

var (
	fingerprintOnce sync.Once
	fingerprint     string
)

// getFingerprint identifies the machine (or container): its machine ID when there is one, else its MAC addresses
func getFingerprint() string {
	fingerprintOnce.Do(func() {
		parts := []string{}
		for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
			if data, err := os.ReadFile(path); err == nil && strings.TrimSpace(string(data)) != "" {
				parts = append(parts, strings.TrimSpace(string(data)))
				break
			}
		}
		if len(parts) == 0 {
			interfaces, _ := net.Interfaces()
			for _, i := range interfaces {
				if i.Flags&net.FlagLoopback == 0 && len(i.HardwareAddr) > 0 {
					parts = append(parts, i.HardwareAddr.String())
				}
			}
			sort.Strings(parts)
		}
		hostname, _ := os.Hostname()
		parts = append(parts, hostname)
		sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
		fingerprint = hex.EncodeToString(sum[:16])
	})
	return fingerprint
}

func getVersion() string {
	r := getReporter()
	if r == nil || r.Version == nil {
		return ""
	}
	return r.Version()
}

// GetActivationCode is what a customer sends us to activate the license of an instance that cannot reach the license server
func GetActivationCode() string {
	instanceId := getInstanceId()
	if instanceId == "" {
		return ""
	}
	hostname, _ := os.Hostname()
	data, _ := json.Marshal(map[string]string{
		"instanceId":  instanceId,
		"fingerprint": getFingerprint(),
		"hostname":    hostname,
		"version":     getVersion(),
	})
	return base64.RawURLEncoding.EncodeToString(data)
}

func getOfflineReportBody(r *Reporter, text string) map[string]interface{} {
	hostname, _ := os.Hostname()
	userCount := int64(0)
	if r.UserCount != nil {
		userCount = r.UserCount()
	}
	var upgradeState interface{}
	if r.UpgradeState != nil {
		upgradeState = r.UpgradeState()
	}
	return map[string]interface{}{
		"license":     text,
		"instanceId":  getInstanceId(),
		"fingerprint": getFingerprint(),
		"hostname":    hostname,
		"version":     getVersion(),
		"edition":     Get(),
		"userCount":   userCount,
		"environment": collectEnvironment(),
		"upgrade":     upgradeState,
		"source":      "offline",
		"time":        time.Now().Format(time.RFC3339),
	}
}

// GetOfflineReport is the report of an instance that cannot reach the license server, which a customer copies to the vendor
func GetOfflineReport() (string, error) {
	r := getReporter()
	if r == nil {
		return "", fmt.Errorf("the database is not ready")
	}
	_, text, _ := getState()
	if text == "" {
		return "", fmt.Errorf("no valid license is configured")
	}
	data, err := json.Marshal(getOfflineReportBody(r, text))
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
