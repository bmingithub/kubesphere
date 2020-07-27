/*
Copyright (C) 2018 Yunify, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this work except in compliance with the License.
You may obtain a copy of the License in the LICENSE file, or at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package capability

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestCSIWatcher(t *testing.T) {
	basePath := filepath.Join(os.TempDir(), "csi-watcher-"+rand.String(8))
	_ = os.MkdirAll(basePath, os.ModePerm)

	cnt := 10
	var driverNames []string
	w := NewCSIWatcher(basePath, func(driverName string) {
		driverNames = append(driverNames, driverName)
	})

	// scene of already run driver
	var runningDriverNames []string
	for i := 0; i < cnt; i++ {
		driverName := "csi-running-" + strconv.Itoa(i)
		_ = os.MkdirAll(filepath.Join(basePath, driverName), os.ModePerm)
		runningDriverNames = append(runningDriverNames, driverName)
	}
	err := w.Start()
	if err != nil {
		t.Error("start watcher error:", err)
		return
	}
	defer func() {
		err := w.Stop()
		if err != nil {
			t.Error("start watcher error:", err)
		}
	}()

	for i:=0; i<100;i++{
		time.Sleep(time.Millisecond * 10)
		if len(driverNames) == cnt{
			break
		}
	}
	sort.Strings(driverNames)
	sort.Strings(runningDriverNames)
	if diff := cmp.Diff(driverNames, runningDriverNames); diff != "" {
		t.Errorf("[runningDrivers] %T differ (-got, +want): %s", runningDriverNames, diff)
	}

	// scene of new driver
	driverNames = nil
	var newDriverNames []string
	for i := 0; i < cnt; i++ {
		driverName := "csi-" + strconv.Itoa(i)
		_ = os.MkdirAll(filepath.Join(basePath, driverName), os.ModePerm)
		newDriverNames = append(newDriverNames, driverName)
	}
	for i:=0; i<100;i++{
		time.Sleep(time.Millisecond * 10)
		if len(driverNames) == cnt{
			break
		}
	}
	if diff := cmp.Diff(driverNames, newDriverNames); diff != "" {
		t.Errorf("[newDrivers] %T differ (-got, +want): %s", newDriverNames, diff)
	}

	// scene of close driver
	driverNames = nil
	var deleteDriverNames []string
	for _, driverName := range newDriverNames{
		_  = os.Remove(filepath.Join(basePath, driverName))
		deleteDriverNames = append(deleteDriverNames, driverName)
	}
	for i:=0; i<100;i++{
		time.Sleep(time.Millisecond * 10)
		if len(driverNames) == cnt{
			break
		}
	}
	if diff := cmp.Diff(driverNames, deleteDriverNames); diff != "" {
		t.Errorf("[closeDrivers] %T differ (-got, +want): %s", newDriverNames, diff)
	}
}

