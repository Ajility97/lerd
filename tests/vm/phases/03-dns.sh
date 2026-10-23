#!/usr/bin/env bash
# Phase 3, the .localhost path and the DNS round trip. Leaves the guest in the
# mode it started in.
source "$(dirname "$0")/../lib.sh"
need_demo
cd "$DEMO_DIR" || exit 1
start_mode=$(tld)
name=$(site_name "$DEMO_DIR")


# A site deliberately kept on plain HTTP, to prove it stays that way.
mkdir -p "$PROJECTS/plainsite/public" && echo '<?php echo "plain";' >"$PROJECTS/plainsite/public/index.php"
(cd "$PROJECTS/plainsite" && lerd link </dev/null >/dev/null 2>&1 && lerd unsecure "$(site_name "$PROJECTS/plainsite")" </dev/null >/dev/null 2>&1)
[ "$start_mode" = test ] && lerd secure "$name" </dev/null >/dev/null 2>&1

if [ "$start_mode" = test ]; then
	check "lerd dns:disable" lerd dns:disable
fi
check_out "3.1 lerd-dns is torn down and sites move to .localhost" 'lerd-dns=(inactive|failed|unknown).*\.localhost' bash -c "echo lerd-dns=\$(systemctl --user is-active lerd-dns) \$(lerd which | grep Site)"
check_out "3.2 lerd sites now shows the .localhost domain" "$name +[^ ]*\\.localhost" lerd sites
expect_200 3.3 "http://$(site_host "$DEMO_DIR")"
check_out "3.4 lerd secure refuses with a message about managed DNS" 'HTTPS requires lerd-managed DNS' lerd secure "$name"
tmp=$(mktemp -d "$HOME/lerd-vm-init.XXXX")
cp -r "$DEMO_DIR/composer.json" "$DEMO_DIR/artisan" "$tmp/" 2>/dev/null
init_out=$(cd "$tmp" && pty_output 8 lerd init --fresh)
echo "$init_out"
check "3.5 lerd init asks its questions but not the HTTPS one" bash -c 'grep -qiE "php|worker|service|framework" <<<"$1" && ! grep -qiE "https" <<<"$1"' _ "$init_out"
rm -rf "$tmp"
check_out "3.6 dns:check prints DNS managed externally" 'DNS managed externally' lerd dns:check
check_out "3.7 [partial] the dashboard's DNS status reads disabled" '"enabled": ?false' bash -c "curl -s http://127.0.0.1:7073/api/status | python3 -c 'import json,sys; print(json.dumps(json.load(sys.stdin)[\"dns\"]))'"
skip "3.8 the per-site HTTPS toggle is a muted lock with an explanation" "dashboard styling, phase 10's browser pass"

check "lerd dns:enable" lerd dns:enable
check_out "3.9 lerd-dns is back up and sites move to .test" 'lerd-dns=active.*\.test' bash -c "echo lerd-dns=\$(systemctl --user is-active lerd-dns) \$(lerd which | grep Site)"
check_out "3.10 [partial] a site that was HTTPS comes back HTTPS with .env synced" "^APP_URL=https://$(site_host "$DEMO_DIR")" grep ^APP_URL= .env
check_out "3.11 a site left on plain HTTP stays HTTP" "^$(site_name "$PROJECTS/plainsite") .* No " lerd sites
expect_200 3.12 "https://$(site_host "$DEMO_DIR")"

# The NetworkManager hook must only route the uplink at lerd-dns while it is
# serving (#2278). NM applies the hook a moment after the connection is up, so
# each reconnect waits a bounded time for 5300 to appear before judging it.
uplink=$(ip route show default | awk '{for(i=1;i<=NF;i++) if($i=="dev"){print $(i+1);exit}}')
conn=$(nmcli -g GENERAL.CONNECTION device show "$uplink" 2>/dev/null)
on_uplink() { resolvectl dns "$uplink" | grep -q '127\.0\.0\.1:5300'; }
off_uplink() { ! on_uplink; }
if [ -f /etc/NetworkManager/dispatcher.d/99-lerd-dns ] && systemctl is-active --quiet systemd-resolved && [ -n "$conn" ]; then
	sudo nmcli connection up "$conn" >/dev/null 2>&1
	wait_for 15 on_uplink
	check_out "3.13 with lerd-dns up, a reconnect routes the uplink at it" '127\.0\.0\.1:5300' resolvectl dns "$uplink"
	systemctl --user stop lerd-watcher lerd-dns
	sudo nmcli connection up "$conn" >/dev/null 2>&1
	# First the reconnect clearing the old route, then a window for the hook to wrongly re-add it.
	wait_for 10 off_uplink
	wait_for 5 on_uplink
	check_not "3.14 with lerd-dns stopped, a reconnect leaves the uplink's DNS alone" '127\.0\.0\.1:5300' resolvectl dns "$uplink"
	check "3.15 internet lookups resolve while lerd-dns is stopped" resolvectl query --legend=no --cache=no example.org
	check "lerd start" lerd start
	check_out "3.16 lerd start routes the uplink back at lerd-dns" '127\.0\.0\.1:5300' resolvectl dns "$uplink"
	expect_200 3.17 "https://$(site_host "$DEMO_DIR")"
else
	for id in 3.13 3.14 3.15 3.16 3.17; do skip "$id the NetworkManager hook leaves a stopped lerd-dns off the uplink" "needs NetworkManager with systemd-resolved"; done
fi

if [ "$start_mode" = localhost ]; then
	check "back to .localhost, as this guest started" lerd dns:disable
	expect_200 "http://$(site_host "$DEMO_DIR")"
fi
