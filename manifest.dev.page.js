function renderDevPage(port) {
  const title = "Backend dev service is not ready";
  const subtitle = "This app is running in dev mode and expects the backend service to be started manually inside the app container.";
  const pageTitle = "Backend Dev";
  const steps = [
    "Sync your latest code into the container with <code>lzc-cli project sync --watch</code> or copy it with <code>lzc-cli project cp</code>.",
    "Open a shell with <code>lzc-cli project exec /bin/sh</code> and start or restart the backend process with <code>/app/run.sh</code>.",
    "Refresh this page after port <code>" + String(port) + "</code> is ready.",
  ];
  const portText = port === undefined || port === null ? "" : String(port);
  const items = steps.map(function (step) {
    return "<li>" + step + "</li>";
  }).join("");
  const portBadge = portText ? "<div class=\"pill\">Expected local port: " + portText + "</div>" : "";
  return [
    "<!doctype html>",
    "<html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>", pageTitle, "</title><style>",
    "body{margin:0;padding:32px;font-family:ui-sans-serif,system-ui,sans-serif;background:#f5f8fc;color:#112033}",
    ".card{max-width:760px;margin:0 auto;background:#fff;border:1px solid #d9e4f2;border-radius:18px;padding:24px;box-shadow:0 16px 40px rgba(17,32,51,.08)}",
    "h1{margin:0 0 12px;font-size:32px}.muted{color:#51627a}.pill{display:inline-block;margin:16px 0 0;padding:6px 10px;border-radius:999px;background:#eef5ff;color:#0c5dd6;font-weight:600}",
    "ol{margin:18px 0 0;padding-left:20px;line-height:1.8}code{background:#f2f6fb;border-radius:8px;padding:2px 6px}button{margin-top:20px;border:0;border-radius:10px;padding:10px 14px;background:#112033;color:#fff;cursor:pointer}",
    "</style></head><body><main class=\"card\"><h1>", title, "</h1><p class=\"muted\">", subtitle, "</p>", portBadge, "<ol>", items, "</ol><button onclick=\"location.reload()\">Refresh</button></main></body></html>",
  ].join("");
}
