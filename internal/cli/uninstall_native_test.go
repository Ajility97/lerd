package cli

import (
	"reflect"
	"testing"
)

// The launchd sweep addresses jobs as com.lerd.<file>, which the native FPM
// pools are not, so uninstall deleted their plists and left them serving. A
// reinstall then raced the still-loaded job and came up with no FPM at all.
func TestUninstallStopsEveryNativePool(t *testing.T) {
	var stopped []string
	stopNativePools([]string{"8.3", "8.4"}, func(v string) error { stopped = append(stopped, v); return nil })
	if want := []string{"8.3", "8.4"}; !reflect.DeepEqual(stopped, want) {
		t.Errorf("stopped %v, want %v", stopped, want)
	}
}
