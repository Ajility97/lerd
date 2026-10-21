package podman

import (
	"reflect"
	"testing"
)

// Trimmed from a real applehv console log: the kernel's uptime stamp starting
// over marks a reboot, and only the kills after it are this boot's.
const oomConsoleLog = `[    0.000000] Booting Linux on physical CPU 0x0000000000 [0x610f0000]
[708961.947962] Out of memory: Killed process 1318 (dbus-broker-lau) total-vm:9292kB, anon-rss:276kB
[709000.465360] Out of memory: Killed process 2357719 (mysqld) total-vm:2178448kB, anon-rss:435684kB
[    0.000000] Booting Linux on physical CPU 0x0000000000 [0x610f0000]
[    2.114023] systemd[1]: Detected virtualization apple.
[root@localhost ~]# [  239.106785] Out of memory: Killed process 4157 (mysqld) total-vm:2112916kB, anon-rss:456404kB
[  919.398937] Out of memory: Killed process 1144 (systemd) total-vm:24620kB, anon-rss:6588kB
[  948.967036] Out of memory: Killed process 3952 (php-fpm) total-vm:268840kB, anon-rss:64160kB
`

func TestParseOOMKills_countsOnlyTheCurrentBoot(t *testing.T) {
	got := ParseOOMKills(oomConsoleLog)
	want := []string{"mysqld", "systemd", "php-fpm"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseOOMKills = %v, want %v", got, want)
	}
}

func TestParseOOMKills_noneOnAHealthyBoot(t *testing.T) {
	if got := ParseOOMKills("[    0.000000] Booting Linux\n[    2.1] systemd[1]: ok\n"); len(got) != 0 {
		t.Errorf("ParseOOMKills = %v, want none", got)
	}
}
