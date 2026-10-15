#!/usr/bin/env php
<?php
// A plugin script can be written in any language its first line names. This
// one is PHP, which lerd runs in its PHP container; LERD_API_URL already points
// at lerd there. Every plugin process may call lerd's own API, the one the
// dashboard uses, with LERD_API_TOKEN. This prints the sites lerd knows.
$ctx = stream_context_create(['http' => [
    'header' => 'X-Lerd-Plugin-Token: ' . getenv('LERD_API_TOKEN'),
    'ignore_errors' => true,
]]);
$body = file_get_contents(getenv('LERD_API_URL') . '/api/sites', false, $ctx);
if (!str_contains($http_response_header[0] ?? '', ' 200')) {
    fwrite(STDERR, 'lerd answered: ' . ($http_response_header[0] ?? 'nothing') . "\n");
    exit(1);
}
foreach (json_decode($body, true) as $site) {
    echo $site['name'], "\n";
}
